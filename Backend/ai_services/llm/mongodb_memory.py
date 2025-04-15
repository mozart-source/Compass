from typing import Dict, List, Any, Optional, Tuple, cast
import logging
import uuid
from datetime import datetime
from langchain_community.chat_message_histories import ChatMessageHistory
from langchain.schema import AIMessage, HumanMessage, SystemMessage, BaseMessage
from ai_services.base.mongo_client import get_mongo_client
from data_layer.models.conversation import Conversation
from app.schemas.message_schemas import ConversationHistory, UserMessage, AssistantMessage, Message

logger = logging.getLogger(__name__)


# Create our own base class to avoid class conflicts
class MongoDBMessageHistory:
    """MongoDB-backed chat message history implementation for LangChain."""

    def __init__(
        self,
        user_id: str,
        session_id: Optional[str] = None,
        conversation_id: Optional[str] = None,
        domain: Optional[str] = None,
        mongo_client=None
    ):
        """Initialize MongoDB chat message history."""
        self.user_id = user_id
        self.session_id = session_id or str(uuid.uuid4())
        self.conversation_id = conversation_id
        self.domain = domain
        # Use provided client or get the singleton
        self.mongo_client = mongo_client or get_mongo_client()
        self.messages: List[BaseMessage] = []

        # Initialize conversation if not yet loaded
        self._init_conversation()

        # Load initial messages
        self._load_messages()

    def _init_conversation(self) -> None:
        """Initialize or retrieve the conversation."""
        try:
            if self.conversation_id:
                # Try to load existing conversation
                conversation = self.mongo_client.conversation_repo.find_by_id(
                    self.conversation_id)
                if not conversation:
                    # Conversation not found, create new one
                    logger.warning(
                        f"Conversation ID {self.conversation_id} not found, creating new conversation")
                    self._create_new_conversation()
            else:
                # Try to find conversation by session ID
                conversation = self.mongo_client.get_conversation_by_session(
                    self.session_id)
                if not conversation:
                    # Conversation not found, create new one
                    self._create_new_conversation()
                else:
                    # Found existing conversation
                    self.conversation_id = conversation.id
                    logger.info(
                        f"Found existing conversation with ID {self.conversation_id}")
        except Exception as e:
            logger.error(f"Error initializing conversation: {str(e)}")
            # Create a new conversation as fallback
            self._create_new_conversation()

    def _create_new_conversation(self) -> None:
        """Create a new conversation in MongoDB."""
        try:
            # Generate default title based on timestamp
            default_title = f"Conversation {datetime.utcnow().strftime('%Y-%m-%d %H:%M')}"

            conversation = self.mongo_client.create_conversation(
                user_id=self.user_id,
                session_id=self.session_id,
                title=default_title,
                domain=self.domain
            )
            self.conversation_id = conversation.id
            logger.info(
                f"Created new conversation with ID {self.conversation_id}")
        except Exception as e:
            logger.error(f"Failed to create new conversation: {str(e)}")
            # Set a placeholder ID - will try to create again on next operation
            self.conversation_id = None

    def _message_to_dict(self, message: BaseMessage) -> Dict[str, Any]:
        """Convert LangChain message to MongoDB dictionary."""
        if isinstance(message, HumanMessage):
            role = "user"
        elif isinstance(message, AIMessage):
            role = "assistant"
        elif isinstance(message, SystemMessage):
            role = "system"
        else:
            role = "unknown"

        return {
            "role": role,
            "content": message.content,
            "timestamp": datetime.utcnow(),
            "metadata": getattr(message, "additional_kwargs", {})
        }

    def _dict_to_message(self, message_dict: Dict[str, Any]) -> BaseMessage:
        """Convert MongoDB dictionary to LangChain message."""
        role = message_dict.get("role", "")
        content = message_dict.get("content", "")
        metadata = message_dict.get("metadata", {})

        if role == "user":
            return HumanMessage(content=content, additional_kwargs=metadata)
        elif role == "assistant":
            return AIMessage(content=content, additional_kwargs=metadata)
        elif role == "system":
            return SystemMessage(content=content, additional_kwargs=metadata)
        else:
            # Default to human message if unknown
            return HumanMessage(content=content, additional_kwargs=metadata)

    def _load_messages(self) -> None:
        """Load messages from MongoDB into the messages attribute."""
        try:
            if not self.conversation_id:
                self.messages = []
                return

            conversation = self.mongo_client.conversation_repo.find_by_id(
                self.conversation_id)
            if not conversation:
                logger.warning(
                    f"Conversation ID {self.conversation_id} not found when retrieving messages")
                self.messages = []
                return

            self.messages = [self._dict_to_message(
                msg) for msg in conversation.messages]
            logger.info(
                f"Loaded {len(self.messages)} messages from conversation {self.conversation_id}")
        except Exception as e:
            logger.error(f"Error loading messages: {str(e)}")
            self.messages = []

    def add_message(self, message: BaseMessage) -> None:
        """Add a message to the conversation."""
        try:
            # If no conversation ID, initialize or retry creation
            if not self.conversation_id:
                self._init_conversation()

            # Check again after initialization
            if not self.conversation_id:
                logger.error("Failed to initialize conversation ID")
                # Still add to local messages for in-memory usage
                self.messages.append(message)
                return

            message_dict = self._message_to_dict(message)
            updated = self.mongo_client.add_message_to_conversation(
                conversation_id=self.conversation_id,
                role=message_dict["role"],
                content=message_dict["content"],
                metadata=message_dict["metadata"]
            )

            if not updated:
                logger.error(
                    f"Failed to add message to conversation {self.conversation_id}")
            else:
                logger.info(
                    f"Added message (role={message_dict['role']}) to conversation {self.conversation_id}")
                # Update the local messages list
                self.messages.append(message)
        except Exception as e:
            logger.error(f"Error adding message to conversation: {str(e)}")
            # Still add to local messages for in-memory usage
            self.messages.append(message)

    def clear(self) -> None:
        """Clear all messages from the conversation."""
        try:
            if not self.conversation_id:
                self.messages = []
                return

            conversation = self.mongo_client.conversation_repo.find_by_id(
                self.conversation_id)
            if not conversation:
                logger.warning(
                    f"Conversation ID {self.conversation_id} not found when clearing messages")
                self.messages = []
                return

            # Update the conversation with empty messages list
            self.mongo_client.conversation_repo.update(
                self.conversation_id,
                {"messages": []}
            )

            # Also clear the local messages
            self.messages = []

            logger.info(
                f"Cleared all messages from conversation {self.conversation_id}")
        except Exception as e:
            logger.error(f"Error clearing messages: {str(e)}")
            # Clear local messages anyway
            self.messages = []

    def add_user_message(self, content: str) -> None:
        """Add a user message to the conversation."""
        self.add_message(HumanMessage(content=content))

    def add_ai_message(self, content: str) -> None:
        """Add an AI message to the conversation."""
        self.add_message(AIMessage(content=content))

    def to_conversation_history(self) -> ConversationHistory:
        """Convert MongoDB messages to ConversationHistory format."""
        history = ConversationHistory()
        for message in self.messages:
            content = message.content
            if isinstance(content, list):
                content = str(content)
            if isinstance(message, HumanMessage):
                history.add_message(UserMessage(content=content))
            elif isinstance(message, AIMessage):
                history.add_message(AssistantMessage(content=content))
        return history

    def get_langchain_messages(self) -> List[Dict[str, str]]:
        """Get messages in the format expected by OpenAI API."""
        messages = []
        for msg in self.messages:
            content = msg.content
            if isinstance(content, list):
                content = str(content)
            if isinstance(msg, HumanMessage):
                messages.append({"role": "user", "content": content})
            elif isinstance(msg, AIMessage):
                messages.append({"role": "assistant", "content": content})
            elif isinstance(msg, SystemMessage):
                messages.append({"role": "system", "content": content})
        return messages


# Use a regular ChatMessageHistory as a wrapper for our custom implementation
def get_mongodb_memory(
    user_id: str,
    session_id: Optional[str] = None,
    conversation_id: Optional[str] = None,
    domain: Optional[str] = None
) -> MongoDBMessageHistory:
    """Get MongoDB-backed chat message history."""
    # Get the shared MongoDB client
    mongo_client = get_mongo_client()

    # Create our MongoDB-backed implementation with the shared client
    mongo_history = MongoDBMessageHistory(
        user_id=user_id,
        session_id=session_id,
        conversation_id=conversation_id,
        domain=domain,
        mongo_client=mongo_client
    )

    return mongo_history
