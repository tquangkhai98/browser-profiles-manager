// ============================================
// BPM Desktop — Frontend Logic
// ============================================

// State
let profiles = [];
let browsers = [];
let refreshInterval = null;

// ============================================
// Initialization
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    // Wails injects window.go and window.runtime before DOMContentLoaded
    // But we add a small safety check
    init();
});

async function init() {
    try {
        // Ensure Wails runtime is available
        if (!window.go || !window.go.main || !window.go.main.App) {
            console.warn('Wails runtime not ready, retrying...');
            await waitForRuntime();
        }

        // Load initial data
        await loadBrowsers();
        await loadProfiles();

        // Auto-refresh every 5 seconds
        refreshInterval = setInterval(loadProfiles, 5000);

        // Bind button events
        bindEvents();
    } catch (err) {
        console.error('Init failed:', err);
        document.getElementById('profile-list').style.display = 'none';
        document.getElementById('empty-state').style.display = 'flex';
        document.getElementById('empty-state').innerHTML = `
            <h3>Failed to initialize</h3>
            <p>${escapeHtml(String(err))}</p>
        `;
    }
}

async function waitForRuntime() {
    let attempts = 0;
    while ((!window.go || !window.go.main || !window.go.main.App) && attempts < 30) {
        await sleep(200);
        attempts++;
    }
    if (!window.go || !window.go.main || !window.go.main.App) {
        throw new Error('Wails runtime failed to load');
    }
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Wrapper to call Go methods safely
function callGo(methodName, ...args) {
    if (!window.go?.main?.App?.[methodName]) {
        return Promise.reject(`Method ${methodName} not available`);
    }
    return window.go.main.App[methodName](...args);
}

// ============================================
// Data Loading
// ============================================

async function loadBrowsers() {
    try {
        browsers = await callGo('DetectBrowsers') || [];
        renderBrowserBadges();
    } catch (err) {
        console.error('Failed to detect browsers:', err);
        browsers = [];
        renderBrowserBadges();
    }
}

async function loadProfiles() {
    try {
        profiles = await callGo('ListProfiles') || [];
        renderProfiles();
    } catch (err) {
        console.error('Failed to load profiles:', err);
    }
}

// ============================================
// Rendering
// ============================================

function renderBrowserBadges() {
    const container = document.getElementById('browser-badges');
    if (!browsers || browsers.length === 0) {
        container.innerHTML = '<span class="browser-badge" style="background:var(--red-dim);color:var(--red)">No browsers found</span>';
        return;
    }
    container.innerHTML = browsers.map(b => `
        <span class="browser-badge" title="${escapeHtml(b.exe_path || '')}">✓ ${escapeHtml(b.name || 'Unknown')}</span>
    `).join('');
}

function renderProfiles() {
    const list = document.getElementById('profile-list');
    const empty = document.getElementById('empty-state');
    const countEl = document.getElementById('profile-count');

    if (!profiles || profiles.length === 0) {
        list.style.display = 'none';
        empty.style.display = 'flex';
        countEl.textContent = '';
        return;
    }

    list.style.display = 'grid';
    empty.style.display = 'none';
    countEl.textContent = `${profiles.length} profile${profiles.length !== 1 ? 's' : ''}`;

    list.innerHTML = profiles.map(p => {
        const statusClass = p.locked ? 'locked' : 'free';
        const statusText = p.locked ? 'Locked' : 'Free';
        const created = formatDate(p.created_at);
        const lastUsed = p.last_used ? timeAgo(p.last_used) : 'Never';
        const lockInfo = p.locked && p.lock_by
            ? `<div class="lock-info">🔒 Locked by ${escapeHtml(p.lock_by)} (PID ${p.lock_pid})</div>`
            : '';

        return `
            <div class="profile-card" data-name="${escapeHtml(p.name)}">
                <div class="profile-card-header">
                    <div class="profile-name-wrap">
                        <span class="status-dot ${statusClass}" title="${statusText}"></span>
                        <span class="profile-name">${escapeHtml(p.name)}</span>
                    </div>
                    <span class="profile-browser-badge">${escapeHtml(p.browser || 'chrome')}</span>
                </div>
                ${lockInfo}
                <div class="profile-meta">
                    <span class="profile-meta-item">
                        <svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M12 2H4a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2V4a2 2 0 00-2-2zM5 7h6M5 10h4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                        Created ${created}
                    </span>
                    <span class="profile-meta-item">
                        <svg width="12" height="12" viewBox="0 0 16 16" fill="none"><circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.5"/><path d="M8 5v3l2 2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
                        Used ${lastUsed}
                    </span>
                </div>
                <div class="profile-actions">
                    <button class="btn btn-primary btn-sm" onclick="launchProfile('${escapeAttr(p.name)}')" ${p.locked ? 'disabled title="Profile is locked"' : ''}>
                        <svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M4 3l9 5-9 5V3z" fill="currentColor"/></svg>
                        Launch
                    </button>
                    <button class="btn-icon" onclick="viewCredentials('${escapeAttr(p.name)}')" title="View credentials">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M2 8s2.5-4 6-4 6 4 6 4-2.5 4-6 4-6-4-6-4z" stroke="currentColor" stroke-width="1.5"/><circle cx="8" cy="8" r="2" stroke="currentColor" stroke-width="1.5"/></svg>
                    </button>
                    <button class="btn-icon danger" onclick="confirmDelete('${escapeAttr(p.name)}')" title="Delete profile">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 5h10M5 5V3h6v2M6 7v4M10 7v4M4 5l1 8h6l1-8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// ============================================
// Event Binding
// ============================================

function bindEvents() {
    // Create profile
    document.getElementById('btn-create').addEventListener('click', openCreateModal);
    document.getElementById('btn-create-confirm').addEventListener('click', handleCreate);

    // Import profile
    document.getElementById('btn-import').addEventListener('click', openImportModal);
    document.getElementById('btn-browse').addEventListener('click', handleBrowse);
    document.getElementById('btn-import-confirm').addEventListener('click', handleImport);

    // Sync credentials
    document.getElementById('btn-sync').addEventListener('click', openSyncModal);
    document.getElementById('btn-sync-confirm').addEventListener('click', handleSync);

    // Delete confirm
    document.getElementById('btn-delete-confirm').addEventListener('click', handleDeleteConfirm);

    // Enter key on inputs
    document.getElementById('create-name').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleCreate();
    });
    document.getElementById('import-name').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleImport();
    });

    // Close modals on overlay click
    document.querySelectorAll('.modal-overlay').forEach(overlay => {
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) overlay.style.display = 'none';
        });
    });

    // ESC key closes modals
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            document.querySelectorAll('.modal-overlay').forEach(el => el.style.display = 'none');
        }
    });
}

// ============================================
// Modal Management
// ============================================

function openModal(id) {
    document.getElementById(id).style.display = 'flex';
}

function closeModal(id) {
    document.getElementById(id).style.display = 'none';
}

function openCreateModal() {
    document.getElementById('create-name').value = '';
    populateBrowserSelect('create-browser');
    openModal('modal-create');
    setTimeout(() => document.getElementById('create-name').focus(), 100);
}

function openImportModal() {
    document.getElementById('import-path').value = '';
    document.getElementById('import-name').value = '';
    openModal('modal-import');
}

function openSyncModal() {
    if (profiles.length < 2) {
        showToast('Need at least 2 profiles to sync', 'error');
        return;
    }
    populateProfileSelect('sync-source', profiles);
    populateProfileSelect('sync-target', profiles);
    openModal('modal-sync');
}

function populateBrowserSelect(selectId) {
    const select = document.getElementById(selectId);
    if (browsers && browsers.length > 0) {
        select.innerHTML = browsers.map(b =>
            `<option value="${escapeAttr(b.id)}">${escapeHtml(b.name)}</option>`
        ).join('');
    } else {
        select.innerHTML = '<option value="chrome">Chrome (default)</option>';
    }
}

function populateProfileSelect(selectId, profileList) {
    const select = document.getElementById(selectId);
    select.innerHTML = profileList.map(p =>
        `<option value="${escapeAttr(p.name)}">${escapeHtml(p.name)} (${escapeHtml(p.browser || 'chrome')})</option>`
    ).join('');
}

// ============================================
// Actions
// ============================================

async function handleCreate() {
    const name = document.getElementById('create-name').value.trim();
    const browserVal = document.getElementById('create-browser').value;

    if (!name) {
        showToast('Please enter a profile name', 'error');
        return;
    }

    showLoading(true);
    try {
        await callGo('CreateProfile', name, browserVal);
        closeModal('modal-create');
        showToast(`Profile "${name}" created`, 'success');
        await loadProfiles();
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

async function launchProfile(name) {
    showLoading(true);
    try {
        await callGo('LaunchBrowser', name);
        showToast(`Launched browser for "${name}"`, 'success');
        setTimeout(loadProfiles, 1000);
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

let deleteTargetName = '';

function confirmDelete(name) {
    deleteTargetName = name;
    document.getElementById('delete-profile-name').textContent = name;
    openModal('modal-delete');
}

async function handleDeleteConfirm() {
    if (!deleteTargetName) return;

    showLoading(true);
    try {
        await callGo('DeleteProfile', deleteTargetName, false);
        closeModal('modal-delete');
        showToast(`Profile "${deleteTargetName}" deleted`, 'success');
        deleteTargetName = '';
        await loadProfiles();
    } catch (err) {
        if (String(err).includes('locked')) {
            if (confirm('Profile is locked. Force delete?')) {
                try {
                    await callGo('DeleteProfile', deleteTargetName, true);
                    closeModal('modal-delete');
                    showToast(`Profile "${deleteTargetName}" force deleted`, 'success');
                    deleteTargetName = '';
                    await loadProfiles();
                } catch (err2) {
                    showToast(err2, 'error');
                }
            }
        } else {
            showToast(err, 'error');
        }
    }
    showLoading(false);
}

async function viewCredentials(name) {
    showLoading(true);
    try {
        const result = await callGo('InspectCredentials', name);
        document.getElementById('creds-title').textContent = `Credentials — ${name}`;

        const summary = document.getElementById('creds-summary');
        summary.innerHTML = `
            <div class="creds-stat">
                <span class="creds-stat-value">${result.total_cookies || 0}</span>
                <span class="creds-stat-label">Cookies</span>
            </div>
            <div class="creds-stat">
                <span class="creds-stat-value">${result.total_logins || 0}</span>
                <span class="creds-stat-label">Logins</span>
            </div>
        `;

        const tbody = document.getElementById('creds-tbody');
        const tableWrap = document.getElementById('creds-table-wrap');
        const emptyEl = document.getElementById('creds-empty');

        if (result.sites && result.sites.length > 0) {
            tableWrap.style.display = 'block';
            emptyEl.style.display = 'none';
            tbody.innerHTML = result.sites.map(s => `
                <tr>
                    <td>${escapeHtml(s.domain)}</td>
                    <td>${s.cookies}</td>
                    <td>${s.logins}</td>
                </tr>
            `).join('');
        } else {
            tableWrap.style.display = 'none';
            emptyEl.style.display = 'block';
        }

        openModal('modal-creds');
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

async function handleSync() {
    const src = document.getElementById('sync-source').value;
    const dst = document.getElementById('sync-target').value;

    if (src === dst) {
        showToast('Source and target must be different', 'error');
        return;
    }

    showLoading(true);
    try {
        const result = await callGo('SyncCredentials', src, dst);
        closeModal('modal-sync');
        showToast(result.message, 'success');
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

async function handleBrowse() {
    try {
        const dir = await callGo('SelectDirectory');
        if (dir) {
            document.getElementById('import-path').value = dir;
        }
    } catch (err) {
        showToast(err, 'error');
    }
}

async function handleImport() {
    const srcPath = document.getElementById('import-path').value.trim();
    const name = document.getElementById('import-name').value.trim();

    if (!srcPath) {
        showToast('Please select a directory', 'error');
        return;
    }
    if (!name) {
        showToast('Please enter a profile name', 'error');
        return;
    }

    showLoading(true);
    try {
        await callGo('ImportProfile', srcPath, name);
        closeModal('modal-import');
        showToast(`Imported profile "${name}"`, 'success');
        await loadProfiles();
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

// ============================================
// Toast Notifications
// ============================================

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const icons = { success: '✓', error: '✕', info: 'ℹ' };

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.innerHTML = `<span>${icons[type] || ''}</span><span>${escapeHtml(String(message))}</span>`;
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('fadeOut');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// ============================================
// Loading
// ============================================

function showLoading(show) {
    document.getElementById('loading-overlay').style.display = show ? 'flex' : 'none';
}

// ============================================
// Helpers
// ============================================

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function escapeAttr(text) {
    return String(text).replace(/'/g, "\\'").replace(/"/g, '&quot;');
}

function formatDate(isoString) {
    if (!isoString) return '—';
    try {
        const d = new Date(isoString);
        return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch { return '—'; }
}

function timeAgo(isoString) {
    if (!isoString) return 'never';
    try {
        const now = Date.now();
        const then = new Date(isoString).getTime();
        const diff = now - then;

        const seconds = Math.floor(diff / 1000);
        if (seconds < 60) return 'just now';
        const minutes = Math.floor(seconds / 60);
        if (minutes < 60) return `${minutes}m ago`;
        const hours = Math.floor(minutes / 60);
        if (hours < 24) return `${hours}h ago`;
        const days = Math.floor(hours / 24);
        if (days < 30) return `${days}d ago`;
        return `${Math.floor(days / 30)}mo ago`;
    } catch { return 'unknown'; }
}
