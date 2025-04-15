from typing import Generic, TypeVar, Optional, List, Dict, Any, Type, Union, Tuple, cast
from bson.objectid import ObjectId
from pydantic import BaseModel
from data_layer.models.base_model import MongoBaseModel, T
from data_layer.mongodb.connection import get_collection, get_async_collection
from pymongo.collection import Collection
from pymongo.results import InsertOneResult, UpdateResult, DeleteResult
import logging
from pymongo import ReturnDocument

logger = logging.getLogger(__name__)


class BaseMongoRepository(Generic[T]):
    """Base repository for MongoDB operations with generic CRUD functionality."""

    def __init__(self, model_class: Type[T]):
        """Initialize the repository with a model class."""
        self.model_class = model_class
        self.collection_name = model_class.collection_name
        try:
            self._collection = self.get_collection()
            self._async_collection = self.get_async_collection()
        except Exception as e:
            logger.warning(
                f"MongoDB collection initialization failed for {self.collection_name}: {str(e)}")
            self._collection = None
            self._async_collection = None
        self._model = model_class
        logger.info(
            f"Initialized {self.__class__.__name__} for collection {self.collection_name}")

    @property
    def collection(self) -> Optional[Collection]:
        """Get the MongoDB collection for this repository."""
        if self._collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return None
        return self._collection

    @property
    def async_collection(self) -> Any:
        """Get the async MongoDB collection for this repository."""
        if self._async_collection is None:
            logger.error(
                f"Async collection not available for {self.collection_name}")
            return None
        return self._async_collection

    @property
    def model(self) -> Type[T]:
        """Get the model class for this repository."""
        return self._model

    def get_collection(self) -> Optional[Collection]:
        """Get the MongoDB collection for this repository."""
        collection = get_collection(self.collection_name)
        if collection is None:
            logger.error(
                f"Failed to get MongoDB collection: {self.collection_name}")
            return None
        return collection

    def get_async_collection(self) -> Any:
        """Get the async MongoDB collection for this repository."""
        return get_async_collection(self.collection_name)

    # Synchronous methods

    def find_by_id(self, id: str) -> Optional[T]:
        """Find document by ID."""
        collection = self.get_collection()

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return None

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return None

        result = collection.find_one({"_id": obj_id})
        if result:
            return self.model_class.from_mongodb(result)
        return None

    def find_one(self, filter: Dict[str, Any]) -> Optional[T]:
        """Find one document by filter."""
        collection = self.get_collection()
        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return None

        result = collection.find_one(filter)
        if result:
            return self.model_class.from_mongodb(result)
        return None

    def find_many(self,
                  filter: Optional[Dict[str, Any]] = None,
                  skip: int = 0,
                  limit: int = 100,
                  sort: Optional[List[Tuple[str, int]]] = None) -> List[T]:
        """Find multiple documents with pagination and sorting."""
        collection = self.get_collection()

        # Apply default empty filter if none provided
        if filter is None:
            filter = {}

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return []

        cursor = collection.find(filter).skip(skip).limit(limit)

        # Apply sorting if provided
        if sort:
            cursor = cursor.sort(sort)

        return [self.model_class.from_mongodb(doc) for doc in cursor]

    def count(self, filter: Optional[Dict[str, Any]] = None) -> int:
        """Count documents matching filter."""
        collection = self.get_collection()
        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return 0

        return collection.count_documents(filter or {})

    def insert(self, model: T) -> str:
        """Insert a new document."""
        collection = self.get_collection()
        data = model.dict_for_mongodb()

        # Remove _id if it's None
        if "_id" in data and data["_id"] is None:
            del data["_id"]

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return ""

        result: InsertOneResult = collection.insert_one(data)
        return str(result.inserted_id)

    def update(self, id: str, data: Dict[str, Any]) -> Optional[T]:
        """Update document by ID."""
        collection = self.get_collection()

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return None

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return None

        # Set updated_at timestamp
        if "updated_at" not in data:
            from datetime import datetime
            data["updated_at"] = datetime.utcnow()

        result = collection.find_one_and_update(
            {"_id": obj_id},
            {"$set": data},
            return_document=ReturnDocument.AFTER
        )

        if result:
            return self.model_class.from_mongodb(result)
        return None

    def update_by_filter(self, filter: Dict[str, Any], data: Dict[str, Any]) -> int:
        """Update documents by filter, return number of documents modified."""
        collection = self.get_collection()

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return 0

        # Set updated_at timestamp
        if "updated_at" not in data:
            from datetime import datetime
            data["updated_at"] = datetime.utcnow()

        result: UpdateResult = collection.update_many(
            filter,
            {"$set": data}
        )

        return result.modified_count

    def delete(self, id: str) -> bool:
        """Delete document by ID."""
        collection = self.get_collection()

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return False

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return False

        result: DeleteResult = collection.delete_one({"_id": obj_id})
        return result.deleted_count > 0

    def delete_many(self, filter: Dict[str, Any]) -> int:
        """Delete documents by filter, return number of documents deleted."""
        collection = self.get_collection()

        if collection is None:
            logger.error(
                f"Collection not available for {self.collection_name}")
            return 0

        result: DeleteResult = collection.delete_many(filter)
        return result.deleted_count

    # Async methods

    async def async_find_by_id(self, id: str) -> Optional[T]:
        """Find document by ID (async)."""
        collection = self.get_async_collection()

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return None

        try:
            result = await collection.find_one({"_id": obj_id})
            if result:
                return self.model_class.from_mongodb(result)
            return None
        except Exception as e:
            logger.error(f"Error in async_find_by_id: {str(e)}")
            return None

    async def async_find_one(self, filter: Dict[str, Any]) -> Optional[T]:
        """Find one document by filter (async)."""
        collection = self.get_async_collection()
        try:
            result = await collection.find_one(filter)
            if result:
                return self.model_class.from_mongodb(result)
            return None
        except Exception as e:
            logger.error(f"Error in async_find_one: {str(e)}")
            return None

    async def async_find_many(self,
                              filter: Optional[Dict[str, Any]] = None,
                              skip: int = 0,
                              limit: int = 100,
                              sort: Optional[List[Tuple[str, int]]] = None) -> List[T]:
        """Find multiple documents with pagination and sorting (async)."""
        collection = self.get_async_collection()

        # Apply default empty filter if none provided
        if filter is None:
            filter = {}

        cursor = collection.find(filter).skip(skip).limit(limit)

        # Apply sorting if provided
        if sort:
            cursor = cursor.sort(sort)

        result = []
        try:
            # Properly await the async cursor
            async for doc in cursor:
                result.append(self.model_class.from_mongodb(doc))
        except Exception as e:
            logger.error(
                f"Error iterating cursor in async_find_many: {str(e)}")

        return result

    async def async_count(self, filter: Optional[Dict[str, Any]] = None) -> int:
        """Count documents matching filter (async)."""
        collection = self.get_async_collection()
        try:
            return await collection.count_documents(filter or {})
        except Exception as e:
            logger.error(f"Error in async_count: {str(e)}")
            return 0

    async def async_insert(self, model: T) -> str:
        """Insert a new document (async)."""
        collection = self.get_async_collection()
        data = model.dict_for_mongodb()

        # Remove _id if it's None
        if "_id" in data and data["_id"] is None:
            del data["_id"]

        try:
            result = await collection.insert_one(data)
            return str(result.inserted_id)
        except Exception as e:
            logger.error(f"Error in async_insert: {str(e)}")
            return ""

    async def async_update(self, id: str, data: Dict[str, Any]) -> Optional[T]:
        """Update document by ID (async)."""
        collection = self.get_async_collection()

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return None

        # Set updated_at timestamp
        if "updated_at" not in data:
            from datetime import datetime
            data["updated_at"] = datetime.utcnow()

        try:
            result = await collection.find_one_and_update(
                {"_id": obj_id},
                {"$set": data},
                return_document=ReturnDocument.AFTER
            )

            if result:
                return self.model_class.from_mongodb(result)
            return None
        except Exception as e:
            logger.error(f"Error in async_update: {str(e)}")
            return None

    async def async_update_by_filter(self, filter: Dict[str, Any], data: Dict[str, Any]) -> int:
        """Update documents by filter, return number of documents modified (async)."""
        collection = self.get_async_collection()

        # Set updated_at timestamp
        if "updated_at" not in data:
            from datetime import datetime
            data["updated_at"] = datetime.utcnow()

        try:
            result = await collection.update_many(
                filter,
                {"$set": data}
            )

            return result.modified_count
        except Exception as e:
            logger.error(f"Error in async_update_by_filter: {str(e)}")
            return 0

    async def async_delete(self, id: str) -> bool:
        """Delete document by ID (async)."""
        collection = self.get_async_collection()

        try:
            obj_id = ObjectId(id)
        except:
            logger.warning(f"Invalid ObjectId format: {id}")
            return False

        try:
            result = await collection.delete_one({"_id": obj_id})
            return result.deleted_count > 0
        except Exception as e:
            logger.error(f"Error in async_delete: {str(e)}")
            return False

    async def async_delete_many(self, filter: Dict[str, Any]) -> int:
        """Delete documents by filter, return number of documents deleted (async)."""
        collection = self.get_async_collection()
        try:
            result = await collection.delete_many(filter)
            return result.deleted_count
        except Exception as e:
            logger.error(f"Error in async_delete_many: {str(e)}")
            return 0
