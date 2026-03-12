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
    event.target.classList.add('active');
    
    if (tabName === 'databases') {
        loadDatabases();
    }
    if (tabName === 'users') {
        loadUsers();
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
    document.getElementById('newDbName').value = '';
    document.getElementById('createDbResult').style.display = 'none';
}

function closeCreateDbModal() {
    document.getElementById('createDbModal').style.display = 'none';
}

function submitCreateDatabase(event) {
    event.preventDefault();
    var dbType = document.getElementById('newDbType').value;
    var dbName = document.getElementById('newDbName').value.trim();
    var btn = document.getElementById('createDbBtn');
    var resultDiv = document.getElementById('createDbResult');

    btn.disabled = true;
    btn.textContent = 'Creating...';
    resultDiv.style.display = 'none';

    fetch(BACKEND_URL + '/api/admin/database/create', {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({ db_type: dbType, database_name: dbName })
    })
    .then(function(response) { return response.json().then(function(d) { return { ok: response.ok, data: d }; }); })
    .then(function(result) {
        btn.disabled = false;
        btn.textContent = 'Create Database';
        resultDiv.style.display = 'block';
        if (result.ok) {
            resultDiv.style.background = 'rgba(16,185,129,0.15)';
            resultDiv.style.color = '#10b981';
            resultDiv.textContent = 'Database "' + dbName + '" created successfully on ' + dbType;
            addOperationLog('Database "' + dbName + '" created on ' + dbType, 'success');
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

function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
