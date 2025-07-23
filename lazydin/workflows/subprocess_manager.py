import asyncio
import importlib
import logging
import time
import traceback
import uuid
from typing import Any, Callable, Dict, List, Optional, Tuple


class TaskManager:
    """Manages background tasks using asyncio event loop"""

    tasks: Dict[str, Dict[str, Any]] = {}
    _task_queue = asyncio.Queue()
    _running = False
    _worker_task = None
    _workflows_path = "lazydin.workflows"  # Default workflows path

    @staticmethod
    def generate_task_id() -> str:
        """Generate a unique ID for a task"""
        return str(uuid.uuid4())

    @classmethod
    def initialize(cls, workflows_path: str = "lazydin.workflows"):
        """Initialize the TaskManager with a specific workflows path"""
        cls._workflows_path = workflows_path
        logging.info(f"TaskManager initialized with workflows path: {workflows_path}")
        return cls

    @classmethod
    def _import_function(cls, function_path: str) -> Callable:
        """Import a function from its path"""

        # If the function path doesn't contain a dot, assume it's a module under the workflows path
        if "." not in function_path:
            function_path = f"{cls._workflows_path}.{function_path}"

        parts = function_path.split(".")

        if len(parts) >= 3:
            method_name = parts[-1]
            class_name = parts[-2]
            module_path = f"{cls._workflows_path}.{'.'.join(parts[:-2])}"
            logging.info(f"Importing from module path: {module_path}, class: {class_name}, method: {method_name}")

            try:
                module = importlib.import_module(module_path)
                
                # Check if the attribute is a class or another module
                if hasattr(module, class_name):
                    class_obj = getattr(module, class_name)
                    
                    # Verify it's a class, not a module
                    if isinstance(class_obj, type):
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
                        # It's not a class, might be a submodule
                        submodule = class_obj
                        if hasattr(submodule, method_name):
                            return getattr(submodule, method_name)
                        else:
                            raise ImportError(f"Cannot find method {method_name} in {class_name}")
                else:
                    raise ImportError(f"Cannot find class or module {class_name} in {module_path}")
                    
            except Exception as e:
                logging.error(f"Error importing {function_path}: {str(e)}")
                raise
        else:
            module_path, function_name = function_path.rsplit(".", 1)
            module = importlib.import_module(module_path)
            return getattr(module, function_name)


    
    @classmethod
    async def _task_worker(cls):
        """Worker that processes tasks from the queue"""
        cls._running = True
        pending_tasks = set()
        
        while cls._running:
            try:
                # Get a task from the queue with a timeout to allow checking _running flag
                try:
                    task_info = await asyncio.wait_for(cls._task_queue.get(), timeout=0.5)
                except asyncio.TimeoutError:
                    # Check if we should continue running
                    if not cls._running:
                        break
                    # Clean up completed tasks
                    if pending_tasks:
                        done, pending_tasks = await asyncio.wait(
                            pending_tasks, timeout=0, return_when=asyncio.FIRST_COMPLETED
                        )
                    continue
                
                # Process the task
                task_id = task_info["task_id"]
                function_path = task_info["function"]
                args = task_info["args"]
                kwargs = task_info["kwargs"]
                
                # Create a task and add it to our pending tasks
                task = asyncio.create_task(
                    cls._run_async_function(task_id, function_path, *args, **kwargs)
                )
                pending_tasks.add(task)
                task.add_done_callback(lambda t: pending_tasks.discard(t))
                
                # Mark the queue task as done
                cls._task_queue.task_done()
                
            except asyncio.CancelledError:
                break
            except Exception as e:
                logging.exception(f"Error in task worker: {e}")
        
        # Wait for all pending tasks to complete when shutting down
        if pending_tasks:
            await asyncio.gather(*pending_tasks, return_exceptions=True)
    
    @classmethod
    async def _run_async_function(cls, task_id: str, function_path: str, *args, **kwargs):
        """Run an async function and store its result"""
        try:
            func = cls._import_function(function_path)

            try:
                coro = func(*args, **kwargs)
                if not asyncio.iscoroutine(coro):
                    coro = func.__call__(*args, **kwargs)

                result = await coro

                cls.tasks[task_id]["status"] = "completed"
                cls.tasks[task_id]["result"] = result
                cls.tasks[task_id]["end_time"] = time.time()
                # Remove the task from the dictionary once completed
                cls.tasks.pop(task_id)
            except Exception as e:
                logging.exception(f"Error executing async function: {e}")
                cls.tasks[task_id]["status"] = "error"
                cls.tasks[task_id]["error"] = str(e)
                cls.tasks[task_id]["traceback"] = traceback.format_exc()
                cls.tasks[task_id]["end_time"] = time.time()
                # Remove the task from the dictionary once it has an error
                cls.tasks.pop(task_id)

        except Exception as e:
            logging.exception(f"Error in task {task_id}: {e}")
            cls.tasks[task_id]["status"] = "error"
            cls.tasks[task_id]["error"] = str(e)
            cls.tasks[task_id]["traceback"] = traceback.format_exc()
            cls.tasks[task_id]["end_time"] = time.time()
            # Remove the task from the dictionary once it has an error
            cls.tasks.pop(task_id)

    @classmethod
    async def start_task_loop(cls):
        """Start the background task worker if it's not already running"""
        if cls._worker_task is None or cls._worker_task.done():
            cls._running = True
            cls._worker_task = asyncio.create_task(cls._task_worker())
            logging.info("Started background task worker")
            return True
        return True
    
    @classmethod
    async def stop_task_loop(cls):
        """Stop the background task worker"""
        if cls._worker_task and not cls._worker_task.done():
            cls._running = False
            try:
                await cls._worker_task
                logging.info("Stopped background task worker")
            except Exception as e:
                logging.error(f"Error stopping task worker: {e}")
    
    @classmethod
    async def run_function(cls, function_path: str, *args, **kwargs) -> str:
        """
        Run any Python async function in a background task loop

        Args:
            function_path: Import path to the function (e.g., 'module.submodule.function')
            *args: Positional arguments to pass to the function
            **kwargs: Keyword arguments to pass to the function

        Returns:
            task_id: Unique ID for the task
        """
        # Make sure the task worker is running
        await cls.start_task_loop()
        
        task_id = cls.generate_task_id()

        cls.tasks[task_id] = {
            "task_id": task_id,
            "function": function_path,
            "args": args,
            "kwargs": kwargs,
            "status": "queued",  # Start as queued, not running
            "start_time": time.time(),
            "end_time": None,
            "result": None,
        }
        
        # Add the task to the queue
        task_info = cls.tasks[task_id].copy()
        try:
            await cls._task_queue.put(task_info)
            
            # Update status to running
            cls.tasks[task_id]["status"] = "running"
            
            logging.info(f"Queued async function {function_path} as task {task_id}")
            return task_id
        except Exception as e:
            error_msg = f"Failed to queue task: {str(e)}"
            logging.error(error_msg)
            cls.tasks[task_id]["status"] = "error"
            cls.tasks[task_id]["error"] = error_msg
            cls.tasks[task_id]["end_time"] = time.time()

        return task_id

    @classmethod
    def get_task(cls, task_id: str) -> Optional[Dict[str, Any]]:
        """Get task information by its ID"""
        return cls.tasks.get(task_id)

    @classmethod
    def get_task_status(cls, task_id: str) -> Tuple[str, Optional[Any]]:
        """
        Get the status of a task

        Returns:
            (status, result_or_error)
        """
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
        Mark a task as cancelled (can't truly cancel already running tasks)
        """
        task = cls.tasks.get(task_id)
        if not task:
            return False

        if task["status"] == "running":
            task["status"] = "cancelled"
            task["end_time"] = time.time()
            # Remove the task from the dictionary once cancelled
            cls.tasks.pop(task_id)
            return True
        return False

    @classmethod
    def get_all_tasks(cls) -> List[Dict[str, Any]]:
        """Get information about all tasks"""
        return list(cls.tasks.values())

    @classmethod
    async def cleanup_old_tasks(cls, max_age_seconds: int = 3600):
        """Remove completed tasks older than the specified age"""
        current_time = time.time()
        for task_id in list(cls.tasks.keys()):
            task = cls.tasks[task_id]
            if task["status"] in ("completed", "error", "cancelled"):
                if (
                    task["end_time"]
                    and (current_time - task["end_time"]) > max_age_seconds
                ):
                    cls.tasks.pop(task_id)
    
    @classmethod
    async def shutdown(cls):
        """Shutdown the task manager and clean up resources"""
        await cls.stop_task_loop()
