class ChatApp {
    constructor() {
        this.apiBase = window.location.origin;
        this.currentSessionId = null;
        this.currentProjectId = null;
        this.sessions = [];
        this.models = [];
        this.availableModels = [];
        this.projects = [];
        this.isConnected = false;
        this.currentTab = 'sessions';
        this.isAuthenticated = false;
        this.currentUser = null;
        this.authToken = null;
        
        this.initializeElements();
        this.attachEventListeners();
        this.checkAuthenticationState();
        this.checkConnection();
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
        // Logo replaced the new chat button - keeping reference for compatibility
        this.newChatBtn = null;
        this.sidebar = document.getElementById('sidebar');
        this.sidebarToggleBtn = document.getElementById('sidebar-toggle');
        this.modelSelectorBtn = document.getElementById('model-selector-btn');
        this.modelDropdown = document.getElementById('model-dropdown');
        this.currentModelName = document.getElementById('current-model-name');
        
        // Authentication elements
        this.authModal = document.getElementById('auth-modal');
        this.closeAuthBtn = document.getElementById('close-auth');
        this.loginForm = document.getElementById('login-form');
        this.registerForm = document.getElementById('register-form');
        this.loginFormElement = document.getElementById('login-form-element');
        this.registerFormElement = document.getElementById('register-form-element');
        this.showRegisterBtn = document.getElementById('show-register');
        this.showLoginBtn = document.getElementById('show-login');
        this.loginSubmitBtn = document.getElementById('login-submit-btn');
        this.registerSubmitBtn = document.getElementById('register-submit-btn');
        this.userMenu = document.querySelector('.user-menu');
        this.userMenuDropdown = document.getElementById('user-menu-dropdown');
        this.userMenuName = document.getElementById('user-menu-name');
        this.userMenuEmail = document.getElementById('user-menu-email');
        this.logoutBtn = document.getElementById('logout-btn');
        this.userProfileBtn = document.getElementById('user-profile-btn');
        this.userSettingsBtn = document.getElementById('user-settings-btn');
        
        // Settings elements
        this.settingsBtn = document.getElementById('settings-btn');
        this.settingsModal = document.getElementById('settings-modal');
        this.closeSettingsBtn = document.getElementById('close-settings');
        
        // Settings navigation
        this.settingsNavBtns = document.querySelectorAll('.settings-nav-btn');
        this.generalTab = document.getElementById('general-tab');
        this.modelsTab = document.getElementById('models-tab');
        this.memoryTab = document.getElementById('memory-tab');
        
        // Settings form elements (with null checks)
        this.defaultStreamingToggle = document.getElementById('default-streaming');
        this.autoScrollToggle = document.getElementById('auto-scroll');
        this.themeSelect = document.getElementById('theme-select');
        this.sidebarWidthSelect = document.getElementById('sidebar-width');
        this.temperatureSlider = document.getElementById('temperature-slider');
        this.temperatureValue = document.getElementById('temperature-value');
        this.maxTokensInput = document.getElementById('max-tokens');
        
        // Memory management elements
        this.memoryTabBtns = document.querySelectorAll('.memory-tab-btn');
        this.entitiesTab = document.getElementById('entities-tab');
        this.relationsTab = document.getElementById('relations-tab');
        this.searchTab = document.getElementById('search-tab');
        this.entityNameInput = document.getElementById('entity-name-input');
        this.entityTypeInput = document.getElementById('entity-type-input');
        this.entityObservationsInput = document.getElementById('entity-observations-input');
        this.addEntityBtn = document.getElementById('add-entity-btn');
        this.relationFromInput = document.getElementById('relation-from-input');
        this.relationToInput = document.getElementById('relation-to-input');
        this.relationTypeInput = document.getElementById('relation-type-input');
        this.addRelationBtn = document.getElementById('add-relation-btn');
        this.memorySearchInput = document.getElementById('memory-search-input');
        this.memorySearchBtn = document.getElementById('memory-search-btn');
        this.memoryClearBtn = document.getElementById('memory-clear-btn');
        this.refreshMemoryBtn = document.getElementById('refresh-memory-btn');
        this.clearAllMemoryBtn = document.getElementById('clear-all-memory-btn');
        this.entitiesList = document.getElementById('entities-list');
        this.relationsList = document.getElementById('relations-list');
        this.searchResults = document.getElementById('search-results');
        
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
        
        // Delete confirmation modal elements
        this.deleteModal = document.getElementById('delete-confirmation-modal');
        this.deleteModalTitle = document.getElementById('delete-modal-title');
        this.deleteModalMessage = document.getElementById('delete-modal-message');
        this.deleteModalCancel = document.getElementById('delete-modal-cancel');
        this.deleteModalConfirm = document.getElementById('delete-modal-confirm');
        
        // Project creation modal elements
        this.projectModal = document.getElementById('project-creation-modal');
        this.projectModalInput = document.getElementById('project-name-input');
        this.projectModalCancel = document.getElementById('project-modal-cancel');
        this.projectModalConfirm = document.getElementById('project-modal-confirm');
        
        // Bind the createNewProject method to maintain proper 'this' context
        this.boundCreateNewProject = this.createNewProject.bind(this);
    }

    attachEventListeners() {
        // Core event listeners (required elements)
        if (this.sendBtn) {
            this.sendBtn.addEventListener('click', () => this.sendMessage());
        }
        // New chat button replaced with logo - functionality moved to other UI elements
        // Users can create new sessions through other means
        if (this.sidebarToggleBtn) {
            this.sidebarToggleBtn.addEventListener('click', () => this.toggleSidebar());
        }
        if (this.modelSelectorBtn) {
            this.modelSelectorBtn.addEventListener('click', () => this.toggleModelDropdown());
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
        
        // Settings navigation listeners
        if (this.settingsNavBtns) {
            this.settingsNavBtns.forEach(btn => {
                btn.addEventListener('click', (e) => this.switchSettingsTab(e.target.dataset.tab));
            });
        }
        
        // Memory management listeners
        if (this.memoryTabBtns) {
            this.memoryTabBtns.forEach(btn => {
                btn.addEventListener('click', (e) => this.switchMemoryTab(e.target.dataset.tab));
            });
        }
        
        if (this.addEntityBtn) {
            this.addEntityBtn.addEventListener('click', () => this.createMemorySummary());
        }
        
        if (this.addRelationBtn) {
            this.addRelationBtn.addEventListener('click', () => this.addRelation());
        }
        
        if (this.memorySearchBtn) {
            this.memorySearchBtn.addEventListener('click', () => this.searchMemory());
        }
        
        if (this.memoryClearBtn) {
            this.memoryClearBtn.addEventListener('click', () => this.clearMemorySearch());
        }
        
        if (this.refreshMemoryBtn) {
            this.refreshMemoryBtn.addEventListener('click', () => this.refreshMemory());
        }
        
        if (this.clearAllMemoryBtn) {
            this.clearAllMemoryBtn.addEventListener('click', () => this.clearAllMemory());
        }
        
        if (this.memorySearchInput) {
            this.memorySearchInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    this.searchMemory();
                }
            });
        }
        
        // Delete confirmation modal listeners
        if (this.deleteModalCancel) {
            this.deleteModalCancel.addEventListener('click', () => this.hideDeleteModal());
        }
        if (this.deleteModal) {
            this.deleteModal.addEventListener('click', (e) => {
                if (e.target === this.deleteModal) {
                    this.hideDeleteModal();
                }
            });
        }
        
        // Project creation modal listeners
        if (this.projectModalCancel) {
            this.projectModalCancel.addEventListener('click', () => this.hideProjectModal());
        }
        if (this.projectModalConfirm) {
            this.projectModalConfirm.addEventListener('click', () => this.handleProjectCreation());
        }
        if (this.projectModal) {
            this.projectModal.addEventListener('click', (e) => {
                if (e.target === this.projectModal) {
                    this.hideProjectModal();
                }
            });
        }
        if (this.projectModalInput) {
            this.projectModalInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.handleProjectCreation();
                } else if (e.key === 'Escape') {
                    this.hideProjectModal();
                }
            });
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
        
        // Load sidebar state
        this.loadSidebarState();
        
        // Initialize project menu listeners
        this.initializeProjectMenus();
        
        // Authentication event listeners
        this.attachAuthenticationListeners();
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
        if (!this.isAuthenticated) return;
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/sessions`);
            const data = await response.json();
            
            this.sessions = data.sessions || [];
            
            // Try to restore the last active session from localStorage first
            const savedSessionId = this.getSavedSessionId();
            console.log('Attempting to restore saved session:', savedSessionId);
            
            if (savedSessionId && this.sessions.find(s => s.id === savedSessionId)) {
                console.log('Found saved session in sessions list, selecting it');
                this.currentSessionId = savedSessionId; // Set directly without triggering save again
                this.renderSessions();
                await this.loadMessages(savedSessionId);
            } else if (!this.currentSessionId && this.sessions.length > 0) {
                console.log('No saved session found, selecting first session');
                // If no saved session or saved session doesn't exist, select the first one
                this.selectSession(this.sessions[0].id);
            } else {
                // Just render sessions without selecting any
                this.renderSessions();
            }
        } catch (error) {
            console.error('Failed to load sessions:', error);
            this.showError('Failed to load chat sessions');
        }
    }

    async updateSessionsList() {
        if (!this.isAuthenticated) return;
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/sessions`);
            const data = await response.json();
            
            this.sessions = data.sessions || [];
            
            // Only re-render the sessions list without affecting the current chat
            this.renderSessions();
        } catch (error) {
            console.error('Failed to update sessions list:', error);
            // Don't show error to user as this is a background update
        }
    }

    async loadProjects() {
        if (!this.isAuthenticated) return;
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/projects`);
            const data = await response.json();
            
            this.projects = data.projects || [];
            this.renderProjects();
        } catch (error) {
            console.error('Failed to load projects:', error);
            // Don't show error to user as projects might not be implemented yet
            // Fall back to using the hardcoded projects in HTML
            console.log('Using hardcoded projects from HTML as fallback');
        }
    }

    renderProjects() {
        const projectsList = document.querySelector('.projects-list');
        if (!projectsList) return;
        
        if (this.projects.length === 0) {
            // Show enhanced empty state message when no projects exist
            projectsList.innerHTML = `
                <div class="projects-empty-state">
                    <div class="empty-state-icon">
                        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                            <path d="M12 11v6" stroke-dasharray="2,2" opacity="0.5"></path>
                            <path d="M9 14h6" stroke-dasharray="2,2" opacity="0.5"></path>
                        </svg>
                    </div>
                    <div class="empty-state-content">
                        <h3 class="empty-state-title">No projects yet</h3>
                        <p class="empty-state-description">
                            Projects help you organize your conversations by topic or purpose.
                            Create your first project to get started!
                        </p>
                        <div class="empty-state-actions">
                            <button class="empty-state-btn primary" onclick="chatApp.createNewProject()">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M12 5v14M5 12h14"></path>
                                </svg>
                                Create Your First Project
                            </button>
                        </div>
                        <div class="empty-state-tips">
                            <p class="tip-text">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <circle cx="12" cy="12" r="10"></circle>
                                    <path d="M12 16v-4M12 8h.01"></path>
                                </svg>
                                Tip: You can organize chats by work projects, personal topics, or learning subjects
                            </p>
                        </div>
                    </div>
                </div>
            `;
            // Remove any hardcoded projects that have been deleted
            this.removeDeletedHardcodedProjects();
            // Re-initialize project menu listeners for remaining hardcoded projects
            this.initializeProjectMenus();
            return;
        }

        projectsList.innerHTML = this.projects.map(project => `
            <div class="project-item ${project.id === this.currentProjectId ? 'active' : ''}" data-project-id="${project.id}">
                <div class="project-header">
                    <button class="project-expand-btn" data-project-id="${project.id}" title="Expand/Collapse">
                        <svg class="expand-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="9,18 15,12 9,6"></polyline>
                        </svg>
                    </button>
                    <div class="project-content" data-project-id="${project.id}">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                        </svg>
                        <span>${this.escapeHtml(project.name)}</span>
                    </div>
                    <button class="project-menu-btn" data-project-id="${project.id}" title="Project options">
                        <svg viewBox="0 0 24 24" fill="currentColor">
                            <circle cx="12" cy="12" r="2"></circle>
                            <circle cx="12" cy="5" r="2"></circle>
                            <circle cx="12" cy="19" r="2"></circle>
                        </svg>
                    </button>
                    <div class="project-menu" id="project-menu-${project.id}">
                        <button class="project-menu-item" onclick="chatApp.renameProject('${project.id}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                            </svg>
                            Rename
                        </button>
                        <button class="project-menu-item danger" onclick="chatApp.deleteProject('${project.id}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polyline points="3,6 5,6 21,6"></polyline>
                                <path d="M19,6v14a2,2 0 0,1-2,2H7a2,2 0 0,1-2-2V6m3,0V4a2,2 0 0,1,2-2h4a2,2 0 0,1,2,2v2"></path>
                            </svg>
                            Delete
                        </button>
                    </div>
                </div>
                <div class="project-chats" id="project-chats-${project.id}" style="display: none;">
                    <!-- Project chats will be loaded here when expanded -->
                </div>
            </div>
        `).join('');

        // Re-initialize project menu listeners after rendering
        this.initializeProjectMenus();
    }
    
    // Helper method to remove hardcoded projects that have been deleted
    removeDeletedHardcodedProjects() {
        const deletedProjects = this.getDeletedHardcodedProjects();
        deletedProjects.forEach(projectId => {
            const projectElement = document.querySelector(`[data-project-id="${projectId}"]`);
            if (projectElement) {
                projectElement.remove();
                console.log(`Removed deleted hardcoded project: ${projectId}`);
            }
        });
    }
    
    // Get list of deleted hardcoded projects from localStorage
    getDeletedHardcodedProjects() {
        const deleted = localStorage.getItem('deletedHardcodedProjects');
        return deleted ? JSON.parse(deleted) : [];
    }
    
    // Add a project to the deleted hardcoded projects list
    markHardcodedProjectAsDeleted(projectId) {
        const deleted = this.getDeletedHardcodedProjects();
        if (!deleted.includes(projectId)) {
            deleted.push(projectId);
            localStorage.setItem('deletedHardcodedProjects', JSON.stringify(deleted));
            console.log(`Marked hardcoded project as deleted: ${projectId}`);
        }
    }

    renderSessions() {
        console.log('renderSessions called with', this.sessions.length, 'sessions');
        
        if (this.sessions.length === 0) {
            this.sessionsList.innerHTML = '<div style="padding: 1rem; text-align: center; color: #6b7280;">No chat sessions yet</div>';
            return;
        }

        this.sessionsList.innerHTML = this.sessions.map(session => `
            <div class="session-item ${session.id === this.currentSessionId ? 'active' : ''}"
                 data-session-id="${session.id}">
                <div class="session-content">
                    <div class="session-title">${this.escapeHtml(session.title)}</div>
                </div>
                <button class="session-menu-btn" data-session-id="${session.id}" title="Session options">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="12" r="2"></circle>
                        <circle cx="12" cy="5" r="2"></circle>
                        <circle cx="12" cy="19" r="2"></circle>
                    </svg>
                </button>
                <div class="session-menu" id="session-menu-${session.id}">
                    <button class="session-menu-item" onclick="chatApp.renameSession('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                        Rename
                    </button>
                    <button class="session-menu-item" onclick="chatApp.archiveSession('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="21,8 21,21 3,21 3,8"></polyline>
                            <rect x="1" y="3" width="22" height="5"></rect>
                            <line x1="10" y1="12" x2="14" y2="12"></line>
                        </svg>
                        Archive
                    </button>
                    <button class="session-menu-item danger" onclick="chatApp.deleteSession('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="3,6 5,6 21,6"></polyline>
                            <path d="M19,6v14a2,2 0 0,1-2,2H7a2,2 0 0,1-2-2V6m3,0V4a2,2 0 0,1,2-2h4a2,2 0 0,1,2,2v2"></path>
                        </svg>
                        Delete
                    </button>
                </div>
            </div>
        `).join('');
        
        console.log('Sessions HTML rendered, attaching event listeners...');

        // Add click listeners to session items
        const sessionItems = this.sessionsList.querySelectorAll('.session-item');
        console.log('Found', sessionItems.length, 'session items to attach listeners to');
        
        sessionItems.forEach((item, index) => {
            console.log(`Attaching listener to session item ${index}:`, item.dataset.sessionId);
            item.addEventListener('click', (e) => {
                // Don't select session if clicking on menu button
                if (e.target.closest('.session-menu-btn')) {
                    console.log('Clicked on menu button, ignoring');
                    return;
                }
                const sessionId = item.dataset.sessionId;
                console.log('Session clicked:', sessionId, 'Element:', item);
                if (sessionId) {
                    this.selectSession(sessionId);
                } else {
                    console.error('No session ID found on element:', item);
                }
            });
        });

        // Add click listeners to menu buttons
        this.sessionsList.querySelectorAll('.session-menu-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const sessionId = btn.dataset.sessionId;
                this.toggleSessionMenu(sessionId);
            });
        });
    }

    async selectSession(sessionId) {
        this.currentSessionId = sessionId;
        this.currentProjectId = null; // Clear project selection when selecting a session
        this.saveSessionId(sessionId); // Save to localStorage for persistence
        this.renderSessions(); // Re-render to update active state
        this.updateProjectHighlighting(); // Clear project highlighting
        await this.loadMessages(sessionId);
    }

    async loadMessages(sessionId) {
        if (!this.isAuthenticated) return;
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/sessions/${sessionId}/messages`);
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
                    ${message.role === 'assistant' ? `
                        <div class="message-actions">
                            <button class="message-action-btn copy-btn" title="Copy" onclick="chatApp.copyMessage('${message.id}')">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                                </svg>
                            </button>
                            <button class="message-action-btn tts-btn" title="Read aloud" onclick="chatApp.speakMessage('${message.id}')">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon>
                                    <path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path>
                                </svg>
                            </button>
                            <button class="message-action-btn regenerate-btn" title="Regenerate" onclick="chatApp.regenerateMessage('${message.id}')">
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polyline points="23 4 23 10 17 10"></polyline>
                                    <polyline points="1 20 1 14 7 14"></polyline>
                                    <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"></path>
                                </svg>
                            </button>
                        </div>
                    ` : ''}
                </div>
            </div>
        `).join('');

        this.scrollToBottom();
    }

    async createNewSession() {
        // Generate a new session ID
        const sessionId = 'session-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        this.currentSessionId = sessionId;
        this.currentProjectId = null; // Clear project selection when creating new session
        this.saveSessionId(sessionId); // Save to localStorage for persistence
        
        // Clear messages and show welcome
        this.chatMessages.innerHTML = `
            <div class="welcome-section">
                <h1 class="welcome-title">What's on the agenda today?</h1>
            </div>
        `;
        
        // Update sessions list to remove active state
        this.renderSessions();
        this.updateProjectHighlighting(); // Clear project highlighting
        
        // The session will be created automatically when the first message is sent
        this.messageInput.focus();
    }

    async deleteSession(sessionId) {
        console.log('deleteSession called with:', sessionId, '- SHOWING MODAL');
        
        const session = this.sessions.find(s => s.id === sessionId);
        if (!session) {
            console.log('Session not found:', sessionId);
            return;
        }

        // Show the styled confirmation modal
        this.showDeleteModal(
            'Delete chat?',
            `This will permanently delete "${session.title}" and all its messages. This action cannot be undone.`,
            'Delete chat',
            async () => {
                console.log('Proceeding with actual session deletion via API');
                
                try {
                    const response = await this.authenticatedFetch(`${this.apiBase}/v1/sessions/${sessionId}`, {
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
                        this.clearSavedSessionId(); // Clear from localStorage
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
                    
                    this.showNotification(`Chat "${session.title}" deleted successfully`, 'success');
                    console.log('Deletion complete');
                    
                } catch (error) {
                    console.error('Failed to delete session:', error);
                    this.showNotification(`Failed to delete chat: ${error.message}`, 'error');
                }
            }
        );
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
            
            // Only update the sessions list without reloading messages
            // This prevents clearing the chat messages we just added
            this.updateSessionsList();
        }
    }

    async sendNonStreamingMessage(message, model) {
        const response = await this.authenticatedFetch(`${this.apiBase}/v1/chat`, {
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
        const response = await this.authenticatedFetch(`${this.apiBase}/v1/chat`, {
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
        let buffer = ''; // Buffer to accumulate partial data

        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                // Decode the chunk and add to buffer
                const chunk = decoder.decode(value, { stream: true });
                buffer += chunk;

                // Process complete lines from buffer
                const lines = buffer.split('\n');
                // Keep the last line in buffer as it might be incomplete
                buffer = lines.pop() || '';

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const jsonStr = line.slice(6).trim();
                        if (jsonStr === '') continue; // Skip empty data lines
                        
                        try {
                            const data = JSON.parse(jsonStr);
                            
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
                                return; // Exit the function when done
                            } else if (data.type === 'error') {
                                throw new Error(data.error);
                            }
                        } catch (e) {
                            console.error('Failed to parse SSE data:', e, 'Raw line:', line);
                            // Continue processing other lines instead of breaking
                        }
                    }
                }
            }

            // Process any remaining data in buffer
            if (buffer.trim() && buffer.startsWith('data: ')) {
                const jsonStr = buffer.slice(6).trim();
                if (jsonStr !== '') {
                    try {
                        const data = JSON.parse(jsonStr);
                        if (data.type === 'done' && messageElement && data.metadata) {
                            this.updateMessageMetadata(messageElement, {
                                model: model,
                                tokens: data.metadata.total_tokens
                            });
                        }
                    } catch (e) {
                        console.error('Failed to parse final SSE data:', e, 'Raw buffer:', buffer);
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
                ${role === 'assistant' ? `
                    <div class="message-actions">
                        <button class="message-action-btn copy-btn" title="Copy" onclick="chatApp.copyMessage('${Date.now()}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                            </svg>
                        </button>
                        <button class="message-action-btn tts-btn" title="Read aloud" onclick="chatApp.speakMessage('${Date.now()}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon>
                                <path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path>
                            </svg>
                        </button>
                        <button class="message-action-btn regenerate-btn" title="Regenerate" onclick="chatApp.regenerateMessage('${Date.now()}')">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polyline points="23 4 23 10 17 10"></polyline>
                                <polyline points="1 20 1 14 7 14"></polyline>
                                <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"></path>
                            </svg>
                        </button>
                    </div>
                ` : ''}
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
        
        // Update legacy select for compatibility
        this.modelSelect.innerHTML = availableModels.map(model =>
            `<option value="${model.name}" ${model.is_default ? 'selected' : ''}>${model.display_name}</option>`
        ).join('');
        
        // Update new model dropdown
        this.updateModelDropdown(availableModels);
        
        // Restore previous selection if still available
        if (currentValue && availableModels.some(m => m.name === currentValue)) {
            this.modelSelect.value = currentValue;
        }
    }

    updateModelDropdown(models) {
        if (!this.modelDropdown) return;
        
        const dropdownContent = this.modelDropdown.querySelector('.model-dropdown-content');
        if (!dropdownContent) return;
        
        dropdownContent.innerHTML = models.map(model => `
            <div class="model-option ${model.is_default ? 'active' : ''}" data-model="${model.name}">
                <div class="model-option-info">
                    <div class="model-option-name">${model.display_name}</div>
                    <div class="model-option-description">${this.getModelDescription(model)}</div>
                </div>
            </div>
        `).join('');
        
        // Add click listeners to model options
        dropdownContent.querySelectorAll('.model-option').forEach(option => {
            option.addEventListener('click', (e) => {
                const modelName = option.dataset.model;
                this.selectModel(modelName);
            });
        });
        
        // Update current model name display
        const defaultModel = models.find(m => m.is_default);
        if (defaultModel && this.currentModelName) {
            this.currentModelName.textContent = defaultModel.display_name;
        } else if (models.length > 0 && this.currentModelName) {
            // Fallback to first available model if no default is set
            this.currentModelName.textContent = models[0].display_name;
        } else if (this.currentModelName) {
            // Show a fallback if no models are available
            this.currentModelName.textContent = 'No models available';
        }
    }

    getModelDescription(model) {
        // Provide descriptions based on model names
        const descriptions = {
            // Llama models
            'llama3.2:1b': 'Compact Llama model, fast and efficient',
            'llama3.2:3b': 'Balanced Llama model, good performance',
            'llama3.1:8b': 'Large Llama model, high quality responses',
            'llama3.1:70b': 'Massive Llama model, best quality',
            'llama3:8b': 'Standard Llama 3 model',
            'llama3:70b': 'Large Llama 3 model',
            'llama2:7b': 'Llama 2 base model',
            'llama2:13b': 'Medium Llama 2 model',
            'llama2:70b': 'Large Llama 2 model',
            
            // Code models
            'codellama:7b': 'Code-focused Llama model',
            'codellama:13b': 'Larger code generation model',
            'codellama:34b': 'Large code generation model',
            
            // Mistral models
            'mistral:7b': 'Efficient 7B parameter model',
            'mistral:latest': 'Latest Mistral model version',
            'mixtral:8x7b': 'Mixture of experts model',
            
            // Other popular models
            'phi3:mini': 'Compact Microsoft Phi model',
            'phi3:medium': 'Balanced Microsoft Phi model',
            'gemma:2b': 'Small Google Gemma model',
            'gemma:7b': 'Standard Google Gemma model',
            'qwen2:1.5b': 'Small Qwen model',
            'qwen2:7b': 'Standard Qwen model',
            
            // Legacy fallbacks
            'gpt-5': 'Advanced language model',
            'gpt-4': 'Advanced language model',
            'gpt-4-turbo': 'Advanced language model',
            'gpt-3.5-turbo': 'Advanced language model'
        };
        
        return descriptions[model.name] || model.description || 'Advanced language model';
    }

    toggleModelDropdown() {
        if (!this.modelDropdown || !this.modelSelectorBtn) return;
        
        const isOpen = this.modelDropdown.classList.contains('active');
        
        if (isOpen) {
            this.closeModelDropdown();
        } else {
            this.openModelDropdown();
        }
    }

    openModelDropdown() {
        if (!this.modelDropdown || !this.modelSelectorBtn) return;
        
        // Close any other open dropdowns
        document.querySelectorAll('.model-dropdown.active').forEach(dropdown => {
            if (dropdown !== this.modelDropdown) {
                dropdown.classList.remove('active');
            }
        });
        
        this.modelDropdown.classList.add('active');
        this.modelSelectorBtn.setAttribute('aria-expanded', 'true');
        
        // Close dropdown when clicking outside
        setTimeout(() => {
            const closeHandler = (e) => {
                if (!e.target.closest('.model-selector')) {
                    this.closeModelDropdown();
                    document.removeEventListener('click', closeHandler);
                }
            };
            document.addEventListener('click', closeHandler);
        }, 0);
    }

    closeModelDropdown() {
        if (!this.modelDropdown || !this.modelSelectorBtn) return;
        
        this.modelDropdown.classList.remove('active');
        this.modelSelectorBtn.setAttribute('aria-expanded', 'false');
    }

    selectModel(modelName) {
        // Update the legacy select
        if (this.modelSelect) {
            this.modelSelect.value = modelName;
        }
        
        // Update the dropdown display
        const selectedModel = this.models.find(m => m.name === modelName);
        if (selectedModel && this.currentModelName) {
            this.currentModelName.textContent = selectedModel.display_name;
        }
        
        // Update active state in dropdown
        const modelOptions = this.modelDropdown.querySelectorAll('.model-option');
        modelOptions.forEach(option => {
            option.classList.toggle('active', option.dataset.model === modelName);
        });
        
        // Close dropdown
        this.closeModelDropdown();
        
        // Show notification
        this.showNotification(`Switched to ${selectedModel?.display_name || modelName}`, 'success');
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
        const settings = JSON.parse(localStorage.getItem('ollamaPilotSettings') || '{}');
        
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
        
        localStorage.setItem('ollamaPilotSettings', JSON.stringify(settings));
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
                this.clearSavedSessionId(); // Clear from localStorage
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
        const settings = JSON.parse(localStorage.getItem('ollamaPilotSettings') || '{}');
        const dataStr = JSON.stringify(settings, null, 2);
        const dataBlob = new Blob([dataStr], {type: 'application/json'});
        
        const link = document.createElement('a');
        link.href = URL.createObjectURL(dataBlob);
        link.download = 'ollama-pilot-settings.json';
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
                    localStorage.setItem('ollamaPilotSettings', JSON.stringify(settings));
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
        const settings = JSON.parse(localStorage.getItem('ollamaPilotSettings') || '{}');
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

    // Session menu methods
    toggleSessionMenu(sessionId) {
        // Close all other session menus
        document.querySelectorAll('.session-menu').forEach(menu => {
            if (menu.id !== `session-menu-${sessionId}`) {
                menu.classList.remove('active');
            }
        });
        
        // Toggle the clicked menu
        const menu = document.getElementById(`session-menu-${sessionId}`);
        if (menu) {
            menu.classList.toggle('active');
            
            // Close menu when clicking outside
            if (menu.classList.contains('active')) {
                const closeHandler = (e) => {
                    if (!e.target.closest('.session-menu-btn') && !e.target.closest('.session-menu')) {
                        menu.classList.remove('active');
                        document.removeEventListener('click', closeHandler);
                    }
                };
                setTimeout(() => document.addEventListener('click', closeHandler), 0);
            }
        }
    }

    async renameSession(sessionId) {
        const session = this.sessions.find(s => s.id === sessionId);
        if (!session) return;
        
        const newTitle = prompt('Enter new session name:', session.title);
        if (!newTitle || newTitle === session.title) return;
        
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions/${sessionId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    title: newTitle
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            // Update local session
            session.title = newTitle;
            this.renderSessions();
            this.showNotification('Session renamed successfully', 'success');
        } catch (error) {
            console.error('Failed to rename session:', error);
            this.showNotification('Failed to rename session', 'error');
        }
    }

    async archiveSession(sessionId) {
        const session = this.sessions.find(s => s.id === sessionId);
        if (!session) return;
        
        if (!confirm(`Archive "${session.title}"?`)) return;
        
        try {
            const response = await fetch(`${this.apiBase}/v1/sessions/${sessionId}/archive`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            // Remove from local sessions
            this.sessions = this.sessions.filter(s => s.id !== sessionId);
            
            // If this was the current session, clear it
            if (this.currentSessionId === sessionId) {
                this.currentSessionId = null;
                this.clearSavedSessionId(); // Clear from localStorage
                this.chatMessages.innerHTML = `
                    <div class="welcome-section">
                        <h1 class="welcome-title">What's on the agenda today?</h1>
                    </div>
                `;
            }
            
            this.renderSessions();
            this.showNotification('Session archived', 'success');
        } catch (error) {
            console.error('Failed to archive session:', error);
            this.showNotification('Failed to archive session', 'error');
        }
    }

    // Project menu methods
    toggleProjectMenu(projectId) {
        // Close all other project menus
        document.querySelectorAll('.project-menu').forEach(menu => {
            if (menu.id !== `project-menu-${projectId}`) {
                menu.classList.remove('active');
            }
        });
        
        // Toggle the clicked menu
        const menu = document.getElementById(`project-menu-${projectId}`);
        if (menu) {
            menu.classList.toggle('active');
            
            // Close menu when clicking outside
            if (menu.classList.contains('active')) {
                const closeHandler = (e) => {
                    if (!e.target.closest('.project-menu-btn') && !e.target.closest('.project-menu')) {
                        menu.classList.remove('active');
                        document.removeEventListener('click', closeHandler);
                    }
                };
                setTimeout(() => document.addEventListener('click', closeHandler), 0);
            }
        }
    }

    async renameProject(projectId) {
        const projectElement = document.querySelector(`[data-project-id="${projectId}"] .project-content span`);
        if (!projectElement) return;
        
        const currentName = projectElement.textContent;
        const newName = prompt('Enter new project name:', currentName);
        if (!newName || newName === currentName) return;
        
        try {
            // Make API call to rename the project
            const response = await fetch(`${this.apiBase}/v1/projects/${projectId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: newName
                })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            // Update local projects array
            const project = this.projects.find(p => p.id === projectId);
            if (project) {
                project.name = newName;
            }

            // Update local UI
            projectElement.textContent = newName;
            this.showNotification(`Project renamed to "${newName}"`, 'success');
            
            // Close the menu
            const menu = document.getElementById(`project-menu-${projectId}`);
            if (menu) {
                menu.classList.remove('active');
            }
        } catch (error) {
            console.error('Failed to rename project:', error);
            this.showNotification(`Failed to rename project: ${error.message}`, 'error');
        }
    }

    async deleteProject(projectId) {
        const projectElement = document.querySelector(`[data-project-id="${projectId}"] .project-content span`);
        if (!projectElement) return;
        
        const projectName = projectElement.textContent;
        
        // Check if this is a real project from the backend or a hardcoded one
        const project = this.projects.find(p => p.id === projectId);
        const isHardcodedProject = !project;
        
        // Show the styled confirmation modal
        this.showDeleteModal(
            'Delete project?',
            `This will permanently delete all project files and chats. To save chats, move them to your chat list or another project before deleting.`,
            'Delete project',
            async () => {
                console.log('Proceeding with project deletion');
                
                try {
                    if (isHardcodedProject) {
                        console.log('Deleting hardcoded project - removing from DOM only');
                        
                        // Mark as deleted in localStorage so it doesn't reappear on refresh
                        this.markHardcodedProjectAsDeleted(projectId);
                        
                        // Remove from local UI immediately
                        const projectItem = document.querySelector(`[data-project-id="${projectId}"]`);
                        if (projectItem) {
                            projectItem.remove();
                            console.log('Hardcoded project removed from local UI');
                        }
                        
                        this.showNotification(`Project "${projectName}" deleted successfully`, 'success');
                        console.log('Hardcoded project deletion complete');
                        
                    } else {
                        console.log('Deleting real project via API');
                        
                        // Make API call to delete the project
                        const response = await this.authenticatedFetch(`${this.apiBase}/v1/projects/${projectId}`, {
                            method: 'DELETE'
                        });

                        if (!response.ok) {
                            const errorData = await response.json();
                            throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
                        }

                        console.log('Project deleted successfully from backend');

                        // Remove from local projects array
                        this.projects = this.projects.filter(p => p.id !== projectId);
                        console.log('Project removed from local array');

                        // Remove from local UI
                        const projectItem = document.querySelector(`[data-project-id="${projectId}"]`);
                        if (projectItem) {
                            projectItem.remove();
                            console.log('Project removed from local UI');
                        }
                        
                        this.showNotification(`Project "${projectName}" deleted successfully`, 'success');
                        console.log('API project deletion complete');
                    }
                    
                } catch (error) {
                    console.error('Failed to delete project:', error);
                    this.showNotification(`Failed to delete project: ${error.message}`, 'error');
                }
            }
        );
    }

    // Initialize project menu event listeners
    initializeProjectMenus() {
        // Add click listeners to project menu buttons
        document.querySelectorAll('.project-menu-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const projectId = btn.dataset.projectId;
                this.toggleProjectMenu(projectId);
            });
        });

        // Add click listeners to project expand buttons
        document.querySelectorAll('.project-expand-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const projectId = btn.dataset.projectId;
                this.toggleProjectExpansion(projectId);
            });
        });

        // Add click listeners to project content (project name area)
        document.querySelectorAll('.project-content').forEach(content => {
            content.addEventListener('click', (e) => {
                e.stopPropagation();
                const projectId = content.dataset.projectId;
                this.selectProject(projectId);
            });
        });

        // Initialize project chat listeners for existing chats
        document.querySelectorAll('.project-chats').forEach(projectChats => {
            this.initializeProjectChatListeners(projectChats);
        });

        // Add click listener to the "Add Project" button
        const addProjectBtn = document.querySelector('.projects-add-btn');
        if (addProjectBtn) {
            // Remove any existing listeners to avoid duplicates
            addProjectBtn.removeEventListener('click', this.boundCreateNewProject);
            addProjectBtn.addEventListener('click', this.boundCreateNewProject);
        }
    }

    async createNewProject() {
        this.showProjectModal();
    }

    showProjectModal() {
        if (!this.projectModal) return;
        
        // Clear previous input
        if (this.projectModalInput) {
            this.projectModalInput.value = '';
        }
        
        // Show modal
        this.projectModal.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        // Focus on input
        if (this.projectModalInput) {
            setTimeout(() => this.projectModalInput.focus(), 100);
        }
    }

    hideProjectModal() {
        if (!this.projectModal) return;
        
        this.projectModal.classList.remove('active');
        document.body.style.overflow = '';
    }

    async handleProjectCreation() {
        const projectName = this.projectModalInput?.value?.trim();
        if (!projectName) return;

        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/projects`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: projectName
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
            }

            const newProject = await response.json();
            
            // Add to local projects array
            this.projects.push(newProject);
            
            // Re-render projects
            this.renderProjects();
            
            // Hide modal
            this.hideProjectModal();
            
            this.showNotification(`Project "${projectName}" created successfully`, 'success');
        } catch (error) {
            console.error('Failed to create project:', error);
            this.showNotification(`Failed to create project: ${error.message}`, 'error');
        }
    }

    // Project selection methods
    selectProject(projectId) {
        this.currentProjectId = projectId;
        this.currentSessionId = null; // Clear session selection when selecting a project
        this.updateProjectHighlighting();
        this.updateSessionHighlighting(); // Clear session highlighting
        
        // Clear chat messages when switching to project
        this.chatMessages.innerHTML = `
            <div class="welcome-section">
                <h1 class="welcome-title">What's on the agenda today?</h1>
            </div>
        `;
        
        const projectName = document.querySelector(`[data-project-id="${projectId}"] .project-content span`)?.textContent;
        if (projectName) {
            this.showNotification(`Switched to project: ${projectName}`, 'info');
        }
    }

    updateProjectHighlighting() {
        // Remove active class from all projects
        document.querySelectorAll('.project-item').forEach(p => p.classList.remove('active'));
        
        // Add active class to current project if one is selected
        if (this.currentProjectId) {
            const currentProject = document.querySelector(`[data-project-id="${this.currentProjectId}"]`);
            if (currentProject) {
                currentProject.classList.add('active');
            }
        }
    }

    updateSessionHighlighting() {
        // Re-render sessions to update highlighting based on currentSessionId
        this.renderSessions();
    }

    // Message action methods
    async copyMessage(messageId) {
        const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
        if (!messageElement) {
            // Fallback: find message by content
            const messages = document.querySelectorAll('.message-text');
            for (let msg of messages) {
                if (msg.textContent) {
                    try {
                        await navigator.clipboard.writeText(msg.textContent);
                        this.showNotification('Message copied to clipboard', 'success');
                        return;
                    } catch (error) {
                        console.error('Failed to copy message:', error);
                    }
                }
            }
            return;
        }
        
        const messageText = messageElement.querySelector('.message-text')?.textContent;
        if (messageText) {
            try {
                await navigator.clipboard.writeText(messageText);
                this.showNotification('Message copied to clipboard', 'success');
            } catch (error) {
                console.error('Failed to copy message:', error);
                this.showNotification('Failed to copy message', 'error');
            }
        }
    }

    speakMessage(messageId) {
        const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
        let messageText = '';
        
        if (!messageElement) {
            // Fallback: find the last assistant message
            const assistantMessages = document.querySelectorAll('.message.assistant .message-text');
            if (assistantMessages.length > 0) {
                messageText = assistantMessages[assistantMessages.length - 1].textContent;
            }
        } else {
            messageText = messageElement.querySelector('.message-text')?.textContent;
        }
        
        if (messageText && 'speechSynthesis' in window) {
            // Stop any ongoing speech
            speechSynthesis.cancel();
            
            const utterance = new SpeechSynthesisUtterance(messageText);
            utterance.rate = 0.9;
            utterance.pitch = 1;
            utterance.volume = 0.8;
            
            speechSynthesis.speak(utterance);
            this.showNotification('Reading message aloud', 'info');
        } else {
            this.showNotification('Text-to-speech not supported', 'error');
        }
    }

    async regenerateMessage(messageId) {
        // For now, just show a notification
        // In a full implementation, this would resend the last user message
        this.showNotification('Regenerate feature coming soon', 'info');
    }

    // Sidebar toggle functionality
    toggleSidebar() {
        if (this.sidebar) {
            this.sidebar.classList.toggle('collapsed');
            
            // Save sidebar state to localStorage
            const isCollapsed = this.sidebar.classList.contains('collapsed');
            localStorage.setItem('sidebarCollapsed', isCollapsed.toString());
            
            // Update toggle button icon
            this.updateSidebarToggleIcon(isCollapsed);
        }
    }

    updateSidebarToggleIcon(isCollapsed) {
        if (this.sidebarToggleBtn) {
            const svg = this.sidebarToggleBtn.querySelector('svg');
            if (svg) {
                if (isCollapsed) {
                    // Show expand icon (arrow right)
                    svg.innerHTML = '<path d="M9 18l6-6-6-6"></path>';
                } else {
                    // Show collapse icon (hamburger menu)
                    svg.innerHTML = '<path d="M3 12h18M3 6h18M3 18h18"></path>';
                }
            }
        }
    }

    // Load sidebar state from localStorage
    loadSidebarState() {
        const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
        if (isCollapsed && this.sidebar) {
            this.sidebar.classList.add('collapsed');
            this.updateSidebarToggleIcon(true);
        }
    }

    // Session persistence methods
    saveSessionId(sessionId) {
        try {
            localStorage.setItem('currentSessionId', sessionId);
            console.log('Session ID saved to localStorage:', sessionId);
        } catch (error) {
            console.warn('Failed to save session ID to localStorage:', error);
        }
    }

    getSavedSessionId() {
        try {
            const sessionId = localStorage.getItem('currentSessionId');
            console.log('Retrieved session ID from localStorage:', sessionId);
            return sessionId;
        } catch (error) {
            console.warn('Failed to retrieve session ID from localStorage:', error);
            return null;
        }
    }

    clearSavedSessionId() {
        try {
            localStorage.removeItem('currentSessionId');
            console.log('Session ID cleared from localStorage');
        } catch (error) {
            console.warn('Failed to clear session ID from localStorage:', error);
        }
    }

    // Delete confirmation modal methods
    showDeleteModal(title, message, confirmButtonText, onConfirm) {
        if (!this.deleteModal) return;
        
        // Set modal content
        if (this.deleteModalTitle) {
            this.deleteModalTitle.textContent = title;
        }
        if (this.deleteModalMessage) {
            this.deleteModalMessage.textContent = message;
        }
        if (this.deleteModalConfirm) {
            this.deleteModalConfirm.textContent = confirmButtonText;
        }
        
        // Remove any existing event listeners
        if (this.deleteModalConfirm) {
            const newConfirmBtn = this.deleteModalConfirm.cloneNode(true);
            this.deleteModalConfirm.parentNode.replaceChild(newConfirmBtn, this.deleteModalConfirm);
            this.deleteModalConfirm = newConfirmBtn;
            
            // Add new event listener
            this.deleteModalConfirm.addEventListener('click', () => {
                this.hideDeleteModal();
                if (onConfirm) {
                    onConfirm();
                }
            });
        }
        
        // Show modal
        this.deleteModal.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        // Focus on cancel button for accessibility
        if (this.deleteModalCancel) {
            setTimeout(() => this.deleteModalCancel.focus(), 100);
        }
    }
    
    hideDeleteModal() {
        if (!this.deleteModal) return;
        
        this.deleteModal.classList.remove('active');
        document.body.style.overflow = '';
    }

    // Project expansion functionality
    toggleProjectExpansion(projectId) {
        const projectItem = document.querySelector(`[data-project-id="${projectId}"]`);
        const projectChats = document.getElementById(`project-chats-${projectId}`);
        
        if (!projectItem || !projectChats) return;
        
        const isExpanded = projectItem.classList.contains('expanded');
        
        if (isExpanded) {
            // Collapse
            projectItem.classList.remove('expanded');
            projectChats.style.display = 'none';
            projectChats.classList.remove('expanded');
        } else {
            // Expand
            projectItem.classList.add('expanded');
            projectChats.style.display = 'block';
            projectChats.classList.add('expanded');
            
            // Load project chats if not already loaded
            this.loadProjectChats(projectId);
        }
    }

    // Load chats for a specific project
    async loadProjectChats(projectId) {
        try {
            const response = await fetch(`${this.apiBase}/v1/projects/${projectId}/sessions`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.renderProjectChats(projectId, data.sessions || []);
        } catch (error) {
            console.error('Failed to load project chats:', error);
            // For now, keep the hardcoded chats as fallback
        }
    }

    // Render chats for a specific project
    renderProjectChats(projectId, sessions) {
        const projectChats = document.getElementById(`project-chats-${projectId}`);
        if (!projectChats) return;
        
        if (sessions.length === 0) {
            projectChats.innerHTML = '<div class="project-chat-empty">No chats in this project</div>';
            return;
        }
        
        projectChats.innerHTML = sessions.map(session => `
            <div class="project-chat-item" data-session-id="${session.id}">
                <div class="project-chat-content">
                    <div class="project-chat-title">${this.escapeHtml(session.title)}</div>
                </div>
                <button class="project-chat-menu-btn" data-session-id="${session.id}" title="Chat options">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="12" r="2"></circle>
                        <circle cx="12" cy="5" r="2"></circle>
                        <circle cx="12" cy="19" r="2"></circle>
                    </svg>
                </button>
                <div class="project-chat-menu" id="project-chat-menu-${session.id}">
                    <button class="project-chat-menu-item" onclick="chatApp.renameProjectChat('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                        Rename
                    </button>
                    <button class="project-chat-menu-item" onclick="chatApp.archiveProjectChat('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="21,8 21,21 3,21 3,8"></polyline>
                            <rect x="1" y="3" width="22" height="5"></rect>
                            <line x1="10" y1="12" x2="14" y2="12"></line>
                        </svg>
                        Archive
                    </button>
                    <button class="project-chat-menu-item danger" onclick="chatApp.deleteProjectChat('${session.id}')">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="3,6 5,6 21,6"></polyline>
                            <path d="M19,6v14a2,2 0 0,1-2,2H7a2,2 0 0,1-2-2V6m3,0V4a2,2 0 0,1,2-2h4a2,2 0 0,1,2,2v2"></path>
                        </svg>
                        Delete
                    </button>
                </div>
            </div>
        `).join('');
        
        // Re-attach event listeners for the new chat items
        this.initializeProjectChatListeners(projectChats);
    }

    // Initialize event listeners for project chat items
    initializeProjectChatListeners(projectChats) {
        // Add click listeners to project chat content
        projectChats.querySelectorAll('.project-chat-content').forEach(content => {
            content.addEventListener('click', (e) => {
                e.stopPropagation();
                const chatItem = content.closest('.project-chat-item');
                const sessionId = chatItem.dataset.sessionId;
                this.selectProjectChat(sessionId);
            });
        });

        // Add click listeners to project chat menu buttons
        projectChats.querySelectorAll('.project-chat-menu-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const sessionId = btn.dataset.sessionId;
                this.toggleProjectChatMenu(sessionId);
            });
        });
    }

    // Toggle project chat menu
    toggleProjectChatMenu(sessionId) {
        // Close all other project chat menus
        document.querySelectorAll('.project-chat-menu').forEach(menu => {
            if (menu.id !== `project-chat-menu-${sessionId}`) {
                menu.classList.remove('active');
            }
        });
        
        // Toggle the clicked menu
        const menu = document.getElementById(`project-chat-menu-${sessionId}`);
        if (menu) {
            menu.classList.toggle('active');
            
            // Close menu when clicking outside
            if (menu.classList.contains('active')) {
                const closeHandler = (e) => {
                    if (!e.target.closest('.project-chat-menu-btn') && !e.target.closest('.project-chat-menu')) {
                        menu.classList.remove('active');
                        document.removeEventListener('click', closeHandler);
                    }
                };
                setTimeout(() => document.addEventListener('click', closeHandler), 0);
            }
        }
    }

    // Project chat menu actions
    async renameProjectChat(sessionId) {
        // Use the same logic as regular session rename
        await this.renameSession(sessionId);
    }

    async archiveProjectChat(sessionId) {
        // Use the same logic as regular session archive
        await this.archiveSession(sessionId);
    }

    async deleteProjectChat(sessionId) {
        console.log('deleteProjectChat called with:', sessionId, '- SHOWING MODAL');
        
        // Try to find the session in the loaded sessions array
        let session = this.sessions.find(s => s.id === sessionId);
        let isHardcodedSession = false;
        
        // If not found in sessions array, try to get the title from the DOM element
        let sessionTitle = 'Unknown Chat';
        if (!session) {
            console.log('Session not found in sessions array, checking DOM for title');
            const chatElement = document.querySelector(`[data-session-id="${sessionId}"] .project-chat-title`);
            if (chatElement) {
                sessionTitle = chatElement.textContent;
                console.log('Found session title in DOM:', sessionTitle);
                // This is a hardcoded session that doesn't exist in the backend
                isHardcodedSession = true;
                // Create a temporary session object for the modal
                session = { id: sessionId, title: sessionTitle };
            } else {
                console.log('Session not found in DOM either, using fallback');
                session = { id: sessionId, title: sessionTitle };
                isHardcodedSession = true;
            }
        } else {
            sessionTitle = session.title;
        }

        // Show the styled confirmation modal (same as regular chat delete)
        this.showDeleteModal(
            'Delete chat?',
            `This will permanently delete "${sessionTitle}" and all its messages. This action cannot be undone.`,
            'Delete chat',
            async () => {
                console.log('Proceeding with project chat deletion');
                
                try {
                    // If this is a hardcoded session, just remove it from DOM
                    if (isHardcodedSession) {
                        console.log('Deleting hardcoded session - removing from DOM only');
                        
                        // Remove the chat item from the DOM immediately
                        const chatElement = document.querySelector(`[data-session-id="${sessionId}"]`);
                        if (chatElement) {
                            chatElement.remove();
                            console.log('Hardcoded chat element removed from DOM');
                        }
                        
                        // If this was the current session, clear it
                        if (this.currentSessionId === sessionId) {
                            this.currentSessionId = null;
                            this.clearSavedSessionId(); // Clear from localStorage
                            this.chatMessages.innerHTML = `
                                <div class="welcome-section">
                                    <h1 class="welcome-title">What's on the agenda today?</h1>
                                </div>
                            `;
                            console.log('Current session cleared');
                        }
                        
                        this.showNotification(`Chat "${sessionTitle}" deleted successfully`, 'success');
                        console.log('Hardcoded project chat deletion complete');
                        
                    } else {
                        // This is a real session from the backend, delete via API
                        console.log('Deleting real session via API');
                        
                        const response = await fetch(`${this.apiBase}/v1/sessions/${sessionId}`, {
                            method: 'DELETE'
                        });

                        if (!response.ok) {
                            const errorData = await response.json();
                            throw new Error(errorData.detail || `HTTP ${response.status}: ${response.statusText}`);
                        }

                        console.log('Project chat deleted successfully from backend');

                        // Remove from local sessions array
                        this.sessions = this.sessions.filter(s => s.id !== sessionId);
                        console.log('Project chat removed from local array');
                        
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
                        
                        // Reload project chats for all expanded projects
                        document.querySelectorAll('.project-item.expanded').forEach(projectItem => {
                            const projectId = projectItem.dataset.projectId;
                            if (projectId) {
                                this.loadProjectChats(projectId);
                            }
                        });
                        
                        this.showNotification(`Chat "${sessionTitle}" deleted successfully`, 'success');
                        console.log('API project chat deletion complete');
                    }
                    
                } catch (error) {
                    console.error('Failed to delete project chat:', error);
                    this.showNotification(`Failed to delete chat: ${error.message}`, 'error');
                }
            }
        );
    }

    // Select a chat from within a project
    async selectProjectChat(sessionId) {
        // Clear any active project chat items
        document.querySelectorAll('.project-chat-item').forEach(item => {
            item.classList.remove('active');
        });
        
        // Set the clicked chat as active
        const chatItem = document.querySelector(`[data-session-id="${sessionId}"]`);
        if (chatItem) {
            chatItem.classList.add('active');
        }
        
        // Load the session like a regular session
        this.currentSessionId = sessionId;
        this.currentProjectId = null; // Clear project selection when selecting a chat
        this.renderSessions(); // Re-render to update active state
        await this.loadMessages(sessionId);
        
        this.showNotification(`Opened chat: ${chatItem?.textContent || 'Unknown'}`, 'info');
    }

    // Settings Tab Management
    switchSettingsTab(tabName) {
        // Update navigation buttons
        this.settingsNavBtns.forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });
        
        // Update tab content
        this.generalTab.classList.toggle('active', tabName === 'general');
        this.modelsTab.classList.toggle('active', tabName === 'models');
        this.memoryTab.classList.toggle('active', tabName === 'memory');
        
        // Load data when switching tabs
        if (tabName === 'models') {
            // Load models when switching to models tab
            this.loadModels();
        } else if (tabName === 'memory') {
            // Initialize memory tab if not already done
            const activeMemoryTab = document.querySelector('.memory-tab-btn.active')?.dataset.tab || 'entities';
            this.switchMemoryTab(activeMemoryTab);
        }
    }

    // Memory Management Methods
    switchMemoryTab(tabName) {
        // Update tab buttons
        this.memoryTabBtns.forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });
        
        // Update tab content
        this.entitiesTab.classList.toggle('active', tabName === 'summaries');
        this.relationsTab.classList.toggle('active', tabName === 'gaps');
        this.searchTab.classList.toggle('active', tabName === 'search');
        
        // Load data when switching tabs
        if (tabName === 'summaries') {
            this.loadMemorySummaries();
        } else if (tabName === 'gaps') {
            this.loadMemoryGaps();
        }
    }

    async loadMemorySummaries() {
        if (!this.entitiesList) return;
        
        this.entitiesList.innerHTML = '<div class="memory-loading"><span class="loading-spinner"></span> Loading memory summaries...</div>';
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/memory/summaries`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            this.renderMemorySummaries(data.summaries || []);
        } catch (error) {
            console.error('Failed to load memory summaries:', error);
            this.entitiesList.innerHTML = '<div class="memory-status error">Failed to load memory summaries</div>';
        }
    }

    renderMemorySummaries(summaries) {
        if (!this.entitiesList) return;
        
        if (summaries.length === 0) {
            this.entitiesList.innerHTML = '<div class="memory-empty-state">No memory summaries found. Summaries are automatically created as you chat.</div>';
            return;
        }
        
        this.entitiesList.innerHTML = summaries.map(summary => `
            <div class="entity-item" data-summary-id="${summary.id}">
                <div class="entity-header">
                    <div>
                        <span class="entity-name">${this.escapeHtml(summary.title || 'Untitled Summary')}</span>
                        <span class="entity-type">${this.escapeHtml(summary.summary_type)}</span>
                    </div>
                    <div class="memory-item-actions">
                        <button class="memory-action-btn delete" onclick="chatApp.deleteSummary('${summary.id}')" title="Delete summary">
                            🗑️
                        </button>
                    </div>
                </div>
                <div class="entity-observations">
                    <div class="observation-item">${this.escapeHtml(summary.content)}</div>
                    ${summary.message_count ? `<div class="observation-item">Messages: ${summary.message_count}</div>` : ''}
                    ${summary.created_at ? `<div class="observation-item">Created: ${this.formatDate(summary.created_at)}</div>` : ''}
                </div>
            </div>
        `).join('');
    }

    async loadMemoryGaps() {
        if (!this.relationsList) return;
        
        this.relationsList.innerHTML = '<div class="memory-loading"><span class="loading-spinner"></span> Loading memory gaps...</div>';
        
        // Get current session ID for gaps
        const sessionId = this.currentSessionId;
        if (!sessionId) {
            this.relationsList.innerHTML = '<div class="memory-empty-state">Select a chat session to view memory gaps.</div>';
            return;
        }
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/memory/gaps/${sessionId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            this.renderMemoryGaps(data.gaps || []);
        } catch (error) {
            console.error('Failed to load memory gaps:', error);
            this.relationsList.innerHTML = '<div class="memory-status error">Failed to load memory gaps</div>';
        }
    }

    renderMemoryGaps(gaps) {
        if (!this.relationsList) return;
        
        if (gaps.length === 0) {
            this.relationsList.innerHTML = '<div class="memory-empty-state">No memory gaps detected in this conversation.</div>';
            return;
        }
        
        this.relationsList.innerHTML = gaps.map(gap => `
            <div class="relation-item">
                <div class="relation-header">
                    <div class="relation-description">
                        <strong>Gap Type:</strong> ${this.escapeHtml(gap.gap_type)}
                        <br>
                        <strong>Duration:</strong> ${this.formatDate(gap.gap_start)} - ${this.formatDate(gap.gap_end)}
                    </div>
                    <div class="memory-item-actions">
                        <button class="memory-action-btn delete" onclick="chatApp.deleteGap('${gap.id}')" title="Delete gap">
                            🗑️
                        </button>
                    </div>
                </div>
                ${gap.context_summary ? `<div class="entity-observations"><div class="observation-item">${this.escapeHtml(gap.context_summary)}</div></div>` : ''}
            </div>
        `).join('');
    }

    async createMemorySummary() {
        const title = this.entityNameInput?.value?.trim();
        const content = this.entityObservationsInput?.value?.trim();
        const sessionId = this.currentSessionId;
        
        if (!content) {
            this.showNotification('Please enter summary content', 'error');
            return;
        }
        
        if (!sessionId) {
            this.showNotification('Please select a chat session first', 'error');
            return;
        }
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/memory/summaries`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    session_id: sessionId,
                    summary_type: 'manual',
                    title: title || 'Manual Summary',
                    content: content
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            // Clear inputs
            if (this.entityNameInput) this.entityNameInput.value = '';
            if (this.entityObservationsInput) this.entityObservationsInput.value = '';
            
            // Reload summaries
            this.loadMemorySummaries();
            this.showNotification('Memory summary created successfully', 'success');
        } catch (error) {
            console.error('Failed to create memory summary:', error);
            this.showNotification(`Failed to create summary: ${error.message}`, 'error');
        }
    }

    async searchMemory() {
        const query = this.memorySearchInput?.value?.trim();
        
        if (!query) {
            this.showNotification('Please enter a search query', 'error');
            return;
        }
        
        if (!this.searchResults) return;
        
        this.searchResults.innerHTML = '<div class="memory-loading"><span class="loading-spinner"></span> Searching...</div>';
        
        try {
            const response = await this.authenticatedFetch(`${this.apiBase}/v1/memory/search`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    query: query,
                    session_id: this.currentSessionId,
                    limit: 20
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.renderSearchResults(data.results || []);
        } catch (error) {
            console.error('Failed to search memory:', error);
            this.searchResults.innerHTML = '<div class="memory-status error">Failed to search memory</div>';
        }
    }

    renderSearchResults(results) {
        if (!this.searchResults) return;
        
        if (results.length === 0) {
            this.searchResults.innerHTML = '<div class="memory-empty-state">No results found for your search query.</div>';
            return;
        }
        
        this.searchResults.innerHTML = results.map(result => `
            <div class="search-result-item">
                <div class="entity-header">
                    <div>
                        <span class="entity-name">Message from ${this.escapeHtml(result.role || 'unknown')}</span>
                        <span class="entity-type">Similarity: ${(result.similarity * 100).toFixed(1)}%</span>
                    </div>
                </div>
                <div class="entity-observations">
                    <div class="observation-item">${this.escapeHtml(result.content)}</div>
                    ${result.created_at ? `<div class="observation-item">Date: ${this.formatDate(result.created_at)}</div>` : ''}
                    ${result.session_id ? `<div class="observation-item">Session: ${this.escapeHtml(result.session_id)}</div>` : ''}
                </div>
            </div>
        `).join('');
    }

    clearMemorySearch() {
        if (this.memorySearchInput) {
            this.memorySearchInput.value = '';
        }
        if (this.searchResults) {
            this.searchResults.innerHTML = '<div class="memory-empty-state">Enter a search query to find similar messages.</div>';
        }
    }

    async deleteSummary(summaryId) {
        if (!confirm('Are you sure you want to delete this memory summary?')) {
            return;
        }
        
        this.showNotification('Delete functionality not yet implemented', 'info');
        // TODO: Implement delete summary API endpoint
    }

    async deleteGap(gapId) {
        if (!confirm('Are you sure you want to delete this memory gap?')) {
            return;
        }
        
        this.showNotification('Delete functionality not yet implemented', 'info');
        // TODO: Implement delete gap API endpoint
    }

    async refreshMemory() {
        // Refresh the current active tab
        const activeTab = document.querySelector('.memory-tab-btn.active')?.dataset.tab;
        
        if (activeTab === 'summaries') {
            this.loadMemorySummaries();
        } else if (activeTab === 'gaps') {
            this.loadMemoryGaps();
        }
        
        this.showNotification('Memory data refreshed', 'success');
    }

    async clearAllMemory() {
        this.showNotification('Clear all memory functionality not yet implemented', 'info');
        // TODO: Implement clear all memory functionality
    }

    // Authentication Methods
    attachAuthenticationListeners() {
        // Authentication modal listeners
        if (this.closeAuthBtn) {
            this.closeAuthBtn.addEventListener('click', () => this.hideAuthModal());
        }
        if (this.authModal) {
            this.authModal.addEventListener('click', (e) => {
                if (e.target === this.authModal) {
                    this.hideAuthModal();
                }
            });
        }

        // Form switching listeners
        if (this.showRegisterBtn) {
            this.showRegisterBtn.addEventListener('click', () => this.showRegisterForm());
        }
        if (this.showLoginBtn) {
            this.showLoginBtn.addEventListener('click', () => this.showLoginForm());
        }

        // Form submission listeners
        if (this.loginFormElement) {
            this.loginFormElement.addEventListener('submit', (e) => this.handleLogin(e));
        }
        if (this.registerFormElement) {
            this.registerFormElement.addEventListener('submit', (e) => this.handleRegister(e));
        }

        // User menu listeners
        if (this.userMenu) {
            this.userMenu.addEventListener('click', () => this.toggleUserMenu());
        }
        if (this.logoutBtn) {
            this.logoutBtn.addEventListener('click', () => this.handleLogout());
        }
        if (this.userProfileBtn) {
            this.userProfileBtn.addEventListener('click', () => this.showUserProfile());
        }
        if (this.userSettingsBtn) {
            this.userSettingsBtn.addEventListener('click', () => this.openSettings());
        }

        // Close user menu when clicking outside
        document.addEventListener('click', (e) => {
            if (this.userMenuDropdown && !e.target.closest('.user-menu') && !e.target.closest('.user-menu-dropdown')) {
                this.hideUserMenu();
            }
        });
    }

    checkAuthenticationState() {
        // Check for stored authentication token
        const token = localStorage.getItem('authToken');
        const userData = localStorage.getItem('userData');
        
        if (token && userData) {
            try {
                this.authToken = token;
                this.currentUser = JSON.parse(userData);
                this.isAuthenticated = true;
                this.updateAuthenticationUI();
                this.loadUserData();
            } catch (error) {
                console.error('Failed to parse stored user data:', error);
                this.clearAuthenticationData();
                this.showAuthModal();
            }
        } else {
            this.showAuthModal();
        }
    }

    showAuthModal() {
        if (this.authModal) {
            this.authModal.classList.add('active');
            document.body.style.overflow = 'hidden';
            this.showLoginForm();
        }
    }

    hideAuthModal() {
        if (this.authModal) {
            this.authModal.classList.remove('active');
            document.body.style.overflow = '';
            this.clearFormErrors();
        }
    }

    showLoginForm() {
        if (this.loginForm && this.registerForm) {
            this.loginForm.classList.add('active');
            this.registerForm.classList.remove('active');
            this.clearFormErrors();
        }
    }

    showRegisterForm() {
        if (this.loginForm && this.registerForm) {
            this.loginForm.classList.remove('active');
            this.registerForm.classList.add('active');
            this.clearFormErrors();
        }
    }

    async handleLogin(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const email = formData.get('email');
        const password = formData.get('password');

        if (!this.validateLoginForm(email, password)) {
            return;
        }

        this.setFormLoading(this.loginFormElement, true);

        try {
            const response = await fetch(`${this.apiBase}/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.detail || 'Login failed');
            }

            // Store authentication data
            this.authToken = data.token;
            this.currentUser = data.user;
            this.isAuthenticated = true;

            localStorage.setItem('authToken', data.token);
            localStorage.setItem('userData', JSON.stringify(data.user));

            // Update UI and load user data
            this.updateAuthenticationUI();
            this.hideAuthModal();
            this.loadUserData();
            
            this.showNotification(`Welcome back, ${data.user.username}!`, 'success');

        } catch (error) {
            console.error('Login failed:', error);
            this.showFormError(this.loginFormElement, error.message);
        } finally {
            this.setFormLoading(this.loginFormElement, false);
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const username = formData.get('username');
        const email = formData.get('email');
        const password = formData.get('password');
        const confirmPassword = formData.get('confirmPassword');

        if (!this.validateRegisterForm(username, email, password, confirmPassword)) {
            return;
        }

        this.setFormLoading(this.registerFormElement, true);

        try {
            const response = await fetch(`${this.apiBase}/v1/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, email, password })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.detail || 'Registration failed');
            }

            // Store authentication data
            this.authToken = data.token;
            this.currentUser = data.user;
            this.isAuthenticated = true;

            localStorage.setItem('authToken', data.token);
            localStorage.setItem('userData', JSON.stringify(data.user));

            // Update UI and load user data
            this.updateAuthenticationUI();
            this.hideAuthModal();
            this.loadUserData();
            
            this.showNotification(`Welcome to Ollama Pilot, ${data.user.username}!`, 'success');

        } catch (error) {
            console.error('Registration failed:', error);
            this.showFormError(this.registerFormElement, error.message);
        } finally {
            this.setFormLoading(this.registerFormElement, false);
        }
    }

    async handleLogout() {
        try {
            // Call logout endpoint
            await fetch(`${this.apiBase}/v1/auth/logout`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.authToken}`,
                    'Content-Type': 'application/json',
                }
            });
        } catch (error) {
            console.error('Logout request failed:', error);
            // Continue with local logout even if server request fails
        }

        // Clear authentication data
        this.clearAuthenticationData();
        this.updateAuthenticationUI();
        this.hideUserMenu();
        
        // Clear user-specific data
        this.sessions = [];
        this.projects = [];
        this.currentSessionId = null;
        this.currentProjectId = null;
        this.clearSavedSessionId();
        
        // Reset UI
        this.renderSessions();
        this.renderProjects();
        this.chatMessages.innerHTML = `
            <div class="welcome-section">
                <h1 class="welcome-title">What's on the agenda today?</h1>
            </div>
        `;
        
        // Show authentication modal
        this.showAuthModal();
        this.showNotification('You have been signed out', 'info');
    }

    validateLoginForm(email, password) {
        this.clearFormErrors();
        let isValid = true;

        if (!email || !email.includes('@')) {
            this.showFieldError('login-email', 'Please enter a valid email address');
            isValid = false;
        }

        if (!password || password.length < 6) {
            this.showFieldError('login-password', 'Password must be at least 6 characters');
            isValid = false;
        }

        return isValid;
    }

    validateRegisterForm(username, email, password, confirmPassword) {
        this.clearFormErrors();
        let isValid = true;

        if (!username || username.length < 3) {
            this.showFieldError('register-username', 'Username must be at least 3 characters');
            isValid = false;
        }

        if (!email || !email.includes('@')) {
            this.showFieldError('register-email', 'Please enter a valid email address');
            isValid = false;
        }

        if (!password || password.length < 6) {
            this.showFieldError('register-password', 'Password must be at least 6 characters');
            isValid = false;
        }

        if (password !== confirmPassword) {
            this.showFieldError('register-confirm-password', 'Passwords do not match');
            isValid = false;
        }

        return isValid;
    }

    showFieldError(fieldId, message) {
        const field = document.getElementById(fieldId);
        if (!field) return;

        const formGroup = field.closest('.form-group');
        if (!formGroup) return;

        formGroup.classList.add('error');
        
        let errorElement = formGroup.querySelector('.error-message');
        if (!errorElement) {
            errorElement = document.createElement('div');
            errorElement.className = 'error-message';
            formGroup.appendChild(errorElement);
        }
        
        errorElement.textContent = message;
    }

    showFormError(form, message) {
        let errorElement = form.querySelector('.auth-error');
        if (!errorElement) {
            errorElement = document.createElement('div');
            errorElement.className = 'auth-error';
            form.insertBefore(errorElement, form.firstChild);
        }
        errorElement.textContent = message;
    }

    clearFormErrors() {
        // Clear field errors
        document.querySelectorAll('.form-group.error').forEach(group => {
            group.classList.remove('error');
        });
        document.querySelectorAll('.error-message').forEach(error => {
            error.textContent = '';
        });

        // Clear form errors
        document.querySelectorAll('.auth-error').forEach(error => {
            error.remove();
        });
    }

    setFormLoading(form, loading) {
        const submitBtn = form.querySelector('button[type="submit"]');
        if (!submitBtn) return;

        if (loading) {
            form.classList.add('loading');
            submitBtn.disabled = true;
        } else {
            form.classList.remove('loading');
            submitBtn.disabled = false;
        }
    }

    updateAuthenticationUI() {
        if (this.isAuthenticated && this.currentUser) {
            // Update user menu
            if (this.userMenu) {
                this.userMenu.classList.remove('unauthenticated');
                this.userMenu.classList.add('authenticated');
            }
            
            // Update user info in sidebar
            const userName = document.querySelector('.user-name');
            const userStatus = document.querySelector('.user-status');
            if (userName) userName.textContent = this.currentUser.username;
            if (userStatus) userStatus.textContent = 'Online';
            
            // Update user menu dropdown
            if (this.userMenuName) this.userMenuName.textContent = this.currentUser.username;
            if (this.userMenuEmail) this.userMenuEmail.textContent = this.currentUser.email;
            
        } else {
            // Update user menu for unauthenticated state
            if (this.userMenu) {
                this.userMenu.classList.remove('authenticated');
                this.userMenu.classList.add('unauthenticated');
            }
            
            // Update user info in sidebar
            const userName = document.querySelector('.user-name');
            const userStatus = document.querySelector('.user-status');
            if (userName) userName.textContent = 'Guest';
            if (userStatus) userStatus.textContent = 'Not signed in';
        }
    }

    toggleUserMenu() {
        if (!this.isAuthenticated) {
            this.showAuthModal();
            return;
        }

        if (this.userMenuDropdown) {
            const isActive = this.userMenuDropdown.classList.contains('active');
            if (isActive) {
                this.hideUserMenu();
            } else {
                this.showUserMenu();
            }
        }
    }

    showUserMenu() {
        if (this.userMenuDropdown) {
            this.userMenuDropdown.classList.add('active');
        }
    }

    hideUserMenu() {
        if (this.userMenuDropdown) {
            this.userMenuDropdown.classList.remove('active');
        }
    }

    showUserProfile() {
        this.hideUserMenu();
        this.showNotification('User profile functionality coming soon', 'info');
    }

    clearAuthenticationData() {
        this.isAuthenticated = false;
        this.currentUser = null;
        this.authToken = null;
        localStorage.removeItem('authToken');
        localStorage.removeItem('userData');
    }

    loadUserData() {
        if (this.isAuthenticated) {
            // Load user-specific data
            this.loadSessions();
            this.loadModels();
            this.loadProjects();
        }
    }

    // Override API methods to include authentication headers
    async authenticatedFetch(url, options = {}) {
        if (this.authToken) {
            options.headers = {
                ...options.headers,
                'Authorization': `Bearer ${this.authToken}`
            };
        }
        
        const response = await fetch(url, options);
        
        // Handle authentication errors
        if (response.status === 401) {
            this.clearAuthenticationData();
            this.updateAuthenticationUI();
            this.showAuthModal();
            throw new Error('Authentication required');
        }
        
        return response;
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