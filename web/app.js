class ChatApp {
    constructor() {
        this.apiBase = window.location.origin;
        this.currentSessionId = null;
        this.sessions = [];
        this.models = [];
        this.availableModels = [];
        this.isConnected = false;
        this.currentTab = 'sessions';
        
        this.initializeElements();
        this.attachEventListeners();
        this.checkConnection();
        this.loadSessions();
        this.loadModels();
    }

    initializeElements() {
        this.statusDot = document.getElementById('status-dot');
        this.statusText = document.getElementById('status-text');
        this.sessionsList = document.getElementById('sessions-list');
        this.modelsList = document.getElementById('models-list');
        this.chatMessages = document.getElementById('chat-messages');
        this.messageInput = document.getElementById('message-input');
        this.sendBtn = document.getElementById('send-btn');
        this.modelSelect = document.getElementById('model-select');
        this.streamingToggle = document.getElementById('streaming-toggle');
        
        // Settings elements
        this.settingsBtn = document.getElementById('settings-btn');
        this.settingsModal = document.getElementById('settings-modal');
        this.closeSettingsBtn = document.getElementById('close-settings');
        
        // Settings form elements
        this.defaultStreamingToggle = document.getElementById('default-streaming');
        this.autoScrollToggle = document.getElementById('auto-scroll');
        this.themeSelect = document.getElementById('theme-select');
        this.sidebarWidthSelect = document.getElementById('sidebar-width');
        this.temperatureSlider = document.getElementById('temperature-slider');
        this.temperatureValue = document.getElementById('temperature-value');
        this.maxTokensInput = document.getElementById('max-tokens');
        this.clearAllSessionsBtn = document.getElementById('clear-all-sessions');
        this.exportSettingsBtn = document.getElementById('export-settings');
        this.importSettingsBtn = document.getElementById('import-settings');
        
        // Model management elements
        this.syncModelsBtn = document.getElementById('sync-models-btn');
        this.refreshModelsBtn = document.getElementById('refresh-models-btn');
        this.modelDownloadInput = document.getElementById('model-download-input');
        this.downloadModelBtn = document.getElementById('download-model-btn');
        this.downloadStatus = document.getElementById('download-status');
        this.progressFill = document.getElementById('progress-fill');
        this.downloadText = document.getElementById('download-text');
        
        // Available models elements
        this.showAvailableModelsBtn = document.getElementById('show-available-models-btn');
        this.hideAvailableModelsBtn = document.getElementById('hide-available-models-btn');
        this.refreshAvailableModelsBtn = document.getElementById('refresh-available-models-btn');
        this.refreshCacheBtn = document.getElementById('refresh-cache-btn');
        this.cacheInfoBtn = document.getElementById('cache-info-btn');
        this.availableModelsPanel = document.getElementById('available-models-panel');
        this.availableModelsList = document.getElementById('available-models-list');
        this.modelSearchInput = document.getElementById('model-search-input');
        this.clearSearchBtn = document.getElementById('clear-search-btn');
    }

    attachEventListeners() {
        this.sendBtn.addEventListener('click', () => this.sendMessage());
        
        // Settings modal listeners
        this.settingsBtn.addEventListener('click', () => this.openSettings());
        this.closeSettingsBtn.addEventListener('click', () => this.closeSettings());
        this.settingsModal.addEventListener('click', (e) => {
            if (e.target === this.settingsModal) {
                this.closeSettings();
            }
        });
        
        // Settings form listeners
        this.temperatureSlider.addEventListener('input', () => {
            this.temperatureValue.textContent = this.temperatureSlider.value;
        });
        
        this.sidebarWidthSelect.addEventListener('change', () => {
            this.updateSidebarWidth();
        });
        
        this.themeSelect.addEventListener('change', () => {
            this.applyTheme(this.themeSelect.value);
            this.saveSettings(); // Save immediately when theme changes
        });
        
        this.clearAllSessionsBtn.addEventListener('click', () => this.clearAllSessions());
        this.exportSettingsBtn.addEventListener('click', () => this.exportSettings());
        this.importSettingsBtn.addEventListener('click', () => this.importSettings());
        
        // Model management listeners
        this.syncModelsBtn.addEventListener('click', () => this.syncModels());
        this.refreshModelsBtn.addEventListener('click', () => this.loadModels());
        this.downloadModelBtn.addEventListener('click', () => this.downloadModel());
        
        // Enable download button when input has text
        this.modelDownloadInput.addEventListener('input', () => {
            this.downloadModelBtn.disabled = !this.modelDownloadInput.value.trim();
        });
        
        // Allow Enter key to trigger download
        this.modelDownloadInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && this.modelDownloadInput.value.trim()) {
                this.downloadModel();
            }
        });
        
        // Available models listeners
        this.showAvailableModelsBtn.addEventListener('click', () => this.showAvailableModels());
        this.hideAvailableModelsBtn.addEventListener('click', () => this.hideAvailableModels());
        this.refreshAvailableModelsBtn.addEventListener('click', () => this.loadAvailableModels());
        
        // Cache management listeners
        if (this.refreshCacheBtn) {
            this.refreshCacheBtn.addEventListener('click', () => this.refreshCache());
        }
        if (this.cacheInfoBtn) {
            this.cacheInfoBtn.addEventListener('click', () => this.showCacheInfo());
        }
        
        // Search functionality listeners
        if (this.modelSearchInput) {
            this.modelSearchInput.addEventListener('input', () => this.filterAvailableModels());
            this.modelSearchInput.addEventListener('keydown', (e) => {
                if (e.key === 'Escape') {
                    this.clearSearch();
                }
            });
        }
        if (this.clearSearchBtn) {
            this.clearSearchBtn.addEventListener('click', () => this.clearSearch());
        }
        
        this.messageInput.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'Enter') {
                e.preventDefault();
                this.sendMessage();
            }
        });

        this.messageInput.addEventListener('input', () => {
            this.adjustTextareaHeight();
            this.updateSendButton();
        });

        // Auto-resize textarea
        this.adjustTextareaHeight();
        
        // Load settings on startup
        this.loadSettings();
    }

    adjustTextareaHeight() {
        this.messageInput.style.height = 'auto';
        this.messageInput.style.height = Math.min(this.messageInput.scrollHeight, 120) + 'px';
    }

    updateSendButton() {
        const hasText = this.messageInput.value.trim().length > 0;
        this.sendBtn.disabled = !hasText || !this.isConnected;
    }

    async checkConnection() {
        try {
            const response = await fetch(`${this.apiBase}/health`);
            const data = await response.json();
            
            if (data.status === 'healthy') {
                this.setConnectionStatus('connected', 'Connected');
                this.isConnected = true;
            } else {
                this.setConnectionStatus('error', 'Service Degraded');
                this.isConnected = false;
            }
        } catch (error) {
            this.setConnectionStatus('error', 'Connection Failed');
            this.isConnected = false;
        }
        
        this.updateSendButton();
        
        // Check connection every 30 seconds
        setTimeout(() => this.checkConnection(), 30000);
    }

    setConnectionStatus(status, text) {
        this.statusDot.className = `status-dot ${status}`;
        this.statusText.textContent = text;
    }

    async loadSessions() {
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions`);
            const data = await response.json();
            
            this.sessions = data.sessions || [];
            this.renderSessions();
            
            // If no current session and we have sessions, select the first one
            if (!this.currentSessionId && this.sessions.length > 0) {
                this.selectSession(this.sessions[0].id);
            }
        } catch (error) {
            console.error('Failed to load sessions:', error);
            this.showError('Failed to load chat sessions');
        }
    }

    renderSessions() {
        if (this.sessions.length === 0) {
            this.sessionsList.innerHTML = '<div style="padding: 1rem; text-align: center; color: #6b7280;">No chat sessions yet</div>';
            return;
        }

        this.sessionsList.innerHTML = this.sessions.map(session => `
            <div class="session-item ${session.id === this.currentSessionId ? 'active' : ''}" 
                 data-session-id="${session.id}">
                <div class="session-title">${this.escapeHtml(session.title)}</div>
                <div class="session-meta">
                    ${session.message_count} messages • ${this.formatDate(session.updated_at)}
                </div>
            </div>
        `).join('');

        // Add click listeners to session items
        this.sessionsList.querySelectorAll('.session-item').forEach(item => {
            item.addEventListener('click', () => {
                const sessionId = item.dataset.sessionId;
                this.selectSession(sessionId);
            });
        });
    }

    async selectSession(sessionId) {
        this.currentSessionId = sessionId;
        this.renderSessions(); // Re-render to update active state
        await this.loadMessages(sessionId);
    }

    async loadMessages(sessionId) {
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions/${sessionId}/messages`);
            const data = await response.json();
            
            this.renderMessages(data.messages || []);
        } catch (error) {
            console.error('Failed to load messages:', error);
            this.showError('Failed to load chat messages');
        }
    }

    renderMessages(messages) {
        if (messages.length === 0) {
            this.chatMessages.innerHTML = `
                <div class="welcome-message">
                    <h2>Start chatting!</h2>
                    <p>Type a message below to begin your conversation with the AI.</p>
                </div>
            `;
            return;
        }

        this.chatMessages.innerHTML = messages.map(message => `
            <div class="message ${message.role}">
                <div class="message-avatar">
                    ${message.role === 'user' ? 'U' : '🤖'}
                </div>
                <div class="message-content">
                    <div class="message-text">${this.escapeHtml(message.content)}</div>
                    <div class="message-meta">
                        ${this.formatDate(message.created_at)}
                        ${message.tokens_used ? ` • ${message.tokens_used} tokens` : ''}
                        ${message.model ? ` • ${message.model}` : ''}
                    </div>
                </div>
            </div>
        `).join('');

        this.scrollToBottom();
    }

    async createNewSession() {
        // Generate a new session ID
        const sessionId = 'session-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        this.currentSessionId = sessionId;
        
        // Clear messages
        this.chatMessages.innerHTML = `
            <div class="welcome-message">
                <h2>New Chat Started!</h2>
                <p>Type a message below to begin your conversation.</p>
            </div>
        `;
        
        // The session will be created automatically when the first message is sent
        this.messageInput.focus();
    }

    async sendMessage() {
        const message = this.messageInput.value.trim();
        if (!message || !this.isConnected) return;

        // If no current session, create one
        if (!this.currentSessionId) {
            this.createNewSession();
        }

        const model = this.modelSelect.value;
        const streaming = this.streamingToggle.checked;

        // Clear input and disable send button
        this.messageInput.value = '';
        this.adjustTextareaHeight();
        this.updateSendButton();

        // Add user message to UI immediately
        this.addMessageToUI('user', message);

        // Show typing indicator
        const typingIndicator = this.addTypingIndicator();

        try {
            if (streaming) {
                await this.sendStreamingMessage(message, model);
            } else {
                await this.sendNonStreamingMessage(message, model);
            }
        } catch (error) {
            console.error('Failed to send message:', error);
            this.showError('Failed to send message. Please try again.');
        } finally {
            // Remove typing indicator
            if (typingIndicator && typingIndicator.parentNode) {
                typingIndicator.parentNode.removeChild(typingIndicator);
            }
            
            // Reload sessions to update the list
            this.loadSessions();
        }
    }

    async sendNonStreamingMessage(message, model) {
        const response = await fetch(`${this.apiBase}/v1/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                session_id: this.currentSessionId,
                model: model,
                stream: false
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        this.addMessageToUI('assistant', data.content, {
            model: data.model,
            tokens: data.tokens_used,
            timestamp: data.created_at
        });
    }

    async sendStreamingMessage(message, model) {
        const response = await fetch(`${this.apiBase}/v1/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                session_id: this.currentSessionId,
                model: model,
                stream: true
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let assistantMessage = '';
        let messageElement = null;

        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        try {
                            const data = JSON.parse(line.slice(6));
                            
                            if (data.type === 'token') {
                                assistantMessage += data.content;
                                
                                if (!messageElement) {
                                    messageElement = this.addMessageToUI('assistant', assistantMessage);
                                } else {
                                    this.updateMessageContent(messageElement, assistantMessage);
                                }
                            } else if (data.type === 'done') {
                                // Add metadata to the message
                                if (messageElement && data.metadata) {
                                    this.updateMessageMetadata(messageElement, {
                                        model: model,
                                        tokens: data.metadata.total_tokens
                                    });
                                }
                                break;
                            } else if (data.type === 'error') {
                                throw new Error(data.error);
                            }
                        } catch (e) {
                            console.error('Failed to parse SSE data:', e);
                        }
                    }
                }
            }
        } finally {
            reader.releaseLock();
        }
    }

    addMessageToUI(role, content, metadata = {}) {
        // Remove welcome message if it exists
        const welcomeMessage = this.chatMessages.querySelector('.welcome-message');
        if (welcomeMessage) {
            welcomeMessage.remove();
        }

        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${role}`;
        
        const timestamp = metadata.timestamp || new Date().toISOString();
        const metaText = this.formatDate(timestamp) + 
                        (metadata.tokens ? ` • ${metadata.tokens} tokens` : '') +
                        (metadata.model ? ` • ${metadata.model}` : '');

        messageDiv.innerHTML = `
            <div class="message-avatar">
                ${role === 'user' ? 'U' : '🤖'}
            </div>
            <div class="message-content">
                <div class="message-text">${this.escapeHtml(content)}</div>
                <div class="message-meta">${metaText}</div>
            </div>
        `;

        this.chatMessages.appendChild(messageDiv);
        this.scrollToBottom();
        
        return messageDiv;
    }

    updateMessageContent(messageElement, content) {
        const textElement = messageElement.querySelector('.message-text');
        if (textElement) {
            textElement.textContent = content;
            this.scrollToBottom();
        }
    }

    updateMessageMetadata(messageElement, metadata) {
        const metaElement = messageElement.querySelector('.message-meta');
        if (metaElement) {
            const timestamp = new Date().toISOString();
            const metaText = this.formatDate(timestamp) + 
                            (metadata.tokens ? ` • ${metadata.tokens} tokens` : '') +
                            (metadata.model ? ` • ${metadata.model}` : '');
            metaElement.textContent = metaText;
        }
    }

    addTypingIndicator() {
        const typingDiv = document.createElement('div');
        typingDiv.className = 'message assistant typing-indicator';
        typingDiv.innerHTML = `
            <div class="message-avatar">🤖</div>
            <div class="message-content">
                <div class="typing-dots">
                    <div class="typing-dot"></div>
                    <div class="typing-dot"></div>
                    <div class="typing-dot"></div>
                </div>
            </div>
        `;

        this.chatMessages.appendChild(typingDiv);
        this.scrollToBottom();
        
        return typingDiv;
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;
        
        this.chatMessages.appendChild(errorDiv);
        this.scrollToBottom();
        
        // Remove error after 5 seconds
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.parentNode.removeChild(errorDiv);
            }
        }, 5000);
    }

    scrollToBottom() {
        this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays < 7) return `${diffDays}d ago`;
        
        return date.toLocaleDateString();
    }

    // Settings functionality

    // Model management
    async loadModels() {
        try {
            // Load all models (including removed ones) for management interface
            const response = await fetch(`${this.apiBase}/v1/models`);
            const data = await response.json();
            
            this.models = data.models || [];
            this.renderModels();
            this.updateModelSelect();
        } catch (error) {
            console.error('Failed to load models:', error);
            this.showModelError('Failed to load models');
        }
    }

    renderModels() {
        if (this.models.length === 0) {
            this.modelsList.innerHTML = '<div style="padding: 1rem; text-align: center; color: #6b7280;">No models available</div>';
            return;
        }

        this.modelsList.innerHTML = this.models.map(model => `
            <div class="model-item ${model.is_default ? 'default' : ''}" data-model-id="${model.id}">
                <div class="model-header">
                    <div>
                        <div class="model-name">
                            ${this.escapeHtml(model.name)}
                            ${model.is_default ? '<span class="default-badge">DEFAULT</span>' : ''}
                        </div>
                        <div class="model-display-name">${this.escapeHtml(model.display_name)}</div>
                    </div>
                    <div class="model-status-container">
                        ${model.status === 'downloading' ?
                            `<div class="model-download-progress">
                                <div class="download-progress-bar">
                                    <div class="download-progress-fill" style="width: ${this.getDownloadProgressPercent(model.id)}%"></div>
                                </div>
                                <span class="model-status downloading">
                                    downloading ${this.getDownloadProgress(model.id)}
                                </span>
                            </div>` :
                            `<span class="model-status ${model.status}">${model.status}</span>`
                        }
                        ${model.size > 0 ? `<span class="model-size">${this.formatModelSize(model.size)}</span>` : ''}
                    </div>
                </div>
                <div class="model-meta">
                    <span>${model.last_used_at ? 'Last used: ' + this.formatDate(model.last_used_at) : 'Never used'}</span>
                    <span>${model.is_enabled ? 'Enabled' : 'Disabled'}</span>
                </div>
                <div class="model-actions">
                    ${!model.is_default ? `<button class="model-btn primary" onclick="chatApp.setDefaultModel('${model.id}')">Set Default</button>` : ''}
                    <button class="model-btn" onclick="chatApp.toggleModelConfig('${model.id}')">Config</button>
                    <button class="model-btn" onclick="chatApp.toggleModel('${model.id}', ${!model.is_enabled})">${model.is_enabled ? 'Disable' : 'Enable'}</button>
                    ${model.status === 'removed' ?
                        `<button class="model-btn restore" onclick="chatApp.restoreModel('${model.id}')">Restore</button>
                         <button class="model-btn danger" onclick="chatApp.hardDeleteModel('${model.id}')">Delete Forever</button>` :
                        `<button class="model-btn warning" onclick="chatApp.deleteModel('${model.id}')">Remove</button>`
                    }
                </div>
                <div id="config-${model.id}" class="model-config-panel">
                    <div class="loading-spinner"></div>
                    Loading configuration...
                </div>
            </div>
        `).join('');
    }

    updateModelSelect() {
        const currentValue = this.modelSelect.value;
        const availableModels = this.models.filter(m => m.status === 'available' && m.is_enabled);
        
        this.modelSelect.innerHTML = availableModels.map(model =>
            `<option value="${model.name}" ${model.is_default ? 'selected' : ''}>${model.display_name}</option>`
        ).join('');
        
        // Restore previous selection if still available
        if (currentValue && availableModels.some(m => m.name === currentValue)) {
            this.modelSelect.value = currentValue;
        }
    }

    async syncModels() {
        try {
            const response = await fetch(`${this.apiBase}/v1/models/sync`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.models = data.models || [];
            this.renderModels();
            this.updateModelSelect();
            this.showSyncStatus('Models synchronized successfully', 'success');
        } catch (error) {
            console.error('Failed to sync models:', error);
            this.showSyncStatus('Failed to sync models', 'error');
        }
    }

    async setDefaultModel(modelId) {
        try {
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}/default`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus('Default model updated', 'success');
        } catch (error) {
            console.error('Failed to set default model:', error);
            this.showSyncStatus('Failed to set default model', 'error');
        }
    }

    async toggleModel(modelId, enable) {
        try {
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    is_enabled: enable
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Model ${enable ? 'enabled' : 'disabled'}`, 'success');
        } catch (error) {
            console.error('Failed to toggle model:', error);
            this.showSyncStatus('Failed to update model', 'error');
        }
    }

    async toggleModelConfig(modelId) {
        const configPanel = document.getElementById(`config-${modelId}`);
        
        if (configPanel.classList.contains('active')) {
            configPanel.classList.remove('active');
            return;
        }
        
        // Close other config panels
        document.querySelectorAll('.model-config-panel.active').forEach(panel => {
            panel.classList.remove('active');
        });
        
        configPanel.classList.add('active');
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}/config`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const config = await response.json();
            this.renderModelConfig(configPanel, config);
        } catch (error) {
            console.error('Failed to load model config:', error);
            configPanel.innerHTML = '<div style="color: #dc2626; font-size: 0.75rem;">Failed to load configuration</div>';
        }
    }

    renderModelConfig(panel, config) {
        panel.innerHTML = `
            <div class="config-row">
                <span class="config-label">Temperature:</span>
                <span class="config-value">${config.temperature || 'Default'}</span>
            </div>
            <div class="config-row">
                <span class="config-label">Top P:</span>
                <span class="config-value">${config.top_p || 'Default'}</span>
            </div>
            <div class="config-row">
                <span class="config-label">Top K:</span>
                <span class="config-value">${config.top_k || 'Default'}</span>
            </div>
            <div class="config-row">
                <span class="config-label">Context Length:</span>
                <span class="config-value">${config.context_length || 'Default'}</span>
            </div>
            <div class="config-row">
                <span class="config-label">Max Tokens:</span>
                <span class="config-value">${config.max_tokens || 'Default'}</span>
            </div>
            ${config.system_prompt ? `
                <div class="config-row">
                    <span class="config-label">System Prompt:</span>
                    <span class="config-value">${this.escapeHtml(config.system_prompt.substring(0, 50))}${config.system_prompt.length > 50 ? '...' : ''}</span>
                </div>
            ` : ''}
        `;
    }

    showSyncStatus(message, type) {
        // Find the models container in settings
        const modelsContainer = document.querySelector('.models-container') || document.querySelector('.models-list-settings')?.parentElement;
        if (!modelsContainer) {
            console.log('Sync status:', message, type);
            return;
        }
        
        const existingStatus = modelsContainer.querySelector('.sync-status');
        if (existingStatus) {
            existingStatus.remove();
        }
        
        const statusDiv = document.createElement('div');
        statusDiv.className = `sync-status ${type}`;
        statusDiv.textContent = message;
        
        modelsContainer.insertBefore(statusDiv, modelsContainer.firstChild);
        
        setTimeout(() => {
            if (statusDiv.parentNode) {
                statusDiv.parentNode.removeChild(statusDiv);
            }
        }, 3000);
    }

    showModelError(message) {
        this.modelsList.innerHTML = `<div style="padding: 1rem; text-align: center; color: #dc2626;">${message}</div>`;
    }

    // Settings functionality
    openSettings() {
        this.settingsModal.classList.add('active');
        document.body.style.overflow = 'hidden';
    }

    closeSettings() {
        this.settingsModal.classList.remove('active');
        document.body.style.overflow = '';
        this.saveSettings();
    }

    loadSettings() {
        const settings = JSON.parse(localStorage.getItem('chatOllamaSettings') || '{}');
        
        // Apply default streaming setting
        if (settings.defaultStreaming !== undefined) {
            this.defaultStreamingToggle.checked = settings.defaultStreaming;
            this.streamingToggle.checked = settings.defaultStreaming;
        }
        
        // Apply auto-scroll setting
        if (settings.autoScroll !== undefined) {
            this.autoScrollToggle.checked = settings.autoScroll;
        }
        
        // Apply theme setting
        if (settings.theme) {
            this.themeSelect.value = settings.theme;
            this.applyTheme(settings.theme);
        } else {
            // Default to auto theme
            this.themeSelect.value = 'auto';
            this.applyTheme('auto');
        }
        
        // Listen for system theme changes when auto is selected
        this.setupThemeListener();
        
        // Apply sidebar width
        if (settings.sidebarWidth) {
            this.sidebarWidthSelect.value = settings.sidebarWidth;
            this.updateSidebarWidth();
        }
        
        // Apply temperature setting
        if (settings.temperature !== undefined) {
            this.temperatureSlider.value = settings.temperature;
            this.temperatureValue.textContent = settings.temperature;
        }
        
        // Apply max tokens setting
        if (settings.maxTokens) {
            this.maxTokensInput.value = settings.maxTokens;
        }
    }

    saveSettings() {
        const settings = {
            defaultStreaming: this.defaultStreamingToggle.checked,
            autoScroll: this.autoScrollToggle.checked,
            theme: this.themeSelect.value,
            sidebarWidth: this.sidebarWidthSelect.value,
            temperature: parseFloat(this.temperatureSlider.value),
            maxTokens: parseInt(this.maxTokensInput.value)
        };
        
        localStorage.setItem('chatOllamaSettings', JSON.stringify(settings));
        this.applyTheme(settings.theme);
    }

    applyTheme(theme) {
        // Remove existing theme classes
        document.body.classList.remove('dark-theme');
        
        if (theme === 'dark') {
            document.body.classList.add('dark-theme');
        } else if (theme === 'auto') {
            // Check system preference
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            if (prefersDark) {
                document.body.classList.add('dark-theme');
            }
        }
        // 'light' theme is the default (no class needed)
    }

    setupThemeListener() {
        // Listen for system theme changes
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addListener(() => {
            // Only apply auto theme changes if auto is selected
            if (this.themeSelect.value === 'auto') {
                this.applyTheme('auto');
            }
        });
    }

    updateSidebarWidth() {
        const width = this.sidebarWidthSelect.value + 'px';
        document.querySelector('.sidebar').style.width = width;
    }

    async clearAllSessions() {
        if (!confirm('Are you sure you want to clear all chat sessions? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions`, {
                method: 'DELETE'
            });
            
            if (response.ok) {
                this.sessions = [];
                this.currentSessionId = null;
                this.renderSessions();
                this.chatMessages.innerHTML = `
                    <div class="welcome-message">
                        <h2>Welcome to Chat Ollama!</h2>
                        <p>Start a conversation with your local LLM. Your messages are saved and you can continue conversations anytime.</p>
                    </div>
                `;
                alert('All chat sessions have been cleared.');
            } else {
                throw new Error('Failed to clear sessions');
            }
        } catch (error) {
            console.error('Failed to clear sessions:', error);
            alert('Failed to clear sessions. Please try again.');
        }
    }

    exportSettings() {
        const settings = JSON.parse(localStorage.getItem('chatOllamaSettings') || '{}');
        const dataStr = JSON.stringify(settings, null, 2);
        const dataBlob = new Blob([dataStr], {type: 'application/json'});
        
        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = 'chat-ollama-settings.json';
        link.click();
    }

    importSettings() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.json';
        
        input.onchange = (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            const reader = new FileReader();
            reader.onload = (e) => {
                try {
                    const settings = JSON.parse(e.target.result);
                    localStorage.setItem('chatOllamaSettings', JSON.stringify(settings));
                    this.loadSettings();
                    alert('Settings imported successfully!');
                } catch (error) {
                    alert('Invalid settings file. Please check the file format.');
                }
            };
            reader.readAsText(file);
        };
        
        input.click();
    }

    async downloadModel() {
        const modelName = this.modelDownloadInput.value.trim();
        if (!modelName) return;
        
        // Disable the download button and show progress
        this.downloadModelBtn.disabled = true;
        this.downloadModelBtn.innerHTML = '<span class="loading-spinner"></span> Pulling...';
        this.showDownloadProgress('Initiating download...', 0);
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/download`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: modelName,
                    display_name: this.generateDisplayName(modelName),
                    description: `Model: ${modelName}`
                })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.showDownloadProgress(`Download started for ${modelName}`, 10);
            
            // Start polling for download status
            await this.pollModelDownloadStatus(data.id, modelName);
            
        } catch (error) {
            console.error('Failed to start download:', error);
            this.showDownloadError(`Failed to download ${modelName}: ${error.message}`);
        } finally {
            this.downloadModelBtn.disabled = false;
            this.downloadModelBtn.innerHTML = '📥 Pull Model';
        }
    }
    
    async pollModelDownloadStatus(modelId, modelName) {
        const pollInterval = 2000; // Poll every 2 seconds
        const maxPolls = 900; // Max 30 minutes
        let pollCount = 0;
        
        const poll = async () => {
            try {
                const response = await fetch(`${this.apiBase}/v1/models/${modelId}/download-status`);
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                const data = await response.json();
                
                if (data.status === 'available') {
                    this.showDownloadSuccess(`${modelName} downloaded successfully!`);
                    this.modelDownloadInput.value = '';
                    
                    // Refresh models list
                    await this.loadModels();
                    
                } else if (data.status === 'error') {
                    this.showDownloadError(`Download failed for ${modelName}`);
                    
                } else if (data.status === 'downloading') {
                    // Use actual progress from API if available, otherwise estimate
                    const progress = data.progress > 0 ? data.progress : Math.min(10 + (pollCount * 2), 90);
                    this.showDownloadProgress(`Downloading ${modelName}... ${progress.toFixed(1)}%`, progress);
                    
                    // Update the model in our local list with progress
                    const modelIndex = this.models.findIndex(m => m.id === modelId);
                    if (modelIndex !== -1) {
                        this.models[modelIndex].progress = progress;
                        this.renderModels(); // Re-render to show updated progress
                    }
                    
                    // Continue polling
                    pollCount++;
                    if (pollCount < maxPolls) {
                        setTimeout(poll, pollInterval);
                    } else {
                        this.showDownloadError(`Download timeout for ${modelName}`);
                    }
                }
                
            } catch (error) {
                console.error('Failed to check download status:', error);
                this.showDownloadError(`Failed to check download status for ${modelName}`);
            }
        };
        
        // Start polling
        setTimeout(poll, pollInterval);
    }
    
    showDownloadProgress(message, progress) {
        this.downloadStatus.style.display = 'block';
        this.downloadStatus.className = 'download-status';
        this.downloadText.textContent = message;
        this.progressFill.style.width = `${progress}%`;
    }
    
    showDownloadSuccess(message) {
        this.downloadStatus.style.display = 'block';
        this.downloadStatus.className = 'download-status success';
        this.downloadText.textContent = message;
        this.progressFill.style.width = '100%';
        
        // Hide after 5 seconds
        setTimeout(() => {
            this.downloadStatus.style.display = 'none';
        }, 5000);
    }
    
    showDownloadError(message) {
        this.downloadStatus.style.display = 'block';
        this.downloadStatus.className = 'download-status error';
        this.downloadText.textContent = message;
        this.progressFill.style.width = '0%';
        
        // Hide after 8 seconds
        setTimeout(() => {
            this.downloadStatus.style.display = 'none';
        }, 8000);
    }
    
    generateDisplayName(modelName) {
        // Convert model names like "llama3.2:3b" to "Llama3.2 3b"
        return modelName
            .split(':')
            .map(part => part.charAt(0).toUpperCase() + part.slice(1))
            .join(' ');
    }

    // Available Models Management
    async showAvailableModels() {
        this.availableModelsPanel.classList.add('active');
        await this.loadAvailableModels();
    }
    
    hideAvailableModels() {
        this.availableModelsPanel.classList.remove('active');
        this.clearSearch();
    }
    
    async loadAvailableModels() {
        this.availableModelsList.innerHTML = `
            <div class="loading-spinner"></div>
            Loading available models...
        `;
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/available`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.allAvailableModels = data.models || []; // Store all models for filtering
            this.renderAvailableModels(this.allAvailableModels);
            
        } catch (error) {
            console.error('Failed to load available models:', error);
            this.availableModelsList.innerHTML = `
                <div class="available-models-error">
                    Failed to load available models. Please try again.
                </div>
            `;
        }
    }
    
    renderAvailableModels(models) {
        if (models.length === 0) {
            this.availableModelsList.innerHTML = `
                <div class="available-models-empty">
                    No new models available for download.
                    All popular models are already installed.
                </div>
            `;
            return;
        }
        
        // Categorize models
        const categories = this.categorizeModels(models);
        
        let html = '';
        for (const [categoryName, categoryModels] of Object.entries(categories)) {
            if (categoryModels.length === 0) continue;
            
            const categoryInfo = this.getCategoryInfo(categoryName);
            html += `
                <div class="model-category">
                    <div class="category-header">
                        <span class="category-title">${categoryInfo.title}</span>
                        <span class="category-description">${categoryInfo.description}</span>
                    </div>
                    ${categoryModels.map(model => this.renderAvailableModelItem(model)).join('')}
                </div>
            `;
        }
        
        this.availableModelsList.innerHTML = html;
        
        // Update clear button state
        this.updateClearButtonState();
    }
    
    categorizeModels(models) {
        const categories = {
            llama: [],
            mistral: [],
            codellama: [],
            phi: [],
            gemma: [],
            qwen: [],
            deepseek: [],
            openai: [],
            vision: [],
            embedding: [],
            specialized: [],
            other: []
        };
        
        models.forEach(model => {
            const modelLower = model.toLowerCase();
            
            if (modelLower.includes('llava') || modelLower.includes('bakllava')) {
                categories.vision.push(model);
            } else if (modelLower.includes('codellama')) {
                categories.codellama.push(model);
            } else if (modelLower.includes('llama')) {
                categories.llama.push(model);
            } else if (modelLower.includes('mistral') || modelLower.includes('mixtral')) {
                categories.mistral.push(model);
            } else if (modelLower.includes('phi')) {
                categories.phi.push(model);
            } else if (modelLower.includes('gemma')) {
                categories.gemma.push(model);
            } else if (modelLower.includes('qwen')) {
                categories.qwen.push(model);
            } else if (modelLower.includes('deepseek')) {
                categories.deepseek.push(model);
            } else if (modelLower.includes('embed') || modelLower.includes('minilm') || modelLower.includes('arctic-embed')) {
                categories.embedding.push(model);
            } else if (modelLower.includes('openhermes') || modelLower.includes('neural-chat') ||
                      modelLower.includes('starling') || modelLower.includes('openchat') ||
                      modelLower.includes('vicuna') || modelLower.includes('orca') ||
                      modelLower.includes('dolphin') || modelLower.includes('nous-hermes') ||
                      modelLower.includes('zephyr') || modelLower.includes('command-r')) {
                categories.openai.push(model);
            } else if (modelLower.includes('wizard') || modelLower.includes('sql') ||
                      modelLower.includes('med') || modelLower.includes('falcon') ||
                      modelLower.includes('stablelm') || modelLower.includes('tinyllama') ||
                      modelLower.includes('yi') || modelLower.includes('solar')) {
                categories.specialized.push(model);
            } else {
                categories.other.push(model);
            }
        });
        
        return categories;
    }
    
    getCategoryInfo(categoryName) {
        const categoryInfoMap = {
            llama: {
                title: '🦙 Llama Models',
                description: 'Meta\'s Llama family - versatile general-purpose models'
            },
            mistral: {
                title: '🌪️ Mistral Models',
                description: 'Mistral AI\'s efficient and powerful language models'
            },
            codellama: {
                title: '💻 Code Llama',
                description: 'Specialized models for code generation and programming tasks'
            },
            phi: {
                title: '🔬 Phi Models',
                description: 'Microsoft\'s compact yet capable language models'
            },
            gemma: {
                title: '💎 Gemma Models',
                description: 'Google\'s lightweight and efficient language models'
            },
            qwen: {
                title: '🚀 Qwen Models',
                description: 'Alibaba\'s multilingual and high-performance models'
            },
            deepseek: {
                title: '🧠 DeepSeek Models',
                description: 'Specialized models for coding and technical tasks'
            },
            openai: {
                title: '🤖 OpenAI-Compatible Models',
                description: 'Models trained to be compatible with OpenAI API format'
            },
            vision: {
                title: '👁️ Vision Models',
                description: 'Multimodal models that can process both text and images'
            },
            embedding: {
                title: '🔗 Embedding Models',
                description: 'Models for text embeddings and semantic search'
            },
            specialized: {
                title: '⚡ Specialized Models',
                description: 'Domain-specific models for particular use cases'
            },
            other: {
                title: '📦 Other Models',
                description: 'Additional experimental and miscellaneous models'
            }
        };
        
        return categoryInfoMap[categoryName] || { title: categoryName, description: '' };
    }
    
    renderAvailableModelItem(modelName) {
        const modelInfo = this.getModelInfo(modelName);
        
        return `
            <div class="available-model-item" onclick="chatApp.selectAvailableModel('${modelName}')">
                <div class="available-model-info">
                    <div class="available-model-name">${modelName}</div>
                    <div class="available-model-description">${modelInfo.description}</div>
                    ${modelInfo.size ? `<div class="available-model-size">${modelInfo.size}</div>` : ''}
                </div>
                <button class="download-model-btn" onclick="event.stopPropagation(); chatApp.downloadSpecificModel('${modelName}')">
                    📥 Download
                </button>
            </div>
        `;
    }
    
    getModelInfo(modelName) {
        const modelInfoMap = {
            // Llama Models
            'llama3.2:1b': { description: 'Compact Llama model, fast and efficient', size: '~1.3GB' },
            'llama3.2:3b': { description: 'Balanced Llama model, good performance', size: '~2.0GB' },
            'llama3.1:8b': { description: 'Large Llama model, high quality responses', size: '~4.7GB' },
            'llama3.1:70b': { description: 'Massive Llama model, best quality', size: '~40GB' },
            'llama3.1:405b': { description: 'Largest Llama model, state-of-the-art', size: '~230GB' },
            'llama3:8b': { description: 'Standard Llama 3 model', size: '~4.7GB' },
            'llama3:70b': { description: 'Large Llama 3 model', size: '~40GB' },
            'llama2:7b': { description: 'Llama 2 base model', size: '~3.8GB' },
            'llama2:13b': { description: 'Medium Llama 2 model', size: '~7.3GB' },
            'llama2:70b': { description: 'Large Llama 2 model', size: '~39GB' },
            
            // Code Llama
            'codellama:7b': { description: 'Code-focused Llama model', size: '~3.8GB' },
            'codellama:13b': { description: 'Larger code generation model', size: '~7.3GB' },
            'codellama:34b': { description: 'Large code generation model', size: '~19GB' },
            'codellama:7b-instruct': { description: 'Instruction-tuned code model', size: '~3.8GB' },
            'codellama:13b-instruct': { description: 'Large instruction-tuned code model', size: '~7.3GB' },
            'codellama:7b-python': { description: 'Python-specialized code model', size: '~3.8GB' },
            
            // Mistral Models
            'mistral:7b': { description: 'Efficient 7B parameter model', size: '~4.1GB' },
            'mistral:latest': { description: 'Latest Mistral model version', size: '~4.1GB' },
            'mixtral:8x7b': { description: 'Mixture of experts model', size: '~26GB' },
            'mixtral:8x22b': { description: 'Large mixture of experts model', size: '~87GB' },
            
            // Phi Models
            'phi3:mini': { description: 'Compact Microsoft Phi model', size: '~2.3GB' },
            'phi3:medium': { description: 'Balanced Microsoft Phi model', size: '~7.9GB' },
            'phi3:14b': { description: 'Large Microsoft Phi model', size: '~7.9GB' },
            'phi3.5:latest': { description: 'Latest Phi 3.5 model', size: '~2.3GB' },
            
            // Gemma Models
            'gemma:2b': { description: 'Small Google Gemma model', size: '~1.4GB' },
            'gemma:7b': { description: 'Standard Google Gemma model', size: '~4.8GB' },
            'gemma2:2b': { description: 'Improved small Gemma model', size: '~1.6GB' },
            'gemma2:9b': { description: 'Medium Gemma 2 model', size: '~5.4GB' },
            'gemma2:27b': { description: 'Large Gemma 2 model', size: '~16GB' },
            
            // Qwen Models
            'qwen2:0.5b': { description: 'Ultra-compact Qwen model', size: '~0.4GB' },
            'qwen2:1.5b': { description: 'Small Qwen model', size: '~0.9GB' },
            'qwen2:7b': { description: 'Standard Qwen model', size: '~4.4GB' },
            'qwen2:72b': { description: 'Large Qwen model', size: '~41GB' },
            'qwen2.5:7b': { description: 'Latest Qwen 2.5 model', size: '~4.4GB' },
            'qwen2.5:14b': { description: 'Medium Qwen 2.5 model', size: '~8.2GB' },
            'qwen2.5:32b': { description: 'Large Qwen 2.5 model', size: '~18GB' },
            'qwen2.5:72b': { description: 'Largest Qwen 2.5 model', size: '~41GB' },
            
            // DeepSeek Models
            'deepseek-coder:6.7b': { description: 'Code-specialized model', size: '~3.8GB' },
            'deepseek-coder:33b': { description: 'Large code generation model', size: '~18GB' },
            'deepseek-llm:7b': { description: 'General purpose DeepSeek model', size: '~4.1GB' },
            'deepseek-llm:67b': { description: 'Large DeepSeek model', size: '~38GB' },
            
            // OpenAI-Compatible Models
            'openhermes:latest': { description: 'OpenAI-compatible chat model', size: '~4.1GB' },
            'neural-chat:latest': { description: 'Conversational AI model', size: '~4.1GB' },
            'starling-lm:latest': { description: 'High-quality chat model', size: '~4.1GB' },
            'openchat:latest': { description: 'Open-source ChatGPT alternative', size: '~4.1GB' },
            'vicuna:7b': { description: 'Fine-tuned LLaMA model', size: '~4.1GB' },
            'vicuna:13b': { description: 'Larger Vicuna model', size: '~7.3GB' },
            'orca2:latest': { description: 'Microsoft Orca 2 model', size: '~4.1GB' },
            'dolphin-mixtral:latest': { description: 'Uncensored Mixtral variant', size: '~26GB' },
            'nous-hermes2:latest': { description: 'Advanced reasoning model', size: '~4.1GB' },
            'zephyr:latest': { description: 'Instruction-following model', size: '~4.1GB' },
            'command-r:latest': { description: 'Cohere Command R model', size: '~20GB' },
            'command-r-plus:latest': { description: 'Enhanced Command R model', size: '~52GB' },
            
            // Vision Models
            'llava:7b': { description: 'Vision-language model', size: '~4.5GB' },
            'llava:13b': { description: 'Larger vision-language model', size: '~7.8GB' },
            'llava:34b': { description: 'Large vision-language model', size: '~19GB' },
            'llava-llama3:8b': { description: 'LLaVA with Llama 3 base', size: '~5.2GB' },
            'bakllava:latest': { description: 'Improved vision model', size: '~4.5GB' },
            
            // Embedding Models
            'nomic-embed-text:latest': { description: 'High-quality text embeddings', size: '~0.3GB' },
            'all-minilm:latest': { description: 'Compact embedding model', size: '~0.1GB' },
            'mxbai-embed-large:latest': { description: 'Large embedding model', size: '~0.7GB' },
            'snowflake-arctic-embed:latest': { description: 'Arctic embedding model', size: '~0.5GB' },
            
            // Specialized Models
            'wizard-math:latest': { description: 'Mathematics-specialized model', size: '~4.1GB' },
            'wizard-coder:latest': { description: 'Advanced coding model', size: '~4.1GB' },
            'sqlcoder:latest': { description: 'SQL generation specialist', size: '~4.1GB' },
            'medllama2:latest': { description: 'Medical domain model', size: '~4.1GB' },
            'falcon:7b': { description: 'TII Falcon model', size: '~4.1GB' },
            'falcon:40b': { description: 'Large Falcon model', size: '~23GB' },
            'yi:6b': { description: '01.AI Yi model', size: '~3.4GB' },
            'yi:34b': { description: 'Large Yi model', size: '~19GB' },
            'solar:latest': { description: 'Upstage Solar model', size: '~6.1GB' },
            'tinyllama:latest': { description: 'Ultra-compact model', size: '~0.6GB' },
            'stablelm2:latest': { description: 'Stability AI model', size: '~1.6GB' }
        };
        
        return modelInfoMap[modelName] || {
            description: `${modelName} language model`,
            size: ''
        };
    }
    
    selectAvailableModel(modelName) {
        this.modelDownloadInput.value = modelName;
        this.downloadModelBtn.disabled = false;
        this.hideAvailableModels();
    }
    
    async downloadSpecificModel(modelName) {
        // Directly download the model without populating the input field
        this.showDownloadProgress(`Starting download for ${modelName}...`, 0);
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/download`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: modelName,
                    display_name: this.generateDisplayName(modelName),
                    description: `Model: ${modelName}`
                })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.showDownloadProgress(`Download started for ${modelName}`, 10);
            
            // Start polling for download status
            await this.pollModelDownloadStatus(data.id, modelName);
            
        } catch (error) {
            console.error('Failed to start download:', error);
            this.showDownloadError(`Failed to download ${modelName}: ${error.message}`);
        }
    }

    // Cache management methods
    async refreshCache() {
        if (!this.refreshCacheBtn) return;
        
        const originalText = this.refreshCacheBtn.innerHTML;
        this.refreshCacheBtn.disabled = true;
        this.refreshCacheBtn.innerHTML = '<span class="loading-spinner"></span> Refreshing...';
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/available/refresh`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.showSyncStatus(`Cache refreshed successfully. Found ${data.total} available models.`, 'success');
            
            // If the available models panel is open, refresh it
            if (this.availableModelsPanel.classList.contains('active')) {
                this.renderAvailableModels(data.models || []);
            }
            
        } catch (error) {
            console.error('Failed to refresh cache:', error);
            this.showSyncStatus(`Failed to refresh cache: ${error.message}`, 'error');
        } finally {
            this.refreshCacheBtn.disabled = false;
            this.refreshCacheBtn.innerHTML = originalText;
        }
    }
    
    async showCacheInfo() {
        try {
            const response = await fetch(`${this.apiBase}/v1/models/cache-info`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const cacheInfo = await response.json();
            
            const lastUpdated = cacheInfo.last_updated ?
                new Date(cacheInfo.last_updated).toLocaleString() : 'Never';
            const isExpired = cacheInfo.is_expired ? 'Yes' : 'No';
            const timeUntilExpiry = cacheInfo.time_until_expiry > 0 ?
                this.formatDuration(cacheInfo.time_until_expiry) : 'Expired';
            
            const message = `Cache Information:
• Cached Models: ${cacheInfo.cached_models_count}
• Last Updated: ${lastUpdated}
• TTL: ${cacheInfo.ttl_hours} hours
• Expired: ${isExpired}
• Time Until Expiry: ${timeUntilExpiry}`;
            
            alert(message);
            
        } catch (error) {
            console.error('Failed to get cache info:', error);
            alert(`Failed to get cache information: ${error.message}`);
        }
    }
    
    formatDuration(nanoseconds) {
        const seconds = Math.floor(nanoseconds / 1000000000);
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        } else if (minutes > 0) {
            return `${minutes}m`;
        } else {
            return `${seconds}s`;
        }
    }

    // Search functionality for available models
    filterAvailableModels() {
        const searchTerm = this.modelSearchInput.value.toLowerCase().trim();
        
        if (!searchTerm) {
            // If search is empty, show all models
            this.renderAvailableModels(this.allAvailableModels || []);
            this.updateClearButtonState();
            return;
        }
        
        // Filter models based on search term
        const filteredModels = (this.allAvailableModels || []).filter(modelName => {
            const modelLower = modelName.toLowerCase();
            const modelInfo = this.getModelInfo(modelName);
            const descriptionLower = modelInfo.description.toLowerCase();
            
            // Search in model name, description, and category
            const matchesName = modelLower.includes(searchTerm);
            const matchesDescription = descriptionLower.includes(searchTerm);
            const matchesCategory = this.getModelCategory(modelName).toLowerCase().includes(searchTerm);
            
            return matchesName || matchesDescription || matchesCategory;
        });
        
        this.renderAvailableModels(filteredModels);
        this.updateClearButtonState();
    }
    
    getModelCategory(modelName) {
        const modelLower = modelName.toLowerCase();
        
        if (modelLower.includes('llava') || modelLower.includes('bakllava')) {
            return 'vision';
        } else if (modelLower.includes('codellama')) {
            return 'codellama';
        } else if (modelLower.includes('llama')) {
            return 'llama';
        } else if (modelLower.includes('mistral') || modelLower.includes('mixtral')) {
            return 'mistral';
        } else if (modelLower.includes('phi')) {
            return 'phi';
        } else if (modelLower.includes('gemma')) {
            return 'gemma';
        } else if (modelLower.includes('qwen')) {
            return 'qwen';
        } else if (modelLower.includes('deepseek')) {
            return 'deepseek';
        } else if (modelLower.includes('embed') || modelLower.includes('minilm') || modelLower.includes('arctic-embed')) {
            return 'embedding';
        } else if (modelLower.includes('openhermes') || modelLower.includes('neural-chat') ||
                  modelLower.includes('starling') || modelLower.includes('openchat') ||
                  modelLower.includes('vicuna') || modelLower.includes('orca') ||
                  modelLower.includes('dolphin') || modelLower.includes('nous-hermes') ||
                  modelLower.includes('zephyr') || modelLower.includes('command-r')) {
            return 'openai';
        } else if (modelLower.includes('wizard') || modelLower.includes('sql') ||
                  modelLower.includes('med') || modelLower.includes('falcon') ||
                  modelLower.includes('stablelm') || modelLower.includes('tinyllama') ||
                  modelLower.includes('yi') || modelLower.includes('solar')) {
            return 'specialized';
        } else {
            return 'other';
        }
    }
    
    clearSearch() {
        if (this.modelSearchInput) {
            this.modelSearchInput.value = '';
            this.filterAvailableModels();
        }
    }
    
    updateClearButtonState() {
        if (this.clearSearchBtn) {
            const hasSearchTerm = this.modelSearchInput && this.modelSearchInput.value.trim().length > 0;
            this.clearSearchBtn.disabled = !hasSearchTerm;
            this.clearSearchBtn.style.opacity = hasSearchTerm ? '1' : '0.5';
        }
    }

    // Model deletion methods
    async deleteModel(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (!model) return;
        
        if (!confirm(`Are you sure you want to remove "${model.display_name}"?\n\nThis will mark the model as removed but keep it in Ollama. You can restore it later.`)) {
            return;
        }
        
        try {
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Model "${model.display_name}" removed successfully`, 'success');
        } catch (error) {
            console.error('Failed to delete model:', error);
            this.showSyncStatus(`Failed to remove model: ${error.message}`, 'error');
        }
    }
    
    async hardDeleteModel(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (!model) return;
        
        if (!confirm(`⚠️ PERMANENT DELETION WARNING ⚠️\n\nAre you sure you want to PERMANENTLY delete "${model.display_name}"?\n\nThis will:\n• Remove the model from your database\n• Delete the model files from Ollama\n• Free up disk space\n\nThis action CANNOT be undone!`)) {
            return;
        }
        
        // Double confirmation for hard delete
        if (!confirm(`Last chance! Type "DELETE" to confirm permanent deletion of "${model.display_name}"`)) {
            return;
        }
        
        try {
            this.showSyncStatus(`Permanently deleting "${model.display_name}"...`, 'warning');
            
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}/hard`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Model "${model.display_name}" permanently deleted`, 'success');
        } catch (error) {
            console.error('Failed to hard delete model:', error);
            this.showSyncStatus(`Failed to permanently delete model: ${error.message}`, 'error');
        }
    }
    
    async restoreModel(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (!model) return;
        
        try {
            this.showSyncStatus(`Restoring "${model.display_name}"...`, 'info');
            
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}/restore`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Model "${model.display_name}" restored successfully`, 'success');
        } catch (error) {
            console.error('Failed to restore model:', error);
            this.showSyncStatus(`Failed to restore model: ${error.message}`, 'error');
        }
    }

    // Format model size in human-readable format
    formatModelSize(sizeInBytes) {
        if (sizeInBytes === 0) return '';
        
        const gb = sizeInBytes / (1024 * 1024 * 1024);
        if (gb >= 1) {
            return `${gb.toFixed(1)} GB`;
        }
        
        const mb = sizeInBytes / (1024 * 1024);
        return `${mb.toFixed(0)} MB`;
    }
    
    // Get download progress for a model
    getDownloadProgress(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (model && model.progress > 0) {
            return `(${model.progress.toFixed(1)}%)`;
        }
        return '';
    }
    
    // Get download progress percentage for a model
    getDownloadProgressPercent(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (model && model.progress > 0) {
            return model.progress;
        }
        return 0;
    }

    // Override scrollToBottom to respect auto-scroll setting
    scrollToBottom() {
        const settings = JSON.parse(localStorage.getItem('chatOllamaSettings') || '{}');
        if (settings.autoScroll !== false) { // Default to true
            this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
        }
    }
}

// Initialize the app when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new ChatApp();
});
// Initialize the app when the page loads
let chatApp;
document.addEventListener('DOMContentLoaded', () => {
    chatApp = new ChatApp();
});