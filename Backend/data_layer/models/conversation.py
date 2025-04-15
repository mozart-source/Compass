from typing import List, Optional, Dict, Any, ClassVar
from pydantic import Field, field_validator
from data_layer.models.base_model import MongoBaseModel
from datetime import datetime


class Message(MongoBaseModel):
    """Individual message in a conversation."""

    role: str = Field(...,
                      description="Role of the message sender (user, assistant, system)")
    content: str = Field(..., description="Content of the message")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Time the message was sent")
    metadata: Dict[str, Any] = Field(
        default_factory=dict, description="Additional metadata")

    # Set collection name (though rarely used directly)
    collection_name: ClassVar[str] = "messages"


class Conversation(MongoBaseModel):
    """Conversation model for storing AI chat history."""

    user_id: str = Field(...,
                         description="ID of the user who owns this conversation")
    session_id: str = Field(..., description="Unique session identifier")
    title: Optional[str] = Field(None, description="Title of the conversation")
    messages: List[Dict[str, Any]] = Field(
        default_factory=list, description="List of messages in the conversation")
    is_active: bool = Field(
        default=True, description="Whether the conversation is active")
    last_message_time: Optional[datetime] = Field(
        None, description="Time of the last message")
    domain: Optional[str] = Field(
        None, description="Domain or context of the conversation")

    # Set collection name
    collection_name: ClassVar[str] = "conversations"

    def add_message(self, role: str, content: str, metadata: Optional[Dict[str, Any]] = None) -> None:
        """Add a message to the conversation."""
        message = {
            "role": role,
            "content": content,
            "timestamp": datetime.utcnow(),
            "metadata": metadata or {}
        }
        self.messages.append(message)
        self.last_message_time = message["timestamp"]

    def get_messages(self) -> List[Dict[str, Any]]:
        """Get all messages in the conversation."""
        return self.messages

    def get_message_count(self) -> int:
        """Get the number of messages in the conversation."""
        return len(self.messages)

    def clear_messages(self) -> None:
        """Clear all messages in the conversation."""
        self.messages = []
        self.last_message_time = None
