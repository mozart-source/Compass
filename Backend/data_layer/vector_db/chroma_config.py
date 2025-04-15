from pathlib import Path
from typing import Optional
from pydantic import BaseModel
from Backend.core.config import settings

class ChromaConfig(BaseModel):
    persist_directory: str = settings.CHROMA_DB_PATH
    chroma_db_impl: str = "duckdb+parquet"
    anonymized_telemetry: bool = False
    allow_reset: bool = True
    is_persistent: bool = True

    @property
    def collection_name(self) -> str:
        return f"{settings.APP_NAME.lower()}_collection"

    def get_persist_directory(self) -> Path:
        """Get the absolute path to the ChromaDB persistence directory"""
        persist_dir = Path(self.persist_directory)
        persist_dir.mkdir(parents=True, exist_ok=True)
        return persist_dir

    def get_settings(self) -> dict:
        """Get ChromaDB settings as a dictionary"""
        return {
            "persist_directory": str(self.get_persist_directory()),
            "chroma_db_impl": self.chroma_db_impl,
            "anonymized_telemetry": self.anonymized_telemetry,
            "allow_reset": self.allow_reset,
            "is_persistent": self.is_persistent
        }

# Create a global instance of ChromaConfig
chroma_config = ChromaConfig()