// ============================================
// BPM Desktop — Frontend Logic
// ============================================

// State
let profiles = [];
let browsers = [];
let refreshInterval = null;
let currentPage = 'profiles'; // 'profiles' | 'settings' | 'ai'
let lastProfileHash = '';
let currentLang = localStorage.getItem('bpm-lang') || 'en';
let currentTheme = localStorage.getItem('bpm-theme') || 'dark';

// ============================================
// Internationalization (i18n)
// ============================================

const i18n = {
    en: {
        // Header
        'import': 'Import',
        'create_profile': 'Create Profile',
        // Profile list
        'browser_profiles': 'Browser Profiles',
        'manage_desc': 'Manage and isolate your development environments.',
        'filter_profiles': 'Filter profiles...',
        'sync': 'Sync',
        'launch': 'Launch',
        'locked': 'Locked',
        'free': 'Free',
        // Empty state
        'no_profiles': 'No profiles yet',
        'no_profiles_desc': 'Create your first isolated browser profile to get started.',
        // Settings
        'back': 'Back',
        'settings': 'Settings',
        'general': 'General',
        'default_browser': 'Default Browser',
        'profile_storage': 'Profile storage path',
        'change': 'Change',
        'appearance': 'Appearance',
        'theme': 'Theme',
        'language': 'Language',
        'mcp_server': 'MCP Server',
        'mcp_hint': "ℹ️ Add this to your AI IDE's MCP config (Claude Code, Cursor, etc.)",
        'data': 'Data',
        'open_config': 'Open config folder',
        'export_all': 'Export all profiles',
        'reset_settings': 'Reset all settings',
        // Create modal
        'profile_name': 'Profile Name',
        'profile_name_hint': 'Lowercase, 3-64 chars, letters/numbers/hyphens',
        'browser': 'Browser',
        'create_info': 'Settings like WebGL, Fingerprinting, and Canvas protection will be automatically optimized for the selected browser.',
        'cancel': 'Cancel',
        'create': 'Create',
        // Sync modal
        'sync_credentials': 'Sync credentials',
        'sync_desc': 'Transfer session data between environments.',
        'from': 'From:',
        'to': 'To:',
        'cookies_only': 'Cookies only',
        'cookies_logins': 'Cookies + Logins',
        'sync_warning': 'Target profile will be backed up first',
        // Import modal
        'import_chrome': 'Import Chrome Profile',
        'import_desc': 'Import an existing Chromium profile directory into bpm as a managed profile.',
        'profile_directory': 'Profile Directory',
        'browse': 'Browse',
        // Credentials modal
        'credentials': 'Credentials',
        'sync_creds': 'Sync Creds',
        'cookies': 'Cookies',
        'logins': 'Logins',
        'domain': 'Domain',
        'no_creds': 'No credentials found in this profile.',
        'close': 'Close',
        // Delete modal
        'delete_profile': 'Delete Profile',
        'delete_confirm': 'Are you sure you want to delete',
        'delete_warning': 'This will permanently remove the profile and all its data.',
        'delete': 'Delete',
        // Edit modal
        'edit_profile': 'Edit Profile',
        'edit_name_hint': 'Lowercase, letters/numbers/hyphens/underscores',
        'data_directory': 'Data Directory',
        'save': 'Save',
        // Status bar
        'profiles_count': 'Profiles',
        'browsers_label': 'Browsers',
        // Toasts
        'theme_dark': 'Dark mode activated',
        'theme_light': 'Light mode activated',
        'lang_changed': 'Language changed to English',
        'default_browser_updated': 'Default browser updated',
        'profile_created': 'Profile "{name}" created',
        'profile_deleted': 'Profile "{name}" deleted',
        'profile_renamed': 'Renamed "{old}" → "{new}"',
        'launched': 'Launched browser for "{name}"',
        'mcp_copied': 'MCP config copied to clipboard',
        'settings_reset': 'Settings reset',
        'need_2_profiles': 'Need at least 2 profiles to sync',
        'enter_name': 'Please enter a profile name',
        'select_dir': 'Please select a directory',
        'src_dst_diff': 'Source and target must be different',
        'imported': 'Imported profile "{name}"',
        'copy_failed': 'Failed to copy',
        'dark_only': 'Dark mode is the only theme',
        // AI Integration
        'ai_agent': 'AI Agent',
        'ai_hero_badge': '✨ NEW — The missing link between AI agents and browser sessions',
        'ai_hero_title': 'AI Agent Integration',
        'ai_hero_tagline': 'Login once. Automate forever.',
        'ai_hero_desc': 'Bridge the gap between your persistent browser profiles and modern AI orchestration. BPM allows your agents to bypass 2FA and login barriers by leveraging existing authenticated sessions through Model Context Protocol (MCP) endpoints.',
        'ai_playwright_title': 'Playwright MCP',
        'ai_playwright_desc': 'Stable, official Anthropic integration. Uses accessibility tree for reliable element interaction. Best for E2E testing and form automation.',
        'ai_devtools_title': 'Chrome DevTools MCP',
        'ai_devtools_desc': 'Connect to a running Chrome instance with remote debugging. DOM inspection, Network monitoring, and Console access in real-time.',
        'ai_browseruse_title': 'Browser Use MCP',
        'ai_browseruse_desc': 'AI-native browser automation — describe actions in plain language instead of CSS selectors. The most seamless integration for AI workflows.',
        'ai_workflow_title': 'The Integration Workflow',
        'ai_step_create': 'Create Profile',
        'ai_step_login': 'Login Once',
        'ai_step_agent': 'AI Agent Uses Session',
        'ai_step_automate': 'Automate Forever',
        'ai_comparison_title': 'Comparison',
        'ai_warning': 'Never open the same profile in two browsers simultaneously — bpm uses file locks to prevent profile corruption.',
        // About page
        'about': 'About',
        'about_author': 'Author',
        'about_repo': 'Repository',
        'about_tech': 'Tech Stack',
        'about_open_github': 'Open on GitHub',
        'about_tagline': 'Manage isolated Chromium browser profiles for AI-powered automation',
    },
    vi: {
        // Header
        'import': 'Nhập',
        'create_profile': 'Tạo Profile',
        // Profile list
        'browser_profiles': 'Browser Profiles',
        'manage_desc': 'Quản lý và cách ly các môi trường phát triển.',
        'filter_profiles': 'Lọc profiles...',
        'sync': 'Đồng bộ',
        'launch': 'Mở',
        'locked': 'Đã khóa',
        'free': 'Sẵn sàng',
        // Empty state
        'no_profiles': 'Chưa có profile',
        'no_profiles_desc': 'Tạo profile trình duyệt cách ly đầu tiên để bắt đầu.',
        // Settings
        'back': 'Quay lại',
        'settings': 'Cài đặt',
        'general': 'Chung',
        'default_browser': 'Trình duyệt mặc định',
        'profile_storage': 'Đường dẫn lưu profile',
        'change': 'Đổi',
        'appearance': 'Giao diện',
        'theme': 'Chủ đề',
        'language': 'Ngôn ngữ',
        'mcp_server': 'MCP Server',
        'mcp_hint': 'ℹ️ Thêm cấu hình này vào AI IDE (Claude Code, Cursor, v.v.)',
        'data': 'Dữ liệu',
        'open_config': 'Mở thư mục cấu hình',
        'export_all': 'Xuất tất cả profiles',
        'reset_settings': 'Đặt lại cài đặt',
        // Create modal
        'profile_name': 'Tên Profile',
        'profile_name_hint': 'Chữ thường, 3-64 ký tự, chữ cái/số/gạch ngang',
        'browser': 'Trình duyệt',
        'create_info': 'Các thiết lập WebGL, Fingerprinting, và Canvas sẽ được tối ưu tự động cho trình duyệt đã chọn.',
        'cancel': 'Hủy',
        'create': 'Tạo',
        // Sync modal
        'sync_credentials': 'Đồng bộ thông tin',
        'sync_desc': 'Chuyển dữ liệu phiên giữa các môi trường.',
        'from': 'Từ:',
        'to': 'Đến:',
        'cookies_only': 'Chỉ Cookies',
        'cookies_logins': 'Cookies + Đăng nhập',
        'sync_warning': 'Profile đích sẽ được sao lưu trước',
        // Import modal
        'import_chrome': 'Nhập Chrome Profile',
        'import_desc': 'Nhập thư mục profile Chromium hiện có vào bpm.',
        'profile_directory': 'Thư mục Profile',
        'browse': 'Chọn',
        // Credentials modal
        'credentials': 'Thông tin đăng nhập',
        'sync_creds': 'Đồng bộ',
        'cookies': 'Cookies',
        'logins': 'Đăng nhập',
        'domain': 'Tên miền',
        'no_creds': 'Không tìm thấy thông tin đăng nhập trong profile này.',
        'close': 'Đóng',
        // Delete modal
        'delete_profile': 'Xóa Profile',
        'delete_confirm': 'Bạn có chắc muốn xóa',
        'delete_warning': 'Hành động này sẽ xóa vĩnh viễn profile và toàn bộ dữ liệu.',
        'delete': 'Xóa',
        // Edit modal
        'edit_profile': 'Sửa Profile',
        'edit_name_hint': 'Chữ thường, chữ cái/số/gạch ngang/gạch dưới',
        'data_directory': 'Thư mục dữ liệu',
        'save': 'Lưu',
        // Status bar
        'profiles_count': 'Profiles',
        'browsers_label': 'Trình duyệt',
        // Toasts
        'theme_dark': 'Chế độ tối đã bật',
        'theme_light': 'Chế độ sáng đã bật',
        'lang_changed': 'Đã chuyển sang Tiếng Việt',
        'default_browser_updated': 'Đã cập nhật trình duyệt mặc định',
        'profile_created': 'Đã tạo profile "{name}"',
        'profile_deleted': 'Đã xóa profile "{name}"',
        'profile_renamed': 'Đổi tên "{old}" → "{new}"',
        'launched': 'Đã mở trình duyệt cho "{name}"',
        'mcp_copied': 'Đã sao chép cấu hình MCP',
        'settings_reset': 'Đã đặt lại cài đặt',
        'need_2_profiles': 'Cần ít nhất 2 profiles để đồng bộ',
        'enter_name': 'Vui lòng nhập tên profile',
        'select_dir': 'Vui lòng chọn thư mục',
        'src_dst_diff': 'Nguồn và đích phải khác nhau',
        'imported': 'Đã nhập profile "{name}"',
        'copy_failed': 'Sao chép thất bại',
        // AI Integration
        'ai_agent': 'AI Agent',
        'ai_hero_badge': '✨ MỚI — Cầu nối giữa AI agent và phiên trình duyệt',
        'ai_hero_title': 'Tích hợp AI Agent',
        'ai_hero_tagline': 'Đăng nhập một lần. Tự động hoá mãi mãi.',
        'ai_hero_desc': 'Kết nối browser profiles với AI. BPM cho phép AI agent sử dụng phiên đăng nhập sẵn có, bỏ qua 2FA và rào cản đăng nhập thông qua Model Context Protocol (MCP).',
        'ai_playwright_title': 'Playwright MCP',
        'ai_playwright_desc': 'Tích hợp chính thức từ Anthropic. Sử dụng cây trợ năng cho tương tác ổn định. Tốt nhất cho E2E testing và tự động hoá form.',
        'ai_devtools_title': 'Chrome DevTools MCP',
        'ai_devtools_desc': 'Kết nối tới Chrome đang chạy với remote debugging. Kiểm tra DOM, giám sát Network, và truy cập Console thời gian thực.',
        'ai_browseruse_title': 'Browser Use MCP',
        'ai_browseruse_desc': 'Tự động hoá trình duyệt bằng ngôn ngữ tự nhiên — mô tả hành động thay vì CSS selectors. Tích hợp liền mạch nhất cho AI.',
        'ai_workflow_title': 'Quy trình tích hợp',
        'ai_step_create': 'Tạo Profile',
        'ai_step_login': 'Đăng nhập',
        'ai_step_agent': 'AI Agent dùng phiên',
        'ai_step_automate': 'Tự động hoá',
        'ai_comparison_title': 'So sánh',
        'ai_warning': 'Không mở cùng profile trên hai trình duyệt đồng thời — bpm dùng file lock để ngăn hư hại profile.',
        // About page
        'about': 'Giới thiệu',
        'about_author': 'Tác giả',
        'about_repo': 'Kho mã nguồn',
        'about_tech': 'Công nghệ',
        'about_open_github': 'Mở trên GitHub',
        'about_tagline': 'Quản lý browser profiles cô lập cho tự động hoá bằng AI',
    }
};

function t(key, params = {}) {
    let text = (i18n[currentLang] && i18n[currentLang][key]) || i18n.en[key] || key;
    for (const [k, v] of Object.entries(params)) {
        text = text.replace(`{${k}}`, v);
    }
    return text;
}

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

        // Apply saved theme
        applyTheme(currentTheme);

        await loadBrowsers();
        await loadProfiles();

        refreshInterval = setInterval(loadProfiles, 5000);

        // Render Lucide icons (static ones in header/settings/modals)
        lucide.createIcons();

        // Apply language
        applyLanguage(currentLang);

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

    // Show/hide status bar on non-profiles pages
    document.getElementById('status-bar').style.display =
        pageName === 'profiles' ? 'flex' : 'none';

    // Toggle AI button active state
    const aiBtn = document.getElementById('btn-ai-integration');
    if (aiBtn) aiBtn.classList.toggle('active', pageName === 'ai');

    // Re-render Lucide icons for dynamic pages
    if (pageName === 'ai') lucide.createIcons();
}

async function openSettings() {
    showPage('settings');
    try {
        const settings = await callGo('GetSettings');
        if (settings) {
            document.getElementById('settings-profiles-dir').textContent = settings.profiles_dir;

            // Populate browser dropdown with current selection
            const select = document.getElementById('settings-browser');
            if (browsers && browsers.length > 0) {
                select.innerHTML = browsers.map(b =>
                    `<option value="${escapeAttr(b.id)}" ${b.id === settings.default_browser ? 'selected' : ''}>${escapeHtml(b.name)}</option>`
                ).join('');
            }

            // About section
            document.getElementById('about-version').textContent = `v${settings.version}`;
            const commitShort = settings.commit && settings.commit !== 'none'
                ? settings.commit.substring(0, 7) : 'none';
            document.getElementById('about-commit').textContent = commitShort;
            document.getElementById('about-date').textContent = settings.build_date || 'unknown';
            document.getElementById('about-config-dir').textContent = settings.config_dir || '~/.config/bpm';
        }

        // Sync theme radio state
        syncRadioGroup('settings-theme-group', currentTheme);
        // Sync language radio state
        syncRadioGroup('settings-lang-group', currentLang);

        // Load MCP config
        const mcpConfig = await callGo('GetMCPConfig');
        document.getElementById('mcp-config-code').textContent = mcpConfig;
    } catch (err) {
        console.error('Failed to load settings:', err);
    }
}

function syncRadioGroup(groupId, activeValue) {
    const group = document.getElementById(groupId);
    if (!group) return;
    group.querySelectorAll('.settings-radio-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.value === activeValue);
    });
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

    list.innerHTML = profiles.map((p, index) => {
        const statusClass = p.locked ? 'locked' : 'free';
        const statusText = p.locked ? t('locked') : t('free');
        const lastUsed = p.last_used ? timeAgo(p.last_used) : 'Never';
        const created = p.created_at ? timeAgo(p.created_at) : '—';
        const matchesSearch = !searchTerm || p.name.toLowerCase().includes(searchTerm);
        const lockInfo = p.locked && p.lock_by
            ? `<div class="lock-info">${t('locked').toUpperCase()} (PID: ${p.lock_pid})</div>`
            : '';

        return `
            <div class="profile-card ${matchesSearch ? '' : 'filtered-out'}" data-name="${escapeHtml(p.name)}" data-index="${index}" draggable="false">
                <button class="drag-handle" title="Drag to reorder" aria-label="Drag to reorder">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                        <circle cx="9" cy="5" r="1"/><circle cx="15" cy="5" r="1"/>
                        <circle cx="9" cy="12" r="1"/><circle cx="15" cy="12" r="1"/>
                        <circle cx="9" cy="19" r="1"/><circle cx="15" cy="19" r="1"/>
                    </svg>
                </button>
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
                    <button class="btn btn-primary btn-sm" onclick="launchProfile('${escapeAttr(p.name)}')" ${p.locked ? `disabled title="${t('locked')}"` : ''}>
                        <i data-lucide="play" class="icon-xs"></i>
                        ${t('launch')}
                    </button>
                    <button class="btn-icon" onclick="viewCredentials('${escapeAttr(p.name)}')" title="${t('credentials')}">
                        <i data-lucide="eye" class="icon-sm"></i>
                    </button>
                    <button class="btn-icon" onclick="openEditModal('${escapeAttr(p.name)}')" title="${t('edit_profile')}">
                        <i data-lucide="pencil" class="icon-sm"></i>
                    </button>
                    <button class="btn-icon danger" onclick="confirmDelete('${escapeAttr(p.name)}')" title="${t('delete_profile')}">
                        <i data-lucide="trash-2" class="icon-sm"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Re-render Lucide icons in dynamically created cards
    lucide.createIcons();

    // Bind drag-and-drop events to new cards
    bindDragAndDrop();
}

// ============================================
// Drag & Drop Reordering (dnd-kit inspired)
// ============================================

let dragState = {
    dragging: null,     // DOM element being dragged
    dragName: null,     // profile name of dragged card
    overCard: null,     // DOM element being hovered over
};

function bindDragAndDrop() {
    const cards = document.querySelectorAll('.profile-card');

    cards.forEach(card => {
        const handle = card.querySelector('.drag-handle');
        if (!handle) return;

        // Only initiate drag from the handle
        handle.addEventListener('mousedown', () => {
            card.setAttribute('draggable', 'true');
        });

        // Reset draggable when mouse is released outside a drag
        handle.addEventListener('mouseup', () => {
            card.setAttribute('draggable', 'false');
        });

        card.addEventListener('dragstart', onDragStart);
        card.addEventListener('dragend', onDragEnd);
        card.addEventListener('dragover', onDragOver);
        card.addEventListener('dragenter', onDragEnter);
        card.addEventListener('dragleave', onDragLeave);
        card.addEventListener('drop', onDrop);
    });
}

function onDragStart(e) {
    const card = e.target.closest('.profile-card');
    if (!card) return;

    dragState.dragging = card;
    dragState.dragName = card.dataset.name;

    // Set drag data
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', card.dataset.name);

    // Delay adding class so the ghost image captures the normal state
    requestAnimationFrame(() => {
        card.classList.add('dragging');
    });
}

function onDragEnd(e) {
    const card = e.target.closest('.profile-card');
    if (card) {
        card.classList.remove('dragging');
        card.setAttribute('draggable', 'false');
    }

    // Clean up all drag-over states
    document.querySelectorAll('.profile-card.drag-over').forEach(el => {
        el.classList.remove('drag-over');
    });

    dragState.dragging = null;
    dragState.dragName = null;
    dragState.overCard = null;
}

function onDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
}

function onDragEnter(e) {
    e.preventDefault();
    const card = e.target.closest('.profile-card');
    if (!card || card === dragState.dragging) return;

    // Remove drag-over from previous target
    if (dragState.overCard && dragState.overCard !== card) {
        dragState.overCard.classList.remove('drag-over');
    }

    card.classList.add('drag-over');
    dragState.overCard = card;
}

function onDragLeave(e) {
    const card = e.target.closest('.profile-card');
    if (!card) return;

    // Only remove if we're actually leaving the card (not entering a child)
    const relatedTarget = e.relatedTarget;
    if (relatedTarget && card.contains(relatedTarget)) return;

    card.classList.remove('drag-over');
    if (dragState.overCard === card) {
        dragState.overCard = null;
    }
}

function onDrop(e) {
    e.preventDefault();
    const targetCard = e.target.closest('.profile-card');
    if (!targetCard || !dragState.dragging || targetCard === dragState.dragging) return;

    const fromName = dragState.dragName;
    const toName = targetCard.dataset.name;

    if (!fromName || !toName || fromName === toName) return;

    // Swap positions in local profiles array
    const fromIdx = profiles.findIndex(p => p.name === fromName);
    const toIdx = profiles.findIndex(p => p.name === toName);

    if (fromIdx === -1 || toIdx === -1) return;

    // Move profile from fromIdx to toIdx (insert, not swap)
    const [moved] = profiles.splice(fromIdx, 1);
    profiles.splice(toIdx, 0, moved);

    // Re-render immediately for visual feedback
    lastProfileHash = JSON.stringify(profiles);
    renderProfiles();

    // Persist new order to backend
    saveProfileOrder();
}

async function saveProfileOrder() {
    try {
        const names = profiles.map(p => p.name);
        await callGo('ReorderProfiles', names);
    } catch (err) {
        console.error('Failed to save profile order:', err);
        showToast('Failed to save order', 'error');
    }
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
    document.getElementById('btn-ai-integration').addEventListener('click', () => {
        if (currentPage === 'ai') showPage('profiles');
        else showPage('ai');
    });
    document.getElementById('btn-settings').addEventListener('click', () => {
        if (currentPage === 'settings') showPage('profiles');
        else openSettings();
    });
    document.getElementById('btn-back-settings').addEventListener('click', () => showPage('profiles'));
    document.getElementById('btn-back-ai').addEventListener('click', () => showPage('profiles'));
    document.getElementById('btn-about-page').addEventListener('click', () => {
        document.getElementById('modal-about').style.display = 'flex';
        lucide.createIcons();
    });
    document.getElementById('btn-close-about').addEventListener('click', () => {
        document.getElementById('modal-about').style.display = 'none';
    });
    document.getElementById('modal-about').addEventListener('click', (e) => {
        if (e.target.classList.contains('modal-overlay')) {
            document.getElementById('modal-about').style.display = 'none';
        }
    });

    // About page link handlers — open in external browser
    document.querySelectorAll('#modal-about a[href^="http"]').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const url = link.getAttribute('href');
            if (window.runtime && window.runtime.BrowserOpenURL) {
                window.runtime.BrowserOpenURL(url);
            } else {
                window.open(url, '_blank');
            }
        });
    });
    document.getElementById('btn-sync-header').addEventListener('click', openSyncModal);

    // Language dropdown toggle
    document.getElementById('btn-language').addEventListener('click', (e) => {
        e.stopPropagation();
        const menu = document.getElementById('lang-menu');
        menu.classList.toggle('open');
    });

    // Language option selection
    document.querySelectorAll('.lang-option').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const lang = btn.dataset.lang;
            setLanguage(lang);
            document.getElementById('lang-menu').classList.remove('open');
        });
    });

    // Close lang dropdown on outside click
    document.addEventListener('click', () => {
        document.getElementById('lang-menu').classList.remove('open');
    });

    // Theme toggle
    document.getElementById('btn-theme').addEventListener('click', () => {
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        setTheme(newTheme);
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
    document.getElementById('btn-open-config').addEventListener('click', handleOpenConfig);
    document.getElementById('btn-change-profiles-dir').addEventListener('click', handleChangeProfilesDir);

    document.getElementById('settings-browser').addEventListener('change', async (e) => {
        try {
            await callGo('SaveDefaultBrowser', e.target.value);
            showToast(t('default_browser_updated'), 'success');
        } catch (err) {
            showToast(err, 'error');
        }
    });

    // Appearance: Theme radio group in settings
    document.getElementById('settings-theme-group').addEventListener('click', (e) => {
        const btn = e.target.closest('.settings-radio-btn');
        if (!btn) return;
        const newTheme = btn.dataset.value;
        setTheme(newTheme);
        syncRadioGroup('settings-theme-group', newTheme);
    });

    // Appearance: Language radio group in settings
    document.getElementById('settings-lang-group').addEventListener('click', (e) => {
        const btn = e.target.closest('.settings-radio-btn');
        if (!btn) return;
        const lang = btn.dataset.value;
        setLanguage(lang);
        syncRadioGroup('settings-lang-group', lang);
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

    // AI Integration: Copy config buttons
    document.querySelectorAll('.ai-copy-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const codeBlock = btn.closest('.ai-code-block');
            const code = codeBlock.querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                btn.classList.add('copied');
                const icon = btn.querySelector('i');
                if (icon) {
                    icon.setAttribute('data-lucide', 'check');
                    lucide.createIcons();
                }
                showToast(t('mcp_copied'), 'success');
                setTimeout(() => {
                    btn.classList.remove('copied');
                    if (icon) {
                        icon.setAttribute('data-lucide', 'copy');
                        lucide.createIcons();
                    }
                }, 2000);
            }).catch(() => {
                showToast(t('copy_failed'), 'error');
            });
        });
    });

    // AI Integration: Copy use case command buttons
    document.querySelectorAll('.ai-usecase-copy').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const code = btn.closest('.ai-usecase-code').querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                btn.classList.add('copied');
                const icon = btn.querySelector('i');
                if (icon) { icon.setAttribute('data-lucide', 'check'); lucide.createIcons(); }
                showToast(t('mcp_copied'), 'success');
                setTimeout(() => {
                    btn.classList.remove('copied');
                    if (icon) { icon.setAttribute('data-lucide', 'copy'); lucide.createIcons(); }
                }, 2000);
            }).catch(() => showToast(t('copy_failed'), 'error'));
        });
    });

    // AI Integration: Copy prompt buttons
    document.querySelectorAll('.ai-prompt-copy').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const card = btn.closest('.ai-prompt-card');
            const text = card.querySelector('.ai-prompt-text').textContent;
            navigator.clipboard.writeText(text).then(() => {
                btn.classList.add('copied');
                const icon = btn.querySelector('i');
                if (icon) {
                    icon.setAttribute('data-lucide', 'check');
                    lucide.createIcons();
                }
                showToast(t('mcp_copied'), 'success');
                setTimeout(() => {
                    btn.classList.remove('copied');
                    if (icon) {
                        icon.setAttribute('data-lucide', 'copy');
                        lucide.createIcons();
                    }
                }, 2000);
            }).catch(() => {
                showToast(t('copy_failed'), 'error');
            });
        });
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
        showToast(t('need_2_profiles'), 'error');
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
        showToast(t('enter_name'), 'error');
        return;
    }

    showLoading(true);
    try {
        await callGo('CreateProfile', name, browserVal);
        closeModal('modal-create');
        showToast(t('profile_created', { name }), 'success');
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
        showToast(t('launched', { name }), 'success');
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
        showToast(t('profile_renamed', { old: editTargetName, new: newName }), 'success');
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
        showToast(t('profile_deleted', { name: deleteTargetName }), 'success');
        deleteTargetName = '';
        await loadProfiles();
    } catch (err) {
        if (String(err).includes('locked')) {
            if (confirm('Profile is locked. Force delete?')) {
                try {
                    await callGo('DeleteProfile', deleteTargetName, true);
                    closeModal('modal-delete');
                    showToast(t('profile_deleted', { name: deleteTargetName }), 'success');
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
        showToast(t('src_dst_diff'), 'error');
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
        showToast(t('select_dir'), 'error');
        return;
    }
    if (!name) {
        showToast(t('enter_name'), 'error');
        return;
    }

    showLoading(true);
    try {
        await callGo('ImportProfile', srcPath, name);
        closeModal('modal-import');
        showToast(t('imported', { name }), 'success');
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
        showToast(t('mcp_copied'), 'success');
    } catch {
        showToast(t('copy_failed'), 'error');
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
        await callGo('ResetSettings');
        showToast(t('settings_reset'), 'success');
        await openSettings();
    } catch (err) {
        showToast(err, 'error');
    }
}

async function handleOpenConfig() {
    try {
        await callGo('OpenConfigDir');
    } catch (err) {
        showToast(err, 'error');
    }
}

async function handleChangeProfilesDir() {
    try {
        const newDir = await callGo('ChangeProfilesDir');
        if (newDir) {
            document.getElementById('settings-profiles-dir').textContent = newDir;
            showToast('Profile storage path updated', 'success');
        }
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

// ============================================
// Theme Management
// ============================================

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    currentTheme = theme;
}

function setTheme(theme) {
    applyTheme(theme);
    localStorage.setItem('bpm-theme', theme);
    showToast(t(theme === 'dark' ? 'theme_dark' : 'theme_light'), 'info');
}

// ============================================
// Language Management
// ============================================

function setLanguage(lang) {
    currentLang = lang;
    localStorage.setItem('bpm-lang', lang);
    document.documentElement.setAttribute('lang', lang);
    applyLanguage(lang);
    showToast(t('lang_changed'), 'success');

    // Update active state in dropdown
    document.querySelectorAll('.lang-option').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.lang === lang);
    });

    // Re-render profiles for translated labels
    lastProfileHash = '';
    renderProfiles();
    lucide.createIcons();
}

function applyLanguage(lang) {
    // Static text elements — mapped by data-i18n or known IDs
    const map = {
        // Header buttons
        'btn-import-header': () => setInnerAfterIcon('btn-import-header', t('import')),
        'btn-create-header': () => setInnerAfterIcon('btn-create-header', t('create_profile')),
        // Profile section
        'section-title': () => setText('.section-title', t('browser_profiles')),
        'section-subtitle': () => setText('.section-subtitle', t('manage_desc')),
        'search-input': () => { const el = document.getElementById('search-input'); if (el) el.placeholder = t('filter_profiles'); },
        'btn-sync-header': () => setInnerAfterIcon('btn-sync-header', t('sync')),
        // Empty state
        'empty-h3': () => setText('#empty-state h3', t('no_profiles')),
        'empty-p': () => setText('#empty-state > p', t('no_profiles_desc')),
        // Settings
        'btn-back-text': () => setInnerAfterIcon('btn-back-settings', t('back')),
        'settings-title': () => setText('.settings-page-title', t('settings')),
        'settings-general': () => setCardTitle(0, t('general')),
        'settings-browser-label': () => setLabel('settings-browser', t('default_browser')),
        'settings-storage-label': () => {
            const labels = document.querySelectorAll('#page-settings .settings-card:first-of-type .form-group label');
            if (labels[1]) labels[1].textContent = t('profile_storage');
        },
        'settings-appearance': () => setCardTitle(1, t('appearance')),
        'settings-theme-label': () => {
            const card = document.querySelectorAll('#page-settings .settings-card')[1];
            if (card) { const labels = card.querySelectorAll('.form-group label'); if (labels[0]) labels[0].textContent = t('theme'); }
        },
        'settings-lang-label': () => {
            const card = document.querySelectorAll('#page-settings .settings-card')[1];
            if (card) { const labels = card.querySelectorAll('.form-group label'); if (labels[1]) labels[1].textContent = t('language'); }
        },
        'settings-mcp-title': () => setCardTitle(2, t('mcp_server')),
        'settings-mcp-hint': () => setText('.settings-hint', t('mcp_hint')),
        'settings-data-title': () => setCardTitle(3, t('data')),
        'btn-open-config': () => setInnerAfterIcon('btn-open-config', t('open_config')),
        'btn-export-all': () => setInnerAfterIcon('btn-export-all', t('export_all')),
        'btn-reset-settings': () => setInnerAfterIcon('btn-reset-settings', t('reset_settings')),
        // Create modal
        'modal-create-title': () => setText('#modal-create .modal-header h2', t('create_profile')),
        'create-name-label': () => setLabel('create-name', t('profile_name')),
        'create-name-hint': () => setText('#modal-create .form-hint', t('profile_name_hint')),
        'create-browser-label': () => setLabel('create-browser', t('browser')),
        'create-info': () => setInfoNote('#modal-create', t('create_info')),
        'btn-create-cancel': () => setText('#modal-create .btn-ghost', t('cancel')),
        'btn-create-confirm': () => { const el = document.getElementById('btn-create-confirm'); if (el) el.textContent = t('create'); },
        // Sync modal
        'modal-sync-title': () => setText('#modal-sync .modal-header h2', t('sync_credentials')),
        'sync-desc': () => setText('#modal-sync .modal-desc', t('sync_desc')),
        'sync-from-label': () => setLabel('sync-source', t('from')),
        'sync-to-label': () => setLabel('sync-target', t('to')),
        'sync-cancel': () => setText('#modal-sync .btn-ghost', t('cancel')),
        'btn-sync-confirm': () => { const el = document.getElementById('btn-sync-confirm'); if (el) el.textContent = t('sync'); },
        // Import modal
        'modal-import-title': () => setText('#modal-import .modal-header h2', t('import_chrome')),
        'import-desc': () => setText('#modal-import .modal-desc', t('import_desc')),
        'import-path-label': () => setLabel('import-path', t('profile_directory')),
        'btn-browse': () => { const el = document.getElementById('btn-browse'); if (el) el.textContent = t('browse'); },
        'import-name-label': () => setLabel('import-name', t('profile_name')),
        'import-cancel': () => setText('#modal-import .btn-ghost', t('cancel')),
        'btn-import-confirm': () => { const el = document.getElementById('btn-import-confirm'); if (el) el.textContent = t('import'); },
        // Credentials modal
        'btn-sync-from-creds': () => { const el = document.getElementById('btn-sync-from-creds'); if (el) el.textContent = t('sync_creds'); },
        'creds-th': () => {
            const ths = document.querySelectorAll('.creds-table th');
            if (ths[0]) ths[0].textContent = t('domain');
            if (ths[1]) ths[1].textContent = t('cookies');
            if (ths[2]) ths[2].textContent = t('logins');
        },
        'creds-empty': () => setText('#creds-empty p', t('no_creds')),
        'creds-close': () => setText('#modal-creds .btn-ghost', t('close')),
        // Delete modal
        'modal-delete-title': () => setText('#modal-delete .modal-header h2', t('delete_profile')),
        'delete-warning': () => setText('#modal-delete .text-danger', t('delete_warning')),
        'delete-cancel': () => setText('#modal-delete .btn-ghost', t('cancel')),
        'btn-delete-confirm': () => { const el = document.getElementById('btn-delete-confirm'); if (el) el.textContent = t('delete'); },
        // Edit modal
        'modal-edit-title': () => setText('#modal-edit .modal-header h2', t('edit_profile')),
        'edit-name-label': () => setLabel('edit-name', t('profile_name')),
        'edit-name-hint': () => setText('#modal-edit .form-hint', t('edit_name_hint')),
        'edit-dir-label': () => {
            const labels = document.querySelectorAll('#modal-edit .form-group label');
            if (labels[1]) labels[1].textContent = t('data_directory');
        },
        'edit-cancel': () => setText('#modal-edit .btn-ghost', t('cancel')),
        'btn-edit-confirm': () => { const el = document.getElementById('btn-edit-confirm'); if (el) el.textContent = t('save'); },
        // AI Integration page
        'ai-btn-header': () => setInnerAfterIcon('btn-ai-integration', t('ai_agent')),
        'ai-back-btn': () => setInnerAfterIcon('btn-back-ai', t('back')),
        'ai-warning-text': () => { const el = document.querySelector('.ai-warning-text'); if (el) el.textContent = t('ai_warning'); },
        // About page
        'about-back-btn': () => setInnerAfterIcon('btn-back-about', t('back')),
        'about-tagline': () => { const el = document.querySelector('.about-tagline'); if (el) el.textContent = t('about_tagline'); },
        'about-author-title': () => { const els = document.querySelectorAll('.about-section-title'); if (els[0]) els[0].textContent = t('about_author'); },
        'about-repo-title': () => { const els = document.querySelectorAll('.about-section-title'); if (els[1]) els[1].textContent = t('about_repo'); },
        'about-tech-title': () => { const els = document.querySelectorAll('.about-section-title'); if (els[2]) els[2].textContent = t('about_tech'); },
        'about-repo-btn': () => setInnerAfterIcon('about-repo-link', t('about_open_github')),
        'about-settings-btn': () => setInnerAfterIcon('btn-about-page', t('about')),
        'ai-badge': () => setText('.ai-badge-text', t('ai_hero_badge')),
        'ai-title': () => setText('.ai-hero-title', t('ai_hero_title')),
        'ai-tagline': () => setText('.ai-hero-tagline', t('ai_hero_tagline')),
        'ai-desc': () => setText('.ai-hero-desc', t('ai_hero_desc')),
        'ai-playwright-title': () => setText('#ai-card-playwright .ai-card-title', t('ai_playwright_title')),
        'ai-playwright-desc': () => setText('#ai-card-playwright .ai-card-desc', t('ai_playwright_desc')),
        'ai-devtools-title': () => setText('#ai-card-devtools .ai-card-title', t('ai_devtools_title')),
        'ai-devtools-desc': () => setText('#ai-card-devtools .ai-card-desc', t('ai_devtools_desc')),
        'ai-browseruse-title': () => setText('#ai-card-browseruse .ai-card-title', t('ai_browseruse_title')),
        'ai-browseruse-desc': () => setText('#ai-card-browseruse .ai-card-desc', t('ai_browseruse_desc')),
        'ai-workflow-title': () => setText('.ai-workflow-section .ai-section-title', t('ai_workflow_title')),
        'ai-step-1': () => { const labels = document.querySelectorAll('.ai-step-label'); if (labels[0]) labels[0].textContent = t('ai_step_create'); },
        'ai-step-2': () => { const labels = document.querySelectorAll('.ai-step-label'); if (labels[1]) labels[1].textContent = t('ai_step_login'); },
        'ai-step-3': () => { const labels = document.querySelectorAll('.ai-step-label'); if (labels[2]) labels[2].textContent = t('ai_step_agent'); },
        'ai-step-4': () => { const labels = document.querySelectorAll('.ai-step-label'); if (labels[3]) labels[3].textContent = t('ai_step_automate'); },
        'ai-comparison-title': () => setText('.ai-comparison-section .ai-section-title', t('ai_comparison_title')),
        'ai-warning': () => setText('.ai-warning-box span', t('ai_warning')),
    };

    Object.values(map).forEach(fn => { try { fn(); } catch {} });
}

// --- i18n helper functions ---

function setText(selector, text) {
    const el = document.querySelector(selector);
    if (el) el.textContent = text;
}

function setLabel(forId, text) {
    const el = document.querySelector(`label[for="${forId}"]`);
    if (el) el.textContent = text;
}

function setInnerAfterIcon(id, text) {
    const el = document.getElementById(id);
    if (!el) return;
    // Preserve the icon <i> or <svg>, replace text
    const icon = el.querySelector('i, svg');
    if (icon) {
        el.textContent = '';
        el.appendChild(icon);
        el.append(` ${text}`);
    } else {
        el.textContent = text;
    }
}

function setCardTitle(index, text) {
    const titles = document.querySelectorAll('#page-settings .settings-card-title');
    if (titles[index]) {
        const icon = titles[index].querySelector('i');
        titles[index].textContent = '';
        if (icon) titles[index].appendChild(icon);
        titles[index].append(`\n                ${text}`);
    }
}

function setInfoNote(modalSelector, text) {
    const note = document.querySelector(`${modalSelector} .modal-info-note`);
    if (note) {
        const icon = note.querySelector('i, svg');
        note.textContent = '';
        if (icon) note.appendChild(icon);
        note.append(` ${text}`);
    }
}
