// ─────────────────────────────────────────────────────────────────────────────
// IAM Admin Console — JavaScript
// ─────────────────────────────────────────────────────────────────────────────

const IAM_API = (() => {
    let base = window.BACKEND_URL || 'http://localhost:8000';
    if (base.endsWith('/')) base = base.slice(0, -1);
    return base;
})();

// IAM-specific auth token (separate from platform auth)
let iamToken = localStorage.getItem('iamToken') || '';
let iamRefreshToken = localStorage.getItem('iamRefreshToken') || '';
let iamServiceAccessInfo = null;
let iamLastGeneratedTokenResponse = null;
let iamSelectedRealm = '';

function clearIAMSession() {
    iamToken = '';
    iamRefreshToken = '';
    localStorage.removeItem('iamToken');
    localStorage.removeItem('iamRefreshToken');
    updateIAMLoginBanner();
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function iamHeaders() {
    const h = { 'Content-Type': 'application/json' };
    if (iamToken) h['Authorization'] = 'Bearer ' + iamToken;
    return h;
}

async function iamFetch(path, opts) {
    opts = opts || {};
    opts.headers = Object.assign(iamHeaders(), opts.headers || {});
    let resp = await fetch(IAM_API + path, opts);
    if (resp.status === 401 && iamRefreshToken) {
        const refreshed = await tryIAMRefresh();
        if (refreshed) {
            opts.headers['Authorization'] = 'Bearer ' + iamToken;
            resp = await fetch(IAM_API + path, opts);
        }
    }

    if (resp.status === 401) {
        clearIAMSession();
    }

    return resp;
}

async function tryIAMRefresh() {
    try {
        const resp = await fetch(IAM_API + '/iam/auth/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: iamRefreshToken })
        });
        if (!resp.ok) {
            clearIAMSession();
            return false;
        }
        const data = await resp.json();
        iamToken = data.access_token || '';
        if (!iamToken) {
            clearIAMSession();
            return false;
        }
        iamRefreshToken = data.refresh_token || iamRefreshToken;
        localStorage.setItem('iamToken', iamToken);
        localStorage.setItem('iamRefreshToken', iamRefreshToken);
        return true;
    } catch (_) {
        clearIAMSession();
        return false;
    }
}

function showIAMToast(message, isError) {
    const el = document.createElement('div');
    el.textContent = message;
    el.style.cssText = 'position:fixed;top:20px;right:20px;z-index:100000;padding:12px 24px;border-radius:8px;color:#fff;font-weight:600;font-size:.9rem;box-shadow:0 4px 12px rgba(0,0,0,.3);transition:opacity .3s;' +
        (isError ? 'background:#e74c3c;' : 'background:#27ae60;');
    document.body.appendChild(el);
    setTimeout(() => { el.style.opacity = '0'; setTimeout(() => el.remove(), 300); }, 3000);
}

async function copyToClipboard(text) {
    const value = (text || '').trim();
    if (!value) return false;
    try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            await navigator.clipboard.writeText(value);
            return true;
        }
    } catch (_) {}

    const ta = document.createElement('textarea');
    ta.value = value;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    return ok;
}

async function copyTextFromElement(id) {
    const el = document.getElementById(id);
    if (!el) {
        showIAMToast('Copy source not found', true);
        return;
    }
    const value = typeof el.value === 'string' ? el.value : el.textContent;
    const ok = await copyToClipboard(value || '');
    showIAMToast(ok ? 'Copied' : 'Copy failed', !ok);
}

function setFieldText(id, value) {
    const el = document.getElementById(id);
    if (!el) return;
    if (typeof el.value === 'string') {
        el.value = value || '';
    } else {
        el.textContent = value || '–';
    }
}

function getServiceTokenEndpoint() {
    const endpoints = iamServiceAccessInfo && iamServiceAccessInfo.endpoints ? iamServiceAccessInfo.endpoints : {};
    return endpoints.keycloak_token || endpoints.iam_token || (IAM_API + '/oauth/token');
}

function buildClientCredentialsSnippet(clientID, clientSecret, scope) {
    let cmd = "curl --request POST '" + getServiceTokenEndpoint() + "'" +
        " --header 'Content-Type: application/x-www-form-urlencoded'" +
        " --data-urlencode 'grant_type=client_credentials'" +
        " --data-urlencode 'client_id=" + (clientID || '') + "'" +
        " --data-urlencode 'client_secret=" + (clientSecret || '') + "'";
    if (scope) {
        cmd += " --data-urlencode 'scope=" + scope + "'";
    }
    return cmd;
}

function applyServiceRealm() {
    const input = document.getElementById('iamServiceRealmInput');
    const requestedRealm = input ? (input.value || '').trim() : '';
    loadServiceAccessInfo(requestedRealm);
}

function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str || '';
    return div.innerHTML;
}

function formatDate(iso) {
    if (!iso) return '–';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return iso;
    return d.toLocaleString();
}

// ── Tab Switching ────────────────────────────────────────────────────────────

function switchIAMTab(tabId) {
    document.querySelectorAll('.admin-container .tab-content').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.admin-container .tab-btn').forEach(b => b.classList.remove('active'));

    const target = document.getElementById(tabId);
    if (target) target.classList.add('active');

    const tabs = document.querySelectorAll('.admin-container .tab-btn');
    tabs.forEach(b => {
        if (b.getAttribute('onclick') && b.getAttribute('onclick').includes(tabId)) {
            b.classList.add('active');
        }
    });

    // Lazy-load data on tab switch
    if (tabId === 'iam-users') loadIAMUsers();
    if (tabId === 'iam-clients') {
        loadIAMClients();
        loadServiceAccessInfo();
    }
    if (tabId === 'iam-roles') loadIAMRoles();
    if (tabId === 'iam-bindings') loadIAMBindings();
    if (tabId === 'iam-oidc') loadOIDCInfo();
    if (tabId === 'iam-dashboard') loadIAMDashboard();
}

// ── Modal helpers ────────────────────────────────────────────────────────────

function openIAMModal(id) {
    const m = document.getElementById(id);
    if (m) m.style.display = 'flex';
}

function closeIAMModal(id) {
    const m = document.getElementById(id);
    if (m) m.style.display = 'none';
}

// Close modals on background click
document.addEventListener('click', function(e) {
    if (e.target.classList.contains('modal') && e.target.id && e.target.id.startsWith('iam')) {
        e.target.style.display = 'none';
    }
});

// ── IAM Login ────────────────────────────────────────────────────────────────

async function iamLogin() {
    const identifier = document.getElementById('iamLoginEmail').value.trim();
    const password = document.getElementById('iamLoginPassword').value;
    if (!identifier || !password) { showIAMToast('Username/email and password required', true); return; }

    try {
        const resp = await fetch(IAM_API + '/iam/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email: identifier, password: password })
        });
        const data = await resp.json();
        if (!resp.ok) {
            showIAMToast(data.error || 'Login failed', true);
            return;
        }
        iamToken = data.access_token || '';
        iamRefreshToken = data.refresh_token || '';
        localStorage.setItem('iamToken', iamToken);
        localStorage.setItem('iamRefreshToken', iamRefreshToken);
        showIAMToast('IAM login successful');
        updateIAMLoginBanner();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Login error: ' + err.message, true);
    }
}

function iamLogout() {
    iamToken = '';
    iamRefreshToken = '';
    localStorage.removeItem('iamToken');
    localStorage.removeItem('iamRefreshToken');
    updateIAMLoginBanner();
    showIAMToast('IAM session ended');
}

function updateIAMLoginBanner() {
    const banner = document.getElementById('iamLoginBanner');
    if (banner) banner.style.display = iamToken ? 'none' : 'block';
}

// ── Dashboard ────────────────────────────────────────────────────────────────

async function loadIAMDashboard() {
    if (!iamToken) { updateIAMLoginBanner(); return; }

    // WhoAmI
    try {
        const resp = await iamFetch('/iam/auth/whoami');
        const whoami = document.getElementById('iamWhoAmI');
        if (resp.ok) {
            const data = await resp.json();
            if (whoami) whoami.textContent = JSON.stringify(data, null, 2);
        } else {
            if (whoami) whoami.textContent = 'Error: ' + resp.status;
        }
    } catch (_) {}

    // Counts
    try {
        const [uResp, cResp, rResp] = await Promise.all([
            iamFetch('/iam/admin/users'),
            iamFetch('/iam/admin/clients'),
            iamFetch('/iam/admin/roles')
        ]);
        if (uResp.ok) {
            const users = await uResp.json();
            const count = Array.isArray(users) ? users.length : (users.users ? users.users.length : 0);
            setTextById('dashUserCount', count);
            setTextById('iamStatUsers', count);
        }
        if (cResp.ok) {
            const clients = await cResp.json();
            const count = Array.isArray(clients) ? clients.length : (clients.clients ? clients.clients.length : 0);
            setTextById('dashClientCount', count);
            setTextById('iamStatClients', count);
        }
        if (rResp.ok) {
            const roles = await rResp.json();
            const count = Array.isArray(roles) ? roles.length : (roles.roles ? roles.roles.length : 0);
            setTextById('dashRoleCount', count);
            setTextById('iamStatRoles', count);
        }
    } catch (_) {}
}

function setTextById(id, val) {
    const el = document.getElementById(id);
    if (el) el.textContent = val;
}

// ── Users CRUD ───────────────────────────────────────────────────────────────

let iamUsersCache = [];
let iamClientsCache = [];

async function loadIAMUsers() {
    if (!iamToken) return;
    try {
        const resp = await iamFetch('/iam/admin/users');
        if (!resp.ok) { showIAMToast('Failed to load users: ' + resp.status, true); return; }
        const data = await resp.json();
        iamUsersCache = Array.isArray(data) ? data : (data.users || []);
        renderIAMUsers(iamUsersCache);
    } catch (err) {
        showIAMToast('Error loading users: ' + err.message, true);
    }
}

function renderIAMUsers(users) {
    const tbody = document.getElementById('iamUsersBody');
    if (!tbody) return;
    if (!users.length) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;color:var(--text-muted);">No users found</td></tr>';
        return;
    }
    tbody.innerHTML = users.map(u => {
        const active = u.active !== false;
        const verified = u.email_verified === true;
        return '<tr>' +
            '<td>' + escapeHtml(u.email) + '</td>' +
            '<td>' + escapeHtml(u.display_name || '–') + '</td>' +
            '<td><span class="status-badge ' + (active ? 'status-active' : 'status-inactive') + '">' + (active ? 'Active' : 'Inactive') + '</span></td>' +
            '<td><span class="status-badge ' + (verified ? 'status-active' : 'status-inactive') + '">' + (verified ? 'Yes' : 'No') + '</span></td>' +
        const canRegenerateSecret = !c.public;
            '<td>' + formatDate(u.created_at) + '</td>' +
            '<td style="white-space:nowrap;">' +
                '<button class="btn-sm btn-primary" onclick="viewUser(\'' + escapeHtml(u.id) + '\')">View</button> ' +
                '<button class="btn-sm btn-secondary" onclick="editUser(\'' + escapeHtml(u.id) + '\')">Edit</button> ' +
                '<button class="btn-sm" style="background:var(--danger-color);color:#fff;" onclick="deleteUser(\'' + escapeHtml(u.id) + '\')">Delete</button>' +
            '</td></tr>';
    }).join('');
}

                (canRegenerateSecret ? '<button class="btn-sm btn-secondary" onclick="regenerateClientSecret(\'' + escapeHtml(c.id) + '\')">Regen Secret</button> ' : '') +
                '<button class="btn-sm btn-secondary" onclick="showClientIDChangeModal(\'' + escapeHtml(c.id) + '\')">Change ID</button> ' +
function filterIAMUsers() {
    const q = (document.getElementById('iamUserSearch').value || '').toLowerCase();
    if (!q) { renderIAMUsers(iamUsersCache); return; }
    renderIAMUsers(iamUsersCache.filter(u =>
        (u.email || '').toLowerCase().includes(q) ||
        (u.display_name || '').toLowerCase().includes(q)
    ));
}

function showCreateUserModal() {
    document.getElementById('iamUserEditID').value = '';
    document.getElementById('iamUserEmail').value = '';
    document.getElementById('iamUserDisplayName').value = '';
    document.getElementById('iamUserPassword').value = '';
    document.getElementById('iamUserActive').checked = true;
    document.getElementById('iamUserEmailVerified').checked = false;
    document.getElementById('iamUserPasswordGroup').style.display = '';
    document.getElementById('iamUserModalTitle').textContent = 'Create User';
    document.getElementById('iamUserSubmitBtn').textContent = 'Create';
    openIAMModal('iamUserModal');
}

async function viewUser(id) {
    try {
        const resp = await iamFetch('/iam/admin/users/' + encodeURIComponent(id));
        if (!resp.ok) { showIAMToast('Failed to load user', true); return; }
        const data = await resp.json();
        document.getElementById('iamUserDetailContent').textContent = JSON.stringify(data, null, 2);
        openIAMModal('iamUserDetailModal');
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function editUser(id) {
    try {
        const resp = await iamFetch('/iam/admin/users/' + encodeURIComponent(id));
        if (!resp.ok) { showIAMToast('Failed to load user', true); return; }
        const u = await resp.json();
        document.getElementById('iamUserEditID').value = u.id;
        document.getElementById('iamUserEmail').value = u.email || '';
        document.getElementById('iamUserDisplayName').value = u.display_name || '';
        document.getElementById('iamUserPassword').value = '';
        document.getElementById('iamUserActive').checked = u.active !== false;
        document.getElementById('iamUserEmailVerified').checked = u.email_verified === true;
        document.getElementById('iamUserPasswordGroup').style.display = 'none';
        document.getElementById('iamUserModalTitle').textContent = 'Edit User';
        document.getElementById('iamUserSubmitBtn').textContent = 'Update';
        openIAMModal('iamUserModal');
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function submitUser() {
    const editID = document.getElementById('iamUserEditID').value;
    const body = {
        email: document.getElementById('iamUserEmail').value.trim(),
        display_name: document.getElementById('iamUserDisplayName').value.trim(),
        active: document.getElementById('iamUserActive').checked,
        email_verified: document.getElementById('iamUserEmailVerified').checked
    };
    if (!editID) {
        body.password = document.getElementById('iamUserPassword').value;
        if (!body.email || !body.password) { showIAMToast('Email and password are required', true); return; }
    }

    try {
        const url = editID ? '/iam/admin/users/' + encodeURIComponent(editID) : '/iam/admin/users';
        const method = editID ? 'PUT' : 'POST';
        const resp = await iamFetch(url, { method: method, body: JSON.stringify(body) });
        const data = await resp.json();
        if (!resp.ok) { showIAMToast(data.error || 'Operation failed', true); return; }
        showIAMToast(editID ? 'User updated' : 'User created');
        closeIAMModal('iamUserModal');
        loadIAMUsers();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function deleteUser(id) {
    if (!confirm('Delete this user permanently?')) return;
    try {
        const resp = await iamFetch('/iam/admin/users/' + encodeURIComponent(id), { method: 'DELETE' });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Delete failed', true);
            return;
        }
        showIAMToast('User deleted');
        loadIAMUsers();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

// ── Clients CRUD ─────────────────────────────────────────────────────────────

async function loadIAMClients() {
    if (!iamToken) return;
    try {
        const resp = await iamFetch('/iam/admin/clients');
        if (!resp.ok) { showIAMToast('Failed to load clients: ' + resp.status, true); return; }
        const data = await resp.json();
        const clients = Array.isArray(data) ? data : (data.clients || []);
        iamClientsCache = clients;
        renderIAMClients(clients);
    } catch (err) {
        showIAMToast('Error loading clients: ' + err.message, true);
    }
}

function renderIAMClients(clients) {
    const tbody = document.getElementById('iamClientsBody');
    if (!tbody) return;
    if (!clients.length) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;color:var(--text-muted);">No clients registered</td></tr>';
        return;
    }
    tbody.innerHTML = clients.map(c => {
        const grants = (c.grant_types || []).join(', ') || '–';
        const uris = (c.redirect_uris || []).map(u => escapeHtml(u)).join('<br>') || '–';
        const scopes = (c.scopes || []).join(', ') || '–';
        const serviceRoles = (c.service_roles || []).join(', ') || '–';
        const mode = c.public ? 'Public' : 'Confidential';
        const hasClientCredentials = Array.isArray(c.grant_types) && c.grant_types.includes('client_credentials');
        const canGenerateToken = hasClientCredentials && !c.public;
        return '<tr>' +
            '<td style="font-family:var(--font-mono);font-size:.85rem;">' + escapeHtml(c.id) + '</td>' +
            '<td>' + escapeHtml(c.name) + '<div style="font-size:.75rem;color:var(--text-muted);margin-top:4px;">' + escapeHtml(mode) + '</div></td>' +
            '<td>' + escapeHtml(grants) + '</td>' +
            '<td style="font-size:.8rem;">' + uris + '</td>' +
            '<td>' + escapeHtml(scopes) + '</td>' +
            '<td>' + escapeHtml(serviceRoles) + '</td>' +
            '<td style="white-space:nowrap;">' +
                (canGenerateToken ? '<button class="btn-sm btn-primary" onclick="showGenerateTokenModal(\'' + escapeHtml(c.id) + '\')">Generate Token</button> ' : '') +
                '<button class="btn-sm btn-secondary" onclick="editClient(\'' + escapeHtml(c.id) + '\')">Edit</button> ' +
                '<button class="btn-sm" style="background:var(--danger-color);color:#fff;" onclick="deleteClient(\'' + escapeHtml(c.id) + '\')">Delete</button>' +
            '</td></tr>';
    }).join('');
}

function showCreateClientModal() {
    document.getElementById('iamClientEditID').value = '';
    document.getElementById('iamClientName').value = '';
    document.getElementById('iamClientRedirectURIs').value = '';
    document.getElementById('iamClientGrantTypes').value = 'authorization_code,refresh_token,client_credentials';
    document.getElementById('iamClientScopes').value = 'openid,profile,email';
    document.getElementById('iamClientServiceRoles').value = '';
    document.getElementById('iamClientPublic').checked = false;
    document.getElementById('iamClientModalTitle').textContent = 'Register OAuth Client';
    document.getElementById('iamClientSubmitBtn').textContent = 'Register';
    openIAMModal('iamClientModal');
}

async function editClient(id) {
    try {
        const resp = await iamFetch('/iam/admin/clients/' + encodeURIComponent(id));
        if (!resp.ok) { showIAMToast('Failed to load client', true); return; }
        const c = await resp.json();
        document.getElementById('iamClientEditID').value = c.id;
        document.getElementById('iamClientName').value = c.name || '';
        document.getElementById('iamClientRedirectURIs').value = (c.redirect_uris || []).join('\n');
        document.getElementById('iamClientGrantTypes').value = (c.grant_types || []).join(',');
        document.getElementById('iamClientScopes').value = (c.scopes || []).join(',');
        document.getElementById('iamClientServiceRoles').value = (c.service_roles || []).join(',');
        document.getElementById('iamClientPublic').checked = c.public === true;
        document.getElementById('iamClientModalTitle').textContent = 'Edit Client';
        document.getElementById('iamClientSubmitBtn').textContent = 'Update';
        openIAMModal('iamClientModal');
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function submitClient() {
    const editID = document.getElementById('iamClientEditID').value;
    const name = document.getElementById('iamClientName').value.trim();
    if (!name) { showIAMToast('Client name is required', true); return; }

    const redirectURIs = document.getElementById('iamClientRedirectURIs').value
        .split('\n').map(s => s.trim()).filter(Boolean);
    const grantTypes = document.getElementById('iamClientGrantTypes').value
        .split(',').map(s => s.trim()).filter(Boolean);
    const scopes = document.getElementById('iamClientScopes').value
        .split(',').map(s => s.trim()).filter(Boolean);
    const serviceRoles = document.getElementById('iamClientServiceRoles').value
        .split(',').map(s => s.trim()).filter(Boolean);
    const publicClient = document.getElementById('iamClientPublic').checked;

    if (publicClient && grantTypes.includes('client_credentials')) {
        showIAMToast('Public clients cannot use client_credentials grant', true);
        return;
    }

    const body = {
        name: name,
        redirect_uris: redirectURIs,
        grant_types: grantTypes,
        scopes: scopes,
        service_roles: serviceRoles,
        public: publicClient
    };

    try {
        const url = editID ? '/iam/admin/clients/' + encodeURIComponent(editID) : '/iam/admin/clients';
        const method = editID ? 'PUT' : 'POST';
        const resp = await iamFetch(url, { method: method, body: JSON.stringify(body) });
        const data = await resp.json();
        if (!resp.ok) { showIAMToast(data.error || 'Operation failed', true); return; }

        closeIAMModal('iamClientModal');

        // Show secret for new clients
        if (!editID && data.client_secret) {
            const clientID = data.id || data.client_id || '';
            const clientSecret = data.client_secret || '';
            const defaultScope = (data.scopes || []).join(' ');
            showClientSecretModal(clientID, clientSecret, defaultScope, '🔑 Client Registered');
        } else {
            showIAMToast(editID ? 'Client updated' : 'Client registered');
        }
        loadIAMClients();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

function showClientSecretModal(clientID, clientSecret, scope, title) {
    setFieldText('iamClientSecretModalTitle', title || '🔑 Client Secret');
    setFieldText('iamNewClientID', clientID || '');
    setFieldText('iamNewClientSecret', clientSecret || '');
    setFieldText('iamClientTokenEndpoint', getServiceTokenEndpoint());
    setFieldText('iamClientCredentialsSnippet', buildClientCredentialsSnippet(clientID || '', clientSecret || '', scope || ''));
    openIAMModal('iamClientSecretModal');
}

async function copyClientSecret() {
    const secret = document.getElementById('iamNewClientSecret').value;
    const ok = await copyToClipboard(secret);
    showIAMToast(ok ? 'Secret copied' : 'Copy failed', !ok);
}

async function copyClientCredentialsSnippet() {
    const snippet = document.getElementById('iamClientCredentialsSnippet').textContent;
    const ok = await copyToClipboard(snippet || '');
    showIAMToast(ok ? 'Request copied' : 'Copy failed', !ok);
}

function getClientByID(clientID) {
    return iamClientsCache.find(c => c.id === clientID) || null;
}

function showClientIDChangeModal(clientID) {
    setFieldText('iamCurrentClientID', clientID || '');
    setFieldText('iamNewClientIDInput', clientID || '');
    openIAMModal('iamClientIDChangeModal');
}

async function submitClientIDChange() {
    const currentID = (document.getElementById('iamCurrentClientID').value || '').trim();
    const newClientID = (document.getElementById('iamNewClientIDInput').value || '').trim();

    if (!currentID || !newClientID) {
        showIAMToast('Both current and new client IDs are required.', true);
        return;
    }

    try {
        const resp = await iamFetch('/iam/admin/clients/' + encodeURIComponent(currentID) + '/client-id', {
            method: 'PUT',
            body: JSON.stringify({ new_client_id: newClientID })
        });
        const data = await resp.json();
        if (!resp.ok) {
            showIAMToast(data.error || 'Client ID update failed', true);
            return;
        }

        showIAMToast('Client ID updated');
        closeIAMModal('iamClientIDChangeModal');
        loadIAMClients();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function regenerateClientSecret(clientID) {
    if (!confirm('Regenerate secret for this client? Existing secret integrations will stop working.')) return;

    const client = getClientByID(clientID);
    if (!client) {
        showIAMToast('Client not found. Refresh and try again.', true);
        return;
    }
    if (client.public) {
        showIAMToast('Public clients do not have secrets.', true);
        return;
    }

    try {
        const resp = await iamFetch('/iam/admin/clients/' + encodeURIComponent(clientID) + '/regenerate-secret', {
            method: 'POST'
        });
        const data = await resp.json();
        if (!resp.ok) {
            showIAMToast(data.error || 'Secret regeneration failed', true);
            return;
        }

        const scope = (data.scopes || client.scopes || []).join(' ');
        showClientSecretModal(data.client_id || clientID, data.client_secret || '', scope, '🔁 Client Secret Regenerated');
        showIAMToast('Client secret regenerated');
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

function showGenerateTokenModal(clientID) {
    const client = getClientByID(clientID);
    if (!client) {
        showIAMToast('Client not found. Refresh and try again.', true);
        return;
    }

    if (client.public) {
        showIAMToast('Public clients do not support client_credentials token generation.', true);
        return;
    }

    if (!Array.isArray(client.grant_types) || !client.grant_types.includes('client_credentials')) {
        showIAMToast('This client does not allow client_credentials grant.', true);
        return;
    }

    setFieldText('iamGenerateTokenClientID', client.id || '');
    setFieldText('iamGenerateTokenClientSecret', '');
    setFieldText('iamGenerateTokenScope', (client.scopes || []).join(' '));
    setFieldText('iamGenerateTokenEndpoint', getServiceTokenEndpoint());
    setFieldText('iamGenerateTokenResponse', 'No token generated yet.');
    iamLastGeneratedTokenResponse = null;

    openIAMModal('iamGenerateTokenModal');
}

async function generateClientTestToken() {
    const clientID = (document.getElementById('iamGenerateTokenClientID').value || '').trim();
    const clientSecret = (document.getElementById('iamGenerateTokenClientSecret').value || '').trim();
    const scope = (document.getElementById('iamGenerateTokenScope').value || '').trim();
    const endpoint = (document.getElementById('iamGenerateTokenEndpoint').value || '').trim();

    if (!clientID || !clientSecret || !endpoint) {
        showIAMToast('Client ID, secret, and endpoint are required.', true);
        return;
    }

    const form = new URLSearchParams();
    form.set('grant_type', 'client_credentials');
    form.set('client_id', clientID);
    form.set('client_secret', clientSecret);
    if (scope) form.set('scope', scope);

    try {
        const resp = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: form.toString()
        });

        const raw = await resp.text();
        let parsed = null;
        try {
            parsed = raw ? JSON.parse(raw) : {};
        } catch (_) {
            parsed = { raw_response: raw };
        }

        iamLastGeneratedTokenResponse = parsed;
        setFieldText('iamGenerateTokenResponse', JSON.stringify(parsed, null, 2));

        if (!resp.ok) {
            const message = parsed && parsed.error ? parsed.error : ('Token generation failed: ' + resp.status);
            showIAMToast(message, true);
            return;
        }

        showIAMToast('Test token generated');
    } catch (err) {
        setFieldText('iamGenerateTokenResponse', 'Connection error: ' + err.message);
        showIAMToast('Token request failed: ' + err.message, true);
    }
}

async function copyGeneratedAccessToken() {
    const token = iamLastGeneratedTokenResponse && iamLastGeneratedTokenResponse.access_token ? iamLastGeneratedTokenResponse.access_token : '';
    if (!token) {
        showIAMToast('No generated access token to copy.', true);
        return;
    }
    const ok = await copyToClipboard(token);
    showIAMToast(ok ? 'Access token copied' : 'Copy failed', !ok);
}

async function deleteClient(id) {
    if (!confirm('Delete this OAuth client?')) return;
    try {
        const resp = await iamFetch('/iam/admin/clients/' + encodeURIComponent(id), { method: 'DELETE' });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Delete failed', true);
            return;
        }
        showIAMToast('Client deleted');
        loadIAMClients();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

// ── Roles CRUD ───────────────────────────────────────────────────────────────

async function loadIAMRoles() {
    if (!iamToken) return;
    try {
        const resp = await iamFetch('/iam/admin/roles');
        if (!resp.ok) { showIAMToast('Failed to load roles: ' + resp.status, true); return; }
        const data = await resp.json();
        const roles = Array.isArray(data) ? data : (data.roles || []);
        renderIAMRoles(roles);
    } catch (err) {
        showIAMToast('Error loading roles: ' + err.message, true);
    }
}

function renderIAMRoles(roles) {
    const tbody = document.getElementById('iamRolesBody');
    if (!tbody) return;
    if (!roles.length) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;color:var(--text-muted);">No roles defined</td></tr>';
        return;
    }
    tbody.innerHTML = roles.map(r => {
        const isSystem = r.system === true;
        const perms = (r.permissions || []).map(p =>
            '<span class="status-badge" style="font-size:.75rem;margin:2px;">' + escapeHtml(p.resource + ':' + p.action) + '</span>'
        ).join(' ') || '–';
        return '<tr>' +
            '<td style="font-weight:600;">' + escapeHtml(r.name) + '</td>' +
            '<td>' + escapeHtml(r.description || '–') + '</td>' +
            '<td><span class="status-badge ' + (isSystem ? 'status-active' : '') + '">' + (isSystem ? 'System' : 'Custom') + '</span></td>' +
            '<td>' + perms + '</td>' +
            '<td style="white-space:nowrap;">' +
                (isSystem ? '<span style="color:var(--text-muted);font-size:.8rem;">Protected</span>' :
                    '<button class="btn-sm btn-secondary" onclick="editRole(\'' + escapeHtml(r.id) + '\')">Edit</button> ' +
                    '<button class="btn-sm" style="background:var(--danger-color);color:#fff;" onclick="deleteRole(\'' + escapeHtml(r.id) + '\')">Delete</button>'
                ) +
            '</td></tr>';
    }).join('');
}

function showCreateRoleModal() {
    document.getElementById('iamRoleEditID').value = '';
    document.getElementById('iamRoleName').value = '';
    document.getElementById('iamRoleDescription').value = '';
    document.getElementById('iamRolePermissions').value = '';
    document.getElementById('iamRoleModalTitle').textContent = 'Create Role';
    document.getElementById('iamRoleSubmitBtn').textContent = 'Create';
    openIAMModal('iamRoleModal');
}

async function editRole(id) {
    try {
        const resp = await iamFetch('/iam/admin/roles/' + encodeURIComponent(id));
        if (!resp.ok) { showIAMToast('Failed to load role', true); return; }
        const r = await resp.json();
        document.getElementById('iamRoleEditID').value = r.id;
        document.getElementById('iamRoleName').value = r.name || '';
        document.getElementById('iamRoleDescription').value = r.description || '';
        document.getElementById('iamRolePermissions').value = (r.permissions || [])
            .map(p => p.resource + ':' + p.action).join('\n');
        document.getElementById('iamRoleModalTitle').textContent = 'Edit Role';
        document.getElementById('iamRoleSubmitBtn').textContent = 'Update';
        openIAMModal('iamRoleModal');
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function submitRole() {
    const editID = document.getElementById('iamRoleEditID').value;
    const name = document.getElementById('iamRoleName').value.trim();
    if (!name) { showIAMToast('Role name is required', true); return; }

    const permissions = document.getElementById('iamRolePermissions').value
        .split('\n').map(s => s.trim()).filter(Boolean)
        .map(line => {
            const parts = line.split(':');
            return { resource: parts[0] || '*', action: parts[1] || '*' };
        });

    const body = {
        name: name,
        description: document.getElementById('iamRoleDescription').value.trim(),
        permissions: permissions
    };

    try {
        const url = editID ? '/iam/admin/roles/' + encodeURIComponent(editID) : '/iam/admin/roles';
        const method = editID ? 'PUT' : 'POST';
        const resp = await iamFetch(url, { method: method, body: JSON.stringify(body) });
        const data = await resp.json();
        if (!resp.ok) { showIAMToast(data.error || 'Operation failed', true); return; }
        showIAMToast(editID ? 'Role updated' : 'Role created');
        closeIAMModal('iamRoleModal');
        loadIAMRoles();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function deleteRole(id) {
    if (!confirm('Delete this role?')) return;
    try {
        const resp = await iamFetch('/iam/admin/roles/' + encodeURIComponent(id), { method: 'DELETE' });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Delete failed', true);
            return;
        }
        showIAMToast('Role deleted');
        loadIAMRoles();
        loadIAMDashboard();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

// ── Role Bindings ────────────────────────────────────────────────────────────

async function loadIAMBindings() {
    if (!iamToken) return;
    try {
        const resp = await iamFetch('/iam/admin/role-bindings');
        if (!resp.ok) { showIAMToast('Failed to load bindings: ' + resp.status, true); return; }
        const data = await resp.json();
        const bindings = Array.isArray(data) ? data : (data.bindings || []);
        renderIAMBindings(bindings);
    } catch (err) {
        showIAMToast('Error loading bindings: ' + err.message, true);
    }
}

function renderIAMBindings(bindings) {
    const tbody = document.getElementById('iamBindingsBody');
    if (!tbody) return;
    if (!bindings.length) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;color:var(--text-muted);">No role bindings</td></tr>';
        return;
    }
    tbody.innerHTML = bindings.map(b => {
        return '<tr>' +
            '<td style="font-family:var(--font-mono);font-size:.8rem;">' + escapeHtml(b.id) + '</td>' +
            '<td style="font-family:var(--font-mono);font-size:.8rem;">' + escapeHtml(b.user_id) + '</td>' +
            '<td style="font-family:var(--font-mono);font-size:.8rem;">' + escapeHtml(b.role_id) + '</td>' +
            '<td>' + formatDate(b.assigned_at) + '</td>' +
            '<td><button class="btn-sm" style="background:var(--danger-color);color:#fff;" onclick="revokeBinding(\'' + escapeHtml(b.id) + '\')">Revoke</button></td>' +
            '</tr>';
    }).join('');
}

async function showAssignRoleModal() {
    // Populate user / role dropdowns
    const userSel = document.getElementById('iamBindingUserID');
    const roleSel = document.getElementById('iamBindingRoleID');
    userSel.innerHTML = '<option value="">Loading...</option>';
    roleSel.innerHTML = '<option value="">Loading...</option>';
    openIAMModal('iamBindingModal');

    try {
        const [uResp, rResp] = await Promise.all([
            iamFetch('/iam/admin/users'),
            iamFetch('/iam/admin/roles')
        ]);
        if (uResp.ok) {
            const data = await uResp.json();
            const users = Array.isArray(data) ? data : (data.users || []);
            userSel.innerHTML = '<option value="">— Select user —</option>' +
                users.map(u => '<option value="' + escapeHtml(u.id) + '">' + escapeHtml(u.email) + '</option>').join('');
        }
        if (rResp.ok) {
            const data = await rResp.json();
            const roles = Array.isArray(data) ? data : (data.roles || []);
            roleSel.innerHTML = '<option value="">— Select role —</option>' +
                roles.map(r => '<option value="' + escapeHtml(r.id) + '">' + escapeHtml(r.name) + '</option>').join('');
        }
    } catch (_) {
        showIAMToast('Failed to populate dropdowns', true);
    }
}

async function submitBinding() {
    const userID = document.getElementById('iamBindingUserID').value;
    const roleID = document.getElementById('iamBindingRoleID').value;
    if (!userID || !roleID) { showIAMToast('Select both user and role', true); return; }

    try {
        const resp = await iamFetch('/iam/admin/role-bindings', {
            method: 'POST',
            body: JSON.stringify({ user_id: userID, role_id: roleID })
        });
        const data = await resp.json();
        if (!resp.ok) { showIAMToast(data.error || 'Assignment failed', true); return; }
        showIAMToast('Role assigned');
        closeIAMModal('iamBindingModal');
        loadIAMBindings();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function revokeBinding(id) {
    if (!confirm('Revoke this role assignment?')) return;
    try {
        const resp = await iamFetch('/iam/admin/role-bindings/' + encodeURIComponent(id), { method: 'DELETE' });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Revoke failed', true);
            return;
        }
        showIAMToast('Role binding revoked');
        loadIAMBindings();
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

// ── Token Management ─────────────────────────────────────────────────────────

async function revokeToken() {
    const jti = document.getElementById('revokeTokenJTI').value.trim();
    if (!jti) { showIAMToast('Enter a token JTI', true); return; }
    try {
        const resp = await iamFetch('/iam/admin/tokens/revoke', {
            method: 'POST',
            body: JSON.stringify({ jti: jti })
        });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Revoke failed', true);
            return;
        }
        showIAMToast('Token revoked');
        document.getElementById('revokeTokenJTI').value = '';
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

async function revokeUserTokens() {
    const uid = document.getElementById('revokeUserID').value.trim();
    if (!uid) { showIAMToast('Enter a user ID', true); return; }
    if (!confirm('Revoke ALL tokens for this user?')) return;
    try {
        const resp = await iamFetch('/iam/admin/users/' + encodeURIComponent(uid) + '/revoke-tokens', {
            method: 'POST'
        });
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            showIAMToast(data.error || 'Revoke failed', true);
            return;
        }
        showIAMToast('All user tokens revoked');
        document.getElementById('revokeUserID').value = '';
    } catch (err) {
        showIAMToast('Error: ' + err.message, true);
    }
}

// ── OIDC / JWKS ──────────────────────────────────────────────────────────────

async function loadServiceAccessInfo(realmOverride) {
    const rawEl = document.getElementById('iamServiceAccessJSON');
    const realmInput = document.getElementById('iamServiceRealmInput');
    const requestedRealm = (realmOverride !== undefined && realmOverride !== null)
        ? String(realmOverride).trim().toLowerCase()
        : (iamSelectedRealm || (realmInput ? (realmInput.value || '').trim().toLowerCase() : ''));

    if (!iamToken) {
        setFieldText('iamServiceRealm', 'IAM login required');
        setFieldText('iamServiceIssuer', 'IAM login required');
        setFieldText('iamServiceTokenEndpoint', 'IAM login required');
        setFieldText('iamServiceDiscoveryEndpoint', 'IAM login required');
        setFieldText('iamServiceCertsEndpoint', 'IAM login required');
        setFieldText('iamServiceGrantTypes', 'IAM login required');
        if (realmInput) realmInput.value = '';
        if (rawEl) rawEl.textContent = 'IAM login required';
        return;
    }

    try {
        const path = requestedRealm
            ? '/iam/admin/service-access-info?realm=' + encodeURIComponent(requestedRealm)
            : '/iam/admin/service-access-info';
        const resp = await iamFetch(path);
        if (!resp.ok) {
            const text = 'Failed to load service access info: ' + resp.status;
            if (rawEl) rawEl.textContent = text;
            setFieldText('iamServiceTokenEndpoint', text);
            showIAMToast('Unable to load realm details', true);
            return;
        }

        const info = await resp.json();
        iamServiceAccessInfo = info;
        iamSelectedRealm = info.realm || requestedRealm || '';
        const endpoints = info.endpoints || {};

        setFieldText('iamServiceRealm', info.realm || 'master');
        if (realmInput) realmInput.value = iamSelectedRealm;
        setFieldText('iamServiceIssuer', info.issuer || '–');
        setFieldText('iamServiceTokenEndpoint', endpoints.keycloak_token || endpoints.iam_token || '–');
        setFieldText('iamServiceDiscoveryEndpoint', endpoints.keycloak_openid_configuration || endpoints.iam_openid_configuration || '–');
        setFieldText('iamServiceCertsEndpoint', endpoints.keycloak_certs || endpoints.iam_jwks || '–');
        setFieldText('iamServiceGrantTypes', (info.grant_types_supported || []).join(', ') || '–');

        if (rawEl) rawEl.textContent = JSON.stringify(info, null, 2);
    } catch (err) {
        const text = 'Connection error: ' + err.message;
        if (rawEl) rawEl.textContent = text;
        setFieldText('iamServiceTokenEndpoint', text);
    }
}

async function loadOIDCInfo() {
    try {
        const [oidcResp, jwksResp] = await Promise.all([
            fetch(IAM_API + '/.well-known/openid-configuration'),
            fetch(IAM_API + '/.well-known/jwks.json')
        ]);
        const oidcEl = document.getElementById('iamOIDCConfig');
        const jwksEl = document.getElementById('iamJWKS');
        if (oidcResp.ok) {
            const data = await oidcResp.json();
            if (oidcEl) oidcEl.textContent = JSON.stringify(data, null, 2);
        } else {
            if (oidcEl) oidcEl.textContent = 'Error: ' + oidcResp.status;
        }
        if (jwksResp.ok) {
            const data = await jwksResp.json();
            if (jwksEl) jwksEl.textContent = JSON.stringify(data, null, 2);
        } else {
            if (jwksEl) jwksEl.textContent = 'Error: ' + jwksResp.status;
        }
    } catch (err) {
        const oidcEl = document.getElementById('iamOIDCConfig');
        if (oidcEl) oidcEl.textContent = 'Connection error: ' + err.message;
    }

    loadServiceAccessInfo();
}

// ── Initialisation ───────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', function() {
    // Only run if we're on the IAM admin page
    if (!document.getElementById('iam-dashboard')) return;

    updateIAMLoginBanner();

    if (iamToken) {
        loadIAMDashboard();
        loadServiceAccessInfo();
    }
});
