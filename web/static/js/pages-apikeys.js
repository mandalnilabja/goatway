// API Keys page
// Extends: Pages (must be loaded after pages-dashboard.js)
// Depends on: API, Utils, Modals, Actions

Pages.apikeys = async function() {
    const app = document.getElementById('app');
    app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

    try {
        const data = await API.listAPIKeys();
        const keys = data.data || [];

        app.innerHTML = `
            <div class="page-header">
                <h2>API Keys</h2>
                <button class="btn-primary" onclick="Modals.showAPIKeyForm()">+ Create Key</button>
            </div>

            <p>API keys allow applications to access the proxy endpoints. Use these keys with the OpenAI SDK.</p>

            <div id="apikeys-list" class="item-list">
                ${keys.length ? keys.map(k => this.renderAPIKeyCard(k)).join('') :
                  '<div class="empty-state"><p>No API keys created yet.</p></div>'}
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<div class="error">Error loading API keys: ${err?.message || err}</div>`;
    }
};

Pages.renderAPIKeyCard = function(key) {
    const statusBadge = key.is_active
        ? '<span class="badge badge-success">Active</span>'
        : '<span class="badge badge-muted">Inactive</span>';
    const scopes = (key.scopes || []).join(', ');
    const rateLimit = key.rate_limit ? `${key.rate_limit}/min` : 'Unlimited';
    const expiresAt = key.expires_at ? Utils.formatDate(key.expires_at) : 'Never';

    return `
        <div class="item-card" data-id="${key.id}">
            <div class="item-header">
                <div class="item-title">
                    ${statusBadge}
                    ${key.name}
                </div>
                <div class="btn-group">
                    <button class="btn-sm" onclick="Actions.toggleAPIKey('${key.id}', ${key.is_active})">${key.is_active ? 'Disable' : 'Enable'}</button>
                    <button class="btn-sm" onclick="Actions.rotateAPIKey('${key.id}')">Rotate</button>
                    <button class="btn-sm" onclick="Modals.showAPIKeyForm('${key.id}')">Edit</button>
                    <button class="btn-sm btn-danger" onclick="Actions.deleteAPIKey('${key.id}')">Delete</button>
                </div>
            </div>
            <div class="item-meta">
                Key: ${key.key_prefix}... &middot; Scopes: ${scopes} &middot; Rate: ${rateLimit} &middot; Expires: ${expiresAt} &middot; Created: ${Utils.formatDate(key.created_at)}
            </div>
        </div>
    `;
};
