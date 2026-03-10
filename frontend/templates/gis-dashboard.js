// =============================================
// AxiomNizam GIS Dashboard JavaScript
// =============================================

const GIS_API = (window.BACKEND_URL || 'http://localhost:8000') + '/api/v1/gis';

// State
let gisMap = null;
let gisState = {
    layers: [],
    regions: [],
    markers: [],
    datasets: [],
    activeDataset: null,
    selectedRegion: null,
    mapLayers: {},           // leaflet layer references
    divisionPolygons: {},    // division id -> L.polygon
    districtPolygons: {},
    markerGroup: null,
    currentPage: 1,
    pageSize: 10,
    tableData: [],
    filteredData: [],
    isMeasuring: false,
    measurePoints: [],
    measureLine: null,
};

// =============================================
// Initialization
// =============================================

document.addEventListener('DOMContentLoaded', function () {
    initMap();
    loadGISData();
});

function initMap() {
    gisMap = L.map('gisMap', {
        center: [23.6850, 90.3563],
        zoom: 7,
        zoomControl: false,
        attributionControl: true,
        maxBounds: [[18, 85], [28, 96]],
        minZoom: 5,
        maxZoom: 18,
    });

    // Base tile layers
    const osmLight = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap contributors',
        maxZoom: 19,
    });
    const osmTopo = L.tileLayer('https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenTopoMap',
        maxZoom: 17,
    });
    const cartoDark = L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; CARTO',
        maxZoom: 19,
    });
    const cartoLight = L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; CARTO',
        maxZoom: 19,
    });

    // Use light tiles by default, matching the screenshot style
    cartoLight.addTo(gisMap);

    // Store base layers for controls
    gisState.baseLayers = {
        'CartoDB Light': cartoLight,
        'OpenStreetMap': osmLight,
        'Topographic': osmTopo,
        'CartoDB Dark': cartoDark,
    };

    // Add Layer control
    L.control.layers(gisState.baseLayers, null, { position: 'topright', collapsed: true }).addTo(gisMap);

    // Scale bar
    L.control.scale({ position: 'bottomright', imperial: false }).addTo(gisMap);

    // Track mouse position
    gisMap.on('mousemove', function (e) {
        const el = document.getElementById('mapCoords');
        if (el) el.textContent = e.latlng.lat.toFixed(4) + ', ' + e.latlng.lng.toFixed(4);
    });

    gisMap.on('zoomend', function () {
        const el = document.getElementById('mapZoom');
        if (el) el.textContent = 'Zoom ' + gisMap.getZoom();
    });

    // Measure mode click handler
    gisMap.on('click', function (e) {
        if (!gisState.isMeasuring) return;
        addMeasurePoint(e.latlng);
    });
}

// =============================================
// Data Loading
// =============================================

async function loadGISData() {
    try {
        const [summaryRes, layersRes, regionsRes, markersRes, datasetsRes] = await Promise.all([
            fetch(GIS_API + '/summary'),
            fetch(GIS_API + '/layers'),
            fetch(GIS_API + '/regions'),
            fetch(GIS_API + '/markers'),
            fetch(GIS_API + '/datasets'),
        ]);

        const summary = await summaryRes.json();
        gisState.layers = await layersRes.json();
        gisState.regions = await regionsRes.json();
        gisState.markers = await markersRes.json();
        gisState.datasets = await datasetsRes.json();

        updateSummaryCards(summary);
        renderLayers();
        renderDivisionGrid();
        renderRegionsOnMap();
        renderMarkersOnMap();
        populateFilterDropdowns();
        populateDatasetSelector();

        // Load first dataset by default
        if (gisState.datasets.length > 0) {
            loadDataset(gisState.datasets[0].id);
        }
    } catch (err) {
        console.error('Failed to load GIS data:', err);
    }
}

// =============================================
// Summary Cards
// =============================================

function updateSummaryCards(summary) {
    document.getElementById('valDivisions').textContent = summary.regionsByType?.division || 0;
    document.getElementById('valDistricts').textContent = summary.regionsByType?.district || 0;
    document.getElementById('valMarkers').textContent = summary.totalMarkers || 0;
    document.getElementById('valLayers').textContent = summary.totalLayers || 0;
    document.getElementById('valDatasets').textContent = summary.totalDatasets || 0;

    // Calculate total population from regions
    let totalPop = 0;
    const divisions = gisState.regions.filter(r => r.type === 'division');
    divisions.forEach(r => {
        if (r.properties?.population) totalPop += r.properties.population;
    });
    document.getElementById('valPopulation').textContent = formatPopulation(totalPop);
}

function formatPopulation(num) {
    if (num >= 1e9) return (num / 1e9).toFixed(2) + 'B';
    if (num >= 1e6) return (num / 1e6).toFixed(2) + 'M';
    if (num >= 1e3) return (num / 1e3).toFixed(1) + 'K';
    return num.toString();
}

// =============================================
// Layers Panel
// =============================================

function renderLayers() {
    const container = document.getElementById('layersList');
    if (!container) return;
    container.innerHTML = '';

    gisState.layers.forEach(layer => {
        const item = document.createElement('div');
        item.className = 'gis-layer-item';
        item.innerHTML = `
            <label class="gis-layer-label">
                <input type="checkbox" ${layer.visible ? 'checked' : ''} 
                    onchange="toggleLayer('${layer.id}', this.checked)">
                <span class="layer-icon">${getLayerIcon(layer.type)}</span>
                <span class="layer-name">${layer.name}</span>
            </label>
        `;
        container.appendChild(item);
    });

    // Add base map selector
    const baseItem = document.createElement('div');
    baseItem.className = 'gis-layer-item gis-layer-base';
    baseItem.innerHTML = `<span class="layer-section-title">Base Maps</span>`;
    container.appendChild(baseItem);

    Object.keys(gisState.baseLayers).forEach(name => {
        const item = document.createElement('div');
        item.className = 'gis-layer-item';
        const isActive = gisMap.hasLayer(gisState.baseLayers[name]);
        item.innerHTML = `
            <label class="gis-layer-label">
                <input type="radio" name="baseMap" ${isActive ? 'checked' : ''} 
                    onchange="switchBaseMap('${name}')">
                <span class="layer-name">${name}</span>
            </label>
        `;
        container.appendChild(item);
    });
}

function getLayerIcon(type) {
    const icons = { geojson: '🔷', tile: '🗺️', marker: '📌', heatmap: '🔥' };
    return icons[type] || '📄';
}

function toggleLayer(id, visible) {
    const layer = gisState.layers.find(l => l.id === id);
    if (layer) layer.visible = visible;

    if (id === 'divisions') {
        Object.values(gisState.divisionPolygons).forEach(p => {
            if (visible) p.addTo(gisMap);
            else gisMap.removeLayer(p);
        });
    } else if (id === 'districts') {
        Object.values(gisState.districtPolygons).forEach(p => {
            if (visible) p.addTo(gisMap);
            else gisMap.removeLayer(p);
        });
    } else if (id === 'markers') {
        if (gisState.markerGroup) {
            if (visible) gisState.markerGroup.addTo(gisMap);
            else gisMap.removeLayer(gisState.markerGroup);
        }
    }
}

function switchBaseMap(name) {
    Object.values(gisState.baseLayers).forEach(l => gisMap.removeLayer(l));
    gisState.baseLayers[name].addTo(gisMap);
}

// =============================================
// Division Grid (Quick Navigation)
// =============================================

function renderDivisionGrid() {
    const container = document.getElementById('divisionGrid');
    if (!container) return;
    container.innerHTML = '';

    const divisions = gisState.regions.filter(r => r.type === 'division');
    divisions.sort((a, b) => a.name.localeCompare(b.name));

    divisions.forEach(div => {
        const btn = document.createElement('button');
        btn.className = 'gis-division-btn';
        btn.textContent = div.name;
        btn.onclick = () => flyToDivision(div);
        container.appendChild(btn);
    });
}

function flyToDivision(division) {
    gisMap.flyTo(division.center, 9, { duration: 1 });
    showRegionProperties(division);

    // Load districts for this division
    loadDistrictsForDivision(division.id);
}

async function loadDistrictsForDivision(divisionId) {
    try {
        const res = await fetch(GIS_API + '/regions?type=district&parent=' + divisionId);
        const districts = await res.json();
        if (districts.length > 0) {
            // Update data table with district data
            const tableRows = districts.map(d => ({
                name: d.name,
                population: d.properties?.population || 0,
                area: d.properties?.area_km2 || 0,
            }));
            populateDataTable(
                [
                    { key: 'name', label: 'Title', type: 'string' },
                    { key: 'population', label: 'Value', type: 'number' },
                ],
                tableRows.map(r => ({ name: r.name, population: formatPopulation(r.population) })),
                'District Population'
            );
        }
    } catch (err) {
        console.error('Failed loading districts:', err);
    }
}

// =============================================
// Map Regions (Polygons / Circles)
// =============================================

function renderRegionsOnMap() {
    const divisions = gisState.regions.filter(r => r.type === 'division');
    const districts = gisState.regions.filter(r => r.type === 'district');
    const divLayer = gisState.layers.find(l => l.id === 'divisions');
    const distLayer = gisState.layers.find(l => l.id === 'districts');

    // Division approximate polygons as circles (real GeoJSON can be uploaded later)
    divisions.forEach(div => {
        const color = div.properties?.color || randomColor();
        const pop = div.properties?.population || 0;
        const radius = Math.max(25000, Math.sqrt(pop) * 5);

        const circle = L.circle(div.center, {
            radius: radius,
            color: darkenColor(color),
            weight: 2,
            fillColor: color,
            fillOpacity: 0.3,
            className: 'gis-region-polygon',
        });

        circle.bindTooltip(div.name, {
            permanent: true,
            direction: 'center',
            className: 'gis-division-label',
        });

        circle.on('click', () => {
            showRegionProperties(div);
            flyToDivision(div);
        });

        circle.on('mouseover', function () {
            this.setStyle({ fillOpacity: 0.5, weight: 3 });
        });
        circle.on('mouseout', function () {
            this.setStyle({ fillOpacity: 0.3, weight: 2 });
        });

        if (divLayer?.visible) circle.addTo(gisMap);
        gisState.divisionPolygons[div.id] = circle;
    });

    // District circles (smaller, only shown when zoomed in)
    districts.forEach(dist => {
        const pop = dist.properties?.population || 0;
        const radius = Math.max(5000, Math.sqrt(pop) * 2.5);

        const circle = L.circle(dist.center, {
            radius: radius,
            color: '#666',
            weight: 1,
            fillColor: getPopulationColor(pop),
            fillOpacity: 0.4,
        });

        circle.bindTooltip(dist.name, { direction: 'top' });
        circle.on('click', () => showRegionProperties(dist));

        if (distLayer?.visible) circle.addTo(gisMap);
        gisState.districtPolygons[dist.id] = circle;
    });

    // Build choropleth legend
    renderLegend();
}

function getPopulationColor(pop) {
    if (pop > 10000000) return '#1a5276';
    if (pop > 5000000) return '#2471a3';
    if (pop > 3000000) return '#2e86c1';
    if (pop > 2000000) return '#5dade2';
    if (pop > 1000000) return '#85c1e9';
    return '#d6eaf8';
}

function renderLegend() {
    const body = document.getElementById('legendBody');
    if (!body) return;
    body.innerHTML = '';

    const levels = [
        { color: '#d6eaf8', label: '< 1M' },
        { color: '#85c1e9', label: '1M - 2M' },
        { color: '#5dade2', label: '2M - 3M' },
        { color: '#2e86c1', label: '3M - 5M' },
        { color: '#2471a3', label: '5M - 10M' },
        { color: '#1a5276', label: '> 10M' },
    ];

    levels.forEach(lvl => {
        const item = document.createElement('div');
        item.className = 'gis-legend-item';
        item.innerHTML = `<span class="legend-color" style="background:${lvl.color}"></span><span>${lvl.label}</span>`;
        body.appendChild(item);
    });

    const hint = document.createElement('div');
    hint.className = 'gis-legend-hint';
    hint.textContent = 'Mouse over a region';
    body.appendChild(hint);
}

// =============================================
// Map Markers
// =============================================

function renderMarkersOnMap() {
    if (gisState.markerGroup) {
        gisMap.removeLayer(gisState.markerGroup);
    }
    gisState.markerGroup = L.layerGroup();

    const markerColors = {
        capital: '#e74c3c',
        port: '#3498db',
        airport: '#2ecc71',
        infrastructure: '#9b59b6',
        tourism: '#f39c12',
        nature: '#27ae60',
    };

    gisState.markers.forEach(m => {
        const color = m.color || markerColors[m.category] || '#3388ff';

        const icon = L.divIcon({
            html: `<div class="gis-marker-icon" style="background:${color}">
                     <span>${getMarkerEmoji(m.category)}</span>
                   </div>`,
            className: 'gis-marker-container',
            iconSize: [32, 32],
            iconAnchor: [16, 32],
            popupAnchor: [0, -32],
        });

        const marker = L.marker([m.lat, m.lng], { icon: icon });
        marker.bindPopup(`<strong>${m.name}</strong><br><em>${m.category}</em>`);
        marker.on('click', () => showMarkerProperties(m));
        gisState.markerGroup.addLayer(marker);
    });

    const markerLayer = gisState.layers.find(l => l.id === 'markers');
    if (markerLayer?.visible) gisState.markerGroup.addTo(gisMap);
}

function getMarkerEmoji(category) {
    const emojis = {
        capital: '⭐', port: '⚓', airport: '✈️',
        infrastructure: '🛣️', tourism: '🏖️', nature: '🌳',
    };
    return emojis[category] || '📍';
}

// =============================================
// Properties Panel
// =============================================

function showRegionProperties(region) {
    gisState.selectedRegion = region;
    document.getElementById('propertiesEmpty').style.display = 'none';
    document.getElementById('propertiesContent').style.display = 'block';
    document.getElementById('propTitle').textContent = region.name;

    const grid = document.getElementById('propGrid');
    grid.innerHTML = '';

    const props = {
        'Type': region.type,
        'ID': region.id,
        'Center': region.center ? region.center[0].toFixed(4) + ', ' + region.center[1].toFixed(4) : '-',
    };

    if (region.properties) {
        Object.keys(region.properties).forEach(key => {
            const val = region.properties[key];
            const label = key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            if (key === 'population') {
                props[label] = Number(val).toLocaleString();
            } else if (key === 'color') {
                return; // skip color
            } else {
                props[label] = typeof val === 'number' ? val.toLocaleString() : val;
            }
        });
    }

    Object.entries(props).forEach(([label, value]) => {
        const row = document.createElement('div');
        row.className = 'gis-prop-row';
        row.innerHTML = `<span class="prop-label">${label}</span><span class="prop-value">${value}</span>`;
        grid.appendChild(row);
    });

    // Switch to properties tab
    switchDataTab('properties');
}

function showMarkerProperties(marker) {
    document.getElementById('propertiesEmpty').style.display = 'none';
    document.getElementById('propertiesContent').style.display = 'block';
    document.getElementById('propTitle').textContent = marker.name;

    const grid = document.getElementById('propGrid');
    grid.innerHTML = '';

    const props = {
        'Category': marker.category,
        'Latitude': marker.lat.toFixed(4),
        'Longitude': marker.lng.toFixed(4),
    };

    if (marker.properties) {
        Object.entries(marker.properties).forEach(([k, v]) => {
            props[k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())] = v;
        });
    }

    Object.entries(props).forEach(([label, value]) => {
        const row = document.createElement('div');
        row.className = 'gis-prop-row';
        row.innerHTML = `<span class="prop-label">${label}</span><span class="prop-value">${value}</span>`;
        grid.appendChild(row);
    });

    switchDataTab('properties');
}

// =============================================
// Data Table
// =============================================

function populateDatasetSelector() {
    const sel = document.getElementById('filterDataset');
    if (!sel) return;
    sel.innerHTML = '';

    gisState.datasets.forEach(ds => {
        const opt = document.createElement('option');
        opt.value = ds.id;
        opt.textContent = ds.name;
        sel.appendChild(opt);
    });
}

async function switchDataset() {
    const sel = document.getElementById('filterDataset');
    if (!sel) return;
    const dsId = sel.value;
    if (dsId) await loadDataset(dsId);
}

async function loadDataset(id) {
    try {
        const res = await fetch(GIS_API + '/datasets/' + id);
        const ds = await res.json();
        gisState.activeDataset = ds;
        populateDataTable(ds.columns, ds.rows, ds.name);
        computeStats(ds);
        switchDataTab('data');
    } catch (err) {
        console.error('Failed loading dataset:', err);
    }
}

function populateDataTable(columns, rows, title) {
    document.getElementById('tableTitle').textContent = title || 'Data';

    // Build thead
    const thead = document.getElementById('dataTableHead');
    thead.innerHTML = '<tr>' + columns.map(c => `<th>${c.label}</th>`).join('') + '</tr>';

    // Store data
    gisState.tableData = rows;
    gisState.filteredData = rows;
    gisState.tableColumns = columns;
    gisState.currentPage = 1;

    renderTablePage();
}

function renderTablePage() {
    const tbody = document.getElementById('dataTableBody');
    const data = gisState.filteredData;
    const cols = gisState.tableColumns;
    const start = (gisState.currentPage - 1) * gisState.pageSize;
    const pageData = data.slice(start, start + gisState.pageSize);

    tbody.innerHTML = '';
    pageData.forEach(row => {
        const tr = document.createElement('tr');
        cols.forEach(col => {
            const td = document.createElement('td');
            let val = row[col.key];
            if (col.type === 'number' && typeof val === 'number') {
                val = val >= 1e6 ? formatPopulation(val) : val.toLocaleString();
            }
            td.textContent = val ?? '-';
            tr.appendChild(td);
        });
        tr.onclick = () => highlightRegionFromTable(row);
        tbody.appendChild(tr);
    });

    // Footer
    const totalPages = Math.ceil(data.length / gisState.pageSize);
    document.getElementById('tableInfo').textContent =
        `${start + 1}-${Math.min(start + gisState.pageSize, data.length)} of ${data.length}`;

    const pagination = document.getElementById('tablePagination');
    pagination.innerHTML = '';
    if (totalPages > 1) {
        const prevBtn = createPageBtn('‹', gisState.currentPage > 1, () => { gisState.currentPage--; renderTablePage(); });
        pagination.appendChild(prevBtn);
        for (let i = 1; i <= totalPages; i++) {
            const btn = createPageBtn(String(i), true, () => { gisState.currentPage = i; renderTablePage(); });
            if (i === gisState.currentPage) btn.classList.add('active');
            pagination.appendChild(btn);
        }
        const nextBtn = createPageBtn('›', gisState.currentPage < totalPages, () => { gisState.currentPage++; renderTablePage(); });
        pagination.appendChild(nextBtn);
    }
}

function createPageBtn(text, enabled, onclick) {
    const btn = document.createElement('button');
    btn.className = 'gis-page-btn' + (enabled ? '' : ' disabled');
    btn.textContent = text;
    if (enabled) btn.onclick = onclick;
    return btn;
}

function filterTable() {
    const query = (document.getElementById('tableSearch').value || '').toLowerCase();
    if (!query) {
        gisState.filteredData = gisState.tableData;
    } else {
        gisState.filteredData = gisState.tableData.filter(row =>
            Object.values(row).some(v => String(v).toLowerCase().includes(query))
        );
    }
    gisState.currentPage = 1;
    renderTablePage();
}

function highlightRegionFromTable(row) {
    const name = row.name || row.Name || row.division;
    if (!name) return;
    const region = gisState.regions.find(r => r.name === name);
    if (region) {
        gisMap.flyTo(region.center, 10, { duration: 0.8 });
        showRegionProperties(region);
    }
}

function exportTableCSV() {
    if (!gisState.tableColumns || !gisState.filteredData) return;
    const header = gisState.tableColumns.map(c => c.label).join(',');
    const rows = gisState.filteredData.map(row =>
        gisState.tableColumns.map(c => JSON.stringify(row[c.key] ?? '')).join(',')
    );
    const csv = [header, ...rows].join('\n');
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = (gisState.activeDataset?.name || 'data') + '.csv';
    link.click();
    URL.revokeObjectURL(link.href);
}

// =============================================
// Statistics Panel
// =============================================

function computeStats(ds) {
    const grid = document.getElementById('statsGrid');
    if (!grid) return;
    grid.innerHTML = '';

    const numCols = ds.columns.filter(c => c.type === 'number');
    numCols.forEach(col => {
        const values = ds.rows.map(r => Number(r[col.key]) || 0);
        const sum = values.reduce((a, b) => a + b, 0);
        const min = Math.min(...values);
        const max = Math.max(...values);
        const avg = sum / values.length;

        const card = document.createElement('div');
        card.className = 'gis-stats-card';
        card.innerHTML = `
            <h4>${col.label}</h4>
            <div class="stat-row"><span>Total</span><span>${formatPopulation(sum)}</span></div>
            <div class="stat-row"><span>Average</span><span>${formatPopulation(Math.round(avg))}</span></div>
            <div class="stat-row"><span>Min</span><span>${formatPopulation(min)}</span></div>
            <div class="stat-row"><span>Max</span><span>${formatPopulation(max)}</span></div>
            <div class="stat-row"><span>Count</span><span>${values.length}</span></div>
        `;
        grid.appendChild(card);
    });
}

// =============================================
// Filters
// =============================================

function populateFilterDropdowns() {
    const divSelect = document.getElementById('filterDivision');
    if (!divSelect) return;

    const divisions = gisState.regions.filter(r => r.type === 'division');
    divisions.sort((a, b) => a.name.localeCompare(b.name));

    divSelect.innerHTML = '<option value="">All Divisions</option>';
    divisions.forEach(d => {
        const opt = document.createElement('option');
        opt.value = d.id;
        opt.textContent = d.name;
        divSelect.appendChild(opt);
    });
}

function applyFilters() {
    const divId = document.getElementById('filterDivision').value;
    const regionType = document.getElementById('filterRegionType').value;

    if (divId) {
        const div = gisState.regions.find(r => r.id === divId);
        if (div) {
            gisMap.flyTo(div.center, 9, { duration: 1 });
            showRegionProperties(div);
            loadDistrictsForDivision(divId);
        }
    }

    // Show/hide regions based on type filter
    if (regionType === 'division' || regionType === '') {
        Object.values(gisState.divisionPolygons).forEach(p => p.addTo(gisMap));
    } else {
        Object.values(gisState.divisionPolygons).forEach(p => gisMap.removeLayer(p));
    }

    if (regionType === 'district' || regionType === '') {
        Object.values(gisState.districtPolygons).forEach(p => p.addTo(gisMap));
    } else {
        Object.values(gisState.districtPolygons).forEach(p => gisMap.removeLayer(p));
    }
}

function clearFilters() {
    document.getElementById('filterDivision').value = '';
    document.getElementById('filterRegionType').value = '';
    resetMapView();
}

// =============================================
// Map Tools
// =============================================

function resetMapView() {
    gisMap.flyTo([23.6850, 90.3563], 7, { duration: 1 });
}

function toggleFullscreen() {
    const mapEl = document.querySelector('.gis-map-wrapper');
    if (!document.fullscreenElement) {
        mapEl.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
}

function locateUser() {
    gisMap.locate({ setView: true, maxZoom: 12 });
    gisMap.on('locationfound', function (e) {
        L.marker(e.latlng).addTo(gisMap).bindPopup('You are here').openPopup();
    });
}

function toggleMeasure() {
    gisState.isMeasuring = !gisState.isMeasuring;
    const btn = document.getElementById('measureBtn');
    if (btn) btn.classList.toggle('active', gisState.isMeasuring);

    if (!gisState.isMeasuring) {
        // Clear measurement
        if (gisState.measureLine) {
            gisMap.removeLayer(gisState.measureLine);
            gisState.measureLine = null;
        }
        gisState.measurePoints.forEach(m => gisMap.removeLayer(m));
        gisState.measurePoints = [];
    }
}

function addMeasurePoint(latlng) {
    const marker = L.circleMarker(latlng, { radius: 5, color: '#e74c3c', fillColor: '#e74c3c', fillOpacity: 1 }).addTo(gisMap);
    gisState.measurePoints.push(marker);

    if (gisState.measurePoints.length > 1) {
        const points = gisState.measurePoints.map(m => m.getLatLng());
        if (gisState.measureLine) gisMap.removeLayer(gisState.measureLine);
        gisState.measureLine = L.polyline(points, { color: '#e74c3c', weight: 2, dashArray: '5,10' }).addTo(gisMap);

        // Calculate total distance
        let totalDist = 0;
        for (let i = 1; i < points.length; i++) {
            totalDist += points[i - 1].distanceTo(points[i]);
        }
        const distStr = totalDist > 1000 ? (totalDist / 1000).toFixed(2) + ' km' : Math.round(totalDist) + ' m';
        gisState.measureLine.bindTooltip(distStr, { permanent: true, direction: 'center' });
    }
}

// =============================================
// UI Panels
// =============================================

function toggleSidebar() {
    const sidebar = document.getElementById('gisSidebar');
    sidebar.classList.toggle('collapsed');
    setTimeout(() => gisMap.invalidateSize(), 350);
}

function toggleDataPanel() {
    const panel = document.getElementById('gisDataPanel');
    panel.classList.toggle('collapsed');
    setTimeout(() => gisMap.invalidateSize(), 350);
}

function switchDataTab(tabName) {
    document.querySelectorAll('.gis-data-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.gis-data-content').forEach(p => p.classList.remove('active'));

    const tabBtn = document.getElementById('tab' + tabName.charAt(0).toUpperCase() + tabName.slice(1));
    const panel = document.getElementById('panel' + tabName.charAt(0).toUpperCase() + tabName.slice(1));
    if (tabBtn) tabBtn.classList.add('active');
    if (panel) panel.classList.add('active');
}

function toggleLegend() {
    const body = document.getElementById('legendBody');
    const icon = document.getElementById('legendToggle');
    if (body.style.display === 'none') {
        body.style.display = 'block';
        icon.textContent = '▾';
    } else {
        body.style.display = 'none';
        icon.textContent = '▸';
    }
}

// =============================================
// Helpers
// =============================================

function randomColor() {
    const colors = ['#e8f5e9', '#e3f2fd', '#fff3e0', '#f3e5f5', '#e8eaf6', '#e0f2f1', '#fce4ec', '#fff8e1'];
    return colors[Math.floor(Math.random() * colors.length)];
}

function darkenColor(hex) {
    hex = hex.replace('#', '');
    if (hex.length === 3) hex = hex.split('').map(c => c + c).join('');
    const r = Math.max(0, parseInt(hex.substr(0, 2), 16) - 60);
    const g = Math.max(0, parseInt(hex.substr(2, 2), 16) - 60);
    const b = Math.max(0, parseInt(hex.substr(4, 2), 16) - 60);
    return '#' + [r, g, b].map(v => v.toString(16).padStart(2, '0')).join('');
}
