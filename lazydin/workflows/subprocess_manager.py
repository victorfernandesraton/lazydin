import asyncio
import importlib
import inspect
import io
import logging
import threading
import time
import uuid
from concurrent.futures import ThreadPoolExecutor
from typing import Any, Callable, Dict, List, Optional, Tuple, Union
import sys
import traceback

class TaskManager:
    """Manages background tasks using threads and asyncio"""
    
    # Dictionary to store task references and their outputs
    tasks: Dict[str, Dict[str, Any]] = {}
    _executor = ThreadPoolExecutor(max_workers=10)
    _lock = threading.Lock()
    
    @staticmethod
    def generate_task_id() -> str:
        """Generate a unique ID for a task"""
        return str(uuid.uuid4())
    
    @staticmethod
    def _import_function(function_path: str) -> Callable:
        """Import a function from its path"""
        try:
            if '.' in function_path:
                parts = function_path.split('.')
                
                # Check if it's likely a class method (at least 3 parts)
                if len(parts) >= 3:
                    # The last part is the method name
                    method_name = parts[-1]
                    
                    # The second-to-last part is the class name
                    class_name = parts[-2]
                    
                    # Everything before that is the module path
                    module_path = '.'.join(parts[:-2])
                    
                    # Import the module
                    module = importlib.import_module(module_path)
                    
                    # Get the class
                    class_obj = getattr(module, class_name)
                    
                    # Return a wrapper that will instantiate the class and call the method
                    async def class_method_wrapper(*args, **kwargs):
                        # Create an instance of the class
                        instance = class_obj()
                        
                        # Get the method from the instance
                        method = getattr(instance, method_name)
                        
                        # If there's only one argument and it's a dict, pass it directly
                        if len(args) == 1 and isinstance(args[0], dict):
                            return await method(args[0])
                        # Otherwise, package all args into a dict
                        else:
                            params = {}
                            for i, arg in enumerate(args):
                                params[f"arg{i}"] = arg
                            params.update(kwargs)
                            return await method(params)
                            
                    return class_method_wrapper
                else:
                    # Regular function import
                    module_path, function_name = function_path.rsplit('.', 1)
                    module = importlib.import_module(module_path)
                    return getattr(module, function_name)
            else:
                # Handle case where function is in the global namespace
                return globals()[function_path]
        except (ImportError, AttributeError) as e:
            logging.error(f"Failed to import function {function_path}: {e}")
            raise ImportError(f"Function {function_path} not found: {e}")
    
    @classmethod
    def _capture_output(cls, func, *args, **kwargs):
        """Capture stdout and stderr while running a function"""
        # Redirect stdout and stderr
        stdout_capture = io.StringIO()
        stderr_capture = io.StringIO()
        
        original_stdout = sys.stdout
        original_stderr = sys.stderr
        
        sys.stdout = stdout_capture
        sys.stderr = stderr_capture
        
        result = None
        error = None
        
        try:
            # Run the function
            result = func(*args, **kwargs)
            return {
                "status": "success",
                "result": result,
                "stdout": stdout_capture.getvalue(),
                "stderr": stderr_capture.getvalue()
            }
        except Exception as e:
            logging.exception(f"Error executing function: {e}")
            return {
                "status": "error",
                "error": str(e),
                "traceback": traceback.format_exc(),
                "stdout": stdout_capture.getvalue(),
                "stderr": stderr_capture.getvalue()
            }
        finally:
            # Restore stdout and stderr
            sys.stdout = original_stdout
            sys.stderr = original_stderr
    
    @classmethod
    def _run_sync_function(cls, task_id: str, function_path: str, *args, **kwargs):
        """Run a synchronous function and store its result"""
        try:
            # Import the function
            func = cls._import_function(function_path)
            
            # Check if the function is async
            if inspect.iscoroutinefunction(func) or (hasattr(func, '__call__') and inspect.iscoroutinefunction(func.__call__)):
                # Create a new event loop for this thread
                loop = asyncio.new_event_loop()
                asyncio.set_event_loop(loop)
                
                try:
                    # Run the async function in the new event loop
                    coro = func(*args, **kwargs)
                    if not asyncio.iscoroutine(coro):
                        # If the function didn't return a coroutine directly, it might be a class method
                        # that returns a coroutine when called
                        coro = func.__call__(*args, **kwargs)
                    
                    result = loop.run_until_complete(coro)
                    
                    with cls._lock:
                        cls.tasks[task_id]["status"] = "completed"
                        cls.tasks[task_id]["result"] = result
                        cls.tasks[task_id]["end_time"] = time.time()
                finally:
                    loop.close()
            else:
                # Run the sync function and capture its output
                output = cls._capture_output(func, *args, **kwargs)
                
                with cls._lock:
                    cls.tasks[task_id].update(output)
                    cls.tasks[task_id]["status"] = "completed" if output["status"] == "success" else "error"
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
        Run any Python function in a background thread
        
        Args:
            function_path: Import path to the function (e.g., 'module.submodule.function')
            *args: Positional arguments to pass to the function
            **kwargs: Keyword arguments to pass to the function
            
        Returns:
            task_id: Unique ID for the task
        """
        task_id = cls.generate_task_id()
        
        # Initialize task info
        with cls._lock:
            cls.tasks[task_id] = {
                "task_id": task_id,
                "function": function_path,
                "args": args,
                "kwargs": kwargs,
                "status": "running",
                "start_time": time.time(),
                "end_time": None,
                "result": None
            }
        
        # Submit the task to the thread pool
        cls._executor.submit(cls._run_sync_function, task_id, function_path, *args, **kwargs)
        
        logging.info(f"Started function {function_path} as task {task_id}")
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
                    if task["end_time"] and (current_time - task["end_time"]) > max_age_seconds:
                        cls.tasks.pop(task_id)


