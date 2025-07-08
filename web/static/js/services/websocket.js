// WebSocket client for real-time updates
class WebSocketClient {
    constructor() {
        this.ws = null;
        this.reconnectInterval = 5000; // 5 seconds
        this.maxReconnectAttempts = 10;
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        this.isManualClose = false;
        
        // Event handlers
        this.onOpen = null;
        this.onClose = null;
        this.onError = null;
        this.onMessage = null;
        this.onStatusUpdate = null;
        this.onConnectionChange = null;
    }

    connect() {
        if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
            return;
        }

        this.isConnecting = true;
        this.isManualClose = false;

        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = (event) => {
                console.log('WebSocket connected');
                this.isConnecting = false;
                this.reconnectAttempts = 0;
                this.updateConnectionStatus('connected');
                
                if (this.onOpen) {
                    this.onOpen(event);
                }
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleMessage(data);
                    
                    if (this.onMessage) {
                        this.onMessage(data);
                    }
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected', event.code, event.reason);
                this.isConnecting = false;
                this.ws = null;
                
                if (!this.isManualClose) {
                    this.updateConnectionStatus('disconnected');
                    this.scheduleReconnect();
                }
                
                if (this.onClose) {
                    this.onClose(event);
                }
            };

            this.ws.onerror = (event) => {
                console.error('WebSocket error:', event);
                this.isConnecting = false;
                this.updateConnectionStatus('error');
                
                if (this.onError) {
                    this.onError(event);
                }
            };

        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.isConnecting = false;
            this.scheduleReconnect();
        }
    }

    disconnect() {
        this.isManualClose = true;
        
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        
        this.updateConnectionStatus('disconnected');
    }

    scheduleReconnect() {
        if (this.isManualClose || this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Max reconnect attempts reached or manual close');
            return;
        }

        this.reconnectAttempts++;
        this.updateConnectionStatus('connecting');
        
        console.log(`Reconnecting in ${this.reconnectInterval}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        
        setTimeout(() => {
            if (!this.isManualClose) {
                this.connect();
            }
        }, this.reconnectInterval);
    }

    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        } else {
            console.warn('WebSocket is not connected. Message not sent:', data);
        }
    }

    handleMessage(data) {
        switch (data.type) {
            case 'status_update':
                if (this.onStatusUpdate) {
                    this.onStatusUpdate(data.data);
                }
                break;
            
            case 'notification':
                this.showNotification(data.data);
                break;
            
            case 'error':
                console.error('WebSocket error message:', data.data);
                this.showNotification({
                    message: data.data.message || 'An error occurred',
                    type: 'error'
                });
                break;
            
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    }

    updateConnectionStatus(status) {
        // Update connection indicator in UI
        const indicator = document.querySelector('.connection-status');
        if (indicator) {
            indicator.className = `connection-status ${status}`;
            
            const statusText = {
                connected: 'Connected',
                disconnected: 'Disconnected',
                connecting: 'Connecting...',
                error: 'Connection Error'
            };
            
            indicator.innerHTML = `
                <div class="flex items-center space-x-2">
                    <div class="w-2 h-2 rounded-full ${this.getStatusColor(status)}"></div>
                    <span>${statusText[status] || status}</span>
                </div>
            `;
        }

        if (this.onConnectionChange) {
            this.onConnectionChange(status);
        }
    }

    getStatusColor(status) {
        const colors = {
            connected: 'bg-green-500',
            disconnected: 'bg-red-500',
            connecting: 'bg-yellow-500',
            error: 'bg-red-500'
        };
        
        return colors[status] || 'bg-gray-500';
    }

    showNotification(data) {
        if (window.app) {
            window.app.showNotification(data.message, data.type);
        }
    }

    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN;
    }

    getReadyState() {
        if (!this.ws) return 'CLOSED';
        
        const states = {
            0: 'CONNECTING',
            1: 'OPEN',
            2: 'CLOSING',
            3: 'CLOSED'
        };
        
        return states[this.ws.readyState] || 'UNKNOWN';
    }
}