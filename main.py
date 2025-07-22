import asyncio
import logging
from typing import Any, Dict, Optional

import uvicorn
from fastapi import BackgroundTasks, FastAPI
from fastapi.responses import JSONResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel
from contextlib import asynccontextmanager


from lazydin.workflows.subprocess_manager import TaskManager
from lazydin.workflows.workflows_analyzer import WorkflowsAnalyzer

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

templates = Jinja2Templates(directory="lazydin/template")


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Initialize TaskManager with the workflows path
    TaskManager.initialize(workflows_path="lazydin.workflows")
    TaskManager.start_task_loop()

    background_tasks = BackgroundTasks()
    background_tasks.add_task(cleanup_old_tasks_periodically)
    yield

    TaskManager.shutdown()


app = FastAPI(lifespan=lifespan)


async def cleanup_old_tasks_periodically():
    while True:
        await asyncio.sleep(3600)  # Clean up every hour
        TaskManager.cleanup_old_tasks()


class ScrapeRequest(BaseModel):
    url: str
    username: str = "test@gmail.com"
    password: str = "test"


@app.get("/health")
async def health():
    return ":-)"


@app.get("/tasks")
async def list_all_tasks():
    """Get a list of all tasks"""
    tasks = TaskManager.get_all_tasks()
    return JSONResponse({"tasks": tasks})


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
    params: Dict[str, Any] = {}


@app.post("/task")
async def run_function(request: RunFunctionRequest):
    """Generic endpoint to run any function in the background"""
    task_id = TaskManager.run_function(
        request.function_path,
        request.params
    )

    return JSONResponse({
        "task_id": task_id,
        "status": "running"
    })


@app.get("/workflows")
async def get_workflows_documentation(workflows_path: Optional[str] = None):
    """Get documentation for all workflows in the specified path"""
    path = workflows_path or TaskManager._workflows_path
    analyzer = WorkflowsAnalyzer(workflows_path=path)
    documentation = analyzer.get_workflow_documentation()
    return JSONResponse(documentation)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
