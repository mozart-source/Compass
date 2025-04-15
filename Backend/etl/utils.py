import os
from dotenv import load_dotenv


def load_env(env_path: str = "etl/config/db_secrets.env"):
    """Load environment variables from a .env file if it exists."""
    if os.path.exists(env_path):
        load_dotenv(env_path)

# Add more utility functions as needed
