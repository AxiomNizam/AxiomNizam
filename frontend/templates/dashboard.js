let autoRefreshInterval = null;

// Set backend URL from template
document.getElementById('backendInfo').textContent = 'http://axiomnizam:8000';

// Initial load
window.addEventListener('DOMContentLoaded', () => {
    refreshData();
    loadAPIs();
    startAutoRefresh();
});

function refreshData() {
    fetchHealth();
    fetchStatus();
    updateLastRefreshTime();
}

function loadAPIs() {
    const apisContent = document.getElementById('apisContent');
    apisContent.innerHTML = '';

    const apiCategories = {
        'Health & Status': [
            { method: 'GET', path: '/health', description: 'Health check (no auth)' },
            { method: 'GET', path: '/status', description: 'Check all connections (no auth)' },
        ],
        'MySQL': [
            { method: 'GET', path: '/api/mysql/users', description: 'List all users', auth: 'required' },
            { method: 'GET', path: '/api/mysql/users/:id', description: 'Get user by ID', auth: 'required' },
            { method: 'POST', path: '/api/mysql/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mysql/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mysql/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'MariaDB': [
            { method: 'GET', path: '/api/mariadb/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/mariadb/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mariadb/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mariadb/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'PostgreSQL': [
            { method: 'GET', path: '/api/postgres/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/postgres/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/postgres/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/postgres/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'MongoDB': [
            { method: 'GET', path: '/api/mongodb/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/mongodb/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/mongodb/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/mongodb/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'Firebase': [
            { method: 'GET', path: '/api/firebase/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/firebase/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/firebase/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/firebase/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'Oracle': [
            { method: 'GET', path: '/api/oracle/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/oracle/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/oracle/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/oracle/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'Percona': [
            { method: 'GET', path: '/api/percona/users', description: 'List all users', auth: 'required' },
            { method: 'POST', path: '/api/percona/users', description: 'Create user', auth: 'admin' },
            { method: 'PUT', path: '/api/percona/users/:id', description: 'Update user', auth: 'admin' },
            { method: 'DELETE', path: '/api/percona/users/:id', description: 'Delete user', auth: 'admin' },
        ],
        'Admin Operations': [
            { method: 'POST', path: '/api/admin/database/create', description: 'Create database', auth: 'admin' },
            { method: 'GET', path: '/api/admin/database/list', description: 'List databases', auth: 'admin' },
            { method: 'POST', path: '/api/admin/table/create', description: 'Create table', auth: 'admin' },
            { method: 'GET', path: '/api/admin/table/list', description: 'List tables', auth: 'admin' },
        ],
        'Notifications': [
            { method: 'POST', path: '/api/notifications/send', description: 'Send custom notification', auth: 'required' },
            { method: 'POST', path: '/api/notifications/health', description: 'Send health notification', auth: 'required' },
            { method: 'POST', path: '/api/notifications/status', description: 'Send status notification', auth: 'required' },
            { method: 'GET', path: '/api/notifications/status', description: 'Get service status (no auth)' },
        ],
    };

    let html = '';
    for (const [category, apis] of Object.entries(apiCategories)) {
        html += '<div class="api-category"><div class="api-category-title">' + category + '</div><div class="api-grid">';

        for (const api of apis) {
            const methodClass = api.method.toLowerCase();
            const authClass = api.auth === 'admin' ? 'required' : (api.auth ? 'required' : 'optional');
            const authText = api.auth === 'admin' ? '🔐 Admin Only' : (api.auth ? '🔐 Required' : '✅ No Auth');
            const authHtml = api.auth ? '<div class="api-auth ' + authClass + '">' + authText + '</div>' : '';

            html += '<div class="api-item"><div class="api-method ' + methodClass + '">' + api.method + 
                    '</div><div class="api-path">' + api.path + '</div><div class="api-description">' + 
                    api.description + '</div>' + authHtml + '</div>';
        }

        html += '</div></div>';
    }

    apisContent.innerHTML = html;
}

function fetchHealth() {
    fetch('/api/health')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            const healthContent = document.getElementById('healthContent');
            const statusText = data.status ? data.status.toUpperCase() : 'UNKNOWN';
            const messageText = data.message || 'No message';
            healthContent.innerHTML = '<div class="status-item"><span class="status-label">Status</span>' +
                '<span class="status-value status-ok">' + statusText + '</span></div>' +
                '<div class="status-item"><span class="status-label">Message</span>' +
                '<span class="status-value">' + messageText + '</span></div>';
            showSuccessMessage('Health status retrieved successfully');
        })
        .catch(function(error) {
            document.getElementById('healthContent').innerHTML = '<div class="error-message">Unable to fetch health status</div>';
            showErrorMessage('Failed to fetch health: ' + error.message);
        });
}

function fetchStatus() {
    fetch('/api/status')
        .then(function(response) {
            console.log('Status response:', response);
            return response.json();
        })
        .then(function(data) {
            console.log('Status data:', data);
            const statusContent = document.getElementById('statusContent');
            
            const databases = data.data || data.databases || {};
            
            console.log('Databases extracted:', databases);
            console.log('Databases keys:', Object.keys(databases));
            
            if (databases && Object.keys(databases).length > 0) {
                let dbHtml = '<div class="db-checklist">';
                
                const sortedDbs = Object.entries(databases).sort(function(a, b) {
                    return a[0].localeCompare(b[0]);
                });
                
                for (let i = 0; i < sortedDbs.length; i++) {
                    const dbName = sortedDbs[i][0];
                    const dbStatus = sortedDbs[i][1];
                    const isConnected = dbStatus.toLowerCase() === 'connected';
                    const statusClass = isConnected ? 'connected' : 'disconnected';
                    const icon = isConnected ? '✓' : '✕';
                    const statusIcon = isConnected ? '✔️' : '❌';
                    
                    dbHtml += '<div class="db-item ' + statusClass + '"><div class="db-info">' +
                        '<div class="db-icon">' + icon + '</div>' +
                        '<div class="db-name">' + capitalizeFirstLetter(dbName) + '</div>' +
                        '</div><div class="db-status ' + statusClass + '">' +
                        '<span class="status-icon">' + statusIcon + '</span>' +
                        '<span>' + dbStatus + '</span></div></div>';
                }
                dbHtml += '</div>';
                statusContent.innerHTML = dbHtml;
            } else {
                statusContent.innerHTML = '<div class="error-message">No database information available. Data: ' + JSON.stringify(databases) + '</div>';
            }
            showSuccessMessage('Database status retrieved successfully');
        })
        .catch(function(error) {
            document.getElementById('statusContent').innerHTML = '<div class="error-message">Unable to fetch database status</div>';
            showErrorMessage('Failed to fetch status: ' + error.message);
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
    document.getElementById('lastUpdate').textContent = time;
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
