from __future__ import annotations

import asyncio
import logging
from datetime import datetime, timezone, timedelta
from typing import Optional, List, Dict, Any, Set
from threading import Thread, Event
import random

from core.twitch import Twitch
from core.storage import Storage, Game, Stream, LogEntry, MinerStatus
from core.constants import WATCH_INTERVAL

logger = logging.getLogger("TwitchDrops")

class Miner:
    def __init__(self, storage: Storage, twitch: Twitch):
        self.storage = storage
        self.twitch = twitch
        
        # State management
        self.is_running = False
        self.current_game: Optional[Game] = None
        self.current_stream: Optional[Stream] = None
        self.watch_start_time: Optional[datetime] = None
        
        # Control
        self._stop_event = Event()
        self._thread: Optional[Thread] = None
        
        # WebSocket clients for real-time updates
        self.ws_clients: Set = set()
        
        # Mining loop
        self._loop_task: Optional[asyncio.Task] = None

    def start(self):
        """Start the mining process"""
        if self.is_running:
            return
            
        self.is_running = True
        self._stop_event.clear()
        
        # Start in a separate thread to avoid blocking the web server
        self._thread = Thread(target=self._run_mining_loop, daemon=True)
        self._thread.start()
        
        self._log("INFO", "Miner started")
        self._update_status()

    def stop(self):
        """Stop the mining process"""
        if not self.is_running:
            return
            
        self.is_running = False
        self._stop_event.set()
        
        if self._thread and self._thread.is_alive():
            self._thread.join(timeout=5)
            
        self.current_game = None
        self.current_stream = None
        self.watch_start_time = None
        
        self._log("INFO", "Miner stopped")
        self._update_status()

    def _run_mining_loop(self):
        """Run the main mining loop in a separate thread"""
        # Create new event loop for this thread
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        
        try:
            loop.run_until_complete(self._mining_loop())
        except Exception as e:
            self._log("ERROR", f"Mining loop error: {e}")
        finally:
            loop.close()

    async def _mining_loop(self):
        """Main mining loop"""
        while self.is_running and not self._stop_event.is_set():
            try:
                await self._process_next_game()
                
                # Wait for the watch interval or stop signal
                for _ in range(WATCH_INTERVAL.seconds):
                    if self._stop_event.is_set():
                        break
                    await asyncio.sleep(1)
                    
            except Exception as e:
                self._log("ERROR", f"Mining loop iteration error: {e}")
                await asyncio.sleep(5)  # Brief pause before retrying

    async def _process_next_game(self):
        """Process the next game in the priority queue"""
        if not self.twitch.auth.is_logged_in:
            self._log("WARNING", "Not logged in, cannot mine")
            return
            
        settings = self.storage.get_settings()
        if not settings.games:
            self._log("DEBUG", "No games in watch list")
            return
            
        games = self.storage.get_games()
        
        # Get the next game to watch (first available in priority order)
        next_game = None
        for game_id in settings.games:
            if game_id in games:
                next_game = games[game_id]
                break
                
        if not next_game:
            self._log("WARNING", "No valid games found in watch list")
            return
            
        # Switch games if needed
        if not self.current_game or self.current_game.id != next_game.id:
            await self._switch_to_game(next_game)
            
        # Continue watching current stream or find a new one
        if not self.current_stream or not await self._is_stream_live(self.current_stream):
            await self._find_and_watch_stream(next_game)
        else:
            await self._continue_watching()

    async def _switch_to_game(self, game: Game):
        """Switch to watching a different game"""
        self.current_game = game
        self.current_stream = None
        self.watch_start_time = None
        
        self._log("INFO", f"Switching to game: {game.display_name}", game.id)
        self._update_status()

    async def _find_and_watch_stream(self, game: Game):
        """Find a stream for the current game and start watching"""
        try:
            streams = await self.twitch.get_streams_for_game(game.name, limit=20)
            
            if not streams:
                self._log("WARNING", f"No live streams found for game: {game.display_name}", game.id)
                return
                
            # Select a random stream from the top results (more natural)
            selected_stream_data = random.choice(streams[:5])
            
            # Convert to our Stream object
            selected_stream = Stream(
                id=selected_stream_data.get("id", ""),
                user_id="",  # Not provided in the API response we're using
                user_login=selected_stream_data.get("user_login", ""),
                user_name=selected_stream_data.get("user_name", ""),
                title=selected_stream_data.get("title", ""),
                viewer_count=selected_stream_data.get("viewer_count", 0),
                language=selected_stream_data.get("language", "en"),
                game_id=game.id,
                game_name=game.display_name,
            )
            
            self.current_stream = selected_stream
            self.watch_start_time = datetime.now(timezone.utc)
            
            self._log(
                "INFO", 
                f"Started watching stream: {selected_stream.user_name} ({selected_stream.title})",
                game.id,
                selected_stream.id
            )
            
            await self._watch_stream(selected_stream)
            
        except Exception as e:
            self._log("ERROR", f"Failed to find streams for game {game.display_name}: {e}", game.id)

    async def _continue_watching(self):
        """Continue watching the current stream"""
        if self.current_stream:
            await self._watch_stream(self.current_stream)

    async def _watch_stream(self, stream: Stream):
        """Make the API request to simulate watching"""
        try:
            # Get stream access token (this simulates watching)
            await self.twitch.get_stream_access_token(stream.user_login)
            
            # Update watch duration
            if self.watch_start_time:
                watch_duration = (datetime.now(timezone.utc) - self.watch_start_time).total_seconds()
            else:
                watch_duration = 0
                
            self._log(
                "SUCCESS",
                f"Watching {stream.user_name} - {stream.title} ({int(watch_duration)}s)",
                self.current_game.id if self.current_game else "",
                stream.id
            )
            
            # Check for drops progress
            await self._check_drop_progress(stream)
            
            self._update_status()
            
        except Exception as e:
            self._log(
                "ERROR", 
                f"Failed to watch stream {stream.user_name}: {e}",
                self.current_game.id if self.current_game else "",
                stream.id
            )
            
            # Mark stream as offline and find a new one
            self.current_stream = None
            self.watch_start_time = None

    async def _is_stream_live(self, stream: Stream) -> bool:
        """Check if a stream is still live"""
        try:
            # For now, assume streams stay live for reasonable periods
            # In a full implementation, you'd check the stream status
            if self.watch_start_time:
                time_watching = datetime.now(timezone.utc) - self.watch_start_time
                # Assume streams stay live for at least 30 minutes
                return time_watching < timedelta(minutes=30)
            return True
        except Exception:
            return False

    async def _check_drop_progress(self, stream: Stream):
        """Check for drop progress on the current stream"""
        try:
            if not self.current_game:
                return
                
            # Get current drop context
            # Note: We'd need the channel ID, which isn't in our simplified stream object
            # This is a placeholder for the actual drop checking logic
            
            # In the real implementation, you'd:
            # 1. Get available drops for the game
            # 2. Check current drop progress
            # 3. Auto-claim completed drops if enabled
            
            pass
            
        except Exception as e:
            self._log("WARNING", f"Failed to check drop progress: {e}")

    def _log(self, level: str, message: str, game_id: str = "", stream_id: str = ""):
        """Add a log entry and broadcast it"""
        entry = LogEntry(
            timestamp=datetime.now(timezone.utc).isoformat(),
            level=level,
            message=message,
            game_id=game_id,
            stream_id=stream_id
        )
        
        # Save to storage
        self.storage.add_log(entry)
        
        # Also log to console
        logger.log(getattr(logging, level, logging.INFO), message)
        
        # Broadcast to WebSocket clients
        self._broadcast_log_update(entry)

    def _update_status(self):
        """Update and broadcast status"""
        watch_duration = 0
        if self.watch_start_time and self.is_running:
            watch_duration = int((datetime.now(timezone.utc) - self.watch_start_time).total_seconds())
            
        status = MinerStatus(
            is_running=self.is_running,
            current_game=self.current_game,
            current_stream=self.current_stream,
            watch_duration=watch_duration,
            total_watched=0,  # Would need to track this separately
            games_queue=list(self.storage.get_games().values()),
            last_update=datetime.now(timezone.utc).isoformat()
        )
        
        # Save status
        self.storage.save_miner_status(status)
        
        # Broadcast to WebSocket clients
        self._broadcast_status_update(status)

    def get_status(self) -> MinerStatus:
        """Get current miner status"""
        return self.storage.get_miner_status()

    def get_logs(self, limit: int = 100) -> List[LogEntry]:
        """Get recent log entries"""
        return self.storage.get_logs(limit)

    # WebSocket client management
    def add_ws_client(self, client):
        """Add a WebSocket client for real-time updates"""
        self.ws_clients.add(client)

    def remove_ws_client(self, client):
        """Remove a WebSocket client"""
        self.ws_clients.discard(client)

    def _broadcast_status_update(self, status: MinerStatus):
        """Broadcast status update to WebSocket clients"""
        if not self.ws_clients:
            return
            
        message = {
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
        }
        
        # Send to all connected clients
        for client in self.ws_clients.copy():
            try:
                client.emit("message", message)
            except Exception as e:
                self._log("WARNING", f"Failed to send status update to client: {e}")
                self.ws_clients.discard(client)

    def _broadcast_log_update(self, entry: LogEntry):
        """Broadcast log update to WebSocket clients"""
        if not self.ws_clients:
            return
            
        message = {
            "type": "log",
            "data": {
                "timestamp": entry.timestamp,
                "level": entry.level,
                "message": entry.message,
                "gameId": entry.game_id,
                "streamId": entry.stream_id,
            }
        }
        
        # Send to all connected clients
        for client in self.ws_clients.copy():
            try:
                client.emit("message", message)
            except Exception as e:
                self.ws_clients.discard(client)