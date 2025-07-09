import asyncio
import logging
from typing import Any, Dict, List

import uvicorn
from fastapi import BackgroundTasks, FastAPI, Form, Request
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel

from lazydin.workflows.linkedin.auth import LinkedinAuth
from lazydin.workflows.subprocess_manager import TaskManager

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

templates = Jinja2Templates(directory="lazydin/template")

app = FastAPI()

@app.on_event("startup")
def startup_event():
    background_tasks = BackgroundTasks()
    background_tasks.add_task(cleanup_old_tasks_periodically)

async def cleanup_old_tasks_periodically():
    while True:
        await asyncio.sleep(3600)  # Clean up every hour
        TaskManager.cleanup_old_tasks()

# Modelo de dados para o corpo da requisição
class ScrapeRequest(BaseModel):
    url: str
    username: str = "test@gmail.com"
    password: str = "test"


@app.get("/health")
async def health():
    return ":-)"

@app.post("/auth")
async def auth(request: ScrapeRequest):
    # Run the execute method directly with parameters as a dictionary
    task_id = TaskManager.run_function(
        "lazydin.workflows.linkedin.auth.LinkedinAuth.execute",
        {
            "driver_key": None,  # driver_key will be created inside the method
            "username": request.username,
            "password": request.password
        }
    )
    
    return JSONResponse({
        "task_id": task_id,
        "status": "running"
    })

@app.get("/task/{task_id}")
async def get_task_status(task_id: str):
    task = TaskManager.get_task(task_id)
    
    if not task:
        return JSONResponse({
            "task_id": task_id,
            "status": "not_found"
        })
    
    return JSONResponse(task)

@app.delete("/task/{task_id}")
async def cancel_task(task_id: str):
    success = TaskManager.cancel_task(task_id)
    
    return JSONResponse({
        "task_id": task_id,
        "cancelled": success
    })

class RunFunctionRequest(BaseModel):
    function_path: str
    args: List[Any] = []
    kwargs: Dict[str, Any] = {}

@app.post("/run")
async def run_function(request: RunFunctionRequest):
    """Generic endpoint to run any function in the background"""
    task_id = TaskManager.run_function(
        request.function_path,
        *request.args,
        **request.kwargs
    )
    
    return JSONResponse({
        "task_id": task_id,
        "status": "running"
    })


@app.post("/scrape", response_class=HTMLResponse)
async def scrape(background_tasks: BackgroundTasks, url: str = Form(...)):
    # Run the execute method directly with parameters as a dictionary
    task_id = TaskManager.run_function(
        "lazydin.workflows.linkedin.auth.LinkedinAuth.execute",
        {
            "driver_key": None,  # driver_key will be created inside the method
            "username": "test@gmail.com",
            "password": "test"
        }
    )
    return f"<p><strong>Task ID:</strong> {task_id}</p><p><strong>URL:</strong> {url}</p>"

@app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    return templates.TemplateResponse("index.html", {"request": request})

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
