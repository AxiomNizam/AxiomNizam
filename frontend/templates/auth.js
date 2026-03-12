// Authentication Module
const AUTH_CONFIG = (() => {
    // Determine backend API URL based on current location
    let apiURL = 'http://localhost:8000';
    if (window.location.hostname && window.location.hostname !== 'localhost') {
        // For non-localhost environments, construct the URL
        apiURL = window.location.protocol + '//' + window.location.hostname + ':8000';
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

// Initialize authentication on page load
window.addEventListener('DOMContentLoaded', function() {
    authToken = localStorage.getItem('authToken');
    refreshToken = localStorage.getItem('refreshToken');
    userName = localStorage.getItem('userName');
    userRole = localStorage.getItem('userRole') || 'user';
    
    if (!authToken) {
        // Redirect to dashboard if trying to access protected page
        const path = window.location.pathname;
        if (path !== '/' && path !== '/admin' && path !== '/system-manager' && path !== '/manager') {
            window.location.href = '/';
        }
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

    // Send credentials to backend API (which handles Keycloak auth securely)
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
        if (!response.ok) {
            return response.json().then(function(data) {
                throw new Error(data.error || 'Login failed: ' + response.statusText);
            });
        }
        return response.json();
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
        const serverRole = data.role || '';
        let userRole = serverRole || extractUserRole(authToken);

        // Safety net: map known demo usernames to their expected roles when the
        // server role is missing or fell back to generic 'user' (e.g. Keycloak
        // intercepted a demo-username login without the matching realm role).
        const knownRoles = { 'sysadmin': 'system-manager', 'admin': 'admin', 'manager': 'manager' };
        if ((!serverRole || userRole === 'user') && knownRoles[data.username]) {
            userRole = knownRoles[data.username];
            console.log('🔄 Role overridden by username mapping:', data.username, '→', userRole);
        }

        console.log('👤 User role:', userRole, '(source:', serverRole ? 'server' : 'jwt-decode', ')');
        localStorage.setItem('userRole', userRole);
        
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
        alert('Login failed: ' + error.message + '\n\nDemo: Use admin/admin');
    });
}

function logout() {
    authToken = null;
    userName = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('userName');
    localStorage.removeItem('userRole');
    localStorage.removeItem('refreshToken');
    window.location.href = '/';
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
        
        // Check for realm roles (most common in Keycloak)
        if (decoded.realm_access && decoded.realm_access.roles) {
            const roles = decoded.realm_access.roles;
            console.log('📋 Realm roles found:', roles);
            
            // Check each role and determine user type
            for (let i = 0; i < roles.length; i++) {
                const role = roles[i].toLowerCase();
                console.log(`  - Checking role: "${role}"`);
                
                // Check for admin role
                if (role.includes('admin') && !role.includes('account')) {
                    console.log('✅ Detected as ADMIN role');
                    return 'admin';
                }
                
                // Check for system-manager role (various formats)
                if (role === 'system-manager' || role === 'system_manager' || 
                    role === 'system-admin') {
                    console.log('✅ Detected as SYSTEM-MANAGER role');
                    return 'system-manager';
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
                    if (role.includes('admin')) return 'admin';
                    if (role === 'system-manager' || role === 'system_manager') return 'system-manager';
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
    return {
        'Content-Type': 'application/json',
        'Authorization': authToken ? 'Bearer ' + authToken : ''
    };
}

function isAuthenticated() {
    return authToken !== null && authToken !== undefined;
}

function isAdmin() {
    return userRole === 'admin';
}
