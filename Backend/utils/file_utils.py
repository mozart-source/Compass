import os
import shutil
import mimetypes
import hashlib
from typing import Optional, List, Tuple
import aiofiles
import magic


def get_file_extension(filename: str) -> str:
    """Get file extension from filename."""
    return os.path.splitext(filename)[1].lower()


def get_mime_type(file_path: str) -> str:
    """Get MIME type of file."""
    mime = magic.Magic(mime=True)
    return mime.from_file(file_path)


def calculate_file_hash(file_path: str, algorithm: str = 'sha256') -> str:
    """Calculate file hash using specified algorithm."""
    hash_obj = hashlib.new(algorithm)
    with open(file_path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b''):
            hash_obj.update(chunk)
    return hash_obj.hexdigest()


async def save_uploaded_file(
    file_content: bytes,
    destination: str,
    filename: str,
    allowed_extensions: Optional[List[str]] = None
) -> Tuple[bool, str]:
    """
    Save uploaded file with validation.
    Returns (success, message).
    """
    if allowed_extensions:
        ext = get_file_extension(filename)
        if ext not in allowed_extensions:
            return False, f"File extension {ext} not allowed"

    try:
        os.makedirs(os.path.dirname(destination), exist_ok=True)
        async with aiofiles.open(destination, 'wb') as f:
            await f.write(file_content)
        return True, "File saved successfully"
    except Exception as e:
        return False, str(e)


def create_unique_filename(original_filename: str, directory: str) -> str:
    """Create unique filename to avoid overwrites."""
    base, ext = os.path.splitext(original_filename)
    counter = 1
    new_filename = original_filename

    while os.path.exists(os.path.join(directory, new_filename)):
        new_filename = f"{base}_{counter}{ext}"
        counter += 1

    return new_filename


def get_file_size(file_path: str, unit: str = 'bytes') -> float:
    """
    Get file size in specified unit.
    Units: bytes, kb, mb, gb
    """
    size_bytes = os.path.getsize(file_path)

    if unit == 'bytes':
        return size_bytes
    elif unit == 'kb':
        return size_bytes / 1024
    elif unit == 'mb':
        return size_bytes / (1024 * 1024)
    elif unit == 'gb':
        return size_bytes / (1024 * 1024 * 1024)
    else:
        raise ValueError("Invalid unit. Use 'bytes', 'kb', 'mb', or 'gb'.")


def is_file_empty(file_path: str) -> bool:
    """Check if file is empty."""
    return os.path.getsize(file_path) == 0


def safe_delete_file(file_path: str) -> Tuple[bool, str]:
    """Safely delete file with error handling."""
    try:
        if os.path.exists(file_path):
            os.remove(file_path)
            return True, "File deleted successfully"
        return False, "File does not exist"
    except Exception as e:
        return False, str(e)


def copy_file_safe(source: str, destination: str) -> Tuple[bool, str]:
    """Safely copy file with error handling."""
    try:
        shutil.copy2(source, destination)
        return True, "File copied successfully"
    except Exception as e:
        return False, str(e)


def get_file_info(file_path: str) -> dict:
    """Get comprehensive file information."""
    return {
        'name': os.path.basename(file_path),
        'extension': get_file_extension(file_path),
        'size': get_file_size(file_path),
        'mime_type': get_mime_type(file_path),
        'created_time': os.path.getctime(file_path),
        'modified_time': os.path.getmtime(file_path),
        'is_empty': is_file_empty(file_path)
    }
