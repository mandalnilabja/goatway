// API Key modal forms
// Extends: Modals (must be loaded after modals.js)
// Depends on: API, Pages

Modals.showAPIKeyForm = async function(editId = null) {
    let apiKey = { name: '', scopes: ['proxy'], rate_limit: 0 };

    if (editId) {
        try {
            apiKey = await API.getAPIKey(editId);
        } catch (err) {
            alert('Error loading API key: ' + (err?.message || String(err)));
            return;
        }
    }

    const hasProxyScope = apiKey.scopes?.includes('proxy') ?? true;
    const hasAdminScope = apiKey.scopes?.includes('admin') ?? false;

    this.show(`
        <div class="modal">
            <div class="modal-header">
                <h3>${editId ? 'Edit' : 'Create'} API Key</h3>
                <button onclick="Modals.close()">&times;</button>
            </div>
            <div class="modal-body">
                <form id="apikey-form">
                    <div class="form-group">
                        <label>Name</label>
                        <input type="text" name="name" value="${apiKey.name}" required placeholder="My Application">
                    </div>
                    <div class="form-group">
                        <label>Scopes</label>
                        <div>
                            <label class="checkbox-label"><input type="checkbox" name="scope_proxy" ${hasProxyScope ? 'checked' : ''}> Proxy (access LLM endpoints)</label>
                            <label class="checkbox-label"><input type="checkbox" name="scope_admin" ${hasAdminScope ? 'checked' : ''}> Admin (manage settings)</label>
                        </div>
                    </div>
                    <div class="form-group">
                        <label>Rate Limit (requests/min, 0 = unlimited)</label>
                        <input type="number" name="rate_limit" value="${apiKey.rate_limit || 0}" min="0">
                    </div>
                    ${!editId ? `
                    <div class="form-group">
                        <label>Expires In (seconds, empty = never)</label>
                        <input type="number" name="expires_in" min="0" placeholder="e.g., 86400 for 1 day">
                    </div>
                    ` : ''}
                    ${editId ? `
                    <div class="form-group">
                        <label class="checkbox-label"><input type="checkbox" name="is_active" ${apiKey.is_active ? 'checked' : ''}> Active</label>
                    </div>
                    ` : ''}
                    <div class="modal-footer">
                        <button type="button" onclick="Modals.close()">Cancel</button>
                        <button type="submit" class="btn-primary">${editId ? 'Update' : 'Create'}</button>
                    </div>
                </form>
            </div>
        </div>
    `);

    document.getElementById('apikey-form').onsubmit = async (e) => {
        e.preventDefault();
        const form = e.target;

        const scopes = [];
        if (form.scope_proxy.checked) scopes.push('proxy');
        if (form.scope_admin.checked) scopes.push('admin');

        if (scopes.length === 0) {
            alert('Please select at least one scope');
            return;
        }

        const data = {
            name: form.name.value,
            scopes: scopes,
            rate_limit: parseInt(form.rate_limit.value) || 0
        };

        if (!editId && form.expires_in?.value) {
            data.expires_in = parseInt(form.expires_in.value);
        }

        if (editId && form.is_active) {
            data.is_active = form.is_active.checked;
        }

        try {
            if (editId) {
                await API.updateAPIKey(editId, data);
                this.close();
            } else {
                const result = await API.createAPIKey(data);
                this.close();
                this.showAPIKeyCreated(result.key);
            }
            Pages.apikeys();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    };
};

Modals.showAPIKeyCreated = function(key) {
    this.show(`
        <div class="modal">
            <div class="modal-header">
                <h3>API Key Created</h3>
            </div>
            <div class="modal-body">
                <p><strong>Important:</strong> Copy your API key now. You won't be able to see it again!</p>
                <div class="form-group">
                    <input type="text" id="new-key-input" value="${key}" readonly style="font-family: monospace;">
                </div>
                <div class="modal-footer">
                    <button type="button" onclick="navigator.clipboard.writeText('${key}'); alert('Copied!')">Copy to Clipboard</button>
                    <button type="button" class="btn-primary" onclick="Modals.close()">Done</button>
                </div>
            </div>
        </div>
    `);
};
