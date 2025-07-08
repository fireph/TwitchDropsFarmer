// Settings component for managing drop farming configuration
class Settings {
    constructor(api) {
        this.api = api;
        this.priorityGames = [];
        this.settings = {};
    }

    async init() {
        await this.loadSettings();
        this.setupEventListeners();
        this.updateUI();
    }

    async loadSettings() {
        try {
            this.settings = await this.api.getSettings();
            this.priorityGames = this.settings.priority_games || [];
        } catch (error) {
            console.error('Failed to load settings:', error);
            this.settings = {
                priority_games: [],
                exclude_games: [],
                watch_unlisted: true,
                claim_drops: true
            };
            this.priorityGames = [];
        }
    }

    setupEventListeners() {
        // Add game button
        const addGameBtn = document.getElementById('add-game-btn');
        const gameInput = document.getElementById('game-input');
        
        if (addGameBtn) {
            addGameBtn.addEventListener('click', () => this.addGame());
        }
        
        if (gameInput) {
            gameInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.addGame();
                }
            });
        }

        // Save settings button
        const saveBtn = document.getElementById('save-settings-btn');
        if (saveBtn) {
            saveBtn.addEventListener('click', () => this.saveSettings());
        }

        // Toggle switches
        const autoClaimToggle = document.getElementById('auto-claim-toggle');
        const watchUnlistedToggle = document.getElementById('watch-unlisted-toggle');

        if (autoClaimToggle) {
            autoClaimToggle.addEventListener('change', () => {
                this.settings.claim_drops = autoClaimToggle.checked;
            });
        }

        if (watchUnlistedToggle) {
            watchUnlistedToggle.addEventListener('change', () => {
                this.settings.watch_unlisted = watchUnlistedToggle.checked;
            });
        }
    }

    addGame() {
        const gameInput = document.getElementById('game-input');
        if (!gameInput) return;

        const gameName = gameInput.value.trim();
        if (!gameName) return;

        // Check if game already exists
        if (this.priorityGames.includes(gameName)) {
            if (window.app) {
                window.app.showNotification('Game already in the list!', 'warning');
            }
            return;
        }

        // Add game to the list
        this.priorityGames.push(gameName);
        gameInput.value = '';
        
        // Update UI
        this.updateGamesList();
        
        if (window.app) {
            window.app.showNotification(`Added ${gameName} to farming list`, 'success');
        }
    }

    removeGame(gameName) {
        const index = this.priorityGames.indexOf(gameName);
        if (index > -1) {
            this.priorityGames.splice(index, 1);
            this.updateGamesList();
            
            if (window.app) {
                window.app.showNotification(`Removed ${gameName} from farming list`, 'success');
            }
        }
    }

    updateGamesList() {
        const gamesList = document.getElementById('priority-games-list');
        if (!gamesList) return;

        if (this.priorityGames.length === 0) {
            gamesList.innerHTML = `
                <div class="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
                    No games configured. Add games above to start farming drops.
                </div>
            `;
            return;
        }

        gamesList.innerHTML = this.priorityGames.map(game => `
            <div class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 rounded-lg px-3 py-2">
                <span class="text-gray-900 dark:text-white font-medium">${game}</span>
                <button
                    onclick="window.settingsComponent.removeGame('${game}')"
                    class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 p-1"
                    title="Remove game"
                >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
        `).join('');
    }

    updateUI() {
        // Update games list
        this.updateGamesList();

        // Update toggle switches
        const autoClaimToggle = document.getElementById('auto-claim-toggle');
        const watchUnlistedToggle = document.getElementById('watch-unlisted-toggle');

        if (autoClaimToggle) {
            autoClaimToggle.checked = this.settings.claim_drops !== false;
        }

        if (watchUnlistedToggle) {
            watchUnlistedToggle.checked = this.settings.watch_unlisted !== false;
        }
    }

    async saveSettings() {
        try {
            const saveBtn = document.getElementById('save-settings-btn');
            if (saveBtn) {
                saveBtn.classList.add('btn-loading');
                saveBtn.disabled = true;
            }

            // Update settings with current values
            this.settings.priority_games = this.priorityGames;

            // Save to server
            await this.api.updateSettings(this.settings);
            
            if (window.app) {
                window.app.showNotification('Settings saved successfully!', 'success');
            }
        } catch (error) {
            console.error('Failed to save settings:', error);
            if (window.app) {
                window.app.showNotification('Failed to save settings: ' + error.message, 'error');
            }
        } finally {
            const saveBtn = document.getElementById('save-settings-btn');
            if (saveBtn) {
                saveBtn.classList.remove('btn-loading');
                saveBtn.disabled = false;
            }
        }
    }

    getSettings() {
        return this.settings;
    }

    getPriorityGames() {
        return this.priorityGames;
    }
}