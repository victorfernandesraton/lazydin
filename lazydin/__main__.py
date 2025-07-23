import asyncio
import sys
import uvicorn
from pathlib import Path

def main():
    """
    Main entry point for the lazydin application when installed via pip.
    This allows running the application with the 'lazydin' command.
    """
    # Import the main module
    try:
        # First try to import from the package
        from lazydin.main import app
    except ImportError:
        # If not found in package, try to import from current directory
        main_path = Path.cwd() / "main.py"
        if main_path.exists():
            sys.path.insert(0, str(Path.cwd()))
            from main import app
        else:
            print("Error: Could not find main.py")
            print("Make sure you're in the correct directory or the package is installed correctly.")
            sys.exit(1)
    
    # Run the FastAPI app
    uvicorn.run(app, host="0.0.0.0", port=8000)

if __name__ == "__main__":
    main()
