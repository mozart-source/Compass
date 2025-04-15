from datetime import datetime, timedelta
from typing import Optional, Union
import pytz


def get_utc_now() -> datetime:
    """Get current UTC datetime."""
    return datetime.now(pytz.UTC)


def convert_to_timezone(dt: datetime, timezone: str = "UTC") -> datetime:
    """Convert datetime to specified timezone."""
    target_tz = pytz.timezone(timezone)
    if dt.tzinfo is None:
        dt = pytz.UTC.localize(dt)
    return dt.astimezone(target_tz)


def format_datetime(dt: datetime, format: str = "%Y-%m-%d %H:%M:%S") -> str:
    """Format datetime to string."""
    return dt.strftime(format)


def parse_datetime(dt_str: str, format: str = "%Y-%m-%d %H:%M:%S") -> datetime:
    """Parse datetime from string."""
    return datetime.strptime(dt_str, format)


def get_date_range(
    start_date: Union[str, datetime],
    end_date: Union[str, datetime],
    as_string: bool = False,
    date_format: str = "%Y-%m-%d"
) -> list:
    """Get list of dates between start_date and end_date."""
    if isinstance(start_date, str):
        start_date = datetime.strptime(start_date, date_format)
    if isinstance(end_date, str):
        end_date = datetime.strptime(end_date, date_format)

    date_list = []
    current_date = start_date
    while current_date <= end_date:
        if as_string:
            date_list.append(current_date.strftime(date_format))
        else:
            date_list.append(current_date)
        current_date += timedelta(days=1)

    return date_list


def add_time(
    dt: datetime,
    days: int = 0,
    hours: int = 0,
    minutes: int = 0,
    seconds: int = 0
) -> datetime:
    """Add time to datetime."""
    return dt + timedelta(
        days=days,
        hours=hours,
        minutes=minutes,
        seconds=seconds
    )


def subtract_time(
    dt: datetime,
    days: int = 0,
    hours: int = 0,
    minutes: int = 0,
    seconds: int = 0
) -> datetime:
    """Subtract time from datetime."""
    return dt - timedelta(
        days=days,
        hours=hours,
        minutes=minutes,
        seconds=seconds
    )


def get_time_difference(
    dt1: datetime,
    dt2: datetime,
    unit: str = "seconds"
) -> float:
    """Get time difference between two datetimes."""
    diff = abs(dt1 - dt2)
    if unit == "seconds":
        return diff.total_seconds()
    elif unit == "minutes":
        return diff.total_seconds() / 60
    elif unit == "hours":
        return diff.total_seconds() / 3600
    elif unit == "days":
        return diff.days
    else:
        raise ValueError(
            "Invalid unit. Use 'seconds', 'minutes', 'hours', or 'days'.")
