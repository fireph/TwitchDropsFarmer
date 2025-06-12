#!/usr/bin/env python3
from __future__ import annotations

import os
import sys
import logging
from pathlib import Path

from flask import Flask, send_from_directory
from flask_cors import CORS
from flask_socketio import SocketIO
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Add the backend directory to the Python path
backend_dir = str(Path(__file__).parent)
if backend_dir not in sys.path:
    sys.path.insert(0, backend_dir)

from core.twitch import Twitch
from core.storage import Storage
from core.miner import Miner
from api.routes import api, init_routes
from api.websocket import init_websocket

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("TwitchDrops")

def create_app():
    """Create and configure the Flask application"""
    app = Flask(__name__, static_folder="../frontend/dist")
    
    # Configuration
    app.config['SECRET_KEY'] = os.getenv('SECRET_KEY', 'dev-secret-key-change-in-production')
    app.config['DEBUG'] = os.getenv('FLASK_DEBUG', 'False').lower() == 'true'
    
    # Enable CORS
    CORS(app, origins=["http://localhost:3000", "http://localhost:5173"])
    
    # Initialize SocketIO
    socketio = SocketIO(
        app, 
        cors_allowed_origins=["http://localhost:3000", "http://localhost:5173"],
        async_mode='threading'
    )
    
    # Initialize core components
    storage = Storage()
    twitch = Twitch()
    miner = Miner(storage, twitch)
    
    # Initialize API routes
    init_routes(twitch, storage, miner)
    app.register_blueprint(api)
    
    # Initialize WebSocket handlers
    init_websocket(socketio, miner)
    
    # Serve frontend static files
    @app.route('/')
    def serve_frontend():
        """Serve the frontend index.html"""
        return send_from_directory(app.static_folder, 'index.html')
    
    @app.route('/<path:path>')
    def serve_static(path):
        """Serve frontend static files"""
        file_path = Path(app.static_folder) / path
        if file_path.exists() and file_path.is_file():
            return send_from_directory(app.static_folder, path)
        else:
            # For SPA routing, return index.html for non-API routes
            if not path.startswith('api/'):
                return send_from_directory(app.static_folder, 'index.html')
            return "Not found", 404
    
    # WebSocket endpoint
    @app.route('/api/v1/ws')
    def websocket_endpoint():
        """WebSocket endpoint for real-time updates"""
        return "WebSocket endpoint - use SocketIO client to connect"
    
    # Health check
    @app.route('/health')
    def health():
        """Health check endpoint"""
        return {"status": "ok", "service": "TwitchDropsFarmer"}
    
    return app, socketio

def main():
    """Main entry point"""
    logger.info("Starting TwitchDropsFarmer...")
    
    # Create app and socketio
    app, socketio = create_app()
    
    # Get configuration
    host = os.getenv('HOST', '0.0.0.0')
    port = int(os.getenv('PORT', 8080))
    debug = os.getenv('FLASK_DEBUG', 'False').lower() == 'true'
    
    logger.info(f"Starting server on {host}:{port}")
    logger.info(f"Debug mode: {debug}")
    logger.info(f"Static files: {app.static_folder}")
    
    try:
        # Run the app with SocketIO
        socketio.run(
            app,
            host=host,
            port=port,
            debug=debug,
            allow_unsafe_werkzeug=True
        )
    except KeyboardInterrupt:
        logger.info("Shutting down...")
    except Exception as e:
        logger.error(f"Server error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()