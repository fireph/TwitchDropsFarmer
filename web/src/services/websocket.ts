import type { WSMessage, MinerStatus, Config } from '@/types'
import { useMinerStore } from '@/stores/miner'

class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private isConnected = false

  connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.isConnected = true
      this.reconnectAttempts = 0
    }

    this.ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      this.isConnected = false
      this.reconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  private handleMessage(message: WSMessage) {
    const minerStore = useMinerStore()

    switch (message.type) {
      case 'status_update':
        // The enhanced status data now includes active_drops and total_progress
        const statusData = message.data as any
        
        // Update the status with enhanced data
        minerStore.updateStatus({
          is_running: statusData.is_running,
          current_stream: statusData.current_stream,
          current_campaign: statusData.current_campaign,
          current_progress: statusData.current_progress,
          total_campaigns: statusData.total_campaigns,
          claimed_drops: statusData.claimed_drops,
          last_update: statusData.last_update,
          next_switch: statusData.next_switch,
          error_message: statusData.error_message,
          active_drops: statusData.active_drops || []
        } as MinerStatus)
        
        break
      case 'config_update':
        minerStore.updateConfig(message.data as Config)
        break
      case 'error':
        console.error('WebSocket error message:', message.data.message)
        break
      default:
        console.warn('Unknown WebSocket message type:', message.type)
    }
  }

  private reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`Attempting to reconnect WebSocket (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
      
      setTimeout(() => {
        this.connect()
      }, this.reconnectDelay * this.reconnectAttempts)
    } else {
      console.error('Max WebSocket reconnection attempts reached')
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message: any) {
    if (this.ws && this.isConnected) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket not connected, unable to send message')
    }
  }

  getConnectionStatus() {
    return this.isConnected
  }
}

export const webSocketService = new WebSocketService()