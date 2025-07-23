import importlib
import inspect
import logging
import pkgutil
from typing import Dict, List, Any


class WorkflowsAnalyzer:
    """
    Analyzes workflows by reading submodules from a given path and extracting
    documentation from the classes exposed by those submodules.
    """

    def __init__(self, workflows_path: str = "workflows"):
        """
        Initialize the WorkflowsAnalyzer with a specific workflows path.

        Args:
            workflows_path: The import path to the workflows directory (e.g., 'workflows')
        """
        self.workflows_path = workflows_path
        self.logger = logging.getLogger(__name__)

    def get_submodules(self) -> List[str]:
        """
        Get a list of all submodules in the workflows path.

        Returns:
            List of submodule names
        """
        try:
            # Import the base package
            package = importlib.import_module(self.workflows_path)
            package_path = getattr(package, "__path__", [])

            # Get all submodules
            submodules = []
            for _, name, is_pkg in pkgutil.iter_modules(package_path):
                # Include both packages and individual modules
                submodules.append(name)

            return submodules
        except ImportError as e:
            self.logger.error(
                f"Error importing workflows path {self.workflows_path}: {e}"
            )
            return []

    def get_classes_from_module(self, module_name: str) -> List[Dict[str, Any]]:
        """
        Get all classes from a specific module.

        Args:
            module_name: The name of the module to analyze

        Returns:
            List of dictionaries containing class information
        """
        full_module_path = f"{self.workflows_path}.{module_name}"
        logging.info(full_module_path)
        try:
            # Import the module
            module = importlib.import_module(full_module_path)

            # Get all classes defined in the module
            classes = []
            for name, obj in inspect.getmembers(module, inspect.isclass):
                # Include classes defined in this module or imported in __init__.py
                if obj.__module__ == full_module_path or (
                    obj.__module__.startswith(f"{self.workflows_path}.{module_name}.")
                    and hasattr(module, name)
                ):
                    class_info = {
                        "name": name,
                        "docstring": inspect.getdoc(obj) or "",
                        "methods": self._get_methods_info(obj),
                        "module": module_name,
                        "full_path": f"{full_module_path}.{name}",
                    }
                    classes.append(class_info)

            return classes
        except ImportError as e:
            self.logger.error(f"Error importing module {full_module_path}: {e}")
            return []

    def _get_methods_info(self, cls) -> List[Dict[str, str]]:
        """
        Get information about the methods of a class.

        Args:
            cls: The class to analyze

        Returns:
            List of dictionaries containing method information
        """
        methods = []
        for name, method in inspect.getmembers(cls, inspect.isfunction):
            # Skip private methods
            if name.startswith("_"):
                continue

            method_info = {
                "name": name,
                "docstring": inspect.getdoc(method) or "",
                "signature": str(inspect.signature(method)),
            }
            methods.append(method_info)

        return methods

    def analyze_workflows(self) -> Dict[str, List[Dict[str, Any]]]:
        """
        Analyze all workflows in the specified path.

        Returns:
            Dictionary mapping module names to lists of class information
        """
        result = {}
        submodules = self.get_submodules()

        for module_name in submodules:
            classes = self.get_classes_from_module(module_name)
            if classes:
                result[module_name] = classes

        return result

    def get_workflow_documentation(self) -> Dict[str, Any]:
        """
        Get documentation for all workflows.

        Returns:
            Dictionary containing workflow documentation
        """
        workflows = self.analyze_workflows()

        documentation = {"workflows_path": self.workflows_path, "modules": []}

        for module_name, classes in workflows.items():
            module_info = {"name": module_name, "classes": []}

            for class_info in classes:
                module_info["classes"].append(
                    {
                        "name": class_info["name"],
                        "docstring": class_info["docstring"],
                        "methods": class_info["methods"],
                        "full_path": class_info["full_path"],
                    }
                )

            documentation["modules"].append(module_info)

        return documentation
