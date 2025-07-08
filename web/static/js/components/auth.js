// Authentication component
class Auth {
    constructor(api) {
        this.api = api;
        this.user = null;
        this.isAuthenticated = false;
    }

    async init() {
        try {
            const status = await this.api.getAuthStatus();
            this.isAuthenticated = status.is_logged_in;
            this.user = status.user;
            
            this.updateUI();
        } catch (error) {
            console.error('Failed to check auth status:', error);
            this.isAuthenticated = false;
            this.user = null;
            this.updateUI();
        }
    }

    async login() {
        try {
            const deviceData = await this.api.getAuthURL();
            
            // Store device code for polling
            sessionStorage.setItem('device_code', deviceData.device_code);
            
            // Use the verification_uri directly - Twitch already includes the device code
            const activationUrl = deviceData.verification_uri;
            
            // Show device code to user and open activation page
            this.showDeviceCode(deviceData, activationUrl);
            
            // Start polling for token
            this.startTokenPolling(deviceData);
            
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        }
    }

    showDeviceCode(deviceData, activationUrl) {
        if (window.app) {
            window.app.showNotification(
                `Go to ${activationUrl} and enter code: ${deviceData.user_code}`,
                'info'
            );
        }
        
        // Automatically open the activation page
        window.open(activationUrl, '_blank');
    }

    async startTokenPolling(deviceData) {
        try {
            // Start the polling process
            await this.api.handleAuthCallback({ device_code: deviceData.device_code });
            
            // Poll for authentication status
            const pollInterval = setInterval(async () => {
                try {
                    const status = await this.api.getAuthStatus();
                    
                    if (status.is_logged_in) {
                        clearInterval(pollInterval);
                        this.isAuthenticated = true;
                        this.user = status.user;
                        this.updateUI();
                        
                        if (window.app) {
                            window.app.showNotification('Successfully logged in!', 'success');
                            window.app.showDashboard();
                            window.app.initWebSocket();
                        }
                    }
                } catch (error) {
                    console.error('Polling error:', error);
                }
            }, deviceData.interval * 1000);
            
            // Stop polling after expiry time
            setTimeout(() => {
                clearInterval(pollInterval);
            }, deviceData.expires_in * 1000);
            
        } catch (error) {
            console.error('Token polling failed:', error);
            throw error;
        }
    }

    async logout() {
        try {
            await this.api.logout();
            this.isAuthenticated = false;
            this.user = null;
            this.updateUI();
        } catch (error) {
            console.error('Logout failed:', error);
            throw error;
        }
    }

    updateUI() {
        const loginBtn = document.getElementById('login-btn');
        const userInfo = document.getElementById('user-info');
        const userAvatar = document.getElementById('user-avatar');
        const userName = document.getElementById('user-name');

        if (this.isAuthenticated && this.user) {
            // Hide login button
            if (loginBtn) {
                loginBtn.style.display = 'none';
            }
            
            // Show user info
            if (userInfo) {
                userInfo.classList.remove('hidden');
            }
            
            // Update user details
            if (userAvatar && this.user.profile_image_url) {
                userAvatar.src = this.user.profile_image_url;
                userAvatar.alt = this.user.display_name;
            }
            
            if (userName) {
                userName.textContent = this.user.display_name;
            }
        } else {
            // Show login button
            if (loginBtn) {
                loginBtn.style.display = 'block';
            }
            
            // Hide user info
            if (userInfo) {
                userInfo.classList.add('hidden');
            }
        }
    }

    getUser() {
        return this.user;
    }

    isLoggedIn() {
        return this.isAuthenticated;
    }
}