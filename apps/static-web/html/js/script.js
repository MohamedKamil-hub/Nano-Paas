// ============================================================
// NEBULA - Real-time System Metrics via Netdata API
// ============================================================

// Netdata runs on port 19999 of the same host
var NETDATA_PORT = 19999;
var NETDATA_BASE = window.location.protocol + '//' + window.location.hostname + ':' + NETDATA_PORT;

// --- Real metrics from Netdata ---

async function updateCPU() {
    var el = document.getElementById('cpu');
    try {
        var res = await fetch(
            NETDATA_BASE + '/api/v1/data?chart=system.cpu&after=-1&points=1&format=json&options=absolute',
            { signal: AbortSignal.timeout(3000) }
        );
        var data = await res.json();
        // data.data[0] = [timestamp, user, system, softirq, irq, ...]
        // Sum all CPU dimensions (skip the timestamp at index 0)
        var total = data.data[0].slice(1).reduce(function(a, b) { return a + b; }, 0);
        el.textContent = total.toFixed(1);
        el.classList.remove('metric-error');
    } catch (e) {
        el.textContent = '\u2014';
        el.title = 'Netdata no accesible en ' + NETDATA_BASE;
        el.classList.add('metric-error');
    }
}

async function updateRAM() {
    var el = document.getElementById('ram');
    try {
        var res = await fetch(
            NETDATA_BASE + '/api/v1/data?chart=system.ram&after=-1&points=1&format=json',
            { signal: AbortSignal.timeout(3000) }
        );
        var data = await res.json();
        // Find 'used' column index from labels
        var usedIdx = data.labels.indexOf('used');
        if (usedIdx < 0) usedIdx = 1;
        var usedMB = Math.round(data.data[0][usedIdx]);
        el.textContent = usedMB;
        el.classList.remove('metric-error');
    } catch (e) {
        el.textContent = '\u2014';
        el.title = 'Netdata no accesible en ' + NETDATA_BASE;
        el.classList.add('metric-error');
    }
}

async function updateContainers() {
    var el = document.getElementById('containers');
    try {
        var res = await fetch(
            NETDATA_BASE + '/api/v1/data?chart=docker_engine.engine_daemon_container_states_containers&after=-1&points=1&format=json',
            { signal: AbortSignal.timeout(3000) }
        );
        var data = await res.json();
        var runIdx = data.labels.indexOf('running');
        if (runIdx > 0) {
            el.textContent = Math.round(data.data[0][runIdx]);
        }
    } catch (e) {
        // Not critical - keep the default value from HTML
    }
}

// --- Time & SSL (unchanged) ---

function updateTime() {
    var now = new Date();
    document.getElementById('time').textContent = now.toLocaleTimeString('es-ES', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

function checkSSL() {
    var el = document.getElementById('ssl');
    if (window.location.protocol === 'https:') {
        el.textContent = 'Activo';
        el.style.color = '#2d7a3e';
    } else {
        el.textContent = 'Inactivo';
        el.style.color = '#c53030';
    }
}

// --- Update all metrics in parallel ---

async function updateMetrics() {
    await Promise.allSettled([updateCPU(), updateRAM(), updateContainers()]);
}

// --- Set dynamic Netdata link ---

function setupNetdataLink() {
    var link = document.getElementById('netdata-link');
    if (link) {
        link.href = NETDATA_BASE;
    }
}

// --- Init ---

document.addEventListener('DOMContentLoaded', function() {
    updateTime();
    checkSSL();
    setupNetdataLink();
    updateMetrics();

    setInterval(updateTime, 1000);
    setInterval(updateMetrics, 5000);
});
