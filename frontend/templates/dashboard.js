let autoRefreshInterval = null;

// Get backend URL from template
const BACKEND_URL = (() => {
    if (typeof window.resolveBackendURL === 'function') {
        return window.resolveBackendURL();
    }

    const value = String(window.BACKEND_URL || '').trim();
    if (value) {
        return value.endsWith('/') ? value.slice(0, -1) : value;
    }

    return 'http://localhost:8000';
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
    const url = '/api/health';
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
    const url = '/api/status';
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

function readDashboardCookie(name) {
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

function getDashboardAuthToken() {
    return localStorage.getItem('authToken') || readDashboardCookie('authToken') || '';
}

function normalizeDashboardRole(role) {
    const value = String(role || '').toLowerCase().trim();
    if (value === 'sysadmin' || value === 'system_admin' || value === 'system-admin') return 'system-manager';
    if (value === 'api-manager' || value === 'api_manager') return 'manager';
    if (value === 'superadmin' || value === 'super-admin') return 'admin';
    return value || 'user';
}

function getDashboardRole() {
    return normalizeDashboardRole(localStorage.getItem('userRole') || readDashboardCookie('userRole') || 'user');
}

function toRuntimePath(builderPath) {
    var raw = String(builderPath || '').trim();
    if (!raw) return '/api/custom';
    if (raw === '/api/custom' || raw.indexOf('/api/custom/') === 0) return raw;
    return '/api/custom/' + raw.replace(/^\/+/, '');
}

function loadAPIs() {
    const apisContent = document.getElementById('apisContent');
    if (!apisContent) return;

    const token = getDashboardAuthToken();
    const role = getDashboardRole();

    if (!token) {
        apisContent.innerHTML = '<div class="api-item" style="max-width:none;">' +
            '<div class="api-description">Login required to view runtime APIs allowed by RBAC.</div>' +
            '</div>';
        return;
    }

    apisContent.innerHTML = '<div class="loading">Loading APIs...</div>';

    fetch(BACKEND_URL + '/api/v1/builder/apis', {
        headers: {
            'Authorization': 'Bearer ' + token,
            'Content-Type': 'application/json'
        }
    })
        .then(function(response) {
            if (response.status === 401 || response.status === 403) {
                throw new Error('RBAC access denied for API catalog');
            }
            if (!response.ok) {
                throw new Error('Failed to load APIs: HTTP ' + response.status);
            }
            return response.json();
        })
        .then(function(data) {
            const all = (data.apis || []);
            const active = all.filter(function(api) {
                return String(api.status || '').toLowerCase() === 'active';
            });

            if (active.length === 0) {
                apisContent.innerHTML = '<div class="api-item" style="max-width:none;">' +
                    '<div class="api-description">No active custom APIs found. Create one from Admin > API Builder.</div>' +
                    '</div>';
                return;
            }

            let html = '<div style="margin-bottom: 12px; color:#5f6b7a; font-size:0.9rem;">' +
                'Showing ' + active.length + ' active API Builder endpoints for role: <strong>' + escapeHtmlDash(role) + '</strong></div>';

            html += '<div class="api-grid">';
            for (let i = 0; i < active.length; i++) {
                const api = active[i];
                const apiType = String(api.api_type || 'rest').toLowerCase();
                const displayMethod = (apiType === 'graphql' ? 'POST' : (api.method || 'GET')).toUpperCase();
                const runtimePath = apiType === 'graphql'
                    ? String(api.path || '/api/graphql')
                    : toRuntimePath(api.path || '/');

                const methodClass = displayMethod.toLowerCase();
                const cacheBadge = api.cache_enabled
                    ? '<span class="api-auth" style="background:#d4edda;color:#155724;margin-left:6px;">Cache ' + (api.cache_ttl || 300) + 's</span>'
                    : '';
                const rateText = (api.rate_limit && api.rate_limit > 0)
                    ? (api.rate_limit + '/min')
                    : 'unlimited';

                html += '<div class="api-item">' +
                    '<span class="api-method ' + methodClass + '">' + displayMethod + '</span>' + cacheBadge +
                    '<div class="api-path">' + escapeHtmlDash(runtimePath) + '</div>' +
                    '<div class="api-description">' + escapeHtmlDash(api.description || api.name || 'Custom API') + '</div>' +
                    '<span class="api-auth" style="margin-right:6px;">Type: ' + escapeHtmlDash(apiType) + '</span>' +
                    '<span class="api-auth" style="margin-right:6px;">Category: ' + escapeHtmlDash(api.category || 'custom') + '</span>' +
                    '<span class="api-auth" style="margin-right:6px;">Rate: ' + escapeHtmlDash(rateText) + '</span>' +
                    '<span class="api-auth">Auth: ' + (api.auth_required ? 'required' : 'token-only') + '</span>' +
                    '</div>';
            }
            html += '</div>';
            apisContent.innerHTML = html;
        })
        .catch(function(error) {
            apisContent.innerHTML = '<div class="api-item" style="max-width:none;">' +
                '<div class="api-description">Unable to load API catalog: ' + escapeHtmlDash(error.message) + '</div>' +
                '</div>';
        });
}

function escapeHtmlDash(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
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
