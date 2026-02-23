// Credential modal forms
// Extends: Modals (must be loaded after modals.js)
// Depends on: API, Pages

Modals.updateCredentialFields = function(provider, isEdit) {
    const azureFields = document.getElementById('azure-fields');
    const apiKeyField = document.getElementById('apikey-field');
    const isAzure = provider === 'azure';

    azureFields.style.display = isAzure ? 'block' : 'none';
    apiKeyField.querySelector('input').placeholder = isEdit
        ? 'Leave blank to keep current'
        : (isAzure ? 'Azure API key' : 'sk-...');

    // Update required attributes
    document.querySelector('[name="endpoint"]').required = isAzure && !isEdit;
    document.querySelector('[name="deployment"]').required = isAzure && !isEdit;
};

Modals.buildCredentialData = function(form, isEdit) {
    const provider = form.provider.value;
    const payload = {
        name: form.name.value,
        provider: provider
    };

    if (provider === 'azure') {
        const credData = {};
        if (form.api_key.value) credData.api_key = form.api_key.value;
        if (form.endpoint.value) credData.endpoint = form.endpoint.value;
        if (form.deployment.value) credData.deployment = form.deployment.value;
        if (form.api_version.value) credData.api_version = form.api_version.value;
        if (Object.keys(credData).length > 0 || !isEdit) payload.data = credData;
    } else {
        if (form.api_key.value || !isEdit) {
            payload.data = { api_key: form.api_key.value };
        }
    }
    return payload;
};

Modals.showCredentialForm = async function(editId = null) {
    let credential = { provider: 'openrouter', name: '' };

    if (editId) {
        try {
            credential = await API.getCredential(editId);
        } catch (err) {
            alert('Error loading credential: ' + (err?.message || String(err)));
            return;
        }
    }

    const isAzure = credential.provider === 'azure';
    this.show(`
        <div class="modal">
            <div class="modal-header">
                <h3>${editId ? 'Edit' : 'Add'} Credential</h3>
                <button onclick="Modals.close()">&times;</button>
            </div>
            <div class="modal-body">
                <form id="credential-form">
                    <div class="form-group">
                        <label>Name</label>
                        <input type="text" name="name" value="${credential.name}" required placeholder="My API Key">
                    </div>
                    <div class="form-group">
                        <label>Provider</label>
                        <select name="provider" required>
                            <option value="openrouter" ${credential.provider === 'openrouter' ? 'selected' : ''}>OpenRouter</option>
                            <option value="openai" ${credential.provider === 'openai' ? 'selected' : ''}>OpenAI</option>
                            <option value="anthropic" ${credential.provider === 'anthropic' ? 'selected' : ''}>Anthropic</option>
                            <option value="azure" ${credential.provider === 'azure' ? 'selected' : ''}>Azure OpenAI</option>
                        </select>
                    </div>
                    <div id="azure-fields" style="display: ${isAzure ? 'block' : 'none'}">
                        <div class="form-group">
                            <label>Endpoint</label>
                            <input type="text" name="endpoint" ${isAzure && !editId ? 'required' : ''} placeholder="https://your-resource.openai.azure.com">
                        </div>
                        <div class="form-group">
                            <label>Deployment</label>
                            <input type="text" name="deployment" ${isAzure && !editId ? 'required' : ''} placeholder="gpt-4">
                        </div>
                        <div class="form-group">
                            <label>API Version</label>
                            <input type="text" name="api_version" placeholder="2024-02-15-preview">
                        </div>
                    </div>
                    <div class="form-group" id="apikey-field">
                        <label>API Key</label>
                        <input type="password" name="api_key" ${editId ? '' : 'required'} placeholder="${editId ? 'Leave blank to keep current' : 'sk-...'}">
                    </div>
                    <div class="modal-footer">
                        <button type="button" onclick="Modals.close()">Cancel</button>
                        <button type="submit" class="btn-primary">${editId ? 'Update' : 'Create'}</button>
                    </div>
                </form>
            </div>
        </div>
    `);

    const form = document.getElementById('credential-form');
    form.provider.addEventListener('change', (e) => this.updateCredentialFields(e.target.value, !!editId));

    form.onsubmit = async (e) => {
        e.preventDefault();
        const data = this.buildCredentialData(form, !!editId);

        try {
            if (editId) {
                await API.updateCredential(editId, data);
            } else {
                await API.createCredential(data);
            }
            this.close();
            Pages.credentials();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    };
};
