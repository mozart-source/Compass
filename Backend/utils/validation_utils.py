from typing import Any, Dict, List, Optional
import re
from datetime import datetime
from email_validator import validate_email, EmailNotValidError


def validate_email_format(email: str) -> bool:
    """Validate email format."""
    try:
        validate_email(email)
        return True
    except EmailNotValidError:
        return False


def validate_password_strength(password: str) -> Dict[str, Any]:
    """
    Validate password strength.
    Returns a dictionary with validation results and requirements.
    """
    min_length = 8
    has_upper = bool(re.search(r'[A-Z]', password))
    has_lower = bool(re.search(r'[a-z]', password))
    has_digit = bool(re.search(r'\d', password))
    has_special = bool(re.search(r'[!@#$%^&*(),.?":{}|<>]', password))
    is_long_enough = len(password) >= min_length

    is_valid = all([has_upper, has_lower, has_digit,
                   has_special, is_long_enough])

    return {
        "is_valid": is_valid,
        "requirements": {
            "min_length": min_length,
            "has_upper": has_upper,
            "has_lower": has_lower,
            "has_digit": has_digit,
            "has_special": has_special,
            "is_long_enough": is_long_enough
        }
    }


def validate_date_format(date_str: str, format: str = "%Y-%m-%d") -> bool:
    """Validate if a string matches the specified date format."""
    try:
        datetime.strptime(date_str, format)
        return True
    except ValueError:
        return False


def validate_phone_number(phone: str, country: str = "EG") -> Dict[str, Any]:
    """
    Validate phone number format based on country.
    Currently supports Egyptian phone numbers by default.

    Args:
        phone: Phone number to validate
        country: Country code (default: "EG" for Egypt)

    Returns:
        Dict containing validation result and details
    """
    if country == "EG":
        # Egyptian phone number format: +20XXXXXXXXXX
        # Operators: 10, 11, 12, 15 (Vodafone, Etisalat, Orange, WE)
        pattern = r'^\+20(10|11|12|15)\d{8}$'

        # Remove any spaces or dashes
        phone = ''.join(filter(str.isdigit, phone))
        if not phone.startswith('20'):
            phone = '20' + phone
        phone = '+' + phone

        is_valid = bool(re.match(pattern, phone))
        operator = None
        if is_valid:
            operator_codes = {
                '10': 'Vodafone',
                '11': 'Etisalat',
                '12': 'Orange',
                '15': 'WE'
            }
            operator = operator_codes.get(phone[3:5])

        return {
            "is_valid": is_valid,
            "formatted_number": phone if is_valid else None,
            "operator": operator,
            "country": "Egypt",
            "requirements": {
                "format": "+20XXXXXXXXXX",
                "allowed_operators": ["Vodafone (10)", "Etisalat (11)", "Orange (12)", "WE (15)"],
                "length": 13  # Including +20 and the 9 digits
            }
        }
    else:
        # Default international format for other countries
        pattern = r'^\+?1?\d{9,15}$'
        is_valid = bool(re.match(pattern, phone))
        return {
            "is_valid": is_valid,
            "formatted_number": phone if is_valid else None,
            "country": country,
            "requirements": {
                "format": "International format",
                "length": "9-15 digits"
            }
        }


def validate_username(username: str) -> Dict[str, Any]:
    """
    Validate username format.
    Returns a dictionary with validation results and requirements.
    """
    min_length = 3
    max_length = 30
    pattern = r'^[a-zA-Z0-9_-]+$'

    is_valid = (
        len(username) >= min_length and
        len(username) <= max_length and
        bool(re.match(pattern, username))
    )

    return {
        "is_valid": is_valid,
        "requirements": {
            "min_length": min_length,
            "max_length": max_length,
            "allowed_characters": "letters, numbers, underscore, and hyphen"
        }
    }


def validate_url(url: str) -> bool:
    """Validate URL format."""
    pattern = r'^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$'
    return bool(re.match(pattern, url))
