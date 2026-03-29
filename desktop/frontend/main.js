// ============================================
// BPM Desktop — Frontend Logic
// ============================================

// State
let profiles = [];
let browsers = [];
let refreshInterval = null;
let currentPage = 'profiles'; // 'profiles' | 'settings'
let lastProfileHash = '';

// ============================================
// Initialization
// ============================================

document.addEventListener('DOMContentLoaded', () => {
    init();
});

async function init() {
    try {
        if (!window.go || !window.go.main || !window.go.main.App) {
            console.warn('Wails runtime not ready, retrying...');
            await waitForRuntime();
        }

        await loadBrowsers();
        await loadProfiles();

        refreshInterval = setInterval(loadProfiles, 5000);

        // Render Lucide icons (static ones in header/settings/modals)
        lucide.createIcons();

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
        updateStatusBar();
    } catch (err) {
        console.error('Failed to detect browsers:', err);
        browsers = [];
        updateStatusBar();
    }
}

async function loadProfiles() {
    try {
        const newProfiles = await callGo('ListProfiles') || [];
        // Only re-render if data actually changed (prevents flicker)
        const newHash = JSON.stringify(newProfiles);
        if (newHash === lastProfileHash) return;
        lastProfileHash = newHash;
        profiles = newProfiles;
        renderProfiles();
        updateStatusBar();
    } catch (err) {
        console.error('Failed to load profiles:', err);
    }
}

// ============================================
// Page Navigation
// ============================================

function showPage(pageName) {
    currentPage = pageName;
    document.querySelectorAll('.page').forEach(p => {
        p.classList.remove('active');
        p.style.display = 'none';
    });
    const target = document.getElementById(`page-${pageName}`);
    if (target) {
        target.style.display = 'flex';
        target.classList.add('active');
    }

    // Show/hide status bar on settings page
    document.getElementById('status-bar').style.display =
        pageName === 'settings' ? 'none' : 'flex';
}

async function openSettings() {
    showPage('settings');
    try {
        const settings = await callGo('GetSettings');
        if (settings) {
            document.getElementById('settings-profiles-dir').textContent = settings.profiles_dir;
            document.getElementById('settings-version').textContent = `v${settings.version}`;

            // Populate browser dropdown with current selection
            const select = document.getElementById('settings-browser');
            if (browsers && browsers.length > 0) {
                select.innerHTML = browsers.map(b =>
                    `<option value="${escapeAttr(b.id)}" ${b.id === settings.default_browser ? 'selected' : ''}>${escapeHtml(b.name)}</option>`
                ).join('');
            }
        }

        // Load MCP config
        const mcpConfig = await callGo('GetMCPConfig');
        document.getElementById('mcp-config-code').textContent = mcpConfig;
    } catch (err) {
        console.error('Failed to load settings:', err);
    }
}

// ============================================
// Rendering
// ============================================

function renderProfiles() {
    const list = document.getElementById('profile-list');
    const empty = document.getElementById('empty-state');
    const searchTerm = (document.getElementById('search-input')?.value || '').toLowerCase();

    if (!profiles || profiles.length === 0) {
        list.style.display = 'none';
        empty.style.display = 'flex';
        return;
    }

    list.style.display = 'grid';
    empty.style.display = 'none';

    list.innerHTML = profiles.map(p => {
        const statusClass = p.locked ? 'locked' : 'free';
        const statusText = p.locked ? 'Locked' : 'Free';
        const lastUsed = p.last_used ? timeAgo(p.last_used) : 'Never';
        const created = p.created_at ? timeAgo(p.created_at) : '—';
        const matchesSearch = !searchTerm || p.name.toLowerCase().includes(searchTerm);
        const lockInfo = p.locked && p.lock_by
            ? `<div class="lock-info">LOCKED (PID: ${p.lock_pid})</div>`
            : '';

        return `
            <div class="profile-card ${matchesSearch ? '' : 'filtered-out'}" data-name="${escapeHtml(p.name)}">
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
                        <i data-lucide="calendar" class="icon-xxs"></i>
                        ${created}
                    </span>
                    <span class="profile-meta-item">
                        <i data-lucide="clock" class="icon-xxs"></i>
                        ${lastUsed}
                    </span>
                </div>
                <div class="profile-actions">
                    <button class="btn btn-primary btn-sm" onclick="launchProfile('${escapeAttr(p.name)}')" ${p.locked ? 'disabled title="Profile is locked"' : ''}>
                        <i data-lucide="play" class="icon-xs"></i>
                        Launch
                    </button>
                    <button class="btn-icon" onclick="viewCredentials('${escapeAttr(p.name)}')" title="View credentials">
                        <i data-lucide="eye" class="icon-sm"></i>
                    </button>
                    <button class="btn-icon" onclick="openEditModal('${escapeAttr(p.name)}')" title="Edit profile">
                        <i data-lucide="pencil" class="icon-sm"></i>
                    </button>
                    <button class="btn-icon danger" onclick="confirmDelete('${escapeAttr(p.name)}')" title="Delete profile">
                        <i data-lucide="trash-2" class="icon-sm"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Re-render Lucide icons in dynamically created cards
    lucide.createIcons();
}

function updateStatusBar() {
    const countEl = document.getElementById('status-profile-count');
    const browsersEl = document.getElementById('status-browsers');

    if (countEl) {
        const count = profiles ? profiles.length : 0;
        countEl.innerHTML = `
            <svg width="12" height="12" viewBox="0 0 16 16" fill="none"><path d="M2 4h12M2 8h12M2 12h12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            ${count} Profile${count !== 1 ? 's' : ''}
        `;
    }

    if (browsersEl) {
        const names = browsers && browsers.length > 0
            ? browsers.map(b => b.name).join(', ')
            : '—';
        browsersEl.innerHTML = `
            <svg width="12" height="12" viewBox="0 0 16 16" fill="none"><rect x="2" y="2" width="12" height="10" rx="2" stroke="currentColor" stroke-width="1.5" fill="none"/><path d="M2 10h12" stroke="currentColor" stroke-width="1.5"/></svg>
            Browsers: ${escapeHtml(names)}
        `;
    }
}

// ============================================
// Event Binding
// ============================================

function bindEvents() {
    // Header buttons
    document.getElementById('btn-create-header').addEventListener('click', openCreateModal);
    document.getElementById('btn-import-header').addEventListener('click', openImportModal);
    document.getElementById('btn-settings').addEventListener('click', () => {
        if (currentPage === 'settings') showPage('profiles');
        else openSettings();
    });
    document.getElementById('btn-back-settings').addEventListener('click', () => showPage('profiles'));
    document.getElementById('btn-sync-header').addEventListener('click', openSyncModal);

    // Language toggle (placeholder)
    document.getElementById('btn-language').addEventListener('click', () => {
        showToast('Language switching coming soon', 'info');
    });

    // Theme toggle (visual feedback only)
    document.getElementById('btn-theme').addEventListener('click', () => {
        showToast('Dark mode is the only theme', 'info');
    });

    // Create
    document.getElementById('btn-create-confirm').addEventListener('click', handleCreate);

    // Import
    document.getElementById('btn-browse').addEventListener('click', handleBrowse);
    document.getElementById('btn-import-confirm').addEventListener('click', handleImport);

    // Sync
    document.getElementById('btn-sync-confirm').addEventListener('click', handleSync);
    document.getElementById('btn-sync-from-creds').addEventListener('click', () => {
        closeModal('modal-creds');
        openSyncModal();
    });

    // Delete
    document.getElementById('btn-delete-confirm').addEventListener('click', handleDeleteConfirm);

    // Edit
    document.getElementById('btn-edit-confirm').addEventListener('click', handleRename);
    document.getElementById('edit-name').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleRename();
    });
    document.getElementById('btn-open-profile-dir').addEventListener('click', () => {
        if (editTargetName) callGo('OpenProfileDir', editTargetName);
    });

    // Settings actions
    document.getElementById('btn-copy-mcp').addEventListener('click', handleCopyMCP);
    document.getElementById('btn-export-all').addEventListener('click', handleExportAll);
    document.getElementById('btn-reset-settings').addEventListener('click', handleResetSettings);

    document.getElementById('settings-browser').addEventListener('change', async (e) => {
        try {
            await callGo('SaveDefaultBrowser', e.target.value);
            showToast('Default browser updated', 'success');
        } catch (err) {
            showToast(err, 'error');
        }
    });

    // Search filter
    document.getElementById('search-input').addEventListener('input', () => {
        renderProfiles();
    });

    // Enter key
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

    // ESC key
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
let editTargetName = '';

function confirmDelete(name) {
    deleteTargetName = name;
    document.getElementById('delete-profile-name').textContent = name;
    openModal('modal-delete');
}

function openEditModal(name) {
    editTargetName = name;
    document.getElementById('edit-name').value = name;

    // Find profile to get dataDir
    const p = profiles.find(p => p.name === name);
    document.getElementById('edit-path').textContent = p ? p.data_dir : '';

    openModal('modal-edit');
    setTimeout(() => {
        const input = document.getElementById('edit-name');
        input.focus();
        input.select();
    }, 100);
}

async function handleRename() {
    const newName = document.getElementById('edit-name').value.trim();
    if (!newName || newName === editTargetName) {
        closeModal('modal-edit');
        return;
    }

    showLoading(true);
    try {
        await callGo('RenameProfile', editTargetName, newName);
        closeModal('modal-edit');
        showToast(`Renamed "${editTargetName}" → "${newName}"`, 'success');
        lastProfileHash = ''; // Force refresh
        editTargetName = '';
        await loadProfiles();
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
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
        document.getElementById('creds-title').textContent = `${name} — Credentials`;

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
// Settings Actions
// ============================================

async function handleCopyMCP() {
    try {
        const code = document.getElementById('mcp-config-code').textContent;
        await navigator.clipboard.writeText(code);
        showToast('MCP config copied to clipboard', 'success');
    } catch {
        showToast('Failed to copy', 'error');
    }
}

async function handleExportAll() {
    showLoading(true);
    try {
        const msg = await callGo('ExportAllProfiles');
        if (msg) showToast(msg, 'success');
    } catch (err) {
        showToast(err, 'error');
    }
    showLoading(false);
}

async function handleResetSettings() {
    if (!confirm('Reset all settings to defaults? This will not delete profiles.')) return;
    try {
        await callGo('SaveDefaultBrowser', '');
        showToast('Settings reset', 'success');
        await openSettings();
    } catch (err) {
        showToast(err, 'error');
    }
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
