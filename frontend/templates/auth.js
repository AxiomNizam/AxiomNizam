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
        if (path !== '/' && path !== '/admin' && path !== '/system-manager') {
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
        localStorage.setItem('userRole', 'admin');
        closeLoginModal();
        window.location.reload();
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
    window.location.href = '/';
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
