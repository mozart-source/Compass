from typing import List, Union, Dict, Optional, Any
from sentence_transformers import SentenceTransformer
import numpy as np
from Backend.ai_services.base.ai_service_base import AIServiceBase
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result
from Backend.data_layer.repositories.ai_model_repository import AIModelRepository
from sqlalchemy.ext.asyncio import AsyncSession
from functools import lru_cache
import hashlib
import time

logger = get_logger(__name__)


class EmbeddingService(AIServiceBase):
    # Class variable to hold the single instance
    _instance = None
    _is_initialized = False

    def __new__(cls, model_name: str = 'all-MiniLM-L6-v2', db_session: Optional[AsyncSession] = None):
        """Implement singleton pattern to ensure model is loaded only once."""
        if cls._instance is None:
            logger.info("Creating new EmbeddingService singleton instance")
            cls._instance = super(EmbeddingService, cls).__new__(cls)
        return cls._instance

    def __init__(self, model_name: str = 'all-MiniLM-L6-v2', db_session: Optional[AsyncSession] = None):
        """Initialize the model only once."""
        # Skip initialization if already done
        if not self._is_initialized:
            logger.info(
                f"Initializing EmbeddingService with model: {model_name}")
            super().__init__("embedding")
            self.model_name = model_name
            self.model_version = "1.0.0"
            self.model = None
            self.dimension = None
            self.db_session = db_session
            self.model_repository = AIModelRepository(
                db_session) if db_session else None
            self._current_model_id: Optional[int] = None
            self._initialize_model()
            self._cache = {}
            # Mark as initialized
            self.__class__._is_initialized = True
        else:
            logger.debug("Reusing existing EmbeddingService instance")

    async def _get_or_create_model(self) -> Optional[int]:
        """Get or create AI model record in database."""
        if not self.model_repository:
            return None

        try:
            model = await self.model_repository.get_model_by_name_version(
                name=self.model_name,
                version=self.model_version
            )

            if not model:
                model = await self.model_repository.create_model({
                    "name": self.model_name,
                    "version": self.model_version,
                    "type": "embedding",
                    "provider": "sentence-transformers",
                    "model_metadata": {
                        "dimension": self.dimension,
                        "max_sequence_length": self.model.max_seq_length if self.model else None
                    },
                    "status": "active"
                })

            return model.id
        except Exception as e:
            logger.error(f"Error getting/creating AI model: {str(e)}")
            return None

    async def _update_model_stats(self, latency: float, success: bool = True) -> None:
        """Update model usage statistics."""
        if self.model_repository and self._current_model_id:
            try:
                await self.model_repository.update_model_stats(
                    self._current_model_id,
                    latency,
                    success
                )
            except Exception as e:
                logger.error(f"Error updating model stats: {str(e)}")

    def _initialize_model(self) -> None:
        """Initialize the embedding model with error handling."""
        try:
            logger.info(
                f"Loading SentenceTransformer model: {self.model_name}")
            self.model = SentenceTransformer(self.model_name)
            self.dimension = self.model.get_sentence_embedding_dimension()
            logger.info(
                f"Model loaded successfully with dimension: {self.dimension}")
        except Exception as e:
            logger.error(f"Failed to initialize embedding model: {str(e)}")
            self.model = None
            self.dimension = None
            raise

    @lru_cache(maxsize=1000)
    def _generate_cache_key(self, text: str) -> str:
        """Generate cache key for embeddings."""
        return f"embedding:{hashlib.sha256(text.encode()).hexdigest()}"

    @cache_response(ttl=3600)
    async def get_embedding(
        self,
        text: Union[str, List[str]],
        normalize: bool = True,
        batch_size: int = 32
    ) -> Union[List[float], List[List[float]]]:
        """Generate embeddings with batching and normalization."""
        try:
            if not self._current_model_id:
                self._current_model_id = await self._get_or_create_model()

            start_time = time.time()
            success = True

            if self.model is None:
                raise RuntimeError("Embedding model not initialized")

            if isinstance(text, list):
                result = await self._batch_encode(text, batch_size, normalize)
                latency = time.time() - start_time
                await self._update_model_stats(latency, success)
                return result

            # Check cache for single text
            cache_key = self._generate_cache_key(text)
            if cached_embedding := self._cache.get(cache_key):
                return cached_embedding

            # Generate new embedding
            embedding = self.model.encode(
                text,
                normalize_embeddings=normalize,
                convert_to_tensor=False  # Return numpy array
            )

            # Convert to list and cache
            embedding_list = embedding.tolist()
            self._cache[cache_key] = embedding_list

            latency = time.time() - start_time
            await self._update_model_stats(latency, success)
            return embedding_list

        except Exception as e:
            success = False
            latency = time.time() - start_time
            await self._update_model_stats(latency, success)
            logger.error(f"Error generating embedding: {str(e)}")
            raise

    async def _batch_encode(
        self,
        texts: List[str],
        batch_size: int,
        normalize: bool
    ) -> List[List[float]]:
        """Process texts in batches with caching."""
        if self.model is None:
            raise RuntimeError("Embedding model not initialized")

        all_embeddings = []
        batch_texts = []
        cached_indices = {}

        # Check cache and collect texts needing embedding
        for i, text in enumerate(texts):
            cache_key = self._generate_cache_key(text)
            if cached_embedding := self._cache.get(cache_key):
                all_embeddings.append(cached_embedding)
            else:
                batch_texts.append(text)
                cached_indices[len(batch_texts) - 1] = i

        # Process uncached texts in batches
        if batch_texts:
            for i in range(0, len(batch_texts), batch_size):
                batch = batch_texts[i:i + batch_size]
                embeddings = self.model.encode(
                    batch,
                    normalize_embeddings=normalize,
                    convert_to_tensor=False  # Return numpy array
                )

                # Convert to list and cache
                for j, embedding in enumerate(embeddings):
                    embedding_list = embedding.tolist()
                    text_idx = i + j
                    original_idx = cached_indices[text_idx]
                    cache_key = self._generate_cache_key(batch_texts[text_idx])
                    self._cache[cache_key] = embedding_list
                    all_embeddings.insert(original_idx, embedding_list)

        return all_embeddings

    async def compare_embeddings(
        self,
        embedding1: List[float],
        embedding2: List[float]
    ) -> float:
        """Compare embeddings using cosine similarity."""
        if self.model is None:
            raise RuntimeError("Embedding model not initialized")

        # Convert to numpy arrays for efficient computation
        vec1 = np.array(embedding1)
        vec2 = np.array(embedding2)

        # Compute cosine similarity
        similarity = np.dot(vec1, vec2) / \
            (np.linalg.norm(vec1) * np.linalg.norm(vec2))
        return float(similarity)

    def get_model_info(self) -> Dict[str, Any]:
        """Get model information and configuration."""
        if self.model is None:
            raise RuntimeError("Embedding model not initialized")

        return {
            "model_name": self.model_name,
            "model_version": self.model_version,
            "embedding_dimension": self.dimension,
            "max_sequence_length": self.model.max_seq_length,
            "model_type": "sentence-transformer",
            "cache_size": len(self._cache)
        }

    def clear_cache(self) -> None:
        """Clear the embedding cache."""
        self._cache.clear()
        self._generate_cache_key.cache_clear()
