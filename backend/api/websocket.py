from __future__ import annotations

import logging
from flask_socketio import SocketIO, emit, disconnect
from flask import request

logger = logging.getLogger("TwitchDrops.websocket")

# Global instances (will be injected by main app)
miner_instance = None

def init_websocket(socketio: SocketIO, miner):
    """Initialize WebSocket handlers"""
    global miner_instance
    miner_instance = miner
    
    @socketio.on('connect')
    def handle_connect():
        """Handle client connection"""
        logger.info(f"Client connected: {request.sid}")
        
        # Add client to miner's WebSocket clients
        if miner_instance:
            # Create a simple client wrapper
            class SocketIOClient:
                def __init__(self, sid):
                    self.sid = sid
                    
                def emit(self, event, data):
                    socketio.emit(event, data, room=self.sid)
            
            client = SocketIOClient(request.sid)
            miner_instance.add_ws_client(client)
            
            # Send initial status
            status = miner_instance.get_status()
            emit('message', {
                "type": "status",
                "data": {
                    "isRunning": status.is_running,
                    "currentGame": {
                        "id": status.current_game.id,
                        "name": status.current_game.name,
                        "displayName": status.current_game.display_name,
                        "boxArtURL": status.current_game.box_art_url,
                    } if status.current_game else None,
                    "currentStream": {
                        "id": status.current_stream.id,
                        "userLogin": status.current_stream.user_login,
                        "userName": status.current_stream.user_name,
                        "title": status.current_stream.title,
                        "viewerCount": status.current_stream.viewer_count,
                    } if status.current_stream else None,
                    "watchDuration": status.watch_duration,
                    "totalWatched": status.total_watched,
                    "lastUpdate": status.last_update,
                }
            })

    @socketio.on('disconnect')
    def handle_disconnect():
        """Handle client disconnection"""
        logger.info(f"Client disconnected: {request.sid}")
        
        # Remove client from miner's WebSocket clients
        if miner_instance:
            # Find and remove the client
            for client in list(miner_instance.ws_clients):
                if hasattr(client, 'sid') and client.sid == request.sid:
                    miner_instance.remove_ws_client(client)
                    break

    @socketio.on('ping')
    def handle_ping():
        """Handle ping from client"""
        emit('pong')

    @socketio.on('error')
    def handle_error(error):
        """Handle WebSocket errors"""
        logger.error(f"WebSocket error from {request.sid}: {error}")