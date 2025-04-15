import re
from typing import List, Optional
import unicodedata
import random
import string


def slugify(text: str) -> str:
    """
    Convert text to URL-friendly slug.
    Example: "Hello World!" -> "hello-world"
    """
    # Convert to lowercase and normalize unicode characters
    text = text.lower()
    text = unicodedata.normalize('NFKD', text).encode(
        'ascii', 'ignore').decode('utf-8')

    # Replace any non-word character with a dash
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[-\s]+', '-', text).strip('-')

    return text


def generate_random_string(length: int = 10, include_special: bool = False) -> str:
    """Generate a random string of specified length."""
    characters = string.ascii_letters + string.digits
    if include_special:
        characters += string.punctuation

    return ''.join(random.choice(characters) for _ in range(length))


def truncate(text: str, max_length: int, suffix: str = '...') -> str:
    """Truncate text to specified length with suffix."""
    if len(text) <= max_length:
        return text
    return text[:max_length - len(suffix)] + suffix


def extract_emails(text: str) -> List[str]:
    """Extract all email addresses from text."""
    email_pattern = r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    return re.findall(email_pattern, text)


def extract_urls(text: str) -> List[str]:
    """Extract all URLs from text."""
    url_pattern = r'http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+'
    return re.findall(url_pattern, text)


def remove_extra_whitespace(text: str) -> str:
    """Remove extra whitespace from text."""
    return ' '.join(text.split())


def is_palindrome(text: str, ignore_case: bool = True, ignore_spaces: bool = True) -> bool:
    """Check if text is a palindrome."""
    if ignore_case:
        text = text.lower()
    if ignore_spaces:
        text = ''.join(text.split())
    return text == text[::-1]


def count_words(text: str) -> int:
    """Count number of words in text."""
    return len(text.split())


def mask_string(text: str, mask_char: str = '*', visible_start: int = 4, visible_end: int = 4) -> str:
    """
    Mask part of a string, showing only specified number of characters at start and end.
    Example: mask_string("1234567890", visible_start=4, visible_end=2) -> "1234****90"
    """
    if len(text) <= visible_start + visible_end:
        return text

    masked_length = len(text) - visible_start - visible_end
    return text[:visible_start] + mask_char * masked_length + text[-visible_end:]


def normalize_string(text: str, form: str = 'NFKC') -> str:
    """Normalize unicode string."""
    return unicodedata.normalize(form, text)


def extract_numbers(text: str) -> List[str]:
    """Extract all numbers from text."""
    return re.findall(r'\d+(?:\.\d+)?', text)


def camel_to_snake(text: str) -> str:
    """Convert camelCase to snake_case."""
    pattern = re.compile(r'(?<!^)(?=[A-Z])')
    return pattern.sub('_', text).lower()


def snake_to_camel(text: str) -> str:
    """Convert snake_case to camelCase."""
    components = text.split('_')
    return components[0] + ''.join(x.title() for x in components[1:])
