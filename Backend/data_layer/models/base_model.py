from typing import Optional, Any, Dict, ClassVar, Type, List, TypeVar, Generic, Set, Union
from pydantic import BaseModel, Field, ConfigDict, field_validator, model_validator
from datetime import datetime
import uuid
from bson.objectid import ObjectId

# Custom type for MongoDB ObjectId


class PyObjectId(str):
    """ObjectId field for Pydantic models."""

    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not isinstance(v, ObjectId):
            if not isinstance(v, str):
                raise TypeError("ObjectId required")
            try:
                v = ObjectId(v)
            except ValueError:
                raise ValueError("Invalid ObjectId")
        return str(v)

    @classmethod
    def __get_pydantic_core_schema__(cls, source_type, handler):
        from pydantic_core import core_schema
        return core_schema.union_schema([
            core_schema.is_instance_schema(ObjectId),
            core_schema.chain_schema([
                core_schema.str_schema(),
                core_schema.no_info_plain_validator_function(cls.validate),
            ])
        ])


class MongoBaseModel(BaseModel):
    """Base model for MongoDB documents with automatic ID generation and timestamps."""

    id: Optional[str] = Field(
        default_factory=lambda: str(ObjectId()), alias="_id")
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    # Metadata for the collection name
    collection_name: ClassVar[str] = "base"

    # Config for Pydantic model
    model_config = ConfigDict(
        populate_by_name=True,
        arbitrary_types_allowed=True,
        validate_assignment=True,
        str_strip_whitespace=True,
        json_schema_extra={
            "example": {
                "id": "507f1f77bcf86cd799439011",
                "created_at": "2023-01-01T00:00:00",
                "updated_at": "2023-01-01T00:00:00"
            }
        }
    )

    @model_validator(mode='before')
    @classmethod
    def set_updated_at(cls, data: Any) -> Any:
        """Update the updated_at field on every save."""
        if isinstance(data, dict):
            data["updated_at"] = datetime.utcnow()
        return data

    def dict_for_mongodb(self) -> Dict[str, Any]:
        """Convert to a MongoDB-friendly dict with ObjectId and proper field names."""
        data = self.model_dump(by_alias=True)

        # Convert string ID to ObjectId if necessary
        if "_id" in data and isinstance(data["_id"], str):
            try:
                data["_id"] = ObjectId(data["_id"])
            except:
                pass

        return data

    @classmethod
    def from_mongodb(cls, data: Dict[str, Any]) -> Optional["MongoBaseModel"]:
        """Create model instance from MongoDB document."""
        if not data:
            return None

        # Ensure ObjectId is converted to string
        if "_id" in data and isinstance(data["_id"], ObjectId):
            data["_id"] = str(data["_id"])

        return cls(**data)


# Type variable for generic repositories
T = TypeVar('T', bound=MongoBaseModel)
