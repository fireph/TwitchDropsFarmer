#!/usr/bin/env python3
"""
Startup script for TwitchDropsFarmer

This script sets up the environment and starts the application.
It can be used for development or production deployment.
"""

import os
import sys
from pathlib import Path

def main():
    """Main entry point for the application"""
    # Ensure we're in the correct directory
    project_root = Path(__file__).parent.absolute()
    os.chdir(project_root)
    
    # Add the backend directory to Python path
    backend_dir = project_root / "backend"
    if str(backend_dir) not in sys.path:
        sys.path.insert(0, str(backend_dir))
    
    # Verify backend directory exists
    if not backend_dir.exists():
        print(f"Error: Backend directory not found at {backend_dir}")
        sys.exit(1)
    
    # Verify main.py exists
    main_py = backend_dir / "main.py"
    if not main_py.exists():
        print(f"Error: main.py not found at {main_py}")
        sys.exit(1)
    
    print(f"Starting TwitchDropsFarmer from {project_root}")
    print(f"Backend directory: {backend_dir}")
    
    try:
        # Import and run the main application
        from main import main as app_main
        app_main()
    except ImportError as e:
        print(f"Error importing main module: {e}")
        print("Make sure you're in the project root directory and backend dependencies are installed")
        sys.exit(1)
    except Exception as e:
        print(f"Error starting application: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()