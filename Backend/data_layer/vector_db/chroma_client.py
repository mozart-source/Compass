import chromadb
from chromadb.config import Settings
from chromadb.api.types import (
    Document, Documents, EmbeddingFunction, Embeddings,
    QueryResult, Metadata, GetResult, OneOrMany, ID, IDs
)
from Backend.core.config import settings
import logging
import os
from typing import List, Dict, Any, Optional, Union, Sequence, cast
import numpy as np
from concurrent.futures import ThreadPoolExecutor

logger = logging.getLogger(__name__)


class ChromaClient:
    _instance = None
    _is_initialized = False

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(ChromaClient, cls).__new__(cls)
        return cls._instance

    def __init__(self):
        if not self._is_initialized:
            self.client = None
            self.collection = None
            self.thread_pool = ThreadPoolExecutor(max_workers=4)
            self._initialize_client()
            self.__class__._is_initialized = True

    def _initialize_client(self):
        """Initialize ChromaDB client with proper settings."""
        try:
            # Create data directory if it doesn't exist
            os.makedirs(settings.CHROMA_PERSIST_DIRECTORY, exist_ok=True)

            # Initialize client with persistent storage and optimized settings
            self.client = chromadb.PersistentClient(
                path=settings.CHROMA_PERSIST_DIRECTORY,
                settings=Settings(
                    anonymized_telemetry=settings.CHROMA_TELEMETRY_ENABLED,
                    allow_reset=settings.CHROMA_ALLOW_RESET,
                    is_persistent=True
                )
            )

            # Get or create collection with optimized settings
            self.collection = self.client.get_or_create_collection(
                name=settings.CHROMA_COLLECTION_NAME,
                metadata={
                    "hnsw:space": settings.CHROMA_METADATA_SPACE,
                    "hnsw:M": 16,  # Number of connections per element
                    "hnsw:ef_construction": 100,  # Size of the dynamic candidate list
                    "hnsw:ef": 50  # Size of the dynamic candidate list for search
                }
            )

            # Add a check to ensure collection is initialized
            if self.collection is None:
                raise RuntimeError("Failed to initialize ChromaDB collection")

            logger.info("Successfully initialized ChromaDB client")
        except Exception as e:
            logger.error(f"Failed to initialize ChromaDB client: {str(e)}")
            raise RuntimeError(
                f"Failed to initialize ChromaDB client: {str(e)}")

    def add_documents(
        self,
        documents: Union[str, List[str]],
        metadatas: Optional[Union[Dict[str, Any],
                                  List[Dict[str, Any]]]] = None,
        ids: Optional[Union[str, List[str]]] = None,
        batch_size: int = 32
    ) -> None:
        """Add documents to the collection with batching."""
        try:
            # Ensure collection is initialized
            if self.collection is None:
                self._initialize_client()
                if self.collection is None:
                    raise RuntimeError(
                        "ChromaDB collection could not be initialized")

            # Convert inputs to lists
            docs = [documents] if isinstance(documents, str) else documents
            metas = [metadatas] if isinstance(metadatas, dict) else (
                metadatas or [None] * len(docs))
            doc_ids = [ids] if isinstance(ids, str) else (
                ids or [f"doc_{i}" for i in range(len(docs))])

            # Validate lengths
            if len(metas) != len(docs) or len(doc_ids) != len(docs):
                raise ValueError(
                    "Number of documents, metadatas, and IDs must match")

            # Process in batches
            for i in range(0, len(docs), batch_size):
                batch_end = min(i + batch_size, len(docs))
                batch_docs = docs[i:batch_end]
                batch_metas = metas[i:batch_end] if metas else None
                batch_ids = doc_ids[i:batch_end]

                # Cast types for ChromaDB
                batch_metas_cast = cast(Optional[List[Metadata]], batch_metas)
                batch_ids_cast = cast(List[str], batch_ids)

                self.collection.add(
                    documents=batch_docs,
                    metadatas=batch_metas_cast,
                    ids=batch_ids_cast
                )

            logger.info(
                f"Successfully added {len(docs)} documents in {(len(docs) - 1) // batch_size + 1} batches")
        except Exception as e:
            logger.error(f"Failed to add documents: {str(e)}")
            raise

    def query(
        self,
        query_text: Union[str, List[str]],
        n_results: int = 5,
        where: Optional[Dict[str, Any]] = None
    ) -> QueryResult:
        """Optimized query with type safety."""
        try:
            # Ensure collection is initialized
            if self.collection is None:
                self._initialize_client()
                if self.collection is None:
                    raise RuntimeError(
                        "ChromaDB collection could not be initialized")

            # Convert query to list format
            query_texts = [query_text] if isinstance(
                query_text, str) else query_text

            # Execute query with proper type casting
            results = self.collection.query(
                query_texts=cast(Documents, query_texts),
                n_results=n_results,
                where=where
            )
            return results
        except Exception as e:
            logger.error(f"Failed to query collection: {str(e)}")
            raise

    def delete(
        self,
        ids: Optional[Union[str, List[str]]] = None,
        where: Optional[Dict[str, Any]] = None
    ) -> None:
        """Delete documents with proper type handling."""
        try:
            if self.collection is None:
                raise RuntimeError("Collection not initialized")

            # Convert single ID to list and cast types
            doc_ids = [ids] if isinstance(ids, str) else ids
            doc_ids_cast = cast(Optional[List[str]], doc_ids)

            self.collection.delete(
                ids=doc_ids_cast,
                where=where
            )
            logger.info(
                f"Successfully deleted documents: {len(doc_ids) if doc_ids else 'all matching where clause'}")
        except Exception as e:
            logger.error(f"Failed to delete documents: {str(e)}")
            raise

    def __del__(self):
        """Cleanup resources."""
        if hasattr(self, 'thread_pool'):
            self.thread_pool.shutdown(wait=True)


# Initialize global client
chroma_client = ChromaClient()
