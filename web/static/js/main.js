// Main application JavaScript
class TwitchDropsFarmer {
    constructor() {
        this.api = new API();
        this.websocket = null;
        this.auth = new Auth(this.api);
        this.dashboard = null;
        this.settings = null;
        
        this.init();
    }

    async init() {
        // Initialize theme
        this.initTheme();
        
        // Initialize components
        await this.initComponents();
        
        // Initialize WebSocket if authenticated
        if (this.auth.isAuthenticated) {
            this.initWebSocket();
        }
        
        // Set up event listeners
        this.setupEventListeners();
    }

    async initComponents() {
        // Initialize authentication
        await this.auth.init();
        
        // Initialize dashboard if authenticated
        if (this.auth.isAuthenticated) {
            this.dashboard = new Dashboard(this.api);
            await this.dashboard.init();
            
            // Initialize settings component
            this.settings = new Settings(this.api);
            await this.settings.init();
            
            // Make settings component globally accessible
            window.settingsComponent = this.settings;
            
            this.showDashboard();
        } else {
            this.showLoginRequired();
        }
    }

    initTheme() {
        const themeToggle = document.getElementById('theme-toggle');
        const savedTheme = localStorage.getItem('theme');
        const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        const currentTheme = savedTheme || systemTheme;

        // Apply theme
        if (currentTheme === 'dark') {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }

        // Theme toggle handler
        if (themeToggle) {
            themeToggle.addEventListener('click', () => {
                const isDark = document.documentElement.classList.contains('dark');
                
                if (isDark) {
                    document.documentElement.classList.remove('dark');
                    localStorage.setItem('theme', 'light');
                } else {
                    document.documentElement.classList.add('dark');
                    localStorage.setItem('theme', 'dark');
                }
            });
        }
    }

    initWebSocket() {
        this.websocket = new WebSocketClient();
        this.websocket.onStatusUpdate = (status) => {
            if (this.dashboard) {
                this.dashboard.updateStatus(status);
            }
        };
        this.websocket.connect();
    }

    setupEventListeners() {
        // Login buttons
        const loginBtn = document.getElementById('login-btn');
        const loginBtnCenter = document.getElementById('login-btn-center');
        
        if (loginBtn) {
            loginBtn.addEventListener('click', () => this.handleLogin());
        }
        
        if (loginBtnCenter) {
            loginBtnCenter.addEventListener('click', () => this.handleLogin());
        }

        // Logout button
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.handleLogout());
        }

        // Miner controls
        const startMinerBtn = document.getElementById('start-miner');
        const stopMinerBtn = document.getElementById('stop-miner');
        
        if (startMinerBtn) {
            startMinerBtn.addEventListener('click', () => this.startMiner());
        }
        
        if (stopMinerBtn) {
            stopMinerBtn.addEventListener('click', () => this.stopMiner());
        }
    }

    async handleLogin() {
        try {
            await this.auth.login();
        } catch (error) {
            this.showNotification('Login failed: ' + error.message, 'error');
        }
    }

    async handleLogout() {
        try {
            await this.auth.logout();
            
            // Disconnect WebSocket
            if (this.websocket) {
                this.websocket.disconnect();
                this.websocket = null;
            }
            
            // Show login required
            this.showLoginRequired();
            
            this.showNotification('Logged out successfully', 'success');
        } catch (error) {
            this.showNotification('Logout failed: ' + error.message, 'error');
        }
    }

    async startMiner() {
        try {
            const startBtn = document.getElementById('start-miner');
            if (startBtn) {
                startBtn.classList.add('btn-loading');
                startBtn.disabled = true;
            }

            await this.api.post('/api/miner/start');
            this.showNotification('Miner started successfully', 'success');
        } catch (error) {
            this.showNotification('Failed to start miner: ' + error.message, 'error');
        } finally {
            const startBtn = document.getElementById('start-miner');
            if (startBtn) {
                startBtn.classList.remove('btn-loading');
                startBtn.disabled = false;
            }
        }
    }

    async stopMiner() {
        try {
            const stopBtn = document.getElementById('stop-miner');
            if (stopBtn) {
                stopBtn.classList.add('btn-loading');
                stopBtn.disabled = true;
            }

            await this.api.post('/api/miner/stop');
            this.showNotification('Miner stopped successfully', 'success');
        } catch (error) {
            this.showNotification('Failed to stop miner: ' + error.message, 'error');
        } finally {
            const stopBtn = document.getElementById('stop-miner');
            if (stopBtn) {
                stopBtn.classList.remove('btn-loading');
                stopBtn.disabled = false;
            }
        }
    }

    showDashboard() {
        document.getElementById('login-required').classList.add('hidden');
        document.getElementById('dashboard').classList.remove('hidden');
    }

    showLoginRequired() {
        document.getElementById('dashboard').classList.add('hidden');
        document.getElementById('login-required').classList.remove('hidden');
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        const icon = this.getNotificationIcon(type);
        
        notification.innerHTML = `
            <div class="p-4">
                <div class="flex">
                    <div class="flex-shrink-0">
                        ${icon}
                    </div>
                    <div class="ml-3">
                        <p class="text-sm font-medium text-gray-800 dark:text-gray-200">
                            ${message}
                        </p>
                    </div>
                    <div class="ml-auto pl-3">
                        <div class="-mx-1.5 -my-1.5">
                            <button type="button" class="inline-flex bg-white dark:bg-gray-800 rounded-md p-1.5 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500" onclick="this.closest('.notification').remove()">
                                <span class="sr-only">Dismiss</span>
                                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(notification);
        
        // Show notification
        setTimeout(() => {
            notification.classList.add('show');
        }, 100);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 5000);
    }

    getNotificationIcon(type) {
        const icons = {
            success: `<svg class="h-5 w-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                      </svg>`,
            error: `<svg class="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                     <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z"></path>
                   </svg>`,
            warning: `<svg class="h-5 w-5 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z"></path>
                      </svg>`,
            info: `<svg class="h-5 w-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                     <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                   </svg>`
        };
        
        return icons[type] || icons.info;
    }
}

// Utility functions
function formatTime(date) {
    if (!date) return 'N/A';
    
    const now = new Date();
    const target = new Date(date);
    const diff = target - now;
    
    if (diff < 0) return 'Completed';
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    } else {
        return `${minutes}m`;
    }
}

function formatProgress(current, total) {
    if (!total) return 0;
    return Math.min(100, (current / total) * 100);
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new TwitchDropsFarmer();
});