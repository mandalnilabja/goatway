// Modal management
// Depends on: API, Pages (for refresh after form submit)

const Modals = {
    show(content) {
        const overlay = document.createElement('div');
        overlay.id = 'modal-overlay';
        overlay.innerHTML = content;
        overlay.addEventListener('click', e => {
            if (e.target === overlay) this.close();
        });
        document.body.appendChild(overlay);
    },

    close() {
        document.getElementById('modal-overlay')?.remove();
    },

    async showCredentialForm(editId = null) {
        let credential = { provider: 'openrouter', name: '', api_key: '', is_default: false };

        if (editId) {
            try {
                credential = await API.getCredential(editId);
            } catch (err) {
                alert('Error loading credential: ' + (err?.message || String(err)));
                return;
            }
        }

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
                            </select>
                        </div>
                        <div class="form-group">
                            <label>API Key</label>
                            <input type="password" name="api_key" ${editId ? '' : 'required'} placeholder="${editId ? 'Leave blank to keep current' : 'sk-...'}">
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label"><input type="checkbox" name="is_default" ${credential.is_default ? 'checked' : ''}> Set as default</label>
                        </div>
                        <div class="modal-footer">
                            <button type="button" onclick="Modals.close()">Cancel</button>
                            <button type="submit" class="btn-primary">${editId ? 'Update' : 'Create'}</button>
                        </div>
                    </form>
                </div>
            </div>
        `);

        document.getElementById('credential-form').onsubmit = async (e) => {
            e.preventDefault();
            const form = e.target;
            const data = {
                name: form.name.value,
                provider: form.provider.value,
                is_default: form.is_default.checked
            };
            if (form.api_key.value) data.api_key = form.api_key.value;

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
    }
};
