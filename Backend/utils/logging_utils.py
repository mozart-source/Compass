import logging
import logging.handlers
import os
from datetime import datetime
from typing import Optional, Dict, Any
import json
from pathlib import Path


def get_logger(name: str, log_dir: str = 'logs') -> logging.Logger:
    """
    Get a configured logger with both file and console handlers.

    Args:
        name: Logger name
        log_dir: Directory for log files
    """
    log_file = os.path.join(log_dir, f'{name}.log')
    return setup_logger(
        name=name,
        log_file=log_file,
        level=logging.INFO,
        format_string='[%(asctime)s] %(levelname)s [%(name)s:%(lineno)s] %(message)s'
    )


def setup_logger(
    name: str,
    log_file: str,
    level: int = logging.INFO,
    rotation: str = 'midnight',
    format_string: Optional[str] = None,
    backup_count: int = 30
) -> logging.Logger:
    """
    Set up a logger with file and console handlers.

    Args:
        name: Logger name
        log_file: Path to log file
        level: Logging level
        rotation: When to rotate logs ('midnight' or 'size')
        format_string: Custom format string for logs
        backup_count: Number of backup files to keep
    """
    if format_string is None:
        format_string = '[%(asctime)s] %(levelname)s [%(name)s:%(lineno)s] %(message)s'

    logger = logging.getLogger(name)
    logger.setLevel(level)

    formatter = logging.Formatter(format_string)

    # Ensure log directory exists
    os.makedirs(os.path.dirname(log_file), exist_ok=True)

    # File handler with rotation
    if rotation == 'midnight':
        file_handler = logging.handlers.TimedRotatingFileHandler(
            log_file,
            when='midnight',
            interval=1,
            backupCount=backup_count
        )
    else:
        file_handler = logging.handlers.RotatingFileHandler(
            log_file,
            maxBytes=10*1024*1024,  # 10MB
            backupCount=backup_count
        )

    file_handler.setFormatter(formatter)
    logger.addHandler(file_handler)

    # Console handler
    console_handler = logging.StreamHandler()
    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)

    return logger


class JSONLogger:
    """Logger that formats messages as JSON."""

    def __init__(
        self,
        name: str,
        log_file: str,
        level: int = logging.INFO,
        additional_fields: Optional[Dict[str, Any]] = None
    ):
        self.logger = logging.getLogger(name)
        self.logger.setLevel(level)
        self.additional_fields = additional_fields or {}

        # Ensure log directory exists
        os.makedirs(os.path.dirname(log_file), exist_ok=True)

        # JSON formatter
        formatter = logging.Formatter('%(message)s')

        # File handler
        file_handler = logging.handlers.TimedRotatingFileHandler(
            log_file,
            when='midnight',
            interval=1,
            backupCount=30
        )
        file_handler.setFormatter(formatter)
        self.logger.addHandler(file_handler)

    def _format_message(
        self,
        level: str,
        message: str,
        **kwargs
    ) -> str:
        """Format log message as JSON."""
        log_data = {
            'timestamp': datetime.utcnow().isoformat(),
            'level': level,
            'message': message,
            **self.additional_fields,
            **kwargs
        }
        return json.dumps(log_data)

    def info(self, message: str, **kwargs):
        """Log info message."""
        self.logger.info(self._format_message('INFO', message, **kwargs))

    def error(self, message: str, **kwargs):
        """Log error message."""
        self.logger.error(self._format_message('ERROR', message, **kwargs))

    def warning(self, message: str, **kwargs):
        """Log warning message."""
        self.logger.warning(self._format_message('WARNING', message, **kwargs))

    def debug(self, message: str, **kwargs):
        """Log debug message."""
        self.logger.debug(self._format_message('DEBUG', message, **kwargs))


def get_error_logger(
    name: str,
    log_dir: str = 'logs/errors'
) -> logging.Logger:
    """
    Get a logger specifically for error tracking.
    """
    log_file = os.path.join(log_dir, f'{name}_errors.log')
    logger = setup_logger(
        f'{name}_error',
        log_file,
        level=logging.ERROR,
        format_string='[%(asctime)s] %(levelname)s [%(name)s] %(message)s\nStack Trace:\n%(stack_info)s\n'
    )
    return logger


def get_audit_logger(
    name: str,
    log_dir: str = 'logs/audit'
) -> JSONLogger:
    """
    Get a JSON logger for audit tracking.
    """
    log_file = os.path.join(log_dir, f'{name}_audit.log')
    return JSONLogger(
        f'{name}_audit',
        log_file,
        additional_fields={'service': name}
    )
