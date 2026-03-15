// =============================================
// AxiomNizam GIS Dashboard JavaScript
// Supports: General, Agriculture, Industries, Medical (Domestic BD)
//           Satellite, Airplane, Ship (International)
// =============================================

const GIS_API = (window.BACKEND_URL || 'http://localhost:8000') + '/api/v1/gis';

function readAuthCookie(name) {
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

function buildAuthHeaders(includeJSONContentType) {
    if (typeof getAuthHeaders === 'function') {
        const headers = getAuthHeaders();
        if (!includeJSONContentType) {
            delete headers['Content-Type'];
        }
        return headers;
    }

    const headers = {};
    const token = localStorage.getItem('authToken') || readAuthCookie('authToken');
    if (token) {
        headers.Authorization = token.startsWith('Bearer ') ? token : 'Bearer ' + token;
    }
    if (includeJSONContentType) {
        headers['Content-Type'] = 'application/json';
    }
    return headers;
}

// Current dashboard type
let currentDashType = 'general';

// State
let gisMap = null;
let gisState = {
    layers: [],
    regions: [],
    markers: [],
    datasets: [],
    activeDataset: null,
    selectedRegion: null,
    mapLayers: {},
    divisionPolygons: {},
    districtPolygons: {},
    markerGroup: null,
    currentPage: 1,
    pageSize: 10,
    tableData: [],
    filteredData: [],
    tableColumns: [],
    isMeasuring: false,
    measurePoints: [],
    measureLine: null,
    baseLayers: {},
    dashboardConfig: {},
};

// Dashboard theme configurations
const DASH_THEMES = {
    general:     { icon: '🌍', title: 'GENERAL',     accent: '#3b82f6', legendTitle: 'Population' },
    agriculture: { icon: '🌾', title: 'AGRICULTURE',  accent: '#2ecc71', legendTitle: 'Rice Production (MT)' },
    industries:  { icon: '🏭', title: 'INDUSTRIES',   accent: '#e74c3c', legendTitle: 'Industrial Output (Cr)' },
    medical:     { icon: '🏥', title: 'MEDICAL',      accent: '#e74c3c', legendTitle: 'EPI Coverage %' },
    satellite:   { icon: '🛰️', title: 'SATELLITE',    accent: '#00bcd4', legendTitle: 'Orbit Type' },
    airplane:    { icon: '✈️', title: 'AIRPLANE',      accent: '#ff5722', legendTitle: 'Traffic Density' },
    ship:        { icon: '🚢', title: 'SHIP',          accent: '#0277bd', legendTitle: 'Port Throughput (TEU)' },
};

const MARKER_EMOJIS = {
    general:     { capital: '⭐', port: '⚓', airport: '✈️', infrastructure: '🛣️', tourism: '🏖️', nature: '🌳' },
    agriculture: { research: '🔬', cold_storage: '❄️', processing: '🏭', tea_estate: '🍵', dairy: '🥛', seed_bank: '🌱', market: '🛒' },
    industries:  { epz: '🏗️', industrial_zone: '🏭', textile: '🧵', port_industry: '⚓', tech_park: '💻', sez: '📦', power: '⚡' },
    medical:     { hospital: '🏥', research: '🔬', clinic: '🩺', vaccine: '💉', blood_bank: '🩸' },
    satellite:   { satellite: '🛰️', constellation: '✨', navigation: '📡', weather: '🌤️', ground_station: '📻', launch_site: '🚀', earth_observation: '🌍' },
    airplane:    { airport: '✈️', aircraft: '🛩️' },
    ship:        { port: '⚓', canal: '🌊', strait: '🌊', vessel: '🚢' },
};

// Map configurations per scope
const MAP_CONFIG = {
    domestic:      { center: [23.6850, 90.3563], zoom: 7, maxBounds: [[18, 85], [28, 96]], minZoom: 5 },
    international: { center: [20, 0], zoom: 2, maxBounds: [[-85, -180], [85, 180]], minZoom: 2 },
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
// Dashboard Type Switching
// =============================================

function getDashScope() {
    return (currentDashType === 'general' || currentDashType === 'agriculture' || currentDashType === 'industries' || currentDashType === 'medical') ? 'domestic' : 'international';
}

function switchDashboardType(type) {
    if (type === currentDashType) return;
    currentDashType = type;

    // Update type bar button states
    document.querySelectorAll('.gis-type-btn').forEach(b => b.classList.remove('active'));
    const activeBtn = document.querySelector('.gis-type-btn[data-type="' + type + '"]');
    if (activeBtn) activeBtn.classList.add('active');

    // Update dashboard info panel
    const theme = DASH_THEMES[type] || DASH_THEMES.general;
    const iconEl = document.getElementById('dashInfoIcon');
    const titleEl = document.getElementById('dashInfoTitle');
    if (iconEl) iconEl.textContent = theme.icon;
    if (titleEl) titleEl.textContent = theme.title;

    // Update accent color
    document.documentElement.style.setProperty('--gis-accent', theme.accent);

    // Clear map layers
    clearMapLayers();

    // Reconfigure map for scope
    const scope = getDashScope();
    const mapCfg = MAP_CONFIG[scope];
    gisMap.setMaxBounds(mapCfg.maxBounds);
    gisMap.setMinZoom(mapCfg.minZoom);
    gisMap.flyTo(mapCfg.center, mapCfg.zoom, { duration: 1.2 });

    // Load data for new type
    loadGISData();
}

function clearMapLayers() {
    Object.values(gisState.divisionPolygons).forEach(p => gisMap.removeLayer(p));
    Object.values(gisState.districtPolygons).forEach(p => gisMap.removeLayer(p));
    if (gisState.markerGroup) gisMap.removeLayer(gisState.markerGroup);
    gisState.divisionPolygons = {};
    gisState.districtPolygons = {};
    gisState.markerGroup = null;
    gisState.layers = [];
    gisState.regions = [];
    gisState.markers = [];
    gisState.datasets = [];
    gisState.activeDataset = null;
    gisState.selectedRegion = null;
}

// =============================================
// Data Loading
// =============================================

async function loadGISData() {
    try {
        if (currentDashType === 'general') {
            // Use original GIS endpoints
            const [summaryRes, layersRes, regionsRes, markersRes, datasetsRes] = await Promise.all([
                fetch(GIS_API + '/summary', { headers: buildAuthHeaders(false) }),
                fetch(GIS_API + '/layers', { headers: buildAuthHeaders(false) }),
                fetch(GIS_API + '/regions', { headers: buildAuthHeaders(false) }),
                fetch(GIS_API + '/markers', { headers: buildAuthHeaders(false) }),
                fetch(GIS_API + '/datasets', { headers: buildAuthHeaders(false) }),
            ]);
            if (!summaryRes.ok || !layersRes.ok || !regionsRes.ok || !markersRes.ok || !datasetsRes.ok) {
                throw new Error('One or more API calls failed');
            }
            const summary = await summaryRes.json();
            gisState.layers = await layersRes.json();
            gisState.regions = await regionsRes.json();
            gisState.markers = await markersRes.json();
            gisState.datasets = await datasetsRes.json();

            updateSummaryCards(summary);
            renderDashDescription('Population, divisions, districts and points of interest across Bangladesh');
        } else {
            // Use specialized dashboard endpoint
            const res = await fetch(GIS_API + '/dashboards/' + encodeURIComponent(currentDashType), {
                headers: buildAuthHeaders(false)
            });
            if (!res.ok) throw new Error('Dashboard API failed: ' + res.status);
            const data = await res.json();

            gisState.layers = data.layers || [];
            gisState.regions = data.regions || [];
            gisState.markers = data.markers || [];
            gisState.datasets = data.datasets || [];
            gisState.dashboardConfig = data.config || {};

            const summary = {
                totalLayers: gisState.layers.length,
                totalRegions: gisState.regions.length,
                totalMarkers: gisState.markers.length,
                totalDatasets: gisState.datasets.length,
                regionsByType: {},
            };
            gisState.regions.forEach(r => {
                summary.regionsByType[r.type] = (summary.regionsByType[r.type] || 0) + 1;
            });
            updateSummaryCards(summary);
            renderDashDescription(data.description || '');
        }

        renderLayers();
        renderDivisionGrid();
        renderRegionsOnMap();
        renderMarkersOnMap();
        populateFilterDropdowns();
        populateDatasetSelector();

        if (gisState.datasets.length > 0) {
            loadDataset(gisState.datasets[0].id);
        }

        // Reset properties panel
        const emptyEl = document.getElementById('propertiesEmpty');
        const contentEl = document.getElementById('propertiesContent');
        if (emptyEl) emptyEl.style.display = 'flex';
        if (contentEl) contentEl.style.display = 'none';
    } catch (err) {
        console.error('Failed to load GIS data:', err);
    }
}

function renderDashDescription(desc) {
    const el = document.getElementById('dashInfoDesc');
    if (el) el.textContent = desc;
}

// =============================================
// Summary Cards
// =============================================

function updateSummaryCards(summary) {
    const cards = getSummaryCardConfig(summary);
    const cardElements = document.querySelectorAll('.gis-summary-card');

    cards.forEach((card, i) => {
        if (cardElements[i]) {
            const iconEl = cardElements[i].querySelector('.summary-icon');
            const valEl = cardElements[i].querySelector('.summary-value');
            const lblEl = cardElements[i].querySelector('.summary-label');
            if (iconEl) { iconEl.textContent = card.icon; iconEl.style.background = card.color; }
            if (valEl) valEl.textContent = card.value;
            if (lblEl) lblEl.textContent = card.label;
        }
    });
}

function getSummaryCardConfig(summary) {
    const rbt = summary.regionsByType || {};
    const totalRegions = Object.values(rbt).reduce((a, b) => a + b, 0);

    switch (currentDashType) {
        case 'general':
            return [
                { icon: '🏠', color: '#3b82f6', value: rbt.division || 0, label: 'Divisions' },
                { icon: '🏢', color: '#10b981', value: rbt.district || 0, label: 'Districts' },
                { icon: '👥', color: '#8b5cf6', value: calcTotalPopField('population'), label: 'Population' },
                { icon: '📌', color: '#f59e0b', value: summary.totalMarkers || 0, label: 'Markers' },
                { icon: '🗺️', color: '#ef4444', value: summary.totalLayers || 0, label: 'Layers' },
                { icon: '📁', color: '#06b6d4', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'agriculture':
            return [
                { icon: '🌾', color: '#2ecc71', value: totalRegions, label: 'Agri Zones' },
                { icon: '🌿', color: '#27ae60', value: calcFieldTotal('rice_production'), label: 'Rice (MT)' },
                { icon: '🌻', color: '#f39c12', value: calcFieldTotal('wheat_production'), label: 'Wheat (MT)' },
                { icon: '📌', color: '#3498db', value: summary.totalMarkers || 0, label: 'Facilities' },
                { icon: '💧', color: '#2980b9', value: calcFieldAvg('irrigation_pct') + '%', label: 'Avg Irrigation' },
                { icon: '📁', color: '#8e44ad', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'industries':
            return [
                { icon: '🏭', color: '#e74c3c', value: totalRegions, label: 'Ind. Divisions' },
                { icon: '🏗️', color: '#c0392b', value: calcFieldTotal('factories'), label: 'Factories' },
                { icon: '👷', color: '#f39c12', value: calcFieldTotal('employment'), label: 'Employment' },
                { icon: '📦', color: '#3498db', value: calcFieldTotal('garment_units'), label: 'Garment Units' },
                { icon: '📌', color: '#9b59b6', value: summary.totalMarkers || 0, label: 'Markers' },
                { icon: '📁', color: '#1abc9c', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'medical':
            return [
                { icon: '🏥', color: '#e74c3c', value: calcFieldTotal('hospitals'), label: 'Hospitals' },
                { icon: '🛏️', color: '#3498db', value: calcFieldTotal('beds'), label: 'Hospital Beds' },
                { icon: '👨‍⚕️', color: '#2ecc71', value: calcFieldTotal('doctors'), label: 'Doctors' },
                { icon: '💉', color: '#f39c12', value: calcFieldAvg('epi_coverage') + '%', label: 'Avg EPI Coverage' },
                { icon: '🩺', color: '#9b59b6', value: calcFieldTotal('community_clinics'), label: 'Clinics' },
                { icon: '📁', color: '#1abc9c', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'satellite':
            return [
                { icon: '🛰️', color: '#00bcd4', value: totalRegions, label: 'Orbit Zones' },
                { icon: '📡', color: '#ff9800', value: calcFieldTotal('satellites'), label: 'Satellites' },
                { icon: '🚀', color: '#e91e63', value: countMarkerCategory('launch_site'), label: 'Launch Sites' },
                { icon: '📻', color: '#f44336', value: countMarkerCategory('ground_station'), label: 'Ground Stations' },
                { icon: '📌', color: '#4caf50', value: summary.totalMarkers || 0, label: 'Total Markers' },
                { icon: '📁', color: '#9c27b0', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'airplane':
            return [
                { icon: '✈️', color: '#ff5722', value: totalRegions, label: 'Air Regions' },
                { icon: '🛬', color: '#2196f3', value: countMarkerCategory('airport'), label: 'Airports' },
                { icon: '🛩️', color: '#ff9800', value: countMarkerCategory('aircraft'), label: 'Aircraft' },
                { icon: '👥', color: '#4caf50', value: calcFieldTotal('annual_passengers'), label: 'Passengers/yr' },
                { icon: '🛫', color: '#9c27b0', value: calcFieldTotal('airlines'), label: 'Airlines' },
                { icon: '📁', color: '#607d8b', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        case 'ship':
            return [
                { icon: '🌊', color: '#0277bd', value: totalRegions, label: 'Ocean Zones' },
                { icon: '⚓', color: '#f44336', value: countMarkerCategory('port'), label: 'Ports' },
                { icon: '🚢', color: '#ff9800', value: countMarkerCategory('vessel'), label: 'Vessels' },
                { icon: '🌊', color: '#00897b', value: countMarkerCategory('canal') + countMarkerCategory('strait'), label: 'Chokepoints' },
                { icon: '📌', color: '#4caf50', value: summary.totalMarkers || 0, label: 'Total Markers' },
                { icon: '📁', color: '#607d8b', value: summary.totalDatasets || 0, label: 'Datasets' },
            ];
        default:
            return [
                { icon: '🗺️', color: '#3b82f6', value: totalRegions, label: 'Regions' },
                { icon: '📌', color: '#f59e0b', value: summary.totalMarkers || 0, label: 'Markers' },
                { icon: '🗺️', color: '#ef4444', value: summary.totalLayers || 0, label: 'Layers' },
                { icon: '📁', color: '#06b6d4', value: summary.totalDatasets || 0, label: 'Datasets' },
                { icon: '📊', color: '#10b981', value: '-', label: '-' },
                { icon: '🔷', color: '#8b5cf6', value: '-', label: '-' },
            ];
    }
}

function calcTotalPopField(field) {
    const divs = gisState.regions.filter(r => r.type === 'division');
    let total = 0;
    divs.forEach(r => { if (r.properties && r.properties[field]) total += r.properties[field]; });
    return formatPopulation(total);
}

function calcFieldTotal(field) {
    let total = 0;
    gisState.regions.forEach(r => {
        if (r.properties && r.properties[field]) total += Number(r.properties[field]);
    });
    return formatPopulation(total);
}

function calcFieldAvg(field) {
    let total = 0, count = 0;
    gisState.regions.forEach(r => {
        if (r.properties && r.properties[field] !== undefined) {
            total += Number(r.properties[field]);
            count++;
        }
    });
    return count > 0 ? Math.round(total / count) : 0;
}

function countMarkerCategory(cat) {
    return gisState.markers.filter(m => m.category === cat).length;
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

    // Determine which map group to toggle by layer index/type
    const geoLayers = gisState.layers.filter(l => l.type === 'geojson' || l.type === 'heatmap');
    const markerLayers = gisState.layers.filter(l => l.type === 'marker');

    if (geoLayers.length > 0 && geoLayers[0].id === id) {
        Object.values(gisState.divisionPolygons).forEach(p => {
            if (visible) p.addTo(gisMap); else gisMap.removeLayer(p);
        });
    } else if (geoLayers.length > 1 && geoLayers[1].id === id) {
        Object.values(gisState.districtPolygons).forEach(p => {
            if (visible) p.addTo(gisMap); else gisMap.removeLayer(p);
        });
    } else if (markerLayers.some(l => l.id === id)) {
        if (gisState.markerGroup) {
            if (visible) gisState.markerGroup.addTo(gisMap); else gisMap.removeLayer(gisState.markerGroup);
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

    const primaryTypes = getPrimaryRegionTypes();
    const primaries = gisState.regions.filter(r => primaryTypes.includes(r.type));
    primaries.sort((a, b) => a.name.localeCompare(b.name));

    primaries.forEach(div => {
        const btn = document.createElement('button');
        btn.className = 'gis-division-btn';
        btn.textContent = div.name.length > 18 ? div.name.substring(0, 16) + '…' : div.name;
        btn.title = div.name;
        btn.onclick = () => flyToRegion(div);
        container.appendChild(btn);
    });

    // Update panel header
    const panelHeader = container.closest('.gis-panel');
    if (panelHeader) {
        const h3 = panelHeader.querySelector('h3');
        if (h3) h3.textContent = getRegionGridLabel();
    }
}

function getPrimaryRegionTypes() {
    switch (currentDashType) {
        case 'general': case 'agriculture': case 'industries': case 'medical':
            return ['division'];
        case 'satellite': return ['orbit_zone'];
        case 'airplane': return ['air_region'];
        case 'ship': return ['ocean_zone'];
        default: return ['division'];
    }
}

function getRegionGridLabel() {
    switch (currentDashType) {
        case 'general': case 'agriculture': case 'industries': case 'medical':
            return 'DIVISIONS';
        case 'satellite': return 'ORBIT ZONES';
        case 'airplane': return 'AIR REGIONS';
        case 'ship': return 'OCEAN ZONES';
        default: return 'REGIONS';
    }
}

function flyToRegion(region) {
    const scope = getDashScope();
    const zoomLevel = scope === 'domestic' ? 9 : 4;
    gisMap.flyTo(region.center, zoomLevel, { duration: 1 });
    showRegionProperties(region);
    if (currentDashType === 'general') loadDistrictsForDivision(region.id);
}

async function loadDistrictsForDivision(divisionId) {
    if (currentDashType !== 'general') return;
    try {
        const res = await fetch(GIS_API + '/regions?type=district&parent=' + encodeURIComponent(divisionId), {
            headers: buildAuthHeaders(false)
        });
        if (!res.ok) return;
        const districts = await res.json();
        if (districts.length > 0) {
            const tableRows = districts.map(d => ({
                name: d.name,
                population: d.properties?.population || 0,
            }));
            populateDataTable(
                [
                    { key: 'name', label: 'District', type: 'string' },
                    { key: 'population', label: 'Population', type: 'number' },
                ],
                tableRows,
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
    const primaryTypes = getPrimaryRegionTypes();
    const primaries = gisState.regions.filter(r => primaryTypes.includes(r.type));
    const secondaries = gisState.regions.filter(r => !primaryTypes.includes(r.type));

    const primaryLayerVisible = gisState.layers.length > 0 ? gisState.layers[0].visible : true;
    const secondaryLayerVisible = gisState.layers.length > 1 ? gisState.layers[1].visible : false;

    const scope = getDashScope();

    primaries.forEach(region => {
        const color = region.properties?.color || randomColor();
        const val = region.properties?.population || 0;
        const radius = scope === 'domestic'
            ? Math.max(25000, Math.sqrt(Math.max(val, 0) / 100) * 500)
            : Math.max(500000, Math.sqrt(Math.max(val, 0) / 1000) * 5000);

        const circle = L.circle(region.center, {
            radius: radius,
            color: darkenColor(color),
            weight: 2,
            fillColor: color,
            fillOpacity: 0.3,
        });

        circle.bindTooltip(region.name, {
            permanent: scope === 'domestic',
            direction: 'center',
            className: 'gis-division-label',
        });

        circle.on('click', () => {
            showRegionProperties(region);
            if (currentDashType === 'general') loadDistrictsForDivision(region.id);
        });
        circle.on('mouseover', function () { this.setStyle({ fillOpacity: 0.5, weight: 3 }); });
        circle.on('mouseout', function () { this.setStyle({ fillOpacity: 0.3, weight: 2 }); });

        if (primaryLayerVisible) circle.addTo(gisMap);
        gisState.divisionPolygons[region.id] = circle;
    });

    secondaries.forEach(region => {
        const val = region.properties?.population || 0;
        const radius = scope === 'domestic'
            ? Math.max(5000, Math.sqrt(Math.max(val, 0) / 100) * 250)
            : Math.max(200000, Math.sqrt(Math.max(val, 0) / 1000) * 2000);

        const circle = L.circle(region.center, {
            radius: radius,
            color: '#666',
            weight: 1,
            fillColor: getValueColor(val),
            fillOpacity: 0.4,
        });

        circle.bindTooltip(region.name, { direction: 'top' });
        circle.on('click', () => showRegionProperties(region));

        if (secondaryLayerVisible) circle.addTo(gisMap);
        gisState.districtPolygons[region.id] = circle;
    });

    renderLegend();
}

function getValueColor(val) {
    switch (currentDashType) {
        case 'agriculture':
            if (val > 6000000) return '#1b5e20';
            if (val > 4000000) return '#2e7d32';
            if (val > 3000000) return '#43a047';
            if (val > 2000000) return '#66bb6a';
            return '#a5d6a7';
        case 'industries':
            if (val > 200000) return '#b71c1c';
            if (val > 100000) return '#d32f2f';
            if (val > 50000) return '#e53935';
            if (val > 30000) return '#ef5350';
            return '#ef9a9a';
        case 'medical':
            if (val > 200) return '#1b5e20';
            if (val > 100) return '#388e3c';
            if (val > 50) return '#66bb6a';
            return '#a5d6a7';
        default:
            if (val > 10000000) return '#1a5276';
            if (val > 5000000) return '#2471a3';
            if (val > 3000000) return '#2e86c1';
            if (val > 2000000) return '#5dade2';
            if (val > 1000000) return '#85c1e9';
            return '#d6eaf8';
    }
}

function renderLegend() {
    const body = document.getElementById('legendBody');
    const titleEl = document.querySelector('.gis-legend h4');
    if (!body) return;
    body.innerHTML = '';

    const theme = DASH_THEMES[currentDashType] || DASH_THEMES.general;
    if (titleEl) titleEl.textContent = theme.legendTitle;

    const levels = getLegendLevels();
    levels.forEach(lvl => {
        const item = document.createElement('div');
        item.className = 'gis-legend-item';
        item.innerHTML = '<span class="legend-color" style="background:' + lvl.color + '"></span><span>' + lvl.label + '</span>';
        body.appendChild(item);
    });
}

function getLegendLevels() {
    switch (currentDashType) {
        case 'agriculture':
            return [
                { color: '#a5d6a7', label: '< 3M MT' },
                { color: '#66bb6a', label: '3M - 4M' },
                { color: '#43a047', label: '4M - 5M' },
                { color: '#2e7d32', label: '5M - 6M' },
                { color: '#1b5e20', label: '> 6M MT' },
            ];
        case 'industries':
            return [
                { color: '#ef9a9a', label: '< 30K Cr' },
                { color: '#ef5350', label: '30K - 50K' },
                { color: '#e53935', label: '50K - 100K' },
                { color: '#d32f2f', label: '100K - 200K' },
                { color: '#b71c1c', label: '> 200K Cr' },
            ];
        case 'medical':
            return [
                { color: '#a5d6a7', label: '< 50 Hospitals' },
                { color: '#66bb6a', label: '50 - 100' },
                { color: '#388e3c', label: '100 - 200' },
                { color: '#1b5e20', label: '> 200' },
            ];
        case 'satellite':
            return [
                { color: '#e3f2fd', label: 'LEO (160-2000km)' },
                { color: '#fff3e0', label: 'MEO (2000-35786km)' },
                { color: '#fce4ec', label: 'GEO (35786km)' },
                { color: '#e8eaf6', label: 'SSO (Polar)' },
            ];
        case 'airplane':
            return [
                { color: '#ffecb3', label: 'Asia-Pacific' },
                { color: '#bbdefb', label: 'Europe' },
                { color: '#c8e6c9', label: 'North America' },
                { color: '#f8bbd0', label: 'Middle East' },
                { color: '#d1c4e9', label: 'Africa' },
                { color: '#ffe0b2', label: 'Latin America' },
            ];
        case 'ship':
            return [
                { color: '#b3e5fc', label: 'Indian Ocean' },
                { color: '#c8e6c9', label: 'Pacific Ocean' },
                { color: '#dcedc8', label: 'Atlantic Ocean' },
                { color: '#fff9c4', label: 'Mediterranean' },
                { color: '#ffe0b2', label: 'South China Sea' },
                { color: '#e1bee7', label: 'Bay of Bengal' },
            ];
        default:
            return [
                { color: '#d6eaf8', label: '< 1M' },
                { color: '#85c1e9', label: '1M - 2M' },
                { color: '#5dade2', label: '2M - 3M' },
                { color: '#2e86c1', label: '3M - 5M' },
                { color: '#2471a3', label: '5M - 10M' },
                { color: '#1a5276', label: '> 10M' },
            ];
    }
}

// =============================================
// Map Markers
// =============================================

function renderMarkersOnMap() {
    if (gisState.markerGroup) gisMap.removeLayer(gisState.markerGroup);
    gisState.markerGroup = L.layerGroup();

    const emojis = MARKER_EMOJIS[currentDashType] || MARKER_EMOJIS.general;

    gisState.markers.forEach(m => {
        const color = m.color || '#3388ff';
        const emoji = emojis[m.category] || '📍';

        const icon = L.divIcon({
            html: '<div class="gis-marker-icon" style="background:' + color + '"><span>' + emoji + '</span></div>',
            className: 'gis-marker-container',
            iconSize: [32, 32],
            iconAnchor: [16, 32],
            popupAnchor: [0, -32],
        });

        const marker = L.marker([m.lat, m.lng], { icon: icon });
        marker.bindPopup('<strong>' + m.name + '</strong><br><em>' + m.category + '</em>');
        marker.on('click', () => showMarkerProperties(m));
        gisState.markerGroup.addLayer(marker);
    });

    // Check if any marker layer is visible
    const markerLayers = gisState.layers.filter(l => l.type === 'marker');
    const anyVisible = markerLayers.length === 0 || markerLayers.some(l => l.visible);
    if (anyVisible) gisState.markerGroup.addTo(gisMap);
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
        'Type': region.type.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()),
        'ID': region.id,
        'Center': region.center ? region.center[0].toFixed(4) + ', ' + region.center[1].toFixed(4) : '-',
    };

    if (region.properties) {
        Object.keys(region.properties).forEach(key => {
            if (key === 'color') return;
            const val = region.properties[key];
            const label = key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            props[label] = typeof val === 'number' ? val.toLocaleString() : val;
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
        'Category': marker.category.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()),
        'Latitude': marker.lat.toFixed(4),
        'Longitude': marker.lng.toFixed(4),
    };

    if (marker.properties) {
        Object.entries(marker.properties).forEach(([k, v]) => {
            const label = k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            props[label] = typeof v === 'number' ? v.toLocaleString() : v;
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
        let ds;
        if (currentDashType === 'general') {
            const res = await fetch(GIS_API + '/datasets/' + encodeURIComponent(id), {
                headers: buildAuthHeaders(false)
            });
            if (!res.ok) return;
            ds = await res.json();
        } else {
            // For specialized dashboards, datasets are already in state
            ds = gisState.datasets.find(d => d.id === id);
        }
        if (!ds) return;

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
    if (!cols || !data) return;
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
        data.length > 0 ? `${start + 1}-${Math.min(start + gisState.pageSize, data.length)} of ${data.length}` : '0 rows';

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
        const scope = getDashScope();
        gisMap.flyTo(region.center, scope === 'domestic' ? 10 : 4, { duration: 0.8 });
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

    const primaryTypes = getPrimaryRegionTypes();
    const primaries = gisState.regions.filter(r => primaryTypes.includes(r.type));
    primaries.sort((a, b) => a.name.localeCompare(b.name));

    const label = getRegionGridLabel().replace(/S$/, '');
    divSelect.innerHTML = '<option value="">All ' + label + 's</option>';
    primaries.forEach(d => {
        const opt = document.createElement('option');
        opt.value = d.id;
        opt.textContent = d.name;
        divSelect.appendChild(opt);
    });

    // Update region type dropdown
    const typeSelect = document.getElementById('filterRegionType');
    if (typeSelect) {
        const regionTypes = [...new Set(gisState.regions.map(r => r.type))];
        typeSelect.innerHTML = '<option value="">All Types</option>';
        regionTypes.forEach(t => {
            const opt = document.createElement('option');
            opt.value = t;
            opt.textContent = t.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            typeSelect.appendChild(opt);
        });
    }
}

function applyFilters() {
    const divId = document.getElementById('filterDivision').value;
    const regionType = document.getElementById('filterRegionType').value;

    if (divId) {
        const region = gisState.regions.find(r => r.id === divId);
        if (region) {
            const scope = getDashScope();
            gisMap.flyTo(region.center, scope === 'domestic' ? 9 : 4, { duration: 1 });
            showRegionProperties(region);
            if (currentDashType === 'general') loadDistrictsForDivision(divId);
        }
    }

    const primaryTypes = getPrimaryRegionTypes();
    if (!regionType || primaryTypes.includes(regionType)) {
        Object.values(gisState.divisionPolygons).forEach(p => p.addTo(gisMap));
    } else {
        Object.values(gisState.divisionPolygons).forEach(p => gisMap.removeLayer(p));
    }

    if (!regionType || !primaryTypes.includes(regionType)) {
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
    const scope = getDashScope();
    const cfg = MAP_CONFIG[scope];
    gisMap.flyTo(cfg.center, cfg.zoom, { duration: 1 });
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
