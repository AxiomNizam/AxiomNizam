(function() {
    'use strict';

    // ===== API Configuration =====
    const OS_API = (() => {
        let base = (typeof window.resolveBackendURL === 'function') ?
            window.resolveBackendURL() : String(window.BACKEND_URL || '').trim();
        if (!base) {
            const bh = String(window.location.hostname || '').toLowerCase();
            if (bh && bh !== 'localhost' && bh !== '127.0.0.1') {
                const proto = window.location.protocol || 'https:';
                base = bh.indexOf('axiomnizam.') === 0
                    ? proto + '//axiomnizam-platform.' + bh.substring('axiomnizam.'.length)
                    : proto + '//' + bh;
            } else {
                base = 'http://localhost:8000';
            }
        }
        if (base.endsWith('/')) base = base.slice(0, -1);
        return base + '/api/v1/storage';
    })();

    function osHeaders() {
        const h = { 'Content-Type': 'application/json' };
        const tok = localStorage.getItem('iamToken') || localStorage.getItem('authToken') ||
                    (document.cookie.match(/(?:^|;\s*)authToken=([^;]*)/) || [])[1] || '';
        if (tok) h['Authorization'] = 'Bearer ' + tok;
        return h;
    }

    async function osFetch(path, opts) {
        opts = opts || {};
        if (!opts.headers) opts.headers = osHeaders();
        try {
            return await fetch(OS_API + path, opts);
        } catch (e) {
            osToast('Network error: ' + e.message, true);
            throw e;
        }
    }

    // ===== Utilities =====
    function escHtml(s) { const d=document.createElement('div'); d.textContent=s||''; return d.innerHTML; }

    function fmtSize(bytes) {
        if (!bytes || bytes === 0) return '0 B';
        const k = 1024, units = ['B','KB','MB','GB','TB','PB'];
        const i = Math.floor(Math.log(bytes)/Math.log(k));
        return (bytes / Math.pow(k, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i];
    }

    function fmtDate(s) {
        if (!s) return '-';
        const d = new Date(s);
        return isNaN(d.getTime()) ? s : d.toLocaleString();
    }

    function osToast(msg, isErr) {
        const t = document.createElement('div');
        t.className = 'os-toast ' + (isErr ? 'os-toast-error' : 'os-toast-success');
        t.textContent = msg;
        document.body.appendChild(t);
        setTimeout(() => t.remove(), 4000);
    }

    function phaseBadge(phase) {
        const cls = phase === 'Ready' ? 'os-badge-ready' : phase === 'Pending' ? 'os-badge-pending' : 'os-badge-error';
        return '<span class="os-badge ' + cls + '">' + escHtml(phase || 'Unknown') + '</span>';
    }

    function versionBadge(v) {
        const cls = v === 'Enabled' ? 'os-badge-enabled' : 'os-badge-disabled';
        return '<span class="os-badge ' + cls + '">' + escHtml(v || 'Disabled') + '</span>';
    }

    // ===== Tab Switching =====
    window.osSwitch = function(tabId) {
        document.querySelectorAll('.os-panel').forEach(p => p.classList.remove('active'));
        document.querySelectorAll('.os-tab').forEach(b => b.classList.remove('active'));
        const panel = document.getElementById(tabId);
        if (panel) panel.classList.add('active');
        document.querySelectorAll('.os-tab').forEach(b => {
            if (b.getAttribute('onclick') && b.getAttribute('onclick').indexOf(tabId) !== -1) b.classList.add('active');
        });
        if (tabId === 'os-dashboard') osLoadDashboard();
        if (tabId === 'os-buckets') osLoadBuckets();
        if (tabId === 'os-browser') osPopulateBucketSelects();
        if (tabId === 'os-upload') osPopulateBucketSelects();
        if (tabId === 'os-policies') osLoadPolicies();
        if (tabId === 'os-metrics') osLoadMetrics();
        if (tabId === 'os-events') osLoadEvents();
    };

    // ===== Modals =====
    window.osCloseModal = function(id) { document.getElementById(id).classList.remove('show'); };
    function osOpenModal(id) { document.getElementById(id).classList.add('show'); }
    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('os-modal')) e.target.classList.remove('show');
    });

    // ===== Data Cache =====
    let osBucketsCache = [];
    let osObjectsCache = [];
    let osPoliciesCache = [];

    // ===== Dashboard =====
    window.osLoadDashboard = async function() {
        try {
            const [statsResp, bucketsResp, metricsResp] = await Promise.all([
                osFetch('/stats'),
                osFetch('/buckets'),
                osFetch('/metrics')
            ]);
            if (statsResp.ok) {
                const st = await statsResp.json();
                setText('osStatBuckets', st.totalBuckets || 0);
                setText('osStatObjects', st.totalObjects || 0);
                setText('osStatSize', fmtSize(st.totalSizeBytes));
                setText('osDashBuckets', st.totalBuckets || 0);
                setText('osDashObjects', st.totalObjects || 0);
                setText('osDashSize', fmtSize(st.totalSizeBytes));
            }
            if (bucketsResp.ok) {
                const buckets = await bucketsResp.json();
                osBucketsCache = Array.isArray(buckets) ? buckets : [];
                renderDashRecentBuckets(osBucketsCache.slice(0, 5));
            }
            if (metricsResp.ok) {
                const m = await metricsResp.json();
                setText('osStatRequests', m.totalRequests || 0);
                setText('osStatErrors', m.totalErrors || 0);
                setText('osDashRequests', m.totalRequests || 0);
                setText('osDashErrors', m.totalErrors || 0);
                const avgMs = m.totalRequests > 0 ? ((m.avgLatencyMs || 0)).toFixed(1) : '0';
                setText('osDashLatency', avgMs + ' ms');

                const backendHealthy = !!m.backendHealthy;
                setText('osStatHealth', backendHealthy ? 'Online' : 'Offline');
                document.getElementById('osStatHealth').style.color = backendHealthy ? '#22c55e' : '#ef4444';
                document.getElementById('osDashHealth').textContent = JSON.stringify({
                    status: backendHealthy ? 'healthy' : 'degraded',
                    backendHealthy: backendHealthy,
                    uptime: m.uptime || '-',
                    totalRequests: m.totalRequests || 0,
                    totalErrors: m.totalErrors || 0,
                    avgLatencyMs: Number(m.avgLatencyMs || 0).toFixed(1)
                }, null, 2);
            } else {
                setText('osStatHealth', 'Offline');
                document.getElementById('osStatHealth').style.color = '#ef4444';
                document.getElementById('osDashHealth').textContent = 'Storage metrics unavailable';
            }
        } catch(e) {
            document.getElementById('osDashHealth').textContent = 'Error: ' + e.message;
        }
    };

    function setText(id, val) { const el = document.getElementById(id); if (el) el.textContent = val; }

    function renderDashRecentBuckets(buckets) {
        const el = document.getElementById('osDashRecentBuckets');
        if (!buckets.length) { el.innerHTML = '<p style="color:var(--text-muted);">No buckets yet. Create one to get started.</p>'; return; }
        let html = '<table class="os-table"><thead><tr><th>Name</th><th>Tenant</th><th>Phase</th><th>Objects</th><th>Size</th></tr></thead><tbody>';
        buckets.forEach(b => {
            html += '<tr><td>' + escHtml(b.metadata?.name) + '</td><td>' + escHtml(b.metadata?.tenantId) + '</td><td>' + phaseBadge(b.status?.phase) + '</td><td>' + (b.status?.objectCount||0) + '</td><td class="os-size">' + fmtSize(b.status?.totalSize) + '</td></tr>';
        });
        html += '</tbody></table>';
        el.innerHTML = html;
    }

    // ===== Buckets =====
    window.osLoadBuckets = async function() {
        const tenant = document.getElementById('osTenantFilter').value;
        const qs = tenant ? '?tenantId=' + encodeURIComponent(tenant) : '';
        try {
            const resp = await osFetch('/buckets' + qs);
            if (!resp.ok) { osToast('Failed to load buckets: ' + resp.status, true); return; }
            const data = await resp.json();
            osBucketsCache = Array.isArray(data) ? data : [];
            renderBuckets(osBucketsCache);
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    function renderBuckets(buckets) {
        const tbody = document.getElementById('osBucketsBody');
        if (!buckets.length) {
            tbody.innerHTML = '<tr><td colspan="9" class="os-empty"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg><br>No buckets found</td></tr>';
            return;
        }
        tbody.innerHTML = buckets.map(b => {
            const m = b.metadata || {};
            const s = b.spec || {};
            const st = b.status || {};
            return '<tr>' +
                '<td><strong>' + escHtml(m.name) + '</strong></td>' +
                '<td>' + escHtml(m.tenantId) + '</td>' +
                '<td>' + phaseBadge(st.phase) + '</td>' +
                '<td>' + versionBadge(s.versioning) + '</td>' +
                '<td>' + (st.objectCount||0) + '</td>' +
                '<td class="os-size">' + fmtSize(st.totalSize) + '</td>' +
                '<td><button class="os-btn os-btn-secondary os-btn-sm" onclick="osShowBucketTags(\'' + escHtml(m.name) + '\',\'' + escHtml(m.tenantId) + '\')">Tags</button></td>' +
                '<td>' + fmtDate(m.createdAt) + '</td>' +
                '<td style="white-space:nowrap;">' +
                    '<button class="os-btn os-btn-secondary os-btn-sm" onclick="osBrowseThisBucket(\'' + escHtml(m.name) + '\',\'' + escHtml(m.tenantId) + '\')">Browse</button> ' +
                    '<button class="os-btn os-btn-danger os-btn-sm" onclick="osDeleteBucket(\'' + escHtml(m.name) + '\',\'' + escHtml(m.tenantId) + '\')">Delete</button>' +
                '</td></tr>';
        }).join('');
    }

    window.osFilterBuckets = function() {
        const q = (document.getElementById('osBucketSearch').value || '').toLowerCase();
        renderBuckets(osBucketsCache.filter(b => {
            const n = (b.metadata?.name || '').toLowerCase();
            const t = (b.metadata?.tenantId || '').toLowerCase();
            return n.indexOf(q) !== -1 || t.indexOf(q) !== -1;
        }));
    };

    window.osShowCreateBucket = function() { osOpenModal('osCreateBucketModal'); };

    window.osCreateBucket = async function() {
        const name = document.getElementById('osNewBucketName').value.trim();
        const tenant = document.getElementById('osNewBucketTenant').value.trim();
        if (!name || !tenant) { osToast('Name and Tenant ID required', true); return; }
        const body = {
            name: name,
            tenantId: tenant,
            versioning: document.getElementById('osNewBucketVersioning').value,
            region: document.getElementById('osNewBucketRegion').value.trim()
        };
        try {
            const resp = await osFetch('/buckets', { method: 'POST', body: JSON.stringify(body) });
            const data = await resp.json();
            if (!resp.ok) { osToast(data.error || 'Create failed', true); return; }
            osToast('Bucket "' + name + '" created');
            osCloseModal('osCreateBucketModal');
            document.getElementById('osNewBucketName').value = '';
            osLoadBuckets();
            osLoadDashboard();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osDeleteBucket = async function(name, tenantId) {
        if (!confirm('Delete bucket "' + name + '"? This cannot be undone.')) return;
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(name) + '?tenantId=' + encodeURIComponent(tenantId), { method: 'DELETE' });
            if (!resp.ok) { const d = await resp.json().catch(()=>({})); osToast(d.error || 'Delete failed', true); return; }
            osToast('Bucket deleted');
            osLoadBuckets();
            osLoadDashboard();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    // ===== Object Browser =====
    function osPopulateBucketSelects() {
        const selects = ['osBrowserBucket', 'osUploadBucket', 'osPresignBucket', 'osCopyDestBucket'];
        selects.forEach(id => {
            const sel = document.getElementById(id);
            if (!sel) return;
            const current = sel.value;
            const optStart = '<option value="">Select bucket...</option>';
            sel.innerHTML = optStart + osBucketsCache.map(b => {
                const n = b.metadata?.name || '';
                const t = b.metadata?.tenantId || '';
                return '<option value="' + escHtml(n) + '" data-tenant="' + escHtml(t) + '">' + escHtml(n) + ' (' + escHtml(t) + ')</option>';
            }).join('');
            if (current) sel.value = current;
        });
    }

    window.osBrowseBucket = async function() {
        const sel = document.getElementById('osBrowserBucket');
        const bucket = sel.value;
        if (!bucket) return;
        const opt = sel.options[sel.selectedIndex];
        const tenantId = opt.getAttribute('data-tenant') || 'default';
        const prefix = (document.getElementById('osBrowserPrefix').value || '').trim();

        const qs = '?tenantId=' + encodeURIComponent(tenantId) + (prefix ? '&prefix=' + encodeURIComponent(prefix) : '');
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(bucket) + '/objects' + qs);
            if (!resp.ok) { osToast('Failed to list objects', true); return; }
            const data = await resp.json();
            osObjectsCache = data.objects || [];
            renderObjects(osObjectsCache, bucket, tenantId);
            // Breadcrumb
            const bc = document.getElementById('osBrowserBreadcrumb');
            bc.innerHTML = '<a href="#" onclick="document.getElementById(\'osBrowserPrefix\').value=\'\'; osBrowseBucket(); return false;">Bucket ' + escHtml(bucket) + '</a>';
            if (prefix) bc.innerHTML += ' / <span>' + escHtml(prefix) + '</span>';
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osBrowseThisBucket = function(name, tenantId) {
        osSwitch('os-browser');
        setTimeout(() => {
            const sel = document.getElementById('osBrowserBucket');
            for (let i = 0; i < sel.options.length; i++) {
                if (sel.options[i].value === name) { sel.selectedIndex = i; break; }
            }
            osBrowseBucket();
        }, 100);
    };

    function renderObjects(objects, bucket, tenantId) {
        const tbody = document.getElementById('osBrowserBody');
        if (!objects.length) {
            tbody.innerHTML = '<tr><td colspan="7" class="os-empty"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg><br>No objects in this bucket</td></tr>';
            document.getElementById('osBatchDeleteBtn').style.display = 'none';
            document.getElementById('osCopyObjectBtn').style.display = 'none';
            return;
        }
        document.getElementById('osBatchDeleteBtn').style.display = '';
        document.getElementById('osCopyObjectBtn').style.display = '';
        window._osBrowserBucket = bucket;
        window._osBrowserTenant = tenantId;
        tbody.innerHTML = objects.map(o => {
            return '<tr>' +
                '<td><input type="checkbox" class="os-obj-check" data-key="' + escHtml(o.key) + '"></td>' +
                '<td class="os-mono">' + escHtml(o.key) + '</td>' +
                '<td class="os-size">' + fmtSize(o.size) + '</td>' +
                '<td>' + escHtml(o.contentType || '-') + '</td>' +
                '<td>' + fmtDate(o.lastModified) + '</td>' +
                '<td class="os-mono" style="max-width:120px; overflow:hidden; text-overflow:ellipsis;">' + escHtml(o.etag || '-') + '</td>' +
                '<td style="white-space:nowrap;">' +
                    '<button class="os-btn os-btn-secondary os-btn-sm" onclick="osDownloadObject(\'' + escHtml(bucket) + '\',\'' + escHtml(o.key) + '\',\'' + escHtml(tenantId) + '\')">Download</button> ' +
                    '<button class="os-btn os-btn-secondary os-btn-sm" onclick="osShowCopySingle(\'' + escHtml(bucket) + '\',\'' + escHtml(o.key) + '\',\'' + escHtml(tenantId) + '\')">Copy</button> ' +
                    '<button class="os-btn os-btn-danger os-btn-sm" onclick="osDeleteObjectFromBrowser(\'' + escHtml(bucket) + '\',\'' + escHtml(o.key) + '\',\'' + escHtml(tenantId) + '\')">Delete</button>' +
                '</td></tr>';
        }).join('');
    }

    window.osDownloadObject = function(bucket, key, tenantId) {
        const url = OS_API + '/buckets/' + encodeURIComponent(bucket) + '/objects/' + encodeURIComponent(key) + '?tenantId=' + encodeURIComponent(tenantId);
        const a = document.createElement('a');
        a.href = url;
        a.download = key.split('/').pop();
        a.click();
    };

    window.osDeleteObjectFromBrowser = async function(bucket, key, tenantId) {
        if (!confirm('Delete object "' + key + '"?')) return;
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(bucket) + '/objects/' + encodeURIComponent(key) + '?tenantId=' + encodeURIComponent(tenantId), { method: 'DELETE' });
            if (!resp.ok) { osToast('Delete failed', true); return; }
            osToast('Object deleted');
            osBrowseBucket();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    // ===== Upload =====
    let osSelectedFile = null;

    window.osHandleDrop = function(e) {
        e.preventDefault();
        document.getElementById('osUploadZone').classList.remove('dragover');
        if (e.dataTransfer.files.length > 0) {
            osSelectedFile = e.dataTransfer.files[0];
            document.getElementById('osUploadFileName').textContent = osSelectedFile.name + ' (' + fmtSize(osSelectedFile.size) + ')';
        }
    };
    window.osHandleDragOver = function(e) { e.preventDefault(); document.getElementById('osUploadZone').classList.add('dragover'); };
    window.osHandleDragLeave = function(e) { document.getElementById('osUploadZone').classList.remove('dragover'); };
    window.osFileSelected = function(e) {
        if (e.target.files.length > 0) {
            osSelectedFile = e.target.files[0];
            document.getElementById('osUploadFileName').textContent = osSelectedFile.name + ' (' + fmtSize(osSelectedFile.size) + ')';
        }
    };

    window.osUploadObject = async function() {
        const bucket = document.getElementById('osUploadBucket').value;
        const tenantId = (document.getElementById('osUploadTenant').value || 'default').trim();
        let key = (document.getElementById('osUploadKey').value || '').trim();
        if (!bucket) { osToast('Select a bucket', true); return; }
        if (!osSelectedFile) { osToast('Select a file to upload', true); return; }
        if (!key) key = osSelectedFile.name;

        const url = OS_API + '/buckets/' + encodeURIComponent(bucket) + '/objects/' + encodeURIComponent(key) + '?tenantId=' + encodeURIComponent(tenantId);

        const progress = document.getElementById('osUploadProgress');
        const bar = document.getElementById('osUploadProgressBar');
        progress.style.display = 'block';
        bar.style.width = '0%';

        try {
            const tok = localStorage.getItem('iamToken') || localStorage.getItem('authToken') || '';
            const xhr = new XMLHttpRequest();
            xhr.open('PUT', url, true);
            xhr.setRequestHeader('Content-Type', osSelectedFile.type || 'application/octet-stream');
            if (tok) xhr.setRequestHeader('Authorization', 'Bearer ' + tok);

            xhr.upload.onprogress = function(e) {
                if (e.lengthComputable) bar.style.width = Math.round(e.loaded/e.total*100) + '%';
            };
            xhr.onload = function() {
                if (xhr.status >= 200 && xhr.status < 300) {
                    osToast('Upload complete: ' + key);
                    bar.style.width = '100%';
                    osSelectedFile = null;
                    document.getElementById('osUploadFileName').textContent = '';
                    document.getElementById('osUploadKey').value = '';
                } else {
                    osToast('Upload failed: ' + xhr.status, true);
                }
                setTimeout(() => { progress.style.display = 'none'; }, 2000);
            };
            xhr.onerror = function() { osToast('Upload failed', true); progress.style.display = 'none'; };
            xhr.send(osSelectedFile);
        } catch(e) {
            osToast('Error: ' + e.message, true);
            progress.style.display = 'none';
        }
    };

    // ===== Pre-signed URLs =====
    window.osGeneratePresign = async function() {
        const bucket = document.getElementById('osPresignBucket').value;
        const tenantId = (document.getElementById('osPresignTenant').value || 'default').trim();
        const key = (document.getElementById('osPresignKey').value || '').trim();
        const method = document.getElementById('osPresignMethod').value;
        const expiresIn = parseInt(document.getElementById('osPresignExpiry').value) || 900;

        if (!bucket || !key) { osToast('Bucket and key are required', true); return; }

        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(bucket) + '/presign?tenantId=' + encodeURIComponent(tenantId), {
                method: 'POST',
                body: JSON.stringify({ key: key, method: method, expiresIn: expiresIn })
            });
            if (!resp.ok) { const d = await resp.json().catch(()=>({})); osToast(d.error || 'Failed', true); return; }
            const data = await resp.json();
            document.getElementById('osPresignURL').value = data.url || '';
            document.getElementById('osPresignExpires').textContent = fmtDate(data.expiresAt);
            document.getElementById('osPresignResult').style.display = 'block';
            osToast('Pre-signed URL generated');
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osCopyPresign = function() {
        const el = document.getElementById('osPresignURL');
        el.select();
        document.execCommand('copy');
        osToast('Copied to clipboard');
    };

    // ===== Policies =====
    window.osLoadPolicies = async function() {
        try {
            const resp = await osFetch('/policies');
            if (!resp.ok) { osToast('Failed to load policies', true); return; }
            osPoliciesCache = await resp.json();
            if (!Array.isArray(osPoliciesCache)) osPoliciesCache = [];
            renderPolicies(osPoliciesCache);
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    function renderPolicies(policies) {
        const tbody = document.getElementById('osPoliciesBody');
        if (!policies.length) {
            tbody.innerHTML = '<tr><td colspan="6" class="os-empty">No access policies defined</td></tr>';
            return;
        }
        tbody.innerHTML = policies.map(p => {
            return '<tr>' +
                '<td>' + escHtml(p.tenantId) + '</td>' +
                '<td>' + escHtml(p.userId) + '</td>' +
                '<td class="os-mono">' + escHtml(p.bucketName) + '</td>' +
                '<td><span class="os-badge os-badge-enabled">' + escHtml(p.role) + '</span></td>' +
                '<td class="os-mono">' + escHtml(p.prefix || '-') + '</td>' +
                '<td style="white-space:nowrap;">' +
                    '<button class="os-btn os-btn-secondary os-btn-sm" onclick="osViewPolicyDoc(' + escHtml(JSON.stringify(JSON.stringify(p.policyJson))) + ')">View</button> ' +
                    '<button class="os-btn os-btn-danger os-btn-sm" onclick="osDeletePolicy(\'' + escHtml(p.tenantId) + '\',\'' + escHtml(p.userId) + '\',\'' + escHtml(p.bucketName) + '\')">Delete</button>' +
                '</td></tr>';
        }).join('');
    }

    window.osShowCreatePolicy = function() { osOpenModal('osCreatePolicyModal'); };

    window.osCreatePolicy = async function() {
        const body = {
            tenantId: document.getElementById('osNewPolicyTenant').value.trim(),
            userId: document.getElementById('osNewPolicyUser').value.trim(),
            bucketName: document.getElementById('osNewPolicyBucket').value.trim(),
            role: document.getElementById('osNewPolicyRole').value,
            prefix: document.getElementById('osNewPolicyPrefix').value.trim()
        };
        if (!body.tenantId || !body.userId || !body.bucketName) { osToast('All required fields must be filled', true); return; }
        try {
            const resp = await osFetch('/policies', { method: 'POST', body: JSON.stringify(body) });
            const data = await resp.json();
            if (!resp.ok) { osToast(data.error || 'Create failed', true); return; }
            osToast('Policy created');
            osCloseModal('osCreatePolicyModal');
            osLoadPolicies();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osDeletePolicy = async function(tenantId, userId, bucket) {
        if (!confirm('Delete this access policy?')) return;
        try {
            const resp = await osFetch('/policies/' + encodeURIComponent(tenantId) + '/' + encodeURIComponent(userId) + '/' + encodeURIComponent(bucket), { method: 'DELETE' });
            if (!resp.ok) { osToast('Delete failed', true); return; }
            osToast('Policy deleted');
            osLoadPolicies();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osViewPolicyDoc = function(jsonStr) {
        try {
            const parsed = JSON.parse(jsonStr);
            document.getElementById('osViewPolicyJSON').textContent = typeof parsed === 'string' ? parsed : JSON.stringify(parsed, null, 2);
        } catch(e) {
            document.getElementById('osViewPolicyJSON').textContent = jsonStr;
        }
        osOpenModal('osViewPolicyModal');
    };

    // ===== Metrics =====
    window.osLoadMetrics = async function() {
        try {
            const [metricsResp, bucketsResp] = await Promise.all([
                osFetch('/metrics'),
                osFetch('/buckets')
            ]);
            if (!metricsResp.ok) { osToast('Failed to load metrics', true); return; }

            const m = await metricsResp.json();
            setText('osMetricRequests', m.totalRequests || 0);
            setText('osMetricBytesIn', fmtSize(m.totalBytesIn));
            setText('osMetricBytesOut', fmtSize(m.totalBytesOut));
            setText('osMetricErrors', m.totalErrors || 0);
            const avgMs = m.totalRequests > 0 ? (m.avgLatencyMs || 0).toFixed(1) : '0';
            setText('osMetricAvgLatency', avgMs + ' ms');
            setText('osMetricUptime', m.uptime || '-');

            const tbody = document.getElementById('osMetricsBucketBody');
            if (!bucketsResp.ok) {
                tbody.innerHTML = '<tr><td colspan="10" style="text-align:center; padding:20px; color:var(--text-muted);">Unable to load bucket list</td></tr>';
                return;
            }

            const buckets = await bucketsResp.json();
            const bucketList = Array.isArray(buckets) ? buckets : [];
            if (!bucketList.length) {
                tbody.innerHTML = '<tr><td colspan="10" style="text-align:center; padding:20px; color:var(--text-muted);">No per-bucket metrics yet</td></tr>';
                return;
            }

            const metricRows = await Promise.all(bucketList.map(async (b) => {
                const name = (b.metadata && b.metadata.name) ? b.metadata.name : '';
                const tenantId = (b.metadata && b.metadata.tenantId) ? b.metadata.tenantId : 'default';
                if (!name) return null;
                const resp = await osFetch('/metrics/' + encodeURIComponent(name) + '?tenantId=' + encodeURIComponent(tenantId));
                if (!resp.ok) return null;
                const bm = await resp.json();
                return {
                    bucketName: bm.bucketName || name,
                    requestCount: bm.requestCount || 0,
                    getRequests: bm.getRequests || 0,
                    putRequests: bm.putRequests || 0,
                    deleteRequests: bm.deleteRequests || 0,
                    bytesIn: bm.bytesIn || 0,
                    bytesOut: bm.bytesOut || 0,
                    errorCount: bm.errorCount || 0,
                    avgLatencyMs: Number(bm.avgLatencyMs || 0),
                    lastAccessed: bm.lastAccessed || ''
                };
            }));

            const rows = metricRows.filter(Boolean);
            if (!rows.length) {
                tbody.innerHTML = '<tr><td colspan="10" style="text-align:center; padding:20px; color:var(--text-muted);">No per-bucket metrics available</td></tr>';
                return;
            }

            rows.sort((a, b) => b.requestCount - a.requestCount);
            tbody.innerHTML = rows.map((b) => {
                const bAvg = b.requestCount > 0 ? b.avgLatencyMs.toFixed(1) : '0';
                return '<tr>' +
                    '<td><strong>' + escHtml(b.bucketName) + '</strong></td>' +
                    '<td>' + b.requestCount + '</td>' +
                    '<td>' + b.getRequests + '</td>' +
                    '<td>' + b.putRequests + '</td>' +
                    '<td>' + b.deleteRequests + '</td>' +
                    '<td class="os-size">' + fmtSize(b.bytesIn) + '</td>' +
                    '<td class="os-size">' + fmtSize(b.bytesOut) + '</td>' +
                    '<td style="color:#ef4444;">' + b.errorCount + '</td>' +
                    '<td>' + bAvg + ' ms</td>' +
                    '<td>' + fmtDate(b.lastAccessed) + '</td>' +
                '</tr>';
            }).join('');
        } catch(e) { osToast('Error loading metrics: ' + e.message, true); }
    };

    function fmtDuration(sec) {
        if (!sec || sec <= 0) return '-';
        const d = Math.floor(sec / 86400);
        const h = Math.floor((sec % 86400) / 3600);
        const m = Math.floor((sec % 3600) / 60);
        if (d > 0) return d + 'd ' + h + 'h';
        if (h > 0) return h + 'h ' + m + 'm';
        return m + 'm ' + Math.floor(sec % 60) + 's';
    }

    // ===== Events =====
    window.osLoadEvents = async function() {
        const evType = document.getElementById('osEventsTypeFilter').value;
        const limit = document.getElementById('osEventsLimitFilter').value || '100';
        let qs = '?limit=' + limit;
        if (evType) qs += '&type=' + encodeURIComponent(evType);
        try {
            const resp = await osFetch('/events' + qs);
            if (!resp.ok) { osToast('Failed to load events', true); return; }
            const payload = await resp.json();
            const evts = Array.isArray(payload) ? payload : (Array.isArray(payload.events) ? payload.events : []);
            renderEvents(evts);
        } catch(e) { osToast('Error loading events: ' + e.message, true); }
    };

    function eventBadge(type) {
        const colors = {
            'bucket.created': '#22c55e', 'bucket.deleted': '#ef4444',
            'object.uploaded': '#60a5fa', 'object.downloaded': '#a78bfa',
            'object.deleted': '#ef4444', 'object.copied': '#f59e0b',
            'object.multi-deleted': '#ef4444', 'policy.created': '#22c55e',
            'policy.deleted': '#ef4444', 'presign.generated': '#06b6d4'
        };
        const color = colors[type] || '#94a3b8';
        return '<span class="os-badge" style="background:' + color + '22; color:' + color + ';">' + escHtml(type) + '</span>';
    }

    function renderEvents(events) {
        const tbody = document.getElementById('osEventsBody');
        if (!events.length) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center; padding:30px; color:var(--text-muted);">No events found</td></tr>';
            return;
        }
        tbody.innerHTML = events.map(e => {
            return '<tr>' +
                '<td style="white-space:nowrap;">' + fmtDate(e.timestamp) + '</td>' +
                '<td>' + eventBadge(e.type) + '</td>' +
                '<td>' + escHtml(e.tenantId || '-') + '</td>' +
                '<td class="os-mono">' + escHtml(e.bucket || '-') + '</td>' +
                '<td class="os-mono" style="max-width:200px; overflow:hidden; text-overflow:ellipsis;">' + escHtml(e.key || '-') + '</td>' +
                '<td class="os-size">' + (e.size ? fmtSize(e.size) : '-') + '</td>' +
                '<td>' + escHtml(e.details || '-') + '</td>' +
            '</tr>';
        }).join('');
    }

    // ===== Bucket Tags =====
    let osCurrentTagsBucket = '';
    let osCurrentTagsTenant = '';
    let osCurrentTags = [];

    window.osShowBucketTags = async function(bucket, tenantId) {
        osCurrentTagsBucket = bucket;
        osCurrentTagsTenant = tenantId;
        document.getElementById('osTagsBucketName').textContent = bucket;
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(bucket) + '/tags?tenantId=' + encodeURIComponent(tenantId));
            if (resp.ok) {
                const data = await resp.json();
                osCurrentTags = data.tags || [];
            } else {
                osCurrentTags = [];
            }
        } catch(e) { osCurrentTags = []; }
        renderTags();
        osOpenModal('osBucketTagsModal');
    };

    function renderTags() {
        const el = document.getElementById('osTagsList');
        if (!osCurrentTags.length) {
            el.innerHTML = '<p style="color:var(--text-muted); font-size:.9rem;">No tags set</p>';
            return;
        }
        el.innerHTML = '<table class="os-table"><thead><tr><th>Key</th><th>Value</th><th></th></tr></thead><tbody>' +
            osCurrentTags.map((t, i) => '<tr><td class="os-mono">' + escHtml(t.key) + '</td><td class="os-mono">' + escHtml(t.value) + '</td><td><button class="os-btn os-btn-danger os-btn-sm" onclick="osRemoveTag(' + i + ')">Ã—</button></td></tr>').join('') +
            '</tbody></table>';
    }

    window.osAddTag = function() {
        const k = document.getElementById('osNewTagKey').value.trim();
        const v = document.getElementById('osNewTagValue').value.trim();
        if (!k) { osToast('Tag key required', true); return; }
        osCurrentTags.push({ key: k, value: v });
        document.getElementById('osNewTagKey').value = '';
        document.getElementById('osNewTagValue').value = '';
        renderTags();
    };

    window.osRemoveTag = function(idx) {
        osCurrentTags.splice(idx, 1);
        renderTags();
    };

    window.osSaveTags = async function() {
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(osCurrentTagsBucket) + '/tags?tenantId=' + encodeURIComponent(osCurrentTagsTenant), {
                method: 'PUT',
                body: JSON.stringify({ tags: osCurrentTags })
            });
            if (!resp.ok) { const d = await resp.json().catch(()=>({})); osToast(d.error || 'Save tags failed', true); return; }
            osToast('Tags saved');
            osCloseModal('osBucketTagsModal');
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    window.osDeleteAllTags = async function() {
        if (!confirm('Delete all tags from "' + osCurrentTagsBucket + '"?')) return;
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(osCurrentTagsBucket) + '/tags?tenantId=' + encodeURIComponent(osCurrentTagsTenant), { method: 'DELETE' });
            if (!resp.ok) { osToast('Delete tags failed', true); return; }
            osToast('Tags deleted');
            osCurrentTags = [];
            renderTags();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    // ===== Batch Delete =====
    window.osToggleSelectAll = function(checked) {
        document.querySelectorAll('.os-obj-check').forEach(cb => { cb.checked = checked; });
    };

    function osGetSelectedKeys() {
        const keys = [];
        document.querySelectorAll('.os-obj-check:checked').forEach(cb => keys.push(cb.getAttribute('data-key')));
        return keys;
    }

    window.osBatchDelete = async function() {
        const keys = osGetSelectedKeys();
        if (!keys.length) { osToast('No objects selected', true); return; }
        if (!confirm('Delete ' + keys.length + ' selected object(s)?')) return;
        const bucket = window._osBrowserBucket;
        const tenantId = window._osBrowserTenant || 'default';
        try {
            const resp = await osFetch('/buckets/' + encodeURIComponent(bucket) + '/multi-delete?tenantId=' + encodeURIComponent(tenantId), {
                method: 'POST',
                body: JSON.stringify({ keys: keys })
            });
            if (!resp.ok) { const d = await resp.json().catch(()=>({})); osToast(d.error || 'Batch delete failed', true); return; }
            const data = await resp.json();
            osToast('Deleted ' + (data.deleted || 0) + ' of ' + (data.total || keys.length) + ' objects');
            document.getElementById('osSelectAll').checked = false;
            osBrowseBucket();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    // ===== Copy Object =====
    let osCopySrcBucket = '';
    let osCopySrcKey = '';
    let osCopySrcTenant = '';

    window.osShowCopySingle = function(bucket, key, tenantId) {
        osCopySrcBucket = bucket;
        osCopySrcKey = key;
        osCopySrcTenant = tenantId;
        document.getElementById('osCopySrc').value = bucket + '/' + key;
        document.getElementById('osCopyDestKey').value = key;
        document.getElementById('osCopyTenant').value = tenantId;
        // Populate dest bucket select
        const sel = document.getElementById('osCopyDestBucket');
        sel.innerHTML = '<option value="">Select bucket...</option>' +
            osBucketsCache.map(b => {
                const n = b.metadata?.name || '';
                return '<option value="' + escHtml(n) + '">' + escHtml(n) + '</option>';
            }).join('');
        osOpenModal('osCopyObjectModal');
    };

    window.osShowCopyModal = function() {
        const keys = osGetSelectedKeys();
        if (keys.length !== 1) { osToast('Select exactly one object to copy', true); return; }
        const bucket = window._osBrowserBucket;
        const tenantId = window._osBrowserTenant || 'default';
        osShowCopySingle(bucket, keys[0], tenantId);
    };

    window.osDoCopy = async function() {
        const destBucket = document.getElementById('osCopyDestBucket').value;
        const destKey = document.getElementById('osCopyDestKey').value.trim();
        const tenantId = (document.getElementById('osCopyTenant').value || 'default').trim();
        if (!destBucket || !destKey) { osToast('Destination bucket and key required', true); return; }
        try {
            const resp = await osFetch('/copy', {
                method: 'POST',
                body: JSON.stringify({
                    sourceBucket: osCopySrcBucket,
                    sourceKey: osCopySrcKey,
                    destBucket: destBucket,
                    destKey: destKey,
                    tenantId: tenantId
                })
            });
            if (!resp.ok) { const d = await resp.json().catch(()=>({})); osToast(d.error || 'Copy failed', true); return; }
            osToast('Object copied to ' + destBucket + '/' + destKey);
            osCloseModal('osCopyObjectModal');
            osBrowseBucket();
        } catch(e) { osToast('Error: ' + e.message, true); }
    };

    // ===== Init =====
    async function osInit() {
        await osLoadDashboard();
        await osLoadBuckets();
        osPopulateBucketSelects();
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', osInit);
    } else {
        osInit();
    }

})();
