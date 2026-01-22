// System Manager JS
const BACKEND_URL = (() => {
    const elem = document.getElementById('backendURL');
    if (elem && elem.textContent) {
        return elem.textContent.trim();
    }
    return 'http://localhost:8000';
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
    alert('Create database feature - redirect to admin panel');
}

function backupDatabases() {
    addOperationLog('Backup started for all databases', 'info');
    setTimeout(function() {
        addOperationLog('Backup completed successfully', 'success');
    }, 2000);
}

function restoreDatabases() {
    alert('Restore databases feature - please select backup file');
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
