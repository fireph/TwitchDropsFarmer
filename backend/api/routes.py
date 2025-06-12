from __future__ import annotations

import asyncio
import logging
from datetime import datetime, timezone, timedelta
from flask import Blueprint, request, jsonify
from typing import Dict, Any

from core.twitch import Twitch
from core.storage import Storage, Game, Settings
from core.miner import Miner

logger = logging.getLogger("TwitchDrops")

# Global instances (will be injected by main app)
twitch_client: Twitch = None
storage: Storage = None
miner: Miner = None

def init_routes(twitch: Twitch, store: Storage, miner_instance: Miner):
    """Initialize route dependencies"""
    global twitch_client, storage, miner
    twitch_client = twitch
    storage = store
    miner = miner_instance

# Create blueprint
api = Blueprint('api', __name__, url_prefix='/api/v1')

# Authentication endpoints
@api.route('/auth/url', methods=['GET'])
def get_auth_url():
    """Get device authorization URL"""
    try:
        auth_data = twitch_client.start_device_auth()
        return jsonify(auth_data)
    except Exception as e:
        logger.error(f"Failed to start device auth: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/auth/callback', methods=['POST'])
def handle_auth_callback():
    """Handle auth callback (start polling)"""
    try:
        data = request.get_json()
        device_code = data.get('deviceCode')  # Frontend sends 'deviceCode'
        interval = data.get('interval', 5)
        
        if not device_code:
            return jsonify({"error": "deviceCode is required"}), 400
        
        # Start polling in background
        def poll_for_token():
            logger.info(f"Starting token polling for device code: {device_code[:10]}...")
            token = twitch_client.poll_device_auth(device_code, interval)
            if token:
                logger.info("Token received, saving auth data...")
                # Save auth data
                from core.storage import AuthData
                expires_at = datetime.now(timezone.utc) + timedelta(hours=4)
                auth_data = AuthData(
                    access_token=token,
                    expires_at=expires_at.isoformat(),  # Proper ISO format
                    updated_at=datetime.now(timezone.utc).isoformat()
                )
                storage.save_auth(auth_data)
                logger.info("Authentication successful and saved!")
            else:
                logger.warning("Token polling completed but no token received")
        
        # Run in background thread
        import threading
        threading.Thread(target=poll_for_token, daemon=True).start()
        
        return jsonify({"status": "polling"})
    except Exception as e:
        logger.error(f"Auth callback error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/auth/status', methods=['GET'])
def get_auth_status():
    """Get authentication status"""
    try:
        auth_data = storage.get_auth()
        is_authenticated = storage.is_authenticated()
        
        # Debug logging
        logger.info(f"Auth status check - auth_data exists: {auth_data is not None}")
        if auth_data:
            logger.info(f"Auth data - has token: {bool(auth_data.access_token)}, expires_at: {auth_data.expires_at}")
        logger.info(f"is_authenticated result: {is_authenticated}")
        
        return jsonify({
            "authenticated": is_authenticated,
            "username": auth_data.username if auth_data else "",
            "expiresAt": auth_data.expires_at if auth_data else "",
            "hasToken": bool(auth_data and auth_data.access_token) if auth_data else False,
            "debug": f"Auth exists: {auth_data is not None}, Token exists: {bool(auth_data and auth_data.access_token)}"
        })
    except Exception as e:
        logger.error(f"Auth status error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/auth/logout', methods=['DELETE'])
def logout():
    """Logout user"""
    try:
        storage.clear_auth()
        twitch_client.auth.clear()
        return jsonify({"status": "logged out"})
    except Exception as e:
        logger.error(f"Logout error: {e}")
        return jsonify({"error": str(e)}), 500

# Games management endpoints
@api.route('/games', methods=['GET'])
def get_games():
    """Get games in priority order"""
    try:
        games_dict = storage.get_games()
        settings = storage.get_settings()
        
        # Return games in priority order
        ordered_games = []
        for game_id in settings.games:
            if game_id in games_dict:
                game = games_dict[game_id]
                ordered_games.append({
                    "id": game.id,
                    "name": game.name,
                    "displayName": game.display_name,
                    "boxArtURL": game.box_art_url,
                })
        
        return jsonify({"games": ordered_games})
    except Exception as e:
        logger.error(f"Get games error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/games', methods=['POST'])
def add_game():
    """Add a game to the watch list"""
    try:
        data = request.get_json()
        game_name = data.get('name', '').strip()
        
        if not game_name:
            return jsonify({"error": "Game name is required"}), 400
        
        # Create a simple game object (slug resolution would happen in real impl)
        game_id = game_name.lower().replace(' ', '-').replace("'", "")
        game = Game(
            id=game_id,
            name=game_name,
            display_name=game_name,
            box_art_url=""
        )
        
        # Add to storage
        storage.add_game(game)
        
        return jsonify({
            "game": {
                "id": game.id,
                "name": game.name,
                "displayName": game.display_name,
                "boxArtURL": game.box_art_url,
            },
            "message": f"Game '{game.display_name}' added successfully"
        })
    except Exception as e:
        logger.error(f"Add game error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/games/<game_id>', methods=['DELETE'])
def remove_game(game_id: str):
    """Remove a game from the watch list"""
    try:
        storage.remove_game(game_id)
        return jsonify({"status": "removed"})
    except Exception as e:
        logger.error(f"Remove game error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/games/reorder', methods=['PUT'])
def reorder_games():
    """Reorder games in priority list"""
    try:
        data = request.get_json()
        game_ids = data.get('gameIds', [])
        
        storage.reorder_games(game_ids)
        return jsonify({"status": "reordered"})
    except Exception as e:
        logger.error(f"Reorder games error: {e}")
        return jsonify({"error": str(e)}), 500

# Drops endpoints
@api.route('/drops', methods=['GET'])
def get_drops():
    """Get available drops"""
    try:
        if not storage.is_authenticated():
            return jsonify({"error": "Not authenticated"}), 401
        
        # For now, return empty list - real implementation would fetch from Twitch API
        drops = storage.get_drops()
        drops_data = []
        
        for drop in drops:
            drops_data.append({
                "id": drop.id,
                "name": drop.name,
                "description": drop.description,
                "imageURL": drop.image_url,
                "startAt": drop.start_at,
                "endAt": drop.end_at,
                "requiredMinutes": drop.required_minutes,
                "currentMinutes": drop.current_minutes,
                "gameID": drop.game_id,
                "gameName": drop.game_name,
                "isClaimed": drop.is_claimed,
                "isCompleted": drop.is_completed,
            })
        
        return jsonify({"drops": drops_data})
    except Exception as e:
        logger.error(f"Get drops error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/drops/<drop_id>/claim', methods=['POST'])
def claim_drop(drop_id: str):
    """Claim a drop"""
    try:
        if not storage.is_authenticated():
            return jsonify({"error": "Not authenticated"}), 401
        
        # In real implementation, this would call the Twitch API
        # For now, just mark as claimed in storage
        drops = storage.get_drops()
        for drop in drops:
            if drop.id == drop_id:
                drop.is_claimed = True
                break
        
        storage.save_drops(drops)
        
        return jsonify({
            "status": "claimed",
            "dropId": drop_id
        })
    except Exception as e:
        logger.error(f"Claim drop error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/drops/current', methods=['GET'])
def get_current_drop():
    """Get current drop for a channel"""
    try:
        channel_id = request.args.get('channelId')
        channel_login = request.args.get('channelLogin')
        
        if not channel_id:
            return jsonify({"error": "channelId is required"}), 400
        
        # In real implementation, would check current drop progress
        return jsonify({
            "status": "success",
            "message": "Check logs for drop progress details"
        })
    except Exception as e:
        logger.error(f"Get current drop error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/points/claim', methods=['POST'])
def claim_community_points():
    """Claim community points"""
    try:
        data = request.get_json()
        claim_id = data.get('claimId')
        channel_id = data.get('channelId')
        
        if not claim_id or not channel_id:
            return jsonify({"error": "claimId and channelId are required"}), 400
        
        # In real implementation, would call Twitch API
        return jsonify({
            "status": "claimed",
            "claimId": claim_id
        })
    except Exception as e:
        logger.error(f"Claim community points error: {e}")
        return jsonify({"error": str(e)}), 500

# Miner control endpoints
@api.route('/miner/start', methods=['POST'])
def start_miner():
    """Start the miner"""
    try:
        if not storage.is_authenticated():
            return jsonify({"error": "Not authenticated"}), 401
        
        miner.start()
        return jsonify({"status": "started"})
    except Exception as e:
        logger.error(f"Start miner error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/miner/stop', methods=['POST'])
def stop_miner():
    """Stop the miner"""
    try:
        miner.stop()
        return jsonify({"status": "stopped"})
    except Exception as e:
        logger.error(f"Stop miner error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/miner/status', methods=['GET'])
def get_miner_status():
    """Get miner status"""
    try:
        status = miner.get_status()
        
        return jsonify({
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
            "gamesQueue": [
                {
                    "id": game.id,
                    "name": game.name,
                    "displayName": game.display_name,
                    "boxArtURL": game.box_art_url,
                } for game in status.games_queue
            ],
            "lastUpdate": status.last_update,
        })
    except Exception as e:
        logger.error(f"Get miner status error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/miner/logs', methods=['GET'])
def get_miner_logs():
    """Get miner logs"""
    try:
        limit = int(request.args.get('limit', 100))
        logs = miner.get_logs(limit)
        
        logs_data = []
        for log in logs:
            logs_data.append({
                "timestamp": log.timestamp,
                "level": log.level,
                "message": log.message,
                "gameId": log.game_id,
                "streamId": log.stream_id,
            })
        
        return jsonify({"logs": logs_data})
    except Exception as e:
        logger.error(f"Get miner logs error: {e}")
        return jsonify({"error": str(e)}), 500

# Settings endpoints
@api.route('/settings', methods=['GET'])
def get_settings():
    """Get application settings"""
    try:
        settings = storage.get_settings()
        
        return jsonify({
            "games": settings.games,
            "watchInterval": settings.watch_interval,
            "autoClaimDrops": settings.auto_claim_drops,
            "notificationsEnabled": settings.notifications_enabled,
            "theme": settings.theme,
            "language": settings.language,
            "updatedAt": settings.updated_at,
        })
    except Exception as e:
        logger.error(f"Get settings error: {e}")
        return jsonify({"error": str(e)}), 500

@api.route('/settings', methods=['PUT'])
def update_settings():
    """Update application settings"""
    try:
        data = request.get_json()
        
        settings = Settings(
            games=data.get('games', []),
            watch_interval=data.get('watchInterval', 20),
            auto_claim_drops=data.get('autoClaimDrops', True),
            notifications_enabled=data.get('notificationsEnabled', True),
            theme=data.get('theme', 'dark'),
            language=data.get('language', 'en'),
        )
        
        storage.save_settings(settings)
        
        return jsonify({
            "games": settings.games,
            "watchInterval": settings.watch_interval,
            "autoClaimDrops": settings.auto_claim_drops,
            "notificationsEnabled": settings.notifications_enabled,
            "theme": settings.theme,
            "language": settings.language,
            "updatedAt": settings.updated_at,
        })
    except Exception as e:
        logger.error(f"Update settings error: {e}")
        return jsonify({"error": str(e)}), 500

# Health check
@api.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "ok", "timestamp": datetime.now(timezone.utc).isoformat()})

# Game streams endpoint (for debugging/testing)
@api.route('/games/<game_name>/streams', methods=['GET'])
def get_game_streams(game_name: str):
    """Get streams for a specific game"""
    try:
        if not storage.is_authenticated():
            return jsonify({"error": "Not authenticated"}), 401
        
        limit = int(request.args.get('limit', 20))
        
        # This would use the real Twitch API in production
        # For now, return mock data or empty list
        return jsonify({
            "game": game_name,
            "streamCount": 0,
            "streams": []
        })
    except Exception as e:
        logger.error(f"Get game streams error: {e}")
        return jsonify({"error": str(e)}), 500