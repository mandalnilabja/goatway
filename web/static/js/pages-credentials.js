// Credentials page
// Extends: Pages (must be loaded after pages-dashboard.js)
// Depends on: API, Utils, Modals, Actions

Pages.credentials = async function() {
    const app = document.getElementById('app');
    app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

    try {
        const data = await API.listCredentials();
        const creds = data.credentials || [];

        app.innerHTML = `
            <div class="page-header">
                <h2>API Credentials</h2>
                <button class="btn-primary" onclick="Modals.showCredentialForm()">+ Add New</button>
            </div>

            <div id="credentials-list" class="item-list">
                ${creds.length ? creds.map(c => this.renderCredentialCard(c)).join('') :
                  '<div class="empty-state"><p>No credentials configured yet.</p></div>'}
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<div class="error">Error loading credentials: ${err?.message || err}</div>`;
    }
};

Pages.renderCredentialCard = function(cred) {
    return `
        <div class="item-card" data-id="${cred.id}">
            <div class="item-header">
                <div class="item-title">${cred.name}</div>
                <div class="btn-group">
                    <button class="btn-sm" onclick="Modals.showCredentialForm('${cred.id}')">Edit</button>
                    <button class="btn-sm btn-danger" onclick="Actions.deleteCredential('${cred.id}')">Delete</button>
                </div>
            </div>
            <div class="item-meta">
                Provider: ${cred.provider} &middot; Key: ${cred.api_key_preview || '***'} &middot; Created: ${Utils.formatDate(cred.created_at)}
            </div>
        </div>
    `;
};
