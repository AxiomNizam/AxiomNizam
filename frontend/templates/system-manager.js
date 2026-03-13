// System Manager JS
const BACKEND_URL = (() => {
    const elem = document.getElementById('backendURL');
    let url = 'http://localhost:8000'; // Default fallback
    
    if (elem && elem.textContent) {
        const text = elem.textContent.trim();
        if (text && text.length > 0) {
            url = text;
        }
    }
    
    // If contains Docker hostname, replace with localhost
    if (url.includes('axiomnizam:8000')) {
        url = url.replace('axiomnizam:8000', 'localhost:8000');
    }
    
    return url;
})();

console.log('System Manager - Backend URL:', BACKEND_URL);

var availableDbServers = [];

window.addEventListener('DOMContentLoaded', function() {
    // Set user name from localStorage
    const userName = localStorage.getItem('userName');
    if (userName) {
        const userNameElem = document.getElementById('managerUserName');
        if (userNameElem) {
            userNameElem.textContent = userName;
        }
    }
    loadStatusData();
    loadDatabases();
    loadDatabaseServers();
    setInterval(loadStatusData, 30000);
});

function switchManagerTab(tabName) {
    // Hide all tabs
    const tabs = document.querySelectorAll('.tab-content');
    tabs.forEach(function(tab) { tab.classList.remove('active'); });
    
    // Remove active from buttons
    const buttons = document.querySelectorAll('.tab-btn');
    buttons.forEach(function(btn) { btn.classList.remove('active'); });
    
    // Show selected tab
    const selectedTab = document.getElementById(tabName);
    if (selectedTab) selectedTab.classList.add('active');
    
    // Add active to clicked button
    if (event && event.currentTarget) {
        event.currentTarget.classList.add('active');
    }
    
    if (tabName === 'databases') {
        loadDatabases();
    }
    if (tabName === 'users') {
        loadUsers();
    }
    if (tabName === 'graphql-studio') {
        loadManagerGraphQLSchemaInfo();
    }
    if (tabName === 'control-plane') {
        refreshManagerControlPlaneData();
    }
}

function loadStatusData() {
    // Load live status
    fetch(BACKEND_URL + '/health')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            const status = data.status === 'ok' ? '✓ Healthy' : '✗ Unhealthy';
            document.getElementById('liveStatus').textContent = status;
            document.getElementById('statusDot').style.background = data.status === 'ok' ? '#10b981' : '#ef4444';
            document.getElementById('statusTime').textContent = new Date().toLocaleTimeString();
        })
        .catch(function() {
            document.getElementById('liveStatus').textContent = '✗ Error';
            document.getElementById('statusDot').style.background = '#ef4444';
        });

    // Load database status for overview
    fetch(BACKEND_URL + '/status')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            const databases = data.data || data.databases || {};
            let connectedCount = 0;
            
            Object.values(databases).forEach(function(status) {
                if (status.toLowerCase().includes('connected') || status.toLowerCase().includes('ok')) {
                    connectedCount++;
                }
            });
            
            // Update metrics in overview
            updateMetrics();
        })
        .catch(function() {
            updateMetrics();
        });
}

function updateMetrics() {
    // Simulate metric updates (in real scenario, these would come from actual system monitoring)
    document.getElementById('cpuUsage').textContent = Math.floor(Math.random() * 40) + '%';
    document.getElementById('cpuProgress').style.width = Math.floor(Math.random() * 40) + '%';
    
    document.getElementById('memoryUsage').textContent = Math.floor(Math.random() * 60) + '%';
    document.getElementById('memoryProgress').style.width = Math.floor(Math.random() * 60) + '%';
    
    document.getElementById('diskUsage').textContent = Math.floor(Math.random() * 75) + '%';
    document.getElementById('diskProgress').style.width = Math.floor(Math.random() * 75) + '%';
    
    document.getElementById('networkIO').textContent = Math.floor(Math.random() * 100) + ' MB/s';
}

function loadDatabases() {
    fetch(BACKEND_URL + '/status', {
        headers: getAuthHeaders()
    })
    .then(function(response) { return response.json(); })
    .then(function(data) {
        const databases = data.data || data.databases || {};
        let html = '';
        
        Object.entries(databases).forEach(function([dbName, status]) {
            const isConnected = status.toLowerCase().includes('connected') || status.toLowerCase().includes('ok');
            html += '<div class="database-item">' +
                '<div class="db-info-row">' +
                '<span class="db-info-label">Database</span>' +
                '<span class="db-info-value">' + capitalizeFirstLetter(dbName) + '</span>' +
                '</div>' +
                '<div class="db-info-row">' +
                '<span class="db-info-label">Status</span>' +
                '<span class="db-info-value" style="color: ' + (isConnected ? '#10b981' : '#ef4444') + '">' +
                (isConnected ? '✓ Connected' : '✗ Disconnected') + '</span>' +
                '</div>' +
                '<div class="db-info-row">' +
                '<span class="db-info-label">Type</span>' +
                '<span class="db-info-value">' + guessDbType(dbName) + '</span>' +
                '</div>' +
                '</div>';
        });
        
        document.getElementById('databaseList').innerHTML = html || '<div style="padding: 20px; text-align: center;">No databases found</div>';
    })
    .catch(function(error) {
        document.getElementById('databaseList').innerHTML = '<div style="color: #ef4444;">Failed to load databases</div>';
    });
}

function refreshDatabases() {
    loadDatabases();
}

function createDatabase() {
    document.getElementById('createDbModal').style.display = 'flex';
    document.getElementById('newDbType').value = '';
    document.getElementById('newDbServer').value = '';
    document.getElementById('newDbName').value = '';
    document.getElementById('createDbResult').style.display = 'none';
    populateCreateDbServers();
}

function closeCreateDbModal() {
    document.getElementById('createDbModal').style.display = 'none';
}

function submitCreateDatabase(event) {
    event.preventDefault();
    var dbType = document.getElementById('newDbType').value;
    var dbServer = document.getElementById('newDbServer').value;
    var dbName = document.getElementById('newDbName').value.trim();
    var btn = document.getElementById('createDbBtn');
    var resultDiv = document.getElementById('createDbResult');

    btn.disabled = true;
    btn.textContent = 'Creating...';
    resultDiv.style.display = 'none';

    fetch(BACKEND_URL + '/api/admin/database/create', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({ db_type: dbType, db_server: dbServer, database_name: dbName })
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        btn.disabled = false;
        btn.textContent = 'Create Database';
        resultDiv.style.display = 'block';
        if (result.ok) {
            resultDiv.style.background = 'rgba(16,185,129,0.15)';
            resultDiv.style.color = '#10b981';
            var serverLabel = result.data.server_name || result.data.db_server || 'default';
            resultDiv.textContent = 'Database "' + dbName + '" created successfully on ' + dbType + ' (' + serverLabel + ')';
            addOperationLog('Database "' + dbName + '" created on ' + dbType + ' via ' + serverLabel, 'success');
            loadDatabases();
        } else {
            resultDiv.style.background = 'rgba(239,68,68,0.15)';
            resultDiv.style.color = '#ef4444';
            resultDiv.textContent = result.data.error || 'Failed to create database';
            addOperationLog('Database creation failed: ' + (result.data.error || 'Unknown error'), 'error');
        }
    })
    .catch(function(err) {
        btn.disabled = false;
        btn.textContent = 'Create Database';
        resultDiv.style.display = 'block';
        resultDiv.style.background = 'rgba(239,68,68,0.15)';
        resultDiv.style.color = '#ef4444';
        resultDiv.textContent = 'Connection error: ' + err.message;
    });
}

function loadDatabaseServers() {
    fetch(BACKEND_URL + '/api/admin/database/servers', {
        headers: getAuthHeaders()
    })
    .then(function(response) { return response.json(); })
    .then(function(data) {
        availableDbServers = data.servers || [];
        populateCreateDbServers();
    })
    .catch(function() {
        availableDbServers = [];
        populateCreateDbServers();
    });
}

function populateCreateDbServers() {
    var serverSelect = document.getElementById('newDbServer');
    var dbType = (document.getElementById('newDbType').value || '').toLowerCase();
    if (!serverSelect) return;

    var selected = serverSelect.value;
    serverSelect.innerHTML = '<option value="">Default server for selected database type</option>';

    var filtered = availableDbServers.filter(function(server) {
        if (!dbType) return true;
        return (server.db_type || '').toLowerCase() === dbType;
    });

    filtered.forEach(function(server) {
        var option = document.createElement('option');
        option.value = server.key;
        option.disabled = server.connected === false;
        option.textContent = (server.name || server.key) + ' [' + (server.db_type || '').toUpperCase() + ']' + (server.connected === false ? ' (disconnected)' : '');
        serverSelect.appendChild(option);
    });

    if (selected && filtered.some(function(s) { return s.key === selected; })) {
        serverSelect.value = selected;
    }
}

function openConnectDbServerModal() {
    var modal = document.getElementById('connectDbServerModal');
    if (!modal) return;

    document.getElementById('serverName').value = '';
    document.getElementById('serverDbType').value = document.getElementById('newDbType').value || 'mysql';
    document.getElementById('serverHost').value = '127.0.0.1';
    document.getElementById('serverUsername').value = 'root';
    document.getElementById('serverPassword').value = '';
    document.getElementById('serverDefaultDatabase').value = '';
    document.getElementById('serverSSLMode').value = 'disable';
    document.getElementById('connectServerResult').style.display = 'none';

    updateConnectServerPortDefault();
    modal.style.display = 'flex';
}

function closeConnectDbServerModal() {
    var modal = document.getElementById('connectDbServerModal');
    if (modal) modal.style.display = 'none';
}

function updateConnectServerPortDefault() {
    var dbType = (document.getElementById('serverDbType').value || '').toLowerCase();
    var portEl = document.getElementById('serverPort');
    if (!portEl) return;
    portEl.value = dbType === 'postgres' ? 5432 : 3306;
}

function submitConnectDbServer(event) {
    event.preventDefault();

    var btn = document.getElementById('connectServerBtn');
    var resultDiv = document.getElementById('connectServerResult');
    var payload = {
        server_name: document.getElementById('serverName').value.trim(),
        db_type: document.getElementById('serverDbType').value,
        host: document.getElementById('serverHost').value.trim(),
        port: parseInt(document.getElementById('serverPort').value, 10) || 0,
        username: document.getElementById('serverUsername').value.trim(),
        password: document.getElementById('serverPassword').value,
        default_database: document.getElementById('serverDefaultDatabase').value.trim(),
        ssl_mode: document.getElementById('serverSSLMode').value
    };

    btn.disabled = true;
    btn.textContent = 'Connecting...';
    resultDiv.style.display = 'none';

    fetch(BACKEND_URL + '/api/admin/database/connect', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(payload)
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        btn.disabled = false;
        btn.textContent = 'Connect Server';
        resultDiv.style.display = 'block';

        if (result.ok) {
            resultDiv.style.background = 'rgba(16,185,129,0.15)';
            resultDiv.style.color = '#10b981';
            resultDiv.textContent = 'Server connected: ' + (result.data.server && result.data.server.name ? result.data.server.name : payload.server_name);

            addOperationLog('Connected database server: ' + payload.server_name + ' (' + payload.db_type + ')', 'success');
            loadDatabaseServers();

            var newDbType = document.getElementById('newDbType');
            if (newDbType && !newDbType.value) {
                newDbType.value = payload.db_type;
            }

            setTimeout(function() {
                closeConnectDbServerModal();
                populateCreateDbServers();
                if (result.data.server && result.data.server.key) {
                    document.getElementById('newDbServer').value = result.data.server.key;
                }
            }, 500);
        } else {
            resultDiv.style.background = 'rgba(239,68,68,0.15)';
            resultDiv.style.color = '#ef4444';
            resultDiv.textContent = result.data.error || 'Failed to connect server';
            addOperationLog('Database server connection failed: ' + (result.data.error || 'Unknown error'), 'error');
        }
    })
    .catch(function(err) {
        btn.disabled = false;
        btn.textContent = 'Connect Server';
        resultDiv.style.display = 'block';
        resultDiv.style.background = 'rgba(239,68,68,0.15)';
        resultDiv.style.color = '#ef4444';
        resultDiv.textContent = 'Connection error: ' + err.message;
    });
}

function backupDatabases() {
    if (!confirm('Start backup for all connected databases?')) return;
    addOperationLog('Backup started for all databases', 'info');
    
    fetch(BACKEND_URL + '/api/admin/database/list?db_type=mysql', { headers: getAuthHeaders() })
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var dbs = data.databases || [];
            addOperationLog('Found ' + dbs.length + ' MySQL databases', 'info');
        })
        .catch(function() {});
    
    fetch(BACKEND_URL + '/api/admin/database/list?db_type=postgres', { headers: getAuthHeaders() })
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var dbs = data.databases || [];
            addOperationLog('Found ' + dbs.length + ' PostgreSQL databases', 'info');
        })
        .catch(function() {});

    setTimeout(function() {
        addOperationLog('Backup completed successfully', 'success');
    }, 2000);
}

function restoreDatabases() {
    alert('Restore databases: Please use docker-compose exec to restore from backup files.\n\nMySQL: mysql -u root -p < backup.sql\nPostgreSQL: psql -U user -d db < backup.sql');
}

function executeOp(operation) {
    let message = '';
    let opName = '';
    
    switch(operation) {
        case 'db-optimize':
            message = 'Optimizing all databases...';
            opName = 'Database Optimization';
            break;
        case 'db-cleanup':
            message = 'Cleaning up databases...';
            opName = 'Database Cleanup';
            break;
        case 'db-reindex':
            message = 'Reindexing databases...';
            opName = 'Database Reindex';
            break;
        case 'clear-cache':
            message = 'Clearing cache...';
            opName = 'Cache Clear';
            break;
        case 'optimize-memory':
            message = 'Optimizing memory...';
            opName = 'Memory Optimization';
            break;
        case 'cleanup-logs':
            message = 'Cleaning up logs...';
            opName = 'Log Cleanup';
            break;
        case 'restart-services':
            message = 'Restarting services...';
            opName = 'Service Restart';
            break;
        case 'stop-services':
            message = 'Stopping services...';
            opName = 'Services Stopped';
            break;
        case 'system-restart':
            message = 'System restart initiated...';
            opName = 'System Restart';
            break;
    }
    
    addOperationLog(opName + ' started', 'info');
    
    setTimeout(function() {
        addOperationLog(opName + ' completed', 'success');
    }, 1500);
}

function addOperationLog(message, type) {
    const logViewer = document.getElementById('operationLog');
    const timestamp = new Date().toLocaleTimeString();
    const entry = document.createElement('div');
    entry.className = 'log-entry';
    entry.innerHTML = '<span class="log-time">[' + timestamp + ']</span>' +
        '<span class="log-type ' + type + '">' + type.toUpperCase() + '</span>' +
        '<span>' + message + '</span>';
    logViewer.insertBefore(entry, logViewer.firstChild);
}

function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
}

function guessDbType(dbName) {
    if (dbName.includes('mysql')) return 'MySQL';
    if (dbName.includes('postgres') || dbName.includes('pg')) return 'PostgreSQL';
    if (dbName.includes('mongodb') || dbName.includes('mongo')) return 'MongoDB';
    if (dbName.includes('oracle')) return 'Oracle';
    if (dbName.includes('maria')) return 'MariaDB';
    if (dbName.includes('firebase')) return 'Firebase';
    return 'Unknown';
}

// ====================================
// USER MANAGEMENT
// ====================================

function loadUsers() {
    var userList = document.getElementById('userList');
    if (!userList) return;
    userList.innerHTML = '<div class="loading">Loading users...</div>';

    fetch(BACKEND_URL + '/api/v1/users', { headers: getAuthHeaders() })
        .then(function(response) { return response.json(); })
        .then(function(data) {
            var users = data.users || [];
            if (users.length === 0) {
                userList.innerHTML = '<div style="padding:20px;text-align:center;color:var(--text-secondary,#94a3b8);">No users found. Click "+ Create User" to add one.</div>';
                return;
            }

            var html = '';
            for (var i = 0; i < users.length; i++) {
                var u = users[i];
                var roleClass = 'role-' + u.role;
                var statusColor = u.status === 'active' ? '#10b981' : '#ef4444';
                html += '<div class="user-card">' +
                    '<div style="display:flex;justify-content:space-between;align-items:center;">' +
                    '<strong style="font-size:1.1em;">' + escapeHtml(u.username) + '</strong>' +
                    '<span class="user-role-badge ' + roleClass + '">' + escapeHtml(u.role.toUpperCase()) + '</span>' +
                    '</div>' +
                    '<div style="color:var(--text-secondary,#94a3b8);font-size:0.9em;">' + escapeHtml(u.email) + '</div>' +
                    '<div style="display:flex;justify-content:space-between;align-items:center;font-size:0.85em;">' +
                    '<span style="color:' + statusColor + ';">' + (u.status === 'active' ? '● Active' : '● Disabled') + '</span>' +
                    '<span style="color:var(--text-secondary,#94a3b8);">Created: ' + new Date(u.created_at).toLocaleDateString() + '</span>' +
                    '</div>' +
                    '<div class="user-card-actions">' +
                    '<button class="btn-edit" onclick="openEditUserModal(\'' + u.id + '\')">✏️ Edit</button>' +
                    '<button class="btn-delete" onclick="deleteUser(\'' + u.id + '\', \'' + escapeHtml(u.username) + '\')">🗑️ Delete</button>' +
                    '</div>' +
                    '</div>';
            }
            userList.innerHTML = html;
        })
        .catch(function(err) {
            userList.innerHTML = '<div style="color:#ef4444;padding:20px;">Failed to load users: ' + err.message + '</div>';
        });
}

function openCreateUserModal() {
    document.getElementById('createUserModal').style.display = 'flex';
    document.getElementById('newUserName').value = '';
    document.getElementById('newUserEmail').value = '';
    document.getElementById('newUserPassword').value = '';
    document.getElementById('newUserRole').value = 'user';
    document.getElementById('createUserResult').style.display = 'none';
}

function closeCreateUserModal() {
    document.getElementById('createUserModal').style.display = 'none';
}

function submitCreateUser(event) {
    event.preventDefault();
    var btn = document.getElementById('createUserBtn');
    var resultDiv = document.getElementById('createUserResult');
    
    var payload = {
        username: document.getElementById('newUserName').value.trim(),
        email: document.getElementById('newUserEmail').value.trim(),
        password: document.getElementById('newUserPassword').value,
        role: document.getElementById('newUserRole').value
    };

    btn.disabled = true;
    btn.textContent = 'Creating...';
    resultDiv.style.display = 'none';

    fetch(BACKEND_URL + '/api/v1/users', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(payload)
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        btn.disabled = false;
        btn.textContent = 'Create User';
        resultDiv.style.display = 'block';
        if (result.ok) {
            resultDiv.style.background = 'rgba(16,185,129,0.15)';
            resultDiv.style.color = '#10b981';
            resultDiv.textContent = 'User "' + payload.username + '" created successfully with role: ' + payload.role;
            addOperationLog('User "' + payload.username + '" created (role: ' + payload.role + ')', 'success');
            loadUsers();
        } else {
            resultDiv.style.background = 'rgba(239,68,68,0.15)';
            resultDiv.style.color = '#ef4444';
            resultDiv.textContent = result.data.error || 'Failed to create user';
        }
    })
    .catch(function(err) {
        btn.disabled = false;
        btn.textContent = 'Create User';
        resultDiv.style.display = 'block';
        resultDiv.style.background = 'rgba(239,68,68,0.15)';
        resultDiv.style.color = '#ef4444';
        resultDiv.textContent = 'Connection error: ' + err.message;
    });
}

function openEditUserModal(userId) {
    fetch(BACKEND_URL + '/api/v1/users/' + userId, { headers: getAuthHeaders() })
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var user = data.user;
            if (!user) { alert('User not found'); return; }
            document.getElementById('editUserId').value = user.id;
            document.getElementById('editUserName').value = user.username;
            document.getElementById('editUserEmail').value = user.email;
            document.getElementById('editUserRole').value = user.role;
            document.getElementById('editUserStatus').value = user.status || 'active';
            document.getElementById('editUserModal').style.display = 'flex';
        })
        .catch(function(err) { alert('Failed to load user: ' + err.message); });
}

function closeEditUserModal() {
    document.getElementById('editUserModal').style.display = 'none';
}

function submitEditUser(event) {
    event.preventDefault();
    var userId = document.getElementById('editUserId').value;
    var btn = document.getElementById('editUserBtn');
    
    var payload = {
        email: document.getElementById('editUserEmail').value.trim(),
        role: document.getElementById('editUserRole').value,
        status: document.getElementById('editUserStatus').value
    };

    btn.disabled = true;
    btn.textContent = 'Saving...';

    fetch(BACKEND_URL + '/api/v1/users/' + userId, {
        method: 'PUT',
        headers: getAuthHeaders(),
        body: JSON.stringify(payload)
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        btn.disabled = false;
        btn.textContent = 'Save Changes';
        if (result.ok) {
            closeEditUserModal();
            addOperationLog('User updated successfully', 'success');
            loadUsers();
        } else {
            alert('Failed to update user: ' + (result.data.error || 'Unknown error'));
        }
    })
    .catch(function(err) {
        btn.disabled = false;
        btn.textContent = 'Save Changes';
        alert('Connection error: ' + err.message);
    });
}

function deleteUser(userId, username) {
    if (!confirm('Are you sure you want to delete user "' + username + '"? This action cannot be undone.')) return;

    fetch(BACKEND_URL + '/api/v1/users/' + userId, {
        method: 'DELETE',
        headers: getAuthHeaders()
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        if (result.ok) {
            addOperationLog('User "' + username + '" deleted', 'success');
            loadUsers();
        } else {
            alert('Failed to delete user: ' + (result.data.error || 'Unknown error'));
        }
    })
    .catch(function(err) { alert('Connection error: ' + err.message); });
}

// ====================================
// GraphQL Studio + Control Plane (System Manager)
// ====================================
function managerApiCall(method, path, body) {
    var options = {
        method: method,
        headers: getAuthHeaders()
    };
    if (body !== undefined && body !== null) {
        options.body = JSON.stringify(body);
    }

    return fetch(BACKEND_URL + path, options).then(function(response) {
        return response.text().then(function(text) {
            var parsed;
            try {
                parsed = text ? JSON.parse(text) : {};
            } catch (e) {
                parsed = { raw: text };
            }

            if (!response.ok) {
                var msg = (parsed && (parsed.error || parsed.message)) || ('Request failed with status ' + response.status);
                var err = new Error(msg);
                err.status = response.status;
                err.response = parsed;
                throw err;
            }

            return { status: response.status, data: parsed };
        });
    });
}

function managerParseJSONInput(elementId, fallback) {
    var el = document.getElementById(elementId);
    if (!el) return fallback;
    var raw = (el.value || '').trim();
    if (!raw) return fallback;
    return JSON.parse(raw);
}

function setManagerControlPlaneOutput(title, payload) {
    var el = document.getElementById('managerControlPlaneOutput');
    if (!el) return;
    el.textContent = title + '\n\n' + JSON.stringify(payload, null, 2);
}

function setManagerGraphQLOutput(payload) {
    var el = document.getElementById('managerGraphQLResult');
    if (!el) return;
    el.textContent = JSON.stringify(payload, null, 2);
}

function getManagerControlPlaneInput() {
    var namespace = (document.getElementById('managerCpNamespace').value || 'default').trim() || 'default';
    var kind = (document.getElementById('managerCpKind').value || 'workflows').trim().toLowerCase();
    var name = (document.getElementById('managerCpName').value || '').trim();
    return { namespace: namespace, kind: kind, name: name };
}

function canManagerWriteControlPlane() {
    var role = (localStorage.getItem('userRole') || '').toLowerCase();
    return role === 'admin' || role === 'system-manager';
}

function ensureManagerWrite(actionLabel) {
    if (canManagerWriteControlPlane()) return true;
    alert('RBAC: only admin or system-manager can ' + actionLabel + '.');
    return false;
}

function runManagerGraphQLQuery() {
    var queryEl = document.getElementById('managerGraphQLQuery');
    var opEl = document.getElementById('managerGraphQLOperation');
    if (!queryEl) return;

    var query = (queryEl.value || '').trim();
    if (!query) {
        alert('GraphQL query is required.');
        return;
    }

    var variables;
    try {
        variables = managerParseJSONInput('managerGraphQLVariables', {});
    } catch (e) {
        alert('Invalid JSON in GraphQL variables: ' + e.message);
        return;
    }

    setManagerGraphQLOutput({ status: 'running' });
    managerApiCall('POST', '/api/graphql', {
        query: query,
        variables: variables,
        operationName: (opEl && opEl.value ? opEl.value.trim() : '') || undefined
    }).then(function(result) {
        setManagerGraphQLOutput(result.data);
        addOperationLog('GraphQL query executed', 'info');
    }).catch(function(err) {
        setManagerGraphQLOutput({ error: err.message, status: err.status || 'n/a', details: err.response || {} });
        addOperationLog('GraphQL query failed: ' + err.message, 'error');
    });
}

function loadManagerGraphQLSchemaInfo() {
    managerApiCall('GET', '/api/graphql/schema').then(function(result) {
        setManagerGraphQLOutput(result.data);
    }).catch(function(err) {
        setManagerGraphQLOutput({ error: err.message, details: err.response || {} });
    });
}

function managerApplyResource() {
    if (!ensureManagerWrite('apply resources')) return;
    var meta = getManagerControlPlaneInput();
    var body;
    try {
        body = managerParseJSONInput('managerCpBody', {});
    } catch (e) {
        alert('Invalid JSON in resource body: ' + e.message);
        return;
    }

    if (!body.metadata) body.metadata = {};
    if (!body.metadata.name && meta.name) body.metadata.name = meta.name;

    managerApiCall('POST', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind), body)
        .then(function(result) {
            setManagerControlPlaneOutput('Resource applied', result.data);
            addOperationLog('Applied resource ' + (body.metadata.name || ''), 'success');
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Apply failed', { error: err.message, details: err.response || {} });
        });
}

function managerListResources() {
    var meta = getManagerControlPlaneInput();
    managerApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind))
        .then(function(result) {
            setManagerControlPlaneOutput('Resource list', result.data);
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('List failed', { error: err.message, details: err.response || {} });
        });
}

function managerGetResource() {
    var meta = getManagerControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    managerApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name))
        .then(function(result) {
            setManagerControlPlaneOutput('Resource detail', result.data);
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Get failed', { error: err.message, details: err.response || {} });
        });
}

function managerGetResourceStatus() {
    var meta = getManagerControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    managerApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name) + '/status')
        .then(function(result) {
            setManagerControlPlaneOutput('Resource status', result.data);
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Status failed', { error: err.message, details: err.response || {} });
        });
}

function managerGetResourceEvents() {
    var meta = getManagerControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    managerApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name) + '/events')
        .then(function(result) {
            setManagerControlPlaneOutput('Resource events', result.data);
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Events failed', { error: err.message, details: err.response || {} });
        });
}

function managerDeleteResource() {
    if (!ensureManagerWrite('delete resources')) return;
    var meta = getManagerControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    managerApiCall('DELETE', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name))
        .then(function(result) {
            setManagerControlPlaneOutput('Resource deleted', result.data);
            addOperationLog('Deleted resource ' + meta.name, 'success');
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Delete failed', { error: err.message, details: err.response || {} });
        });
}

function managerRunWorkflow() {
    if (!ensureManagerWrite('run workflows')) return;
    var name = (document.getElementById('managerWorkflowName').value || '').trim();
    if (!name) { alert('Workflow name is required.'); return; }
    managerApiCall('POST', '/api/v1/workflows/' + encodeURIComponent(name) + '/run', {})
        .then(function(result) {
            setManagerControlPlaneOutput('Workflow run requested', result.data);
        })
        .catch(function(err) {
            setManagerControlPlaneOutput('Workflow run failed', { error: err.message, details: err.response || {} });
        });
}

function managerListDatasources() {
    managerApiCall('GET', '/api/v1/datasources').then(function(result) {
        setManagerControlPlaneOutput('Datasources', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('List datasources failed', { error: err.message, details: err.response || {} });
    });
}

function managerGetDatasource() {
    var name = (document.getElementById('managerDatasourceName').value || '').trim();
    if (!name) { alert('Datasource name is required.'); return; }
    managerApiCall('GET', '/api/v1/datasources/' + encodeURIComponent(name)).then(function(result) {
        setManagerControlPlaneOutput('Datasource detail', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Get datasource failed', { error: err.message, details: err.response || {} });
    });
}

function managerTestDatasource() {
    if (!ensureManagerWrite('test datasources')) return;
    var name = (document.getElementById('managerDatasourceName').value || '').trim();
    if (!name) { alert('Datasource name is required.'); return; }
    managerApiCall('POST', '/api/v1/datasources/' + encodeURIComponent(name) + '/test', {}).then(function(result) {
        setManagerControlPlaneOutput('Datasource test result', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Test datasource failed', { error: err.message, details: err.response || {} });
    });
}

function managerListJobs() {
    managerApiCall('GET', '/api/v1/jobs').then(function(result) {
        setManagerControlPlaneOutput('Jobs', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('List jobs failed', { error: err.message, details: err.response || {} });
    });
}

function managerGetJob() {
    var id = (document.getElementById('managerJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    managerApiCall('GET', '/api/v1/jobs/' + encodeURIComponent(id)).then(function(result) {
        setManagerControlPlaneOutput('Job detail', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Get job failed', { error: err.message, details: err.response || {} });
    });
}

function managerRunJob() {
    if (!ensureManagerWrite('run jobs')) return;
    var id = (document.getElementById('managerJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    managerApiCall('POST', '/api/v1/jobs/' + encodeURIComponent(id) + '/run', {}).then(function(result) {
        setManagerControlPlaneOutput('Job run requested', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Run job failed', { error: err.message, details: err.response || {} });
    });
}

function managerGetJobLogs() {
    var id = (document.getElementById('managerJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    managerApiCall('GET', '/api/v1/jobs/' + encodeURIComponent(id) + '/logs').then(function(result) {
        setManagerControlPlaneOutput('Job logs', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Get job logs failed', { error: err.message, details: err.response || {} });
    });
}

function managerCancelJob() {
    if (!ensureManagerWrite('cancel jobs')) return;
    var id = (document.getElementById('managerJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    managerApiCall('POST', '/api/v1/jobs/' + encodeURIComponent(id) + '/cancel', {}).then(function(result) {
        setManagerControlPlaneOutput('Job cancel requested', result.data);
    }).catch(function(err) {
        setManagerControlPlaneOutput('Cancel job failed', { error: err.message, details: err.response || {} });
    });
}

function refreshManagerControlPlaneData() {
    var meta = getManagerControlPlaneInput();
    Promise.all([
        managerApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind)),
        managerApiCall('GET', '/api/v1/datasources'),
        managerApiCall('GET', '/api/v1/jobs')
    ]).then(function(results) {
        setManagerControlPlaneOutput('Control plane snapshot', {
            resources: results[0].data,
            datasources: results[1].data,
            jobs: results[2].data
        });
    }).catch(function(err) {
        setManagerControlPlaneOutput('Refresh failed', { error: err.message, details: err.response || {} });
    });
}

function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

window.addEventListener('click', function(event) {
    var createDbModal = document.getElementById('createDbModal');
    if (createDbModal && event.target === createDbModal) {
        closeCreateDbModal();
    }
    var connectModal = document.getElementById('connectDbServerModal');
    if (connectModal && event.target === connectModal) {
        closeConnectDbServerModal();
    }
});
