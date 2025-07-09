// API client for Twitch Drops Farmer
class API {
    constructor(baseURL = '') {
        this.baseURL = baseURL;
        this.defaultHeaders = {
            'Content-Type': 'application/json',
        };
    }

    async request(url, options = {}) {
        const config = {
            headers: { ...this.defaultHeaders, ...options.headers },
            ...options
        };

        try {
            const response = await fetch(`${this.baseURL}${url}`, config);
            
            // Handle different response types
            const contentType = response.headers.get('content-type');
            let data;
            
            if (contentType && contentType.includes('application/json')) {
                data = await response.json();
            } else {
                data = await response.text();
            }

            if (!response.ok) {
                throw new Error(data.error || `HTTP error! status: ${response.status}`);
            }

            return data;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    async get(url, headers = {}) {
        return this.request(url, { method: 'GET', headers });
    }

    async post(url, data = null, headers = {}) {
        const options = { method: 'POST', headers };
        if (data) {
            options.body = JSON.stringify(data);
        }
        return this.request(url, options);
    }

    async put(url, data = null, headers = {}) {
        const options = { method: 'PUT', headers };
        if (data) {
            options.body = JSON.stringify(data);
        }
        return this.request(url, options);
    }

    async delete(url, headers = {}) {
        return this.request(url, { method: 'DELETE', headers });
    }

    // Authentication endpoints
    async getAuthURL() {
        return this.get('/api/auth/url');
    }

    async handleAuthCallback(data) {
        return this.post('/api/auth/callback', data);
    }

    async logout() {
        return this.post('/api/auth/logout');
    }

    async getAuthStatus() {
        return this.get('/api/auth/status');
    }

    // User endpoints
    async getUserProfile() {
        return this.get('/api/user/profile');
    }

    async getUserInventory() {
        return this.get('/api/user/inventory');
    }

    // Campaign endpoints
    async getCampaigns() {
        return this.get('/api/campaigns/');
    }

    async getCampaign(id) {
        return this.get(`/api/campaigns/${id}`);
    }

    async getCampaignDrops(id) {
        return this.get(`/api/campaigns/${id}/drops`);
    }

    // Miner endpoints
    async getMinerStatus() {
        return this.get('/api/miner/status');
    }

    async getCurrentDrop() {
        return this.get('/api/miner/current-drop');
    }

    async getDropProgress() {
        return this.get('/api/miner/progress');
    }

    async startMiner() {
        return this.post('/api/miner/start');
    }

    async stopMiner() {
        return this.post('/api/miner/stop');
    }

    // Settings endpoints
    async getSettings() {
        return this.get('/api/settings/');
    }

    async updateSettings(settings) {
        return this.put('/api/settings/', settings);
    }

    // Game management endpoints
    async addGame(gameName, toPriority = true) {
        return this.post('/api/games/add', {
            game_name: gameName,
            to_priority: toPriority
        });
    }

    // Stream endpoints
    async getStreamsForGame(gameId, limit = 10) {
        return this.get(`/api/streams/game/${gameId}?limit=${limit}`);
    }

    async getCurrentStream() {
        return this.get('/api/streams/current');
    }
}