import asyncio
import importlib
import logging
import threading
import time
import traceback
import uuid
from typing import Any, Callable, Dict, List, Optional, Tuple
from concurrent.futures import ThreadPoolExecutor


class TaskManager:
    """Manages background tasks using a separate event loop"""

    tasks: Dict[str, Dict[str, Any]] = {}
    _lock = threading.Lock()
    _task_loop = None
    _task_thread = None
    _executor = ThreadPoolExecutor(max_workers=10)
    _running = False
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
        function_path = f"{cls._workflows_path}.{function_path}"

        if "." in function_path:
            parts = function_path.split(".")

            if len(parts) >= 3:
                method_name = parts[-1]
                class_name = parts[-2]
                module_path = ".".join(parts[:-2])
                logging.info(module_path)

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
    def _task_loop_thread(cls):
        """Thread function that runs a separate event loop for background tasks"""
        # Create a new event loop for this thread
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        cls._task_loop = loop
        
        # Create a queue for tasks
        cls._task_queue = asyncio.Queue()
        
        # Start the event loop with our worker
        try:
            loop.run_until_complete(cls._task_worker())
        except Exception as e:
            logging.error(f"Task loop thread error: {e}")
        finally:
            loop.close()
            cls._task_loop = None
            logging.info("Task loop thread stopped")
    
    @classmethod
    async def _task_worker(cls):
        """Worker that processes tasks from the queue"""
        cls._running = True
        while cls._running:
            try:
                # Get a task from the queue
                task_info = await cls._task_queue.get()
                task_id = task_info["task_id"]
                function_path = task_info["function"]
                args = task_info["args"]
                kwargs = task_info["kwargs"]
                
                # Run the task
                await cls._run_async_function(task_id, function_path, *args, **kwargs)
                
                # Mark the task as done
                cls._task_queue.task_done()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logging.exception(f"Error in task worker: {e}")
    
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

                with cls._lock:
                    cls.tasks[task_id]["status"] = "completed"
                    cls.tasks[task_id]["result"] = result
                    cls.tasks[task_id]["end_time"] = time.time()
                    # Remove the task from the dictionary once completed
                    cls.tasks.pop(task_id)
            except Exception as e:
                logging.exception(f"Error executing async function: {e}")
                with cls._lock:
                    cls.tasks[task_id]["status"] = "error"
                    cls.tasks[task_id]["error"] = str(e)
                    cls.tasks[task_id]["traceback"] = traceback.format_exc()
                    cls.tasks[task_id]["end_time"] = time.time()
                    # Remove the task from the dictionary once it has an error
                    cls.tasks.pop(task_id)

        except Exception as e:
            logging.exception(f"Error in task {task_id}: {e}")
            with cls._lock:
                cls.tasks[task_id]["status"] = "error"
                cls.tasks[task_id]["error"] = str(e)
                cls.tasks[task_id]["traceback"] = traceback.format_exc()
                cls.tasks[task_id]["end_time"] = time.time()
                # Remove the task from the dictionary once it has an error
                cls.tasks.pop(task_id)

    @classmethod
    def start_task_loop(cls):
        """Start the background task loop if it's not already running"""
        if cls._task_thread is None or not cls._task_thread.is_alive():
            cls._running = True
            cls._task_thread = threading.Thread(target=cls._task_loop_thread, daemon=True)
            cls._task_thread.start()
            logging.info("Started background task loop thread")
    
    @classmethod
    def stop_task_loop(cls):
        """Stop the background task loop"""
        if cls._task_loop and cls._task_thread and cls._task_thread.is_alive():
            cls._running = False
            # Signal the loop to stop
            future = asyncio.run_coroutine_threadsafe(
                asyncio.sleep(0), cls._task_loop
            )
            future.result(timeout=5)  # Wait up to 5 seconds for clean shutdown
            cls._task_thread.join(timeout=5)
            logging.info("Stopped background task loop thread")
    
    @classmethod
    def run_function(cls, function_path: str, *args, **kwargs) -> str:
        """
        Run any Python async function in a background task loop

        Args:
            function_path: Import path to the function (e.g., 'module.submodule.function')
            *args: Positional arguments to pass to the function
            **kwargs: Keyword arguments to pass to the function

        Returns:
            task_id: Unique ID for the task
        """
        # Make sure the task loop is running
        cls.start_task_loop()
        
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
        
        # Add the task to the queue in the background thread
        if cls._task_loop:
            task_info = cls.tasks[task_id].copy()
            future = asyncio.run_coroutine_threadsafe(
                cls._task_queue.put(task_info), cls._task_loop
            )
            # Ensure the task is added to the queue
            future.result(timeout=5)
        else:
            logging.error("Task loop is not running, cannot add task")
            with cls._lock:
                cls.tasks[task_id]["status"] = "error"
                cls.tasks[task_id]["error"] = "Task loop is not running"
                cls.tasks[task_id]["end_time"] = time.time()
                # Remove the task from the dictionary once it has an error
                cls.tasks.pop(task_id)

        logging.info(f"Queued async function {function_path} as task {task_id}")
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
                # Remove the task from the dictionary once cancelled
                cls.tasks.pop(task_id)
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
    
    @classmethod
    def shutdown(cls):
        """Shutdown the task manager and clean up resources"""
        cls.stop_task_loop()
        if cls._executor:
            cls._executor.shutdown(wait=False)
