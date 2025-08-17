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
        // Core elements (required)
        this.statusDot = document.getElementById('status-dot');
        this.statusText = document.getElementById('status-text');
        this.sessionsList = document.getElementById('sessions-list');
        this.modelsList = document.getElementById('models-list');
        this.chatMessages = document.getElementById('chat-messages');
        this.messageInput = document.getElementById('message-input');
        this.sendBtn = document.getElementById('send-btn');
        this.modelSelect = document.getElementById('model-select');
        this.newChatBtn = document.getElementById('new-chat-btn');
        
        // Settings elements
        this.settingsBtn = document.getElementById('settings-btn');
        this.settingsModal = document.getElementById('settings-modal');
        this.closeSettingsBtn = document.getElementById('close-settings');
        
        // Settings form elements (with null checks)
        this.defaultStreamingToggle = document.getElementById('default-streaming');
        this.autoScrollToggle = document.getElementById('auto-scroll');
        this.themeSelect = document.getElementById('theme-select');
        this.sidebarWidthSelect = document.getElementById('sidebar-width');
        this.temperatureSlider = document.getElementById('temperature-slider');
        this.temperatureValue = document.getElementById('temperature-value');
        this.maxTokensInput = document.getElementById('max-tokens');
        
        // Optional elements that may not exist
        this.clearAllSessionsBtn = document.getElementById('clear-all-sessions');
        this.exportSettingsBtn = document.getElementById('export-settings');
        this.importSettingsBtn = document.getElementById('import-settings');
        this.refreshCacheBtn = document.getElementById('refresh-cache-btn');
        this.cacheInfoBtn = document.getElementById('cache-info-btn');
        
        // Model management elements
        this.syncModelsBtn = document.getElementById('sync-models-btn');
        this.modelDownloadInput = document.getElementById('model-download-input');
        this.downloadModelBtn = document.getElementById('download-model-btn');
        this.downloadStatus = document.getElementById('download-status');
        this.progressFill = document.getElementById('progress-fill');
        this.downloadText = document.getElementById('download-text');
        
        // Model tabs
        this.modelTabBtns = document.querySelectorAll('.model-tab-btn');
        this.installedModelsTab = document.getElementById('installed-models-tab');
        this.addModelsTab = document.getElementById('add-models-tab');
        
        // Available models elements
        this.refreshAvailableModelsBtn = document.getElementById('refresh-available-models-btn');
        this.availableModelsList = document.getElementById('available-models-list');
        this.modelSearchInput = document.getElementById('model-search-input');
        this.clearSearchBtn = document.getElementById('clear-search-btn');
    }

    attachEventListeners() {
        // Core event listeners (required elements)
        if (this.sendBtn) {
            this.sendBtn.addEventListener('click', () => this.sendMessage());
        }
        if (this.newChatBtn) {
            this.newChatBtn.addEventListener('click', () => this.createNewSession());
        }
        
        // Settings modal listeners
        if (this.settingsBtn) {
            this.settingsBtn.addEventListener('click', () => this.openSettings());
        }
        if (this.closeSettingsBtn) {
            this.closeSettingsBtn.addEventListener('click', () => this.closeSettings());
        }
        if (this.settingsModal) {
            this.settingsModal.addEventListener('click', (e) => {
                if (e.target === this.settingsModal) {
                    this.closeSettings();
                }
            });
        }
        
        // Settings form listeners
        if (this.temperatureSlider && this.temperatureValue) {
            this.temperatureSlider.addEventListener('input', () => {
                this.temperatureValue.textContent = this.temperatureSlider.value;
            });
        }
        
        if (this.sidebarWidthSelect) {
            this.sidebarWidthSelect.addEventListener('change', () => {
                this.updateSidebarWidth();
            });
        }
        
        if (this.themeSelect) {
            this.themeSelect.addEventListener('change', () => {
                this.applyTheme(this.themeSelect.value);
                this.saveSettings(); // Save immediately when theme changes
            });
        }
        
        // Optional settings listeners
        if (this.clearAllSessionsBtn) {
            this.clearAllSessionsBtn.addEventListener('click', () => this.clearAllSessions());
        }
        if (this.exportSettingsBtn) {
            this.exportSettingsBtn.addEventListener('click', () => this.exportSettings());
        }
        if (this.importSettingsBtn) {
            this.importSettingsBtn.addEventListener('click', () => this.importSettings());
        }
        
        // Model management listeners
        if (this.syncModelsBtn) {
            this.syncModelsBtn.addEventListener('click', () => this.syncModels());
        }
        if (this.downloadModelBtn) {
            this.downloadModelBtn.addEventListener('click', () => this.downloadModel());
        }
        
        // Model tabs listeners
        if (this.modelTabBtns) {
            this.modelTabBtns.forEach(btn => {
                btn.addEventListener('click', (e) => this.switchModelTab(e.target.dataset.tab));
            });
        }
        
        // Enable download button when input has text
        if (this.modelDownloadInput && this.downloadModelBtn) {
            this.modelDownloadInput.addEventListener('input', () => {
                this.downloadModelBtn.disabled = !this.modelDownloadInput.value.trim();
            });
            
            // Allow Enter key to trigger download
            this.modelDownloadInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && this.modelDownloadInput.value.trim()) {
                    this.downloadModel();
                }
            });
        }
        
        // Available models listeners
        if (this.refreshAvailableModelsBtn) {
            this.refreshAvailableModelsBtn.addEventListener('click', () => this.loadAvailableModels());
        }
        
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
        
        // Message input listeners
        if (this.messageInput) {
            this.messageInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
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
        }
        
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
            
            if (data.status === 'healthy' || data.status === 'ok') {
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
                <div class="session-content">
                    <div class="session-title">${this.escapeHtml(session.title)}</div>
                    <div class="session-meta">
                        ${session.message_count} messages • ${this.formatDate(session.updated_at)}
                    </div>
                </div>
                <button class="session-delete-btn" data-session-id="${session.id}" title="Delete session">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M19,4H15.5L14.5,3H9.5L8.5,4H5V6H19M6,19A2,2 0 0,0 8,21H16A2,2 0 0,0 18,19V7H6V19Z" />
                    </svg>
                </button>
            </div>
        `).join('');

        // Add click listeners to session items
        this.sessionsList.querySelectorAll('.session-item').forEach(item => {
            item.addEventListener('click', (e) => {
                // Don't select session if clicking on delete button
                if (e.target.closest('.session-delete-btn')) {
                    return;
                }
                const sessionId = item.dataset.sessionId;
                this.selectSession(sessionId);
            });
        });

        // Add click listeners to delete buttons
        this.sessionsList.querySelectorAll('.session-delete-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const sessionId = btn.dataset.sessionId;
                console.log('Delete button clicked for session:', sessionId);
                this.deleteSession(sessionId);
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
                <div class="welcome-section">
                    <h1 class="welcome-title">What's on the agenda today?</h1>
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
        
        // Clear messages and show welcome
        this.chatMessages.innerHTML = `
            <div class="welcome-section">
                <h1 class="welcome-title">What's on the agenda today?</h1>
            </div>
        `;
        
        // Update sessions list to remove active state
        this.renderSessions();
        
        // The session will be created automatically when the first message is sent
        this.messageInput.focus();
    }

    async deleteSession(sessionId) {
        console.log('deleteSession called with:', sessionId, '- ACTUAL DELETION');
        
        const session = this.sessions.find(s => s.id === sessionId);
        if (!session) {
            console.log('Session not found:', sessionId);
            return;
        }

        if (!confirm(`Are you sure you want to delete "${session.title}"?\n\nThis will permanently delete the session and all its messages. This action cannot be undone.`)) {
            console.log('User cancelled deletion');
            return;
        }

        console.log('Proceeding with actual session deletion via API');
        
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions/${sessionId}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }

            console.log('Session deleted successfully from backend');

            // Remove from local sessions array
            this.sessions = this.sessions.filter(s => s.id !== sessionId);
            console.log('Session removed from local array');
            
            // If this was the current session, clear it
            if (this.currentSessionId === sessionId) {
                this.currentSessionId = null;
                this.chatMessages.innerHTML = `
                    <div class="welcome-section">
                        <h1 class="welcome-title">What's on the agenda today?</h1>
                    </div>
                `;
                console.log('Current session cleared');
            }

            // Re-render sessions list
            this.renderSessions();
            console.log('Sessions list re-rendered');
            
            this.showNotification(`Session "${session.title}" deleted successfully`, 'success');
            console.log('Deletion complete');
            
        } catch (error) {
            console.error('Failed to delete session:', error);
            this.showNotification(`Failed to delete session: ${error.message}`, 'error');
        }
    }

    async sendMessage() {
        const message = this.messageInput.value.trim();
        if (!message || !this.isConnected) return;

        // If no current session, create one
        if (!this.currentSessionId) {
            this.createNewSession();
        }

        const model = this.modelSelect.value;
        const streaming = true; // Always use streaming

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
        if (!this.chatMessages) return null;
        
        // Remove welcome section if it exists
        const welcomeSection = this.chatMessages.querySelector('.welcome-section');
        if (welcomeSection) {
            welcomeSection.remove();
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
        if (!this.chatMessages) return null;
        
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
        if (!this.chatMessages) return;
        
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
        if (this.chatMessages) {
            this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
        }
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
                    <div class="model-info">
                        <div class="model-name">
                            ${this.escapeHtml(model.name)}
                            ${model.is_default ? '<span class="default-badge">DEFAULT</span>' : ''}
                        </div>
                        <div class="model-display-name">${this.escapeHtml(model.display_name)}</div>
                        <div class="model-meta">
                            <span class="model-status ${model.status}">
                                ${model.status === 'downloading' ?
                                    `downloading ${this.getDownloadProgress(model.id)}` :
                                    model.status
                                }
                            </span>
                            ${model.size > 0 ? `<span class="model-size">${this.formatModelSize(model.size)}</span>` : ''}
                            <span>${model.last_used_at ? 'Last used: ' + this.formatDate(model.last_used_at) : 'Never used'}</span>
                        </div>
                        ${model.status === 'downloading' ?
                            `<div class="model-download-progress">
                                <div class="download-progress-bar">
                                    <div class="download-progress-fill" style="width: ${this.getDownloadProgressPercent(model.id)}%"></div>
                                </div>
                            </div>` : ''
                        }
                    </div>
                    <div class="model-actions">
                        <div class="model-actions-dropdown">
                            <button class="model-actions-btn" onclick="chatApp.toggleModelActions('${model.id}')">
                                Actions ▼
                            </button>
                            <div id="actions-${model.id}" class="model-actions-menu">
                                ${model.status === 'downloading' ?
                                    `<button class="warning" onclick="chatApp.cancelDownload('${model.id}')">Cancel Download</button>
                                     <button class="danger" onclick="chatApp.forceRemoveModel('${model.id}')">Force Remove</button>` :
                                    model.status === 'removed' ?
                                        `<button onclick="chatApp.restoreModel('${model.id}')">Restore</button>
                                         <button class="danger" onclick="chatApp.hardDeleteModel('${model.id}')">Delete Forever</button>` :
                                        `${!model.is_default ? `<button class="primary" onclick="chatApp.setDefaultModel('${model.id}')">Set as Default</button>` : ''}
                                         <button onclick="chatApp.toggleModelConfig('${model.id}')">Configuration</button>
                                         <button onclick="chatApp.toggleModel('${model.id}', ${!model.is_enabled})">${model.is_enabled ? 'Disable' : 'Enable'}</button>
                                         <button class="danger" onclick="chatApp.deleteModel('${model.id}')">Remove</button>`
                                }
                            </div>
                        </div>
                    </div>
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
                    <div class="welcome-section">
                        <h1 class="welcome-title">What's on the agenda today?</h1>
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
        
        // Store original button state
        const originalBtnContent = this.downloadModelBtn.innerHTML;
        const originalBtnDisabled = this.downloadModelBtn.disabled;
        
        // Show immediate loading state
        this.downloadModelBtn.disabled = true;
        this.downloadModelBtn.innerHTML = '<span class="loading-spinner"></span> Starting...';
        this.downloadModelBtn.style.background = '#f59e0b';
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
            
            // Update button to show download in progress
            this.downloadModelBtn.innerHTML = '<span class="loading-spinner"></span> Downloading...';
            this.showDownloadProgress(`Download started for ${modelName}`, 10);
            
            // Show success notification
            this.showNotification(`Download started for ${modelName}`, 'success');
            
            // Start polling for download status
            await this.pollModelDownloadStatus(data.id, modelName);
            
        } catch (error) {
            console.error('Failed to start download:', error);
            this.showDownloadError(`Failed to download ${modelName}: ${error.message}`);
            this.showNotification(`Failed to start download: ${error.message}`, 'error');
            
            // Restore original button state on error
            this.downloadModelBtn.disabled = originalBtnDisabled;
            this.downloadModelBtn.innerHTML = originalBtnContent;
            this.downloadModelBtn.style.background = '';
        }
    }
    
    async pollModelDownloadStatus(modelId, modelName) {
        const pollInterval = 3000; // Poll every 3 seconds (reduced frequency)
        const maxPolls = 1800; // Max 90 minutes (3000ms * 1800 = 90 minutes)
        let pollCount = 0;
        let lastProgress = 0;
        let stuckCount = 0;
        let consecutiveErrors = 0;
        let lastSuccessfulPoll = Date.now();
        
        const poll = async () => {
            try {
                const response = await fetch(`${this.apiBase}/v1/models/${modelId}/download-status`);
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                const data = await response.json();
                consecutiveErrors = 0; // Reset error count on successful response
                lastSuccessfulPoll = Date.now();
                
                if (data.status === 'available') {
                    this.showDownloadSuccess(`${modelName} downloaded successfully!`);
                    this.showNotification(`${modelName} is now available!`, 'success');
                    this.modelDownloadInput.value = '';
                    
                    // Reset download button
                    this.downloadModelBtn.disabled = false;
                    this.downloadModelBtn.innerHTML = '📥 Download';
                    this.downloadModelBtn.style.background = '';
                    
                    // Refresh models list
                    await this.loadModels();
                    
                } else if (data.status === 'error') {
                    // Get error details from model description if available
                    const model = this.models.find(m => m.id === modelId);
                    const errorDetails = model && model.description.includes('Download failed:') ?
                        model.description : `Download failed for ${modelName}`;
                    
                    this.showDownloadError(errorDetails);
                    this.showNotification(`Download failed for ${modelName}`, 'error');
                    
                    // Reset download button
                    this.downloadModelBtn.disabled = false;
                    this.downloadModelBtn.innerHTML = '❌ Failed';
                    this.downloadModelBtn.style.background = '#dc2626';
                    
                    // Refresh models list to show error state
                    await this.loadModels();
                    
                } else if (data.status === 'downloading') {
                    // Use actual progress from API, with better fallback logic
                    let progress = data.progress || 0;
                    
                    // If no progress from API, estimate based on time and poll count
                    if (progress === 0 && pollCount > 5) {
                        // Estimate progress: start at 5%, increase slowly
                        progress = Math.min(5 + (pollCount * 0.5), 85);
                    }
                    
                    // Enhanced stuck detection
                    const progressDiff = Math.abs(progress - lastProgress);
                    if (progressDiff < 0.1 && progress > 0) {
                        stuckCount++;
                        
                        // Progressive warnings
                        if (stuckCount === 20) { // 1 minute of no progress
                            this.showNotification(`${modelName} download progress seems slow...`, 'warning');
                        } else if (stuckCount === 40) { // 2 minutes of no progress
                            this.showDownloadProgress(`${modelName} download may be stuck at ${progress.toFixed(1)}%...`, progress);
                            this.showNotification(`${modelName} download appears stuck. This may be normal for large models.`, 'warning');
                        } else if (stuckCount >= 60) { // 3 minutes of no progress
                            this.showDownloadError(`Download appears stuck for ${modelName} at ${progress.toFixed(1)}%. Use Actions menu to cancel if needed.`);
                            this.showNotification(`Download stuck for ${modelName}. Check Actions menu for options.`, 'warning');
                            
                            // Update download button but continue polling
                            this.downloadModelBtn.innerHTML = '⚠️ Stuck';
                            this.downloadModelBtn.style.background = '#f59e0b';
                        }
                    } else {
                        stuckCount = 0; // Reset stuck counter if progress is made
                        lastProgress = progress;
                    }
                    
                    this.showDownloadProgress(`Downloading ${modelName}... ${progress.toFixed(1)}%`, progress);
                    
                    // Update download button with progress
                    if (stuckCount < 60) {
                        this.downloadModelBtn.innerHTML = `<span class="loading-spinner"></span> ${progress.toFixed(0)}%`;
                    }
                    
                    // Update the model in our local list with progress
                    const modelIndex = this.models.findIndex(m => m.id === modelId);
                    if (modelIndex !== -1) {
                        this.models[modelIndex].progress = progress;
                        this.renderModels(); // Re-render to show updated progress
                    }
                    
                    // Continue polling with adaptive interval
                    pollCount++;
                    let nextPollInterval = pollInterval;
                    
                    // Slow down polling if stuck to reduce server load
                    if (stuckCount > 40) {
                        nextPollInterval = 10000; // Poll every 10 seconds if stuck
                    } else if (stuckCount > 20) {
                        nextPollInterval = 5000; // Poll every 5 seconds if slow
                    }
                    
                    if (pollCount < maxPolls) {
                        setTimeout(poll, nextPollInterval);
                    } else {
                        this.showDownloadError(`Download timeout for ${modelName} after 90 minutes. Use Actions menu to cancel.`);
                        this.showNotification(`Download timeout for ${modelName}`, 'error');
                        
                        // Reset download button
                        this.downloadModelBtn.disabled = false;
                        this.downloadModelBtn.innerHTML = '⏰ Timeout';
                        this.downloadModelBtn.style.background = '#dc2626';
                    }
                }
                
            } catch (error) {
                console.error('Failed to check download status:', error);
                consecutiveErrors++;
                
                // Handle network errors gracefully
                if (consecutiveErrors >= 5) {
                    this.showDownloadError(`Lost connection while monitoring ${modelName}. Download may still be in progress.`);
                    this.showNotification(`Connection lost while monitoring download`, 'error');
                    
                    // Reset download button on persistent error
                    this.downloadModelBtn.disabled = false;
                    this.downloadModelBtn.innerHTML = '🔌 Connection Lost';
                    this.downloadModelBtn.style.background = '#dc2626';
                    return; // Stop polling after too many errors
                }
                
                // Check if we've been unable to poll for too long
                if (Date.now() - lastSuccessfulPoll > 300000) { // 5 minutes
                    this.showDownloadError(`Unable to monitor ${modelName} download for 5 minutes. Check connection.`);
                    this.showNotification(`Download monitoring failed`, 'error');
                    
                    this.downloadModelBtn.disabled = false;
                    this.downloadModelBtn.innerHTML = '❌ Monitor Failed';
                    this.downloadModelBtn.style.background = '#dc2626';
                    return;
                }
                
                // Continue polling despite errors, but with longer interval
                pollCount++;
                if (pollCount < maxPolls) {
                    setTimeout(poll, pollInterval * 2); // Double the interval on error
                }
            }
        };
        
        // Start polling after a short delay
        setTimeout(poll, 1000);
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

    // Model Tab Management
    switchModelTab(tabName) {
        // Update tab buttons
        this.modelTabBtns.forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });
        
        // Update tab content
        this.installedModelsTab.classList.toggle('active', tabName === 'installed');
        this.addModelsTab.classList.toggle('active', tabName === 'add');
        
        // Load available models when switching to add tab
        if (tabName === 'add') {
            this.loadAvailableModels();
        }
    }
    
    // Model Actions Dropdown
    toggleModelActions(modelId) {
        // Close all other dropdowns first
        document.querySelectorAll('.model-actions-menu').forEach(menu => {
            if (menu.id !== `actions-${modelId}`) {
                menu.classList.remove('active');
            }
        });
        
        // Toggle the clicked dropdown
        const menu = document.getElementById(`actions-${modelId}`);
        if (menu) {
            menu.classList.toggle('active');
        }
        
        // Close dropdown when clicking outside
        if (menu && menu.classList.contains('active')) {
            const closeHandler = (e) => {
                if (!e.target.closest('.model-actions-dropdown')) {
                    menu.classList.remove('active');
                    document.removeEventListener('click', closeHandler);
                }
            };
            setTimeout(() => document.addEventListener('click', closeHandler), 0);
        }
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
        
        // Show top models first, then categorize the rest
        const topModels = [
            'llama3.2:1b', 'llama3.2:3b', 'llama3.1:8b', 'mistral:7b',
            'codellama:7b', 'phi3:mini', 'gemma:2b', 'qwen2:1.5b'
        ];
        
        const availableTopModels = models.filter(model => topModels.includes(model));
        const otherModels = models.filter(model => !topModels.includes(model));
        
        let html = '';
        
        // Render top models section
        if (availableTopModels.length > 0) {
            html += `
                <div class="model-category">
                    <div class="category-header">
                        <span class="category-title">⭐ Recommended Models</span>
                        <span class="category-description">Popular and well-tested models</span>
                    </div>
                    ${availableTopModels.map(model => this.renderAvailableModelItem(model)).join('')}
                </div>
            `;
        }
        
        // Render other models if any
        if (otherModels.length > 0) {
            const categories = this.categorizeModels(otherModels);
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
        // Switch to the installed models tab to show the input
        this.switchModelTab('add');
    }
    
    async downloadSpecificModel(modelName) {
        // Find the download button for this model and show loading state
        const modelItem = document.querySelector(`[onclick*="downloadSpecificModel('${modelName}')"]`)?.closest('.available-model-item');
        const downloadBtn = modelItem?.querySelector('.download-model-btn');
        
        if (downloadBtn) {
            // Store original button content
            const originalContent = downloadBtn.innerHTML;
            
            // Show immediate loading state
            downloadBtn.innerHTML = '<span class="loading-spinner"></span> Downloading...';
            downloadBtn.disabled = true;
            downloadBtn.style.opacity = '0.7';
        }
        
        // Show download progress immediately
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
            
            // Update button to show download in progress
            if (downloadBtn) {
                downloadBtn.innerHTML = '<span class="loading-spinner"></span> Downloading...';
                downloadBtn.style.background = '#f59e0b';
                downloadBtn.style.color = 'white';
            }
            
            // Start polling for download status
            await this.pollModelDownloadStatus(data.id, modelName);
            
        } catch (error) {
            console.error('Failed to start download:', error);
            this.showDownloadError(`Failed to download ${modelName}: ${error.message}`);
            
            // Restore button state on error
            if (downloadBtn) {
                downloadBtn.innerHTML = '📥 Download';
                downloadBtn.disabled = false;
                downloadBtn.style.opacity = '1';
                downloadBtn.style.background = '';
                downloadBtn.style.color = '';
            }
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

    // Add notification system for better user feedback
    showNotification(message, type = 'info') {
        // Remove existing notifications
        const existingNotifications = document.querySelectorAll('.notification');
        existingNotifications.forEach(notification => notification.remove());
        
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <span class="notification-message">${message}</span>
                <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;
        
        // Add to page
        document.body.appendChild(notification);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 5000);
        
        // Animate in
        setTimeout(() => {
            notification.classList.add('show');
        }, 100);
    }

    // Cancel download functionality
    async cancelDownload(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (!model) return;
        
        if (!confirm(`Cancel download for "${model.display_name}"?\n\nThis will stop the download and mark the model as error state.`)) {
            return;
        }
        
        try {
            // Update model status to error to stop polling
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    status: 'error'
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Download cancelled for "${model.display_name}"`, 'success');
            this.showNotification(`Download cancelled for ${model.display_name}`, 'info');
        } catch (error) {
            console.error('Failed to cancel download:', error);
            this.showSyncStatus(`Failed to cancel download: ${error.message}`, 'error');
            this.showNotification(`Failed to cancel download: ${error.message}`, 'error');
        }
    }

    // Force remove model functionality
    async forceRemoveModel(modelId) {
        const model = this.models.find(m => m.id === modelId);
        if (!model) return;
        
        if (!confirm(`⚠️ FORCE REMOVE WARNING ⚠️\n\nAre you sure you want to force remove "${model.display_name}"?\n\nThis will:\n• Stop any ongoing download\n• Remove the model from the database\n• May leave partial files in Ollama\n\nThis action cannot be undone!`)) {
            return;
        }
        
        try {
            this.showSyncStatus(`Force removing "${model.display_name}"...`, 'warning');
            
            const response = await fetch(`${this.apiBase}/v1/models/${modelId}/hard`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            await this.loadModels(); // Refresh the list
            this.showSyncStatus(`Model "${model.display_name}" force removed`, 'success');
            this.showNotification(`${model.display_name} has been force removed`, 'success');
        } catch (error) {
            console.error('Failed to force remove model:', error);
            this.showSyncStatus(`Failed to force remove model: ${error.message}`, 'error');
            this.showNotification(`Failed to force remove model: ${error.message}`, 'error');
        }
    }
}

// Initialize the app when the page loads
let chatApp;
document.addEventListener('DOMContentLoaded', () => {
    chatApp = new ChatApp();
    
    // Override any cached global references to ensure new method is used
    window.chatApp = chatApp;
    
    // Force override any old deleteSession methods
    if (window.chatApp && window.chatApp.deleteSession) {
        console.log('Global chatApp.deleteSession method is available and updated');
    }
});

// Global function to handle any remaining onclick calls
window.deleteSession = function(sessionId) {
    console.log('Global deleteSession called - redirecting to chatApp instance');
    if (window.chatApp && window.chatApp.deleteSession) {
        window.chatApp.deleteSession(sessionId);
    } else {
        console.error('chatApp instance not available');
    }
};