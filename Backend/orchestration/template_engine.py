from jinja2 import Environment, Template
from typing import Dict, Any


def render_template(template_str: str, context: Dict[str, Any]) -> str:
    """
    Render a template string with the provided context using Jinja2.
    
    Supports advanced Jinja2 features including:
    - Conditional statements (if/else)
    - Loops (for)
    - Filters
    - Macros
    
    Args:
        template_str: The template string containing Jinja2 syntax
        context: Dictionary of variables to use in the template
        
    Returns:
        The rendered template as a string
    """
    # Create a Jinja2 environment with autoescape disabled for better performance
    # Since we're not rendering HTML, we don't need autoescaping
    env = Environment(autoescape=False, trim_blocks=True, lstrip_blocks=True)
    
    # Create a template from the template string
    template = env.from_string(template_str)
    
    # Render the template with the provided context
    return template.render(**context)


def format_list_for_template(items, bullet_char="•") -> str:
    """
    Format a list of items as a bulleted list for templates.
    
    Args:
        items: List of items to format
        bullet_char: Character to use as bullet point
        
    Returns:
        Formatted string with bullet points
    """
    if not items:
        return ""
        
    return "\n".join(f"{bullet_char} {item}" for item in items)
