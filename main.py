# This file is kept for backward compatibility
# The actual implementation has been moved to lazydin/main.py

from lazydin.main import app

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
