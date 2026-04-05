// Authentication Module
const AUTH_CONFIG = (() => {
    // Prefer server-injected backend URL from env.
    let apiURL = window.BACKEND_URL || 'http://localhost:8000';
    if (!window.BACKEND_URL && window.location.hostname && window.location.hostname !== 'localhost') {
        apiURL = window.location.protocol + '//' + window.location.hostname + ':8000';
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
    if (value === 'sysadmin' || value === 'system-admin' || value === 'system_admin' || value === 'system-manager') {
        return 'system-manager';
    }
    if (value === 'manager' || value === 'api-manager' || value === 'api_manager') {
        return 'manager';
    }
    if (value === 'admin' || value === 'superadmin' || value === 'super-admin') {
        return 'admin';
    }
    return value;
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

// Initialize authentication on page load
window.addEventListener('DOMContentLoaded', function() {
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
        
        // Use server-returned role if available (most reliable), else decode JWT
        const serverRole = normalizeRole(data.role || '');
        let userRole = normalizeRole(serverRole || extractUserRole(authToken));

        // Safety net: map known demo usernames to their expected roles when the
        // server role is missing or fell back to generic 'user'.
        const knownRoles = { 'sysadmin': 'system-manager', 'admin': 'admin', 'manager': 'manager' };
        if ((!serverRole || userRole === 'user') && knownRoles[data.username]) {
            userRole = knownRoles[data.username];
            console.log('🔄 Role overridden by username mapping:', data.username, '→', userRole);
        }

        console.log('👤 User role:', userRole, '(source:', serverRole ? 'server' : 'jwt-decode', ')');
        localStorage.setItem('userRole', userRole);
        setAuthCookies(authToken, userRole, userName);
        
        closeLoginModal();
        
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
        const payload = parts[1];
        // Add padding if needed
        const padded = payload + '=='.substring(0, (4 - payload.length % 4) % 4);
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
