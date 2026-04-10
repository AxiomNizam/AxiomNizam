// Authentication Module
const AUTH_CONFIG = (() => {
    // Prefer shared backend resolver to avoid stale localhost fallbacks.
    let apiURL = (typeof window.resolveBackendURL === 'function')
        ? window.resolveBackendURL()
        : String(window.BACKEND_URL || '').trim();

    if (!apiURL) {
        const host = String(window.location.hostname || '').toLowerCase();
        if (host && host !== 'localhost' && host !== '127.0.0.1' && host !== '0.0.0.0') {
            const protocol = window.location.protocol || 'https:';
            if (host.indexOf('axiomnizam.') === 0) {
                apiURL = protocol + '//axiomnizam-platform.' + host.substring('axiomnizam.'.length);
            } else {
                apiURL = protocol + '//' + host;
            }
        } else {
            apiURL = 'http://localhost:8000';
        }
    }

    if (apiURL.length > 1 && apiURL.endsWith('/')) {
        apiURL = apiURL.slice(0, -1);
    }
    return {
        apiURL: apiURL,
        loginEndpoint: '/auth/login',
        refreshEndpoint: '/auth/refresh'
    };
})();

console.log('🔐 Auth Config:', AUTH_CONFIG);

let authToken = null;
let refreshToken = null;
let userName = null;
let userRole = 'user';

function normalizeRole(role) {
    const value = String(role || '').toLowerCase().trim();
    if (!value) return 'user';
    if (value === 'sysadmin' || value === 'sys-admin' || value === 'system-admin' || value === 'system_admin' || value === 'system-manager' || value === 'system manager' || value === 'systemadministrator' || value === 'system-administrator' || value === 'system administrator') {
        return 'system-manager';
    }
    if (value === 'manager' || value === 'api-manager' || value === 'api_manager') {
        return 'manager';
    }
    if (value === 'admin' || value === 'administrator' || value === 'superadmin' || value === 'super-admin') {
        return 'admin';
    }
    if (value.indexOf('system') !== -1 && value.indexOf('admin') !== -1) {
        return 'system-manager';
    }
    if (value.indexOf('admin') !== -1 && value.indexOf('account') === -1) {
        return 'admin';
    }
    return value;
}

function roleFromAlias(value) {
    const raw = String(value || '').toLowerCase().trim();
    if (!raw) return '';

    const localPart = raw.indexOf('@') !== -1 ? raw.split('@')[0] : raw;
    const compact = localPart.replace(/[^a-z0-9]/g, '');

    if (compact === 'sysadmin' || compact === 'systemadmin' || compact === 'systemadministrator' || compact === 'systemmanager') {
        return 'system-manager';
    }
    if (compact === 'admin' || compact === 'administrator' || compact === 'superadmin') {
        return 'admin';
    }
    if (compact === 'manager' || compact === 'apimanager' || compact === 'mgr') {
        return 'manager';
    }
    return '';
}

function roleFromRoleList(roles) {
    if (!Array.isArray(roles)) {
        return normalizeRole(roles || '');
    }

    for (let i = 0; i < roles.length; i++) {
        const normalized = normalizeRole(roles[i]);
        if (normalized === 'system-manager') return 'system-manager';
    }
    for (let i = 0; i < roles.length; i++) {
        const normalized = normalizeRole(roles[i]);
        if (normalized === 'admin') return 'admin';
    }
    for (let i = 0; i < roles.length; i++) {
        const normalized = normalizeRole(roles[i]);
        if (normalized === 'manager') return 'manager';
    }
    return 'user';
}

function resolveBestRole(loginData, token, submittedUsername) {
    const candidates = [
        normalizeRole(loginData && loginData.role),
        roleFromRoleList(loginData && loginData.user && loginData.user.roles),
        extractUserRole(token),
        roleFromAlias(loginData && loginData.username),
        roleFromAlias(loginData && loginData.user && loginData.user.email),
        roleFromAlias(loginData && loginData.user && loginData.user.display_name),
        roleFromAlias(submittedUsername),
    ];

    for (let i = 0; i < candidates.length; i++) {
        const role = normalizeRole(candidates[i]);
        if (role && role !== 'user') {
            return role;
        }
    }

    return 'user';
}

function readCookie(name) {
    const prefix = name + '=';
    const parts = document.cookie.split(';');
    for (let i = 0; i < parts.length; i++) {
        const item = parts[i].trim();
        if (item.startsWith(prefix)) {
            return decodeURIComponent(item.substring(prefix.length));
        }
    }
    return '';
}

function setAuthCookies(token, role, name) {
    const maxAge = 60 * 60 * 12;
    document.cookie = 'authToken=' + encodeURIComponent(token || '') + '; path=/; max-age=' + maxAge + '; SameSite=Lax';
    document.cookie = 'userRole=' + encodeURIComponent(normalizeRole(role)) + '; path=/; max-age=' + maxAge + '; SameSite=Lax';
    document.cookie = 'userName=' + encodeURIComponent(name || '') + '; path=/; max-age=' + maxAge + '; SameSite=Lax';
}

function clearAuthCookies() {
    document.cookie = 'authToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax';
    document.cookie = 'userRole=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax';
    document.cookie = 'userName=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax';
}

function defaultPathForRole(role) {
    const normalized = normalizeRole(role);
    if (normalized === 'system-manager') return '/system-manager';
    if (normalized === 'admin') return '/admin';
    if (normalized === 'manager') return '/manager';
    return '/';
}

function canAccessPath(path, role) {
    const normalized = normalizeRole(role);
    if (path === '/governance') return normalized === 'admin' || normalized === 'system-manager';
    if (path === '/operations-center') return normalized === 'admin' || normalized === 'system-manager' || normalized === 'manager';
    if (path === '/lineage-version') return normalized === 'admin' || normalized === 'system-manager';
    if (path === '/admin') return normalized === 'admin' || normalized === 'system-manager';
    if (path === '/system-manager') return normalized === 'system-manager';
    if (path === '/iam-admin') return normalized === 'system-manager';
    if (path === '/manager') return normalized === 'manager';
    return true;
}

function isProtectedPath(path) {
    return path === '/' ||
        path === '/admin' ||
        path === '/system-manager' ||
        path === '/manager' ||
        path === '/governance' ||
        path === '/operations-center' ||
        path === '/lineage-version' ||
        path === '/iam-admin';
}

function consumeOAuthErrorQuery() {
    if (window.location.pathname !== '/login') return;
    const params = new URLSearchParams(window.location.search || '');
    const oauthError = params.get('oauth_error');
    if (!oauthError) return;

    alert('OAuth login failed: ' + oauthError);
    params.delete('oauth_error');

    if (window.history && window.history.replaceState) {
        const nextQuery = params.toString();
        const cleanedURL = window.location.pathname + (nextQuery ? ('?' + nextQuery) : '') + (window.location.hash || '');
        window.history.replaceState({}, document.title, cleanedURL);
    }
}

function completeOAuthLoginFromFragment() {
    const hash = String(window.location.hash || '').replace(/^#/, '');
    if (!hash || hash.indexOf('oauth_access_token=') === -1) {
        return false;
    }

    const params = new URLSearchParams(hash);
    const token = String(params.get('oauth_access_token') || '').trim();
    if (!token) {
        return false;
    }

    const oauthRefresh = String(params.get('oauth_refresh_token') || '').trim();
    const oauthUser = String(params.get('oauth_username') || '').trim();
    const oauthRole = normalizeRole(String(params.get('oauth_role') || '').trim() || extractUserRole(token) || 'user');
    const oauthReturnTo = String(params.get('oauth_return_to') || '').trim();

    authToken = token;
    refreshToken = oauthRefresh || null;
    userName = oauthUser || readCookie('userName') || 'OAuth User';
    userRole = oauthRole;

    localStorage.setItem('authToken', authToken);
    localStorage.setItem('userRole', userRole);
    localStorage.setItem('userName', userName);
    if (refreshToken) {
        localStorage.setItem('refreshToken', refreshToken);
    } else {
        localStorage.removeItem('refreshToken');
    }
    setAuthCookies(authToken, userRole, userName);

    if (window.history && window.history.replaceState) {
        const cleaned = window.location.pathname + (window.location.search || '');
        window.history.replaceState({}, document.title, cleaned);
    }

    if (oauthReturnTo && oauthReturnTo.charAt(0) === '/' && oauthReturnTo !== '/login' && canAccessPath(oauthReturnTo, userRole)) {
        window.location.href = oauthReturnTo;
    } else {
        window.location.href = defaultPathForRole(userRole);
    }
    return true;
}

// Initialize authentication on page load
window.addEventListener('DOMContentLoaded', function() {
    if (completeOAuthLoginFromFragment()) {
        return;
    }

    consumeOAuthErrorQuery();

    authToken = localStorage.getItem('authToken');
    refreshToken = localStorage.getItem('refreshToken');
    userName = localStorage.getItem('userName');
    userRole = normalizeRole(localStorage.getItem('userRole') || 'user');

    if (!authToken) {
        authToken = readCookie('authToken');
        userRole = normalizeRole(readCookie('userRole') || userRole);
        userName = readCookie('userName') || userName;

        if (authToken) {
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('userRole', userRole);
            if (userName) {
                localStorage.setItem('userName', userName);
            }
        }
    } else {
        setAuthCookies(authToken, userRole, userName);
    }

    const path = window.location.pathname;
    if (!authToken && isProtectedPath(path)) {
        window.location.href = '/login';
        return;
    }

    if (authToken && !canAccessPath(path, userRole)) {
        window.location.href = defaultPathForRole(userRole);
    }
});

function handleLogin(event) {
    event.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    const loginBtn = event.target.querySelector('button[type="submit"]');
    const originalText = loginBtn.textContent;
    loginBtn.disabled = true;
    loginBtn.textContent = 'Logging in...';

    const loginURL = AUTH_CONFIG.apiURL + AUTH_CONFIG.loginEndpoint;
    console.log('🔐 Attempting login to:', loginURL);

    // Send credentials to backend API (which proxies IAM auth securely)
    fetch(loginURL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            username: username,
            password: password
        })
    })
    .then(function(response) {
        console.log('📡 Login response status:', response.status);
        return response.text().then(function(rawBody) {
            var data = {};
            if (rawBody) {
                try {
                    data = JSON.parse(rawBody);
                } catch (parseErr) {
                    var compactBody = rawBody.replace(/\s+/g, ' ').trim();
                    if (compactBody.length > 180) {
                        compactBody = compactBody.substring(0, 180) + '...';
                    }
                    data = {
                        error: 'Login endpoint returned non-JSON response (status ' + response.status + '): ' + compactBody
                    };
                }
            }

            if (!response.ok) {
                throw new Error(data.error || 'Login failed: ' + response.statusText);
            }

            if (!data || !data.access_token) {
                throw new Error(data.error || 'Login succeeded but no access token was returned');
            }

            return data;
        });
    })
    .then(function(data) {
        console.log('✅ Login successful:', data);
        authToken = data.access_token;
        refreshToken = data.refresh_token || null;
        userName = data.username;
        localStorage.setItem('authToken', authToken);
        if (refreshToken) {
            localStorage.setItem('refreshToken', refreshToken);
        }
        localStorage.setItem('userName', userName);
        
        const userRole = resolveBestRole(data, authToken, username);

        console.log('👤 User role:', userRole, '(server role:', normalizeRole(data.role || ''), ', user.roles:', (data.user && data.user.roles) || [], ')');
        localStorage.setItem('userRole', userRole);
        setAuthCookies(authToken, userRole, userName);

        // Redirect based on user role
        if (userRole === 'system-manager') {
            window.location.href = '/system-manager';
        } else if (userRole === 'admin') {
            window.location.href = '/admin';
        } else if (userRole === 'manager') {
            window.location.href = '/manager';
        } else {
            // Normal users go to dashboard (view-only)
            window.location.href = '/';
        }
    })
    .catch(function(error) {
        console.error('❌ Login error:', error);
        loginBtn.disabled = false;
        loginBtn.textContent = originalText;
        alert('Login failed: ' + error.message);
    });
}

function logout() {
    authToken = null;
    userName = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('userName');
    localStorage.removeItem('userRole');
    localStorage.removeItem('refreshToken');
    clearAuthCookies();
    window.location.href = '/login';
}

// Decode JWT token and extract user role
function extractUserRole(token) {
    try {
        // JWT format: header.payload.signature
        const parts = token.split('.');
        if (parts.length !== 3) {
            console.warn('⚠️ Invalid token format');
            return 'user';
        }
        
        // Decode the payload (second part)
        const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/');
        // Add padding if needed
        const padded = payload + '='.repeat((4 - payload.length % 4) % 4);
        const decoded = JSON.parse(atob(padded));
        
        console.log('🔐 Full token payload:', JSON.stringify(decoded, null, 2));
        
        // Check for IAM top-level roles first.
        if (Array.isArray(decoded.roles)) {
            const directRoles = decoded.roles;
            console.log('📋 IAM roles found:', directRoles);
            for (let i = 0; i < directRoles.length; i++) {
                const role = String(directRoles[i] || '').toLowerCase();
                if (role === 'system-manager' || role === 'system_manager' || role === 'system-admin' || role === 'sysadmin') {
                    return 'system-manager';
                }
                if (role.includes('admin') && !role.includes('account')) {
                    return 'admin';
                }
                if (role === 'manager' || role === 'api-manager' || role === 'api_manager') {
                    return 'manager';
                }
            }
        }

        // Check realm roles for legacy compatibility.
        if (decoded.realm_access && decoded.realm_access.roles) {
            const roles = decoded.realm_access.roles;
            console.log('📋 Realm roles found:', roles);
            
            // Check each role and determine user type
            for (let i = 0; i < roles.length; i++) {
                const role = roles[i].toLowerCase();
                console.log(`  - Checking role: "${role}"`);
                
                // Check for admin role
                // Check for system-manager role (various formats)
                if (role === 'system-manager' || role === 'system_manager' || role === 'system-admin' || role === 'sysadmin') {
                    console.log('✅ Detected as SYSTEM-MANAGER role');
                    return 'system-manager';
                }

                if (role.includes('admin') && !role.includes('account')) {
                    console.log('✅ Detected as ADMIN role');
                    return 'admin';
                }

                // Check for manager role (must come AFTER system-manager check)
                if (role === 'manager' || role === 'api-manager' || role === 'api_manager') {
                    console.log('✅ Detected as MANAGER role');
                    return 'manager';
                }
            }
        }
        
        // Check for client roles
        if (decoded.resource_access) {
            const clientRoles = decoded.resource_access['axiomnizam-backend'];
            if (clientRoles && clientRoles.roles) {
                console.log('📋 Client roles found:', clientRoles.roles);
                
                for (let i = 0; i < clientRoles.roles.length; i++) {
                    const role = clientRoles.roles[i].toLowerCase();
                    if (role === 'system-manager' || role === 'system_manager' || role === 'sysadmin') return 'system-manager';
                    if (role.includes('admin')) return 'admin';
                }
            }
        }
        
        console.log('ℹ️ No admin/manager roles found, defaulting to user');
        return 'user';
    } catch (error) {
        console.error('❌ Error decoding token:', error);
        return 'user';
    }
}

function getAuthHeaders() {
    const headers = {
        'Content-Type': 'application/json'
    };

    // Always resolve the latest token from storage/cookies to avoid stale in-memory state.
    const latestToken = localStorage.getItem('authToken') || readCookie('authToken') || authToken;
    if (latestToken) {
        authToken = latestToken;
        headers['Authorization'] = 'Bearer ' + latestToken;
    }

    return headers;
}

function isAuthenticated() {
    return authToken !== null && authToken !== undefined;
}

function isAdmin() {
    return userRole === 'admin' || userRole === 'system-manager';
}
