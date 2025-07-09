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

        // Remove game buttons (using event delegation)
        document.addEventListener('click', async (e) => {
            if (e.target.closest('.remove-game-btn')) {
                const button = e.target.closest('.remove-game-btn');
                const gameName = button.getAttribute('data-game-name');
                if (gameName) {
                    await this.removeGame(gameName);
                }
            }
        });

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

    async addGame() {
        const gameInput = document.getElementById('game-input');
        if (!gameInput) return;

        const gameName = gameInput.value.trim();
        if (!gameName) return;

        // Check if game already exists
        const existingGame = this.priorityGames.find(game => 
            (typeof game === 'string' && game === gameName) ||
            (typeof game === 'object' && game.name === gameName)
        );
        
        if (existingGame) {
            if (window.app) {
                window.app.showNotification('Game already in the list!', 'warning');
            }
            return;
        }

        try {
            // Show loading state
            const addBtn = document.getElementById('add-game-btn');
            if (addBtn) {
                addBtn.disabled = true;
                addBtn.textContent = 'Adding...';
            }

            // Add game with slug resolution via API
            const response = await this.api.addGame(gameName, true); // true for priority games
            
            if (response.success) {
                // Add the game with resolved slug to the list
                this.priorityGames.push(response.game);
                gameInput.value = '';
                
                // Update UI
                this.updateGamesList();
                
                if (window.app) {
                    window.app.showNotification(`Added ${gameName} to farming list`, 'success');
                }
            }
        } catch (error) {
            console.error('Failed to add game:', error);
            if (window.app) {
                window.app.showNotification(`Failed to add game: ${error.message}`, 'error');
            }
        } finally {
            // Reset button state
            const addBtn = document.getElementById('add-game-btn');
            if (addBtn) {
                addBtn.disabled = false;
                addBtn.textContent = 'Add Game';
            }
        }
    }

    async removeGame(gameName) {
        console.log('removeGame called with:', gameName);
        
        const index = this.priorityGames.findIndex(game => 
            (typeof game === 'string' && game === gameName) ||
            (typeof game === 'object' && game.name === gameName)
        );
        
        console.log('Found game at index:', index);
        
        if (index > -1) {
            try {
                // Remove from local array
                this.priorityGames.splice(index, 1);
                console.log('Removed from local array, new length:', this.priorityGames.length);
                
                // Update the UI immediately
                this.updateGamesList();
                console.log('Updated games list UI');
                
                // Save the updated settings to persist the change
                await this.saveSettings();
                console.log('Saved settings successfully');
                
                if (window.app) {
                    window.app.showNotification(`Removed ${gameName} from farming list`, 'success');
                }
            } catch (error) {
                console.error('Failed to remove game:', error);
                if (window.app) {
                    window.app.showNotification(`Failed to remove game: ${error.message}`, 'error');
                }
                // Reload settings to restore previous state
                await this.loadSettings();
                this.updateGamesList();
            }
        } else {
            console.log('Game not found in priorityGames array');
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

        gamesList.innerHTML = this.priorityGames.map(game => {
            const gameName = typeof game === 'string' ? game : game.name;
            const slug = typeof game === 'object' ? game.slug : '';
            const id = typeof game === 'object' ? game.id : '';
            
            // HTML escape function for safe display
            const escapeHtml = (text) => {
                const div = document.createElement('div');
                div.textContent = text;
                return div.innerHTML;
            };
            
            return `
            <div class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 rounded-lg px-3 py-2">
                <div class="flex flex-col">
                    <span class="text-gray-900 dark:text-white font-medium">${escapeHtml(gameName)}</span>
                    ${slug ? `<span class="text-xs text-gray-500 dark:text-gray-400">Slug: ${escapeHtml(slug)}</span>` : ''}
                    ${id ? `<span class="text-xs text-gray-500 dark:text-gray-400">ID: ${escapeHtml(id)}</span>` : ''}
                </div>
                <button
                    class="remove-game-btn text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 p-1"
                    data-game-name="${escapeHtml(gameName)}"
                    title="Remove game"
                >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
            </div>
            `;
        }).join('');
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

            // Don't update priority_games here since they're now managed via /api/games/add
            // Only save other settings
            const settingsToSave = {
                ...this.settings,
                priority_games: this.priorityGames // Keep current games as-is
            };

            // Save to server
            await this.api.updateSettings(settingsToSave);
            
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