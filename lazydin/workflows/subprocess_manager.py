import asyncio
import importlib
import logging
import threading
import time
import traceback
import uuid
from typing import Any, Callable, Dict, List, Optional, Tuple


class TaskManager:
    """Manages background tasks using threads and asyncio"""

    tasks: Dict[str, Dict[str, Any]] = {}
    _lock = threading.Lock()

    @staticmethod
    def generate_task_id() -> str:
        """Generate a unique ID for a task"""
        return str(uuid.uuid4())

    @staticmethod
    def _import_function(function_path: str) -> Callable:
        """Import a function from its path"""
        if "." in function_path:
            parts = function_path.split(".")

            if len(parts) >= 3:
                method_name = parts[-1]
                class_name = parts[-2]
                module_path = ".".join(parts[:-2])

                module = importlib.import_module(module_path)

                class_obj = getattr(module, class_name)

                async def class_method_wrapper(*args, **kwargs):
                    instance = class_obj()

                    method = getattr(instance, method_name)

                    if len(args) == 1 and isinstance(args[0], dict):
                        return await method(args[0])
                    else:
                        params = {}
                        for i, arg in enumerate(args):
                            params[f"arg{i}"] = arg
                        params.update(kwargs)
                        return await method(params)

                return class_method_wrapper
            else:
                module_path, function_name = function_path.rsplit(".", 1)
                module = importlib.import_module(module_path)
                return getattr(module, function_name)
        else:
            return globals()[function_path]

    @classmethod
    async def _run_async_function(
        cls, task_id: str, function_path: str, *args, **kwargs
    ):
        """Run an async function and store its result"""
        try:
            func = cls._import_function(function_path)

            try:
                coro = func(*args, **kwargs)
                if not asyncio.iscoroutine(coro):
                    coro = func.__call__(*args, **kwargs)

                result = await coro

                with cls._lock:
                    cls.tasks[task_id]["status"] = "completed"
                    cls.tasks[task_id]["result"] = result
                    cls.tasks[task_id]["end_time"] = time.time()
            except Exception as e:
                logging.exception(f"Error executing async function: {e}")
                with cls._lock:
                    cls.tasks[task_id]["status"] = "error"
                    cls.tasks[task_id]["error"] = str(e)
                    cls.tasks[task_id]["traceback"] = traceback.format_exc()
                    cls.tasks[task_id]["end_time"] = time.time()

        except Exception as e:
            logging.exception(f"Error in task {task_id}: {e}")
            with cls._lock:
                cls.tasks[task_id]["status"] = "error"
                cls.tasks[task_id]["error"] = str(e)
                cls.tasks[task_id]["traceback"] = traceback.format_exc()
                cls.tasks[task_id]["end_time"] = time.time()

    @classmethod
    def run_function(cls, function_path: str, *args, **kwargs) -> str:
        """
        Run any Python async function in a background task

        Args:
            function_path: Import path to the function (e.g., 'module.submodule.function')
            *args: Positional arguments to pass to the function
            **kwargs: Keyword arguments to pass to the function

        Returns:
            task_id: Unique ID for the task
        """
        task_id = cls.generate_task_id()

        with cls._lock:
            cls.tasks[task_id] = {
                "task_id": task_id,
                "function": function_path,
                "args": args,
                "kwargs": kwargs,
                "status": "running",
                "start_time": time.time(),
                "end_time": None,
                "result": None,
            }

        asyncio.create_task(
            cls._run_async_function(task_id, function_path, *args, **kwargs)
        )

        logging.info(f"Started async function {function_path} as task {task_id}")
        return task_id

    @classmethod
    def get_task(cls, task_id: str) -> Optional[Dict[str, Any]]:
        """Get task information by its ID"""
        with cls._lock:
            return cls.tasks.get(task_id)

    @classmethod
    def get_task_status(cls, task_id: str) -> Tuple[str, Optional[Any]]:
        """
        Get the status of a task

        Returns:
            (status, result_or_error)
        """
        with cls._lock:
            task = cls.tasks.get(task_id)
            if not task:
                return "not_found", None

            status = task.get("status", "unknown")
            if status == "completed":
                return status, task.get("result")
            elif status == "error":
                return status, task.get("error")
            else:
                return status, None

    @classmethod
    def cancel_task(cls, task_id: str) -> bool:
        """
        Mark a task as cancelled (can't truly cancel already running threads)
        """
        with cls._lock:
            task = cls.tasks.get(task_id)
            if not task:
                return False

            if task["status"] == "running":
                task["status"] = "cancelled"
                task["end_time"] = time.time()
                return True
            return False

    @classmethod
    def get_all_tasks(cls) -> List[Dict[str, Any]]:
        """Get information about all tasks"""
        with cls._lock:
            return list(cls.tasks.values())

    @classmethod
    def cleanup_old_tasks(cls, max_age_seconds: int = 3600):
        """Remove completed tasks older than the specified age"""
        current_time = time.time()
        with cls._lock:
            for task_id in list(cls.tasks.keys()):
                task = cls.tasks[task_id]
                if task["status"] in ("completed", "error", "cancelled"):
                    if (
                        task["end_time"]
                        and (current_time - task["end_time"]) > max_age_seconds
                    ):
                        cls.tasks.pop(task_id)
