let autoRefreshInterval = null;

// Get backend URL from template
const BACKEND_URL = (() => {
    const elem = document.getElementById('backendURL');
    let url = 'http://localhost:8000'; // Default fallback
    
    if (elem && elem.textContent) {
        const text = elem.textContent.trim();
        // Ensure URL has protocol
        if (text && text.length > 0) {
            if (text.startsWith('http://') || text.startsWith('https://')) {
                url = text;
            } else if (text.startsWith('localhost') || text.startsWith('127.0.0.1')) {
                url = 'http://' + text;
            } else {
                url = text; // Use as-is if it looks complete
            }
        }
    }
    
    // If contains Docker hostname, replace with localhost
    if (url.includes('axiomnizam:8000')) {
        url = url.replace('axiomnizam:8000', 'localhost:8000');
    }
    
    return url;
})();

console.log('Backend URL:', BACKEND_URL);

// Initial load
window.addEventListener('DOMContentLoaded', () => {
    refreshData();
    loadAPIs();
    displayBackendInfo();
    startAutoRefresh();
});

function refreshData() {
    fetchHealth();
    fetchStatus();
    updateLastRefreshTime();
}

function fetchHealth() {
    const url = BACKEND_URL + '/health';
    console.log('Fetching health from:', url);
    console.log('BACKEND_URL value:', BACKEND_URL);
    
    fetch(url, { mode: 'cors' })
        .then(function(response) { 
            console.log('Health response status:', response.status);
            if (!response.ok) throw new Error('HTTP ' + response.status);
            return response.json(); 
        })
        .then(function(data) {
            console.log('Health data:', data);
            const healthContent = document.getElementById('healthContent');
            const statusText = (data.status || 'unknown').toUpperCase();
            const messageText = data.message || 'No message';
            
            if (healthContent) {
                healthContent.innerHTML = '<div class="status-item"><span class="status-label">Status</span>' +
                    '<span class="status-value status-ok">' + statusText + '</span></div>' +
                    '<div class="status-item"><span class="status-label">Message</span>' +
                    '<span class="status-value">' + messageText + '</span></div>';
            }
        })
        .catch(function(error) {
            console.error('Health fetch error:', error);
            const healthContent = document.getElementById('healthContent');
            if (healthContent) {
                healthContent.innerHTML = '<div class="error-message">❌ Unable to fetch health status: ' + error.message + '</div>';
            }
        });
}

function fetchStatus() {
    const url = BACKEND_URL + '/status';
    console.log('Fetching status from:', url);
    
    fetch(url, { mode: 'cors' })
        .then(function(response) {
            console.log('Status response status:', response.status);
            if (!response.ok) throw new Error('HTTP ' + response.status);
            return response.json();
        })
        .then(function(data) {
            console.log('Status data:', data);
            // Backend returns data.data with the database statuses
            const databases = data.data || data.Data || {};
            
            console.log('Databases extracted:', databases);
            
            // Update the status content div with database list
            const statusContent = document.getElementById('statusContent');
            if (statusContent) {
                if (databases && Object.keys(databases).length > 0) {
                    let dbHtml = '<div class="db-checklist">';
                    
                    const sortedDbs = Object.entries(databases).sort(function(a, b) {
                        return a[0].localeCompare(b[0]);
                    });
                    
                    for (let i = 0; i < sortedDbs.length; i++) {
                        const dbName = sortedDbs[i][0];
                        const dbStatus = sortedDbs[i][1];
                        const isConnected = dbStatus === 'connected' || dbStatus === 'Connected' || dbStatus === 'ok' || dbStatus === 'OK';
                        const statusIcon = isConnected ? '✓' : '✕';
                        const statusBadge = isConnected ? '✔️ connected' : '❌ disconnected';
                        const statusClass = isConnected ? 'connected' : 'disconnected';
                        
                        dbHtml += '<div class="db-item ' + statusClass + '"><div class="db-info">' +
                            '<div class="db-icon">' + statusIcon + '</div>' +
                            '<div class="db-name">' + capitalizeFirstLetter(dbName) + '</div>' +
                            '</div><div class="db-status ' + statusClass + '">' +
                            '<span>' + statusBadge + '</span></div></div>';
                    }
                    dbHtml += '</div>';
                    statusContent.innerHTML = dbHtml;
                } else {
                    statusContent.innerHTML = '<div class="error-message">No database information available</div>';
                }
            }
        })
        .catch(function(error) {
            console.error('Status fetch error:', error);
            const statusContent = document.getElementById('statusContent');
            if (statusContent) {
                statusContent.innerHTML = '<div class="error-message">❌ Unable to fetch status: ' + error.message + '</div>';
            }
        });
}

function toggleAutoRefresh() {
    const checkbox = document.getElementById('autoRefresh');
    if (checkbox.checked) {
        startAutoRefresh();
    } else {
        stopAutoRefresh();
    }
}

function startAutoRefresh() {
    if (!autoRefreshInterval) {
        autoRefreshInterval = setInterval(refreshData, 5000);
    }
}

function stopAutoRefresh() {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
    }
}

function updateLastRefreshTime() {
    const now = new Date();
    const time = now.toLocaleTimeString();
    const elem = document.getElementById('lastUpdate');
    if (elem) {
        elem.textContent = time;
    }
}

function displayBackendInfo() {
    const backendInfoElem = document.getElementById('backendInfo');
    if (backendInfoElem) {
        backendInfoElem.textContent = BACKEND_URL;
    }
}

function loadAPIs() {
    const apiCategories = {
        'Health & Status': [
            { method: 'GET', path: '/health', description: 'Health check (no auth)' },
            { method: 'GET', path: '/status', description: 'Check all connections (no auth)' },
        ],
        'MySQL': [
            { method: 'GET', path: '/api/mysql/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/mysql/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mysql/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mysql/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'PostgreSQL': [
            { method: 'GET', path: '/api/postgres/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/postgres/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/postgres/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/postgres/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'MariaDB': [
            { method: 'GET', path: '/api/mariadb/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/mariadb/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mariadb/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mariadb/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'MongoDB': [
            { method: 'GET', path: '/api/mongodb/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/mongodb/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mongodb/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mongodb/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'Notifications': [
            { method: 'POST', path: '/api/notifications/send', description: 'Send custom notification', auth: 'required' },
            { method: 'POST', path: '/api/notifications/health', description: 'Send health notification', auth: 'required' },
            { method: 'POST', path: '/api/notifications/status', description: 'Send status notification', auth: 'required' },
        ],
    };

    let html = '';
    for (const [category, apis] of Object.entries(apiCategories)) {
        html += '<div style="margin-bottom: 20px;"><strong style="font-size: 1.1em;">' + category + '</strong>';
        html += '<div class="api-grid">';
        
        for (let i = 0; i < apis.length; i++) {
            const api = apis[i];
            const methodClass = api.method.toLowerCase();
            html += '<div class="api-item">' +
                '<span class="api-method ' + methodClass + '">' + api.method + '</span>' +
                '<div class="api-path">' + api.path + '</div>' +
                '<div class="api-description">' + api.description + '</div>';
            if (api.auth) {
                html += '<span class="api-auth" style="background: rgba(239, 68, 68, 0.2); color: var(--danger-color); padding: 3px 8px; border-radius: 3px; font-size: 0.75em;">Auth: ' + api.auth + '</span>';
            }
            html += '</div>';
        }
        
        html += '</div></div>';
    }
    
    const apisContent = document.getElementById('apisContent');
    if (apisContent) {
        apisContent.innerHTML = html;
    }
}

function showErrorMessage(message) {
    const messagesDiv = document.getElementById('messages');
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.textContent = '❌ ' + message;
    messagesDiv.innerHTML = '';
    messagesDiv.appendChild(errorDiv);
    setTimeout(function() { errorDiv.remove(); }, 5000);
}

function showSuccessMessage(message) {
    // Uncomment to show success messages
    // const messagesDiv = document.getElementById('messages');
    // const successDiv = document.createElement('div');
    // successDiv.className = 'success-message';
    // successDiv.textContent = '✅ ' + message;
    // messagesDiv.innerHTML = '';
    // messagesDiv.appendChild(successDiv);
    // setTimeout(() => successDiv.remove(), 3000);
}

function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
}
