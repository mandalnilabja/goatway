// Settings and Logs pages
// Extends: Pages (must be loaded after pages-dashboard.js)
// Depends on: API, Utils, Actions, Router

Pages.logs = async function(params = {}) {
    const app = document.getElementById('app');
    const limit = params.limit || 50;
    const offset = params.offset || 0;

    app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

    try {
        const data = await API.getRequestLogs({ limit, offset, ...params });

        app.innerHTML = `
            <div class="page-header"><h2>Request Logs</h2></div>

            <div class="card">
                <div class="table-container">
                    ${this.renderLogsTable(data.logs || [], true)}
                </div>
                ${this.renderPagination(data.total, limit, offset)}
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<div class="error">Error loading logs: ${err?.message || err}</div>`;
    }
};

Pages.settings = async function() {
    const app = document.getElementById('app');
    app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

    try {
        const info = await API.getInfo();

        app.innerHTML = `
            <div class="page-header"><h2>Settings</h2></div>

            <div class="section card">
                <h3>System Information</h3>
                <table class="info-table">
                    <tr><td><strong>Version</strong></td><td>${info.version || 'dev'}</td></tr>
                    <tr><td><strong>Uptime</strong></td><td>${info.uptime || 'N/A'}</td></tr>
                    <tr><td><strong>Data Directory</strong></td><td>${info.data_dir || 'N/A'}</td></tr>
                </table>
            </div>

            <div class="section danger-zone">
                <h3>Danger Zone</h3>
                <p>Clear old request logs to free up disk space.</p>
                <button class="btn-danger" onclick="Actions.clearOldLogs()">Clear Logs Older Than 30 Days</button>
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<div class="error">Error loading settings: ${err?.message || err}</div>`;
    }
};
