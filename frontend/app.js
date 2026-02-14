// API Base URL
const API_BASE = '/api';

// State
let networkServers = [];
let gateways = {};
let devices = {};
let map = null;
let markers = [];
let expandedServers = new Set();
let expandedGateways = new Set();
let expandedDevices = new Set();

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initMap();
    refreshData();
    // Removed auto-refresh - user can use the refresh button
});

// Initialize Leaflet map
function initMap() {
    // Chieti Scalo coordinates
    map = L.map('map').setView([42.3511, 14.1674], 13);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
    }).addTo(map);
}

// Fetch all data
async function refreshData() {
    const btn = document.getElementById('refresh-btn');
    btn.classList.add('spinning');
    
    try {
        const servers = await fetchAPI('/network-servers');
        networkServers = servers;
        
        // Fetch gateways and devices for each server
        for (const server of servers) {
            try {
                const [gws, devs] = await Promise.all([
                    fetchAPI(`/network-servers/${server.name}/gateways`),
                    fetchAPI(`/network-servers/${server.name}/devices`)
                ]);
                gateways[server.name] = gws;
                devices[server.name] = devs;
            } catch (err) {
                console.error(`Error fetching data for ${server.name}:`, err);
                gateways[server.name] = [];
                devices[server.name] = [];
            }
        }
        
        renderServers();
        renderGateways();
    } catch (err) {
        console.error('Error refreshing data:', err);
    } finally {
        btn.classList.remove('spinning');
    }
}

// API Helper
async function fetchAPI(endpoint, options = {}) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        }
    });
    
    if (response.status === 204) {
        return null;
    }
    
    if (!response.ok) {
        const error = await response.json().catch(() => ({ message: 'Request failed' }));
        throw new Error(error.message || 'Request failed');
    }
    
    return response.json();
}

// Render network servers
function renderServers() {
    const container = document.getElementById('servers-list');
    
    if (networkServers.length === 0) {
        container.innerHTML = '<div class="empty-state">No network servers. Click + to add one.</div>';
        return;
    }
    
    container.innerHTML = networkServers.map(server => {
        const isExpanded = expandedServers.has(server.name);
        const serverGateways = gateways[server.name] || [];
        const serverDevices = devices[server.name] || [];
        
        return `
            <div class="server-item">
                <div class="server-header" onclick="toggleServer('${server.name}')">
                    <span class="chevron ${isExpanded ? 'expanded' : ''}">▶</span>
                    <span style="flex: 1;">${server.name}</span>
                    <span class="server-type-badge">${server.config?.type || 'generic'}</span>
                    <button class="icon-button" onclick="event.stopPropagation(); deleteNetworkServer('${server.name}')" title="Delete Server">×</button>
                </div>
                <div class="server-content ${isExpanded ? 'expanded' : ''}">
                    ${server.config?.url ? `
                        <div class="server-config-info">
                            <span>URL: ${server.config.url}</span>
                            <button class="icon-button icon-button-small" onclick="syncNetworkServer('${server.name}')" title="Sync with Network Server">↻</button>
                        </div>
                    ` : ''}
                    <div class="subsection">
                        <div class="subsection-header" onclick="toggleGateways('${server.name}')">
                            <span class="chevron ${expandedGateways.has(server.name) ? 'expanded' : ''}">▶</span>
                            <span style="flex: 1;">Gateways</span>
                            <span class="count-badge">${serverGateways.length}</span>
                            <button class="icon-button" onclick="event.stopPropagation(); showAddGatewayModal('${server.name}')" title="Add Gateway">+</button>
                        </div>
                        <div class="item-list ${expandedGateways.has(server.name) ? 'expanded' : ''}" id="gateways-${server.name}">
                            ${renderGatewayList(server.name, serverGateways)}
                        </div>
                    </div>
                    <div class="subsection">
                        <div class="subsection-header" onclick="toggleDevices('${server.name}')">
                            <span class="chevron ${expandedDevices.has(server.name) ? 'expanded' : ''}">▶</span>
                            <span style="flex: 1;">Devices</span>
                            <span class="count-badge">${serverDevices.length}</span>
                            <button class="icon-button" onclick="event.stopPropagation(); showAddDeviceModal('${server.name}')" title="Add Device">+</button>
                        </div>
                        <div class="item-list ${expandedDevices.has(server.name) ? 'expanded' : ''}" id="devices-${server.name}">
                            ${renderDeviceList(server.name, serverDevices)}
                        </div>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// Render gateway list
function renderGatewayList(serverName, gateways) {
    if (gateways.length === 0) {
        return '<div class="empty-state">No gateways</div>';
    }
    
    return gateways.map(gw => {
        const isConnected = gw.dataState === "connected";
        const isDisconnected = gw.discoveryState === "disconnected" && gw.dataState === "disconnected";
        
        // Determine status light color: green (connected), red (disconnected), yellow (other states)
        let statusClass = 'pending'; // yellow by default
        if (isConnected) {
            statusClass = 'connected'; // green
        } else if (isDisconnected) {
            statusClass = 'disconnected'; // red
        }
        
        return `
            <div class="gateway-item">
                <div class="gateway-info">
                    <div class="status-light ${statusClass}"></div>
                    <span class="eui">${gw.eui}</span>
                    <button class="icon-button" onclick="deleteGateway('${serverName}', '${gw.eui}')" title="Delete Gateway">×</button>
                </div>
                <div class="button-group">
                    <button 
                        class="btn ${isConnected ? 'btn-danger' : 'btn-success'}"
                        onclick="${isConnected ? 'disconnectGateway' : 'connectGateway'}('${serverName}', '${gw.eui}')"
                    >
                        ${isConnected ? 'Disconnect' : 'Connect'}
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// Check if device has session keys (all must be non-zero)
function hasSessionKeys(dev) {
    // Helper to check if a value is zero (works for both arrays and hex strings)
    const isZero = (val) => {
        if (!val) return true;
        
        // If it's an array
        if (Array.isArray(val)) {
            return val.every(byte => byte === 0);
        }
        
        // If it's a string (hex)
        if (typeof val === 'string') {
            // Remove any whitespace and check if it's all zeros
            const cleaned = val.replace(/\s/g, '');
            return cleaned === '' || /^0+$/.test(cleaned);
        }
        
        return true;
    };
    
    return !isZero(dev.devaddr) && 
           !isZero(dev.appskey) && 
           !isZero(dev.nwkskey);
}

// Check if device can join (appkey and joineui must be non-zero)
function canJoin(dev) {
    const isZero = (val) => {
        if (!val) return true;
        
        // If it's an array
        if (Array.isArray(val)) {
            return val.every(byte => byte === 0);
        }
        
        // If it's a string (hex)
        if (typeof val === 'string') {
            const cleaned = val.replace(/\s/g, '');
            return cleaned === '' || /^0+$/.test(cleaned);
        }
        
        return true;
    };
    
    return !isZero(dev.appkey) && !isZero(dev.joineui);
}

// Render device list
function renderDeviceList(serverName, devices) {
    if (devices.length === 0) {
        return '<div class="empty-state">No devices</div>';
    }
    
    return devices.map(dev => {
        const canUplink = hasSessionKeys(dev);
        const canDoJoin = canJoin(dev);
        
        return `
            <div class="device-item">
                <div class="device-info">
                    <span class="eui">${dev.deveui}</span>
                    <div class="device-stats">
                        <span>↑ ${dev.fcntup}</span>
                        <span>↓ ${dev.fcntdn}</span>
                    </div>
                    <button class="icon-button" onclick="deleteDevice('${serverName}', '${dev.deveui}')" title="Delete Device">×</button>
                </div>
                <div class="button-group">
                    <button class="btn btn-primary" 
                            onclick="sendJoin('${serverName}', '${dev.deveui}')"
                            ${!canDoJoin ? 'disabled title="ABP device"' : ''}>
                        Join
                    </button>
                    <button class="btn btn-primary" 
                            onclick="sendUplink('${serverName}', '${dev.deveui}')"
                            ${!canUplink ? 'disabled title="Device must join first"' : ''}>
                        Uplink
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// Render gateways on map
function renderGateways() {
    // Clear existing markers
    markers.forEach(marker => map.removeLayer(marker));
    markers = [];
    
    // Note: Map markers are disabled. Location feature will be added later.
}

// Toggle functions
function toggleServer(name) {
    if (expandedServers.has(name)) {
        expandedServers.delete(name);
    } else {
        expandedServers.add(name);
    }
    renderServers();
}

function toggleGateways(name) {
    if (expandedGateways.has(name)) {
        expandedGateways.delete(name);
    } else {
        expandedGateways.add(name);
    }
    renderServers();
}

function toggleDevices(name) {
    if (expandedDevices.has(name)) {
        expandedDevices.delete(name);
    } else {
        expandedDevices.add(name);
    }
    renderServers();
}

// Modal functions
function showAddServerModal() {
    const modal = document.getElementById('modal');
    const modalTitle = modal.querySelector('.modal-header h2');
    const modalBody = modal.querySelector('.modal-body');
    
    // Clear the modal body first to ensure fresh content
    modalBody.innerHTML = '';
    modalTitle.textContent = 'Add Network Server';
    
    modalBody.innerHTML = `
        <label>Network Server Type</label>
        <select id="server-type" onchange="updateServerTypeFields()">
            <option value="generic">Generic</option>
            <option value="loriot">LORIOT</option>
            <option value="chirpstack">ChirpStack</option>
            <option value="ttn">The Things Network (TTN)</option>
        </select>

        <label>Server Name</label>
        <input type="text" id="server-name" placeholder="my-server">
        
        <div id="server-type-fields"></div>
        
        <div class="modal-actions">
            <button onclick="closeModal()" class="btn btn-secondary">Cancel</button>
            <button onclick="createServer()" class="btn btn-primary">Create</button>
        </div>
    `;

    modal.classList.add('show');
    setTimeout(() => {
        const input = document.getElementById('server-name');
        input?.focus();
    }, 0);
}

function updateServerTypeFields() {
    const serverType = document.getElementById('server-type').value;
    const fieldsContainer = document.getElementById('server-type-fields');
    
    let fieldsHTML = '';
    
    if (serverType === 'loriot') {
        fieldsHTML = `
            <label>URL</label>
            <input type="text" id="server-url" placeholder="https://eu1.loriot.io">
            <label>Authorization Header</label>
            <input type="text" id="server-auth" placeholder="Bearer your-token-here">
        `;
    } else if (serverType === 'chirpstack') {
        fieldsHTML = `
            <label>URL</label>
            <input type="text" id="server-url" placeholder="https://chirpstack.example.com">
            <label>API Key</label>
            <input type="text" id="server-apikey" placeholder="your-apikey">
        `;
    } else if (serverType === 'ttn') {
        fieldsHTML = `
            <label>URL</label>
            <input type="text" id="server-url" placeholder="https://eu1.cloud.thethings.network">
            <label>API Key</label>
            <input type="text" id="server-apikey" placeholder="your-api-key">
        `;
    }
    
    fieldsContainer.innerHTML = fieldsHTML;
}

function closeModal() {
    document.getElementById('modal').classList.remove('show');
}

// API Actions
async function createServer() {
    const name = document.getElementById('server-name').value.trim();
    const serverType = document.getElementById('server-type').value;
    
    if (!name) {
        alert('Please enter a server name');
        return;
    }
    
    const config = {
        type: serverType
    };
    
    // Add type-specific fields
    if (serverType === 'loriot') {
        const url = document.getElementById('server-url')?.value.trim();
        const authHeader = document.getElementById('server-auth')?.value.trim();
        
        if (!url || !authHeader) {
            alert('Please fill in all LORIOT fields');
            return;
        }
        
        config.url = url;
        config.authHeader = authHeader;
    } else if (serverType === 'chirpstack') {
        const url = document.getElementById('server-url')?.value.trim();
        const apiKey = document.getElementById('server-apikey')?.value.trim();
        
        if (!url || !apiKey) {
            alert('Please fill in all ChirpStack fields');
            return;
        }
        
        config.url = url;
        config.apiKey = apiKey;
    } else if (serverType === 'ttn') {
        const url = document.getElementById('server-url')?.value.trim();
        const apiKey = document.getElementById('server-apikey')?.value.trim();
        
        if (!url || !apiKey) {
            alert('Please fill in all TTN fields');
            return;
        }
        
        config.url = url;
        config.apiKey = apiKey;
    }
    
    try {
        await fetchAPI('/network-servers', {
            method: 'POST',
            body: JSON.stringify({ name, config })
        });
        closeModal();
        await refreshData();
    } catch (err) {
        alert('Error creating server: ' + err.message);
    }
}

async function syncNetworkServer(serverName) {
    try {
        await fetchAPI(`/network-servers/${serverName}/sync`, {
            method: 'POST'
        });
        await refreshData();
        alert(`Successfully synced ${serverName}`);
    } catch (err) {
        alert('Error syncing network server: ' + err.message);
    }
}

async function connectGateway(serverName, eui) {
    try {
        await fetchAPI(`/network-servers/${serverName}/gateways/${eui}/connect`, {
            method: 'POST'
        });
        await refreshData();
    } catch (err) {
        alert('Error connecting gateway: ' + err.message);
    }
}

async function disconnectGateway(serverName, eui) {
    try {
        await fetchAPI(`/network-servers/${serverName}/gateways/${eui}/disconnect`, {
            method: 'POST'
        });
        await refreshData();
    } catch (err) {
        alert('Error disconnecting gateway: ' + err.message);
    }
}

async function sendJoin(serverName, eui) {
    try {
        await fetchAPI(`/network-servers/${serverName}/devices/${eui}/join`, {
            method: 'POST'
        });
        await refreshData();
    } catch (err) {
        alert('Error sending join: ' + err.message);
    }
}

async function sendUplink(serverName, eui) {
    try {
        await fetchAPI(`/network-servers/${serverName}/devices/${eui}/uplink`, {
            method: 'POST'
        });
        await refreshData();
    } catch (err) {
        alert('Error sending uplink: ' + err.message);
    }
}

// Modal functions for adding gateway and device
function showAddGatewayModal(serverName) {
    console.log('showAddGatewayModal called with serverName:', serverName);
    const modal = document.getElementById('modal');
    const modalTitle = modal.querySelector('.modal-header h2');
    const modalBody = modal.querySelector('.modal-body');
    
    // Clear the modal body first to ensure fresh content
    modalBody.innerHTML = '';
    modalTitle.textContent = 'Add Gateway';
    
    modalBody.innerHTML = `
        <label>Gateway EUI (16 hex characters)</label>
        <input type="text" id="gateway-eui" placeholder="AABBCCDDEEFF0011" maxlength="16">
        <label>Discovery URI</label>
        <input type="text" id="discovery-uri" placeholder="ws://localhost:3001" onkeypress="if(event.key==='Enter')createGateway('${serverName}')">
        <div class="modal-actions">
            <button onclick="closeModal()" class="btn btn-secondary">Cancel</button>
            <button onclick="createGateway('${serverName}')" class="btn btn-primary">Create</button>
        </div>
    `;
    
    console.log('Gateway modal title:', modalTitle.textContent);
    
    modal.classList.add('show');
    setTimeout(() => {
        document.getElementById('gateway-eui')?.focus();
    }, 0);
}

function showAddDeviceModal(serverName) {
    const modal = document.getElementById('modal');
    const modalTitle = modal.querySelector('.modal-header h2');
    const modalBody = modal.querySelector('.modal-body');
    
    // Clear the modal body first to ensure fresh content
    modalBody.innerHTML = '';
    modalTitle.textContent = 'Add Device';
    
    modalBody.innerHTML = `
        <label>Device EUI (16 hex characters)</label>
        <input type="text" id="device-eui" placeholder="0011223344556677" maxlength="16">
        <label>Join EUI (16 hex characters)</label>
        <input type="text" id="join-eui" placeholder="0011223344556677" maxlength="16">
        <label>App Key (32 hex characters)</label>
        <input type="text" id="app-key" placeholder="00112233445566770011223344556677" maxlength="32">
        <label>Dev Nonce</label>
        <input type="number" id="dev-nonce" placeholder="0" value="0" onkeypress="if(event.key==='Enter')createDevice('${serverName}')">
        <div class="modal-actions">
            <button onclick="closeModal()" class="btn btn-secondary">Cancel</button>
            <button onclick="createDevice('${serverName}')" class="btn btn-primary">Create</button>
        </div>
    `;
    
    modal.classList.add('show');
    setTimeout(() => {
        document.getElementById('device-eui')?.focus();
    }, 0);
}

async function createGateway(serverName) {
    const eui = document.getElementById('gateway-eui').value.trim();
    const discoveryUri = document.getElementById('discovery-uri').value.trim();
    
    if (!eui || !discoveryUri) {
        alert('Please fill in all fields');
        return;
    }
    
    try {
        await fetchAPI(`/network-servers/${serverName}/gateways`, {
            method: 'POST',
            body: JSON.stringify({ eui, discoveryUri })
        });
        closeModal();
        await refreshData();
    } catch (err) {
        alert('Error creating gateway: ' + err.message);
    }
}

async function createDevice(serverName) {
    const deveui = document.getElementById('device-eui').value.trim();
    const joineui = document.getElementById('join-eui').value.trim();
    const appkey = document.getElementById('app-key').value.trim();
    const devnonce = parseInt(document.getElementById('dev-nonce').value) || 0;
    
    if (!deveui || !joineui || !appkey) {
        alert('Please fill in all fields');
        return;
    }
    
    try {
        await fetchAPI(`/network-servers/${serverName}/devices`, {
            method: 'POST',
            body: JSON.stringify({ deveui, joineui, appkey, devnonce })
        });
        closeModal();
        await refreshData();
    } catch (err) {
        alert('Error creating device: ' + err.message);
    }
}

// Delete functions
async function deleteNetworkServer(serverName) {
    if (!confirm(`Are you sure you want to delete network server "${serverName}"?`)) {
        return;
    }
    
    try {
        await fetchAPI(`/network-servers/${serverName}`, {
            method: 'DELETE'
        });
        await refreshData();
    } catch (err) {
        alert('Error deleting network server: ' + err.message);
    }
}

async function deleteGateway(serverName, eui) {
    if (!confirm(`Are you sure you want to delete gateway "${eui}"?`)) {
        return;
    }
    
    try {
        await fetchAPI(`/network-servers/${serverName}/gateways/${eui}`, {
            method: 'DELETE'
        });
        await refreshData();
    } catch (err) {
        alert('Error deleting gateway: ' + err.message);
    }
}

async function deleteDevice(serverName, eui) {
    if (!confirm(`Are you sure you want to delete device "${eui}"?`)) {
        return;
    }
    
    try {
        await fetchAPI(`/network-servers/${serverName}/devices/${eui}`, {
            method: 'DELETE'
        });
        await refreshData();
    } catch (err) {
        alert('Error deleting device: ' + err.message);
    }
}

// Close modal on ESC key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        closeModal();
    }
});

// Close modal on outside click
document.getElementById('modal').addEventListener('click', (e) => {
    if (e.target.id === 'modal') {
        closeModal();
    }
});
