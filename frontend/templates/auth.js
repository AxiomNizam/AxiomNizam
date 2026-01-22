// Authentication Module
const AUTH_CONFIG = {
    keycloakURL: 'http://localhost:8080',
    keycloakRealm: 'axiomnizam',
    keycloakClient: 'axiomnizam-frontend',
    clientSecret: ''
};

let authToken = null;
let userName = null;
let userRole = 'user';

// Initialize authentication on page load
window.addEventListener('DOMContentLoaded', function() {
    authToken = localStorage.getItem('authToken');
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

    const body = new URLSearchParams();
    body.append('client_id', AUTH_CONFIG.keycloakClient);
    body.append('client_secret', AUTH_CONFIG.clientSecret);
    body.append('grant_type', 'password');
    body.append('username', username);
    body.append('password', password);

    const loginBtn = event.target.querySelector('button[type="submit"]');
    const originalText = loginBtn.textContent;
    loginBtn.disabled = true;
    loginBtn.textContent = 'Logging in...';

    fetch(AUTH_CONFIG.keycloakURL + '/realms/' + AUTH_CONFIG.keycloakRealm + '/protocol/openid-connect/token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: body
    })
    .then(function(response) {
        if (!response.ok) {
            throw new Error('Login failed: ' + response.statusText);
        }
        return response.json();
    })
    .then(function(data) {
        authToken = data.access_token;
        userName = username;
        localStorage.setItem('authToken', authToken);
        localStorage.setItem('userName', userName);
        localStorage.setItem('userRole', 'admin');
        closeLoginModal();
        window.location.reload();
    })
    .catch(function(error) {
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
