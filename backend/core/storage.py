from __future__ import annotations

import json
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Any, Optional
from dataclasses import dataclass, asdict

from core.constants import WORKING_DIR

# Ensure working directory exists
WORKING_DIR.mkdir(exist_ok=True)

@dataclass
class Game:
    id: str
    name: str
    display_name: str
    box_art_url: str = ""

@dataclass
class Drop:
    id: str
    name: str
    description: str
    image_url: str
    start_at: str
    end_at: str
    required_minutes: int
    current_minutes: int
    game_id: str
    game_name: str
    is_claimed: bool = False
    is_completed: bool = False

@dataclass
class Stream:
    id: str
    user_id: str
    user_login: str
    user_name: str
    title: str
    viewer_count: int
    language: str
    game_id: str = ""
    game_name: str = ""

@dataclass
class AuthData:
    access_token: str = ""
    refresh_token: str = ""
    expires_at: str = ""
    user_id: str = ""
    username: str = ""
    updated_at: str = ""

@dataclass
class Settings:
    games: List[str]
    watch_interval: int = 20
    auto_claim_drops: bool = True
    notifications_enabled: bool = True
    theme: str = "dark"
    language: str = "en"
    updated_at: str = ""

@dataclass
class MinerStatus:
    is_running: bool = False
    current_game: Optional[Game] = None
    current_stream: Optional[Stream] = None
    watch_duration: int = 0  # in seconds
    total_watched: int = 0   # in seconds
    games_queue: List[Game] = None
    last_update: str = ""

    def __post_init__(self):
        if self.games_queue is None:
            self.games_queue = []

@dataclass
class LogEntry:
    timestamp: str
    level: str
    message: str
    game_id: str = ""
    stream_id: str = ""

class Storage:
    def __init__(self):
        self.data_dir = WORKING_DIR
        self.auth_file = self.data_dir / "auth.json"
        self.settings_file = self.data_dir / "settings.json"
        self.games_file = self.data_dir / "games.json"
        self.drops_file = self.data_dir / "drops.json"
        self.logs_file = self.data_dir / "logs.json"
        self.status_file = self.data_dir / "status.json"
        
        # Initialize with defaults if files don't exist
        self._ensure_files()

    def _ensure_files(self):
        """Create default files if they don't exist"""
        if not self.settings_file.exists():
            default_settings = Settings(games=[])
            self.save_settings(default_settings)
            
        if not self.games_file.exists():
            self._save_json(self.games_file, {})
            
        if not self.drops_file.exists():
            self._save_json(self.drops_file, [])
            
        if not self.logs_file.exists():
            self._save_json(self.logs_file, [])

    def _load_json(self, file_path: Path, default=None):
        """Load JSON from file with error handling"""
        try:
            if file_path.exists():
                with open(file_path, 'r', encoding='utf-8') as f:
                    return json.load(f)
        except (json.JSONDecodeError, IOError) as e:
            print(f"Error loading {file_path}: {e}")
        return default or {}

    def _save_json(self, file_path: Path, data):
        """Save JSON to file with error handling"""
        try:
            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
        except IOError as e:
            print(f"Error saving {file_path}: {e}")

    # Authentication methods
    def get_auth(self) -> Optional[AuthData]:
        """Get authentication data"""
        data = self._load_json(self.auth_file)
        if data:
            return AuthData(**data)
        return None

    def save_auth(self, auth: AuthData):
        """Save authentication data"""
        auth.updated_at = datetime.now(timezone.utc).isoformat()
        self._save_json(self.auth_file, asdict(auth))

    def clear_auth(self):
        """Clear authentication data"""
        if self.auth_file.exists():
            self.auth_file.unlink()

    def is_authenticated(self) -> bool:
        """Check if user is authenticated"""
        auth = self.get_auth()
        if not auth or not auth.access_token:
            return False
        
        # Check if token is expired
        if auth.expires_at:
            try:
                expires_at = datetime.fromisoformat(auth.expires_at.replace('Z', '+00:00'))
                return datetime.now(timezone.utc) < expires_at
            except ValueError:
                return False
        
        return True

    # Settings methods
    def get_settings(self) -> Settings:
        """Get application settings"""
        data = self._load_json(self.settings_file, {})
        if not data.get('games'):
            data['games'] = []
        return Settings(**data)

    def save_settings(self, settings: Settings):
        """Save application settings"""
        settings.updated_at = datetime.now(timezone.utc).isoformat()
        self._save_json(self.settings_file, asdict(settings))

    # Games methods
    def get_games(self) -> Dict[str, Game]:
        """Get all games"""
        data = self._load_json(self.games_file, {})
        games = {}
        for game_id, game_data in data.items():
            games[game_id] = Game(**game_data)
        return games

    def get_game(self, game_id: str) -> Optional[Game]:
        """Get a specific game"""
        games = self.get_games()
        return games.get(game_id)

    def add_game(self, game: Game):
        """Add a game"""
        games = self.get_games()
        games[game.id] = game
        self._save_games(games)
        
        # Add to settings games list if not already there
        settings = self.get_settings()
        if game.id not in settings.games:
            settings.games.append(game.id)
            self.save_settings(settings)

    def remove_game(self, game_id: str):
        """Remove a game"""
        games = self.get_games()
        if game_id in games:
            del games[game_id]
            self._save_games(games)
        
        # Remove from settings games list
        settings = self.get_settings()
        if game_id in settings.games:
            settings.games.remove(game_id)
            self.save_settings(settings)

    def reorder_games(self, game_ids: List[str]):
        """Reorder games in priority list"""
        settings = self.get_settings()
        settings.games = game_ids
        self.save_settings(settings)

    def _save_games(self, games: Dict[str, Game]):
        """Save games dictionary"""
        data = {}
        for game_id, game in games.items():
            data[game_id] = asdict(game)
        self._save_json(self.games_file, data)

    # Drops methods
    def get_drops(self) -> List[Drop]:
        """Get all drops"""
        data = self._load_json(self.drops_file, [])
        return [Drop(**drop_data) for drop_data in data]

    def save_drops(self, drops: List[Drop]):
        """Save drops"""
        data = [asdict(drop) for drop in drops]
        self._save_json(self.drops_file, data)

    def update_drop_progress(self, drop_id: str, minutes: int):
        """Update drop progress"""
        drops = self.get_drops()
        for drop in drops:
            if drop.id == drop_id:
                drop.current_minutes = minutes
                drop.is_completed = minutes >= drop.required_minutes
                break
        self.save_drops(drops)

    # Logs methods
    def get_logs(self, limit: int = 100) -> List[LogEntry]:
        """Get recent logs"""
        data = self._load_json(self.logs_file, [])
        logs = [LogEntry(**log_data) for log_data in data]
        return logs[-limit:] if limit > 0 else logs

    def add_log(self, entry: LogEntry):
        """Add a log entry"""
        logs_data = self._load_json(self.logs_file, [])
        logs_data.append(asdict(entry))
        
        # Keep only last 1000 entries
        if len(logs_data) > 1000:
            logs_data = logs_data[-1000:]
        
        self._save_json(self.logs_file, logs_data)

    # Status methods
    def get_miner_status(self) -> MinerStatus:
        """Get miner status"""
        data = self._load_json(self.status_file, {})
        if not data:
            return MinerStatus()
        
        # Convert nested objects
        if 'current_game' in data and data['current_game']:
            data['current_game'] = Game(**data['current_game'])
        if 'current_stream' in data and data['current_stream']:
            data['current_stream'] = Stream(**data['current_stream'])
        if 'games_queue' in data and data['games_queue']:
            data['games_queue'] = [Game(**game_data) for game_data in data['games_queue']]
        
        return MinerStatus(**data)

    def save_miner_status(self, status: MinerStatus):
        """Save miner status"""
        status.last_update = datetime.now(timezone.utc).isoformat()
        data = asdict(status)
        self._save_json(self.status_file, data)

    # Statistics methods
    def get_drop_progress(self, drop_id: str) -> int:
        """Get drop progress minutes"""
        drops = self.get_drops()
        for drop in drops:
            if drop.id == drop_id:
                return drop.current_minutes
        return 0

    def save_drop_progress(self, drop_id: str, minutes: int):
        """Save drop progress"""
        self.update_drop_progress(drop_id, minutes)

    # Utility methods
    def create_backup(self) -> str:
        """Create a backup of all data"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = f"backup_{timestamp}.json"
        backup_path = self.data_dir / "backups"
        backup_path.mkdir(exist_ok=True)
        backup_file = backup_path / backup_name
        
        backup_data = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "auth": self._load_json(self.auth_file),
            "settings": self._load_json(self.settings_file),
            "games": self._load_json(self.games_file),
            "drops": self._load_json(self.drops_file),
            "logs": self._load_json(self.logs_file)[-100:],  # Only last 100 logs
            "status": self._load_json(self.status_file),
        }
        
        self._save_json(backup_file, backup_data)
        return backup_name

    def restore_backup(self, backup_name: str) -> bool:
        """Restore from backup"""
        backup_path = self.data_dir / "backups" / backup_name
        if not backup_path.exists():
            return False
        
        try:
            backup_data = self._load_json(backup_path)
            
            # Restore each file
            if "auth" in backup_data and backup_data["auth"]:
                self._save_json(self.auth_file, backup_data["auth"])
            if "settings" in backup_data:
                self._save_json(self.settings_file, backup_data["settings"])
            if "games" in backup_data:
                self._save_json(self.games_file, backup_data["games"])
            if "drops" in backup_data:
                self._save_json(self.drops_file, backup_data["drops"])
            if "logs" in backup_data:
                self._save_json(self.logs_file, backup_data["logs"])
            if "status" in backup_data:
                self._save_json(self.status_file, backup_data["status"])
            
            return True
        except Exception as e:
            print(f"Error restoring backup: {e}")
            return False