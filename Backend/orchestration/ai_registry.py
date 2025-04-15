import json
import os
from typing import Dict, Any, Optional


def load_config(file_name: str) -> Dict[str, Any]:
    """Load configuration from a JSON file."""
    current_dir = os.path.dirname(os.path.abspath(__file__))
    config_path = os.path.join(os.path.dirname(
        current_dir), "core", "configs", file_name)
    try:
        with open(config_path, "r") as f:
            return json.load(f)
    except FileNotFoundError:
        return {}


class AIRegistry:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if not hasattr(self, 'initialized'):
            self.domain_config = load_config("domain_config.json")
            self.cache_config = load_config("cache_config.json")
            self.llm_config = load_config("llm_config.json")
            self.logging_config = load_config("logging_config.json")
            self.initialized = True

    def get_prompt_template(self, domain: str, variant: str = "default") -> str:
        """Get prompt template for a domain and intent variant."""
        config = self.domain_config.get(
            domain, self.domain_config.get('default', {}))
        templates = config.get('prompt_templates', {})

        # First try to get the specific variant
        if variant in templates:
            return templates[variant]

        # If not found, try to get a default template
        if "default" in templates:
            return templates["default"]

        # Fallback to a generic template
        return """
        User Input: {{ user_prompt }}
        Intent: {{ intent }} on {{ target }}
        Context: {{ context_data }}
        
        Task:
        - If 'retrieve', extract the requested information.
        - If 'analyze', provide deep insights and trends.
        - If 'plan', organize and propose an actionable plan.
        - If 'summarize', provide a concise summary.
        
        Additional Knowledge (if available):
        {{ rag_data }}
        """

    def get_domain_config(self, domain: str) -> Dict[str, Any]:
        """Get configuration for a specific domain."""
        return self.domain_config.get(domain, self.domain_config.get("default", {}))

    def get_llm_config(self) -> Dict[str, Any]:
        """Get LLM configuration."""
        return self.llm_config

    def get_cache_config(self) -> Dict[str, Any]:
        """Get cache configuration."""
        return self.cache_config

    def get_rag_settings(self, domain: str) -> Dict[str, Any]:
        """Get RAG settings for a specific domain."""
        domain_config = self.get_domain_config(domain)
        return domain_config.get("rag_settings", {})

    def get_logging_config(self) -> Dict[str, Any]:
        """Get logging configuration."""
        return self.logging_config


# Create singleton instance
ai_registry = AIRegistry()
