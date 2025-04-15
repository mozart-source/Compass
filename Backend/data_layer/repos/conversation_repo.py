from typing import List, Optional, Dict, Any
from data_layer.repos.base_repo import BaseMongoRepository
from data_layer.models.conversation import Conversation
import logging
from datetime import datetime

logger = logging.getLogger(__name__)


class ConversationRepository(BaseMongoRepository[Conversation]):
    """Repository for managing conversation history in MongoDB."""

    def __init__(self):
        """Initialize the repository with the Conversation model."""
        super().__init__(Conversation)

    def find_by_session(self, session_id: str) -> Optional[Conversation]:
        """Find conversation by session ID."""
        return self.find_one({"session_id": session_id})

    def find_by_user(
        self,
        user_id: str,
        skip: int = 0,
        limit: int = 50,
        active_only: bool = False
    ) -> List[Conversation]:
        """Find conversations by user ID with pagination."""
        filter_dict: Dict[str, Any] = {"user_id": user_id}

        if active_only:
            filter_dict["is_active"] = True

        return self.find_many(
            filter=filter_dict,
            skip=skip,
            limit=limit,
            # Sort by last message time, newest first
            sort=[("last_message_time", -1)]
        )

    def create_conversation(
        self,
        user_id: str,
        session_id: str,
        title: Optional[str] = None,
        domain: Optional[str] = None
    ) -> Conversation:
        """Create a new conversation."""
        conversation = Conversation(
            user_id=user_id,
            session_id=session_id,
            title=title,
            domain=domain,
            last_message_time=datetime.utcnow()
        )

        conversation_id = self.insert(conversation)
        logger.info(f"Created conversation with ID {conversation_id}")

        # Return the created conversation
        return conversation

    def add_message_to_conversation(
        self,
        conversation_id: str,
        role: str,
        content: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Optional[Conversation]:
        """Add a message to a conversation by its ID."""
        conversation = self.find_by_id(conversation_id)
        if not conversation:
            logger.warning(f"Conversation with ID {conversation_id} not found")
            return None

        # Add message
        conversation.add_message(role, content, metadata)

        # Update in database
        updated = self.update(
            conversation_id,
            {
                "messages": conversation.messages,
                "last_message_time": conversation.last_message_time
            }
        )

        return updated

    def archive_conversation(self, conversation_id: str) -> bool:
        """Archive a conversation (mark as inactive)."""
        result = self.update(conversation_id, {"is_active": False})
        return result is not None

    def delete_user_conversations(self, user_id: str) -> int:
        """Delete all conversations for a user."""
        return self.delete_many({"user_id": user_id})

    # Async methods

    async def async_find_by_session(self, session_id: str) -> Optional[Conversation]:
        """Find conversation by session ID (async)."""
        return await self.async_find_one({"session_id": session_id})

    async def async_find_by_user(
        self,
        user_id: str,
        skip: int = 0,
        limit: int = 50,
        active_only: bool = False
    ) -> List[Conversation]:
        """Find conversations by user ID with pagination (async)."""
        filter_dict: Dict[str, Any] = {"user_id": user_id}

        if active_only:
            filter_dict["is_active"] = True

        return await self.async_find_many(
            filter=filter_dict,
            skip=skip,
            limit=limit,
            sort=[("last_message_time", -1)]
        )

    async def async_create_conversation(
        self,
        user_id: str,
        session_id: str,
        title: Optional[str] = None,
        domain: Optional[str] = None
    ) -> Conversation:
        """Create a new conversation (async)."""
        conversation = Conversation(
            user_id=user_id,
            session_id=session_id,
            title=title,
            domain=domain,
            last_message_time=datetime.utcnow()
        )

        conversation_id = await self.async_insert(conversation)
        logger.info(f"Created conversation with ID {conversation_id}")

        # Return the created conversation
        return conversation

    async def async_add_message_to_conversation(
        self,
        conversation_id: str,
        role: str,
        content: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Optional[Conversation]:
        """Add a message to a conversation by its ID (async)."""
        conversation = await self.async_find_by_id(conversation_id)
        if not conversation:
            logger.warning(f"Conversation with ID {conversation_id} not found")
            return None

        # Add message
        conversation.add_message(role, content, metadata)

        # Update in database
        updated = await self.async_update(
            conversation_id,
            {
                "messages": conversation.messages,
                "last_message_time": conversation.last_message_time
            }
        )

        return updated
