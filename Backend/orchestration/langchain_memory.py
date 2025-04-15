"""
LangChain Memory Manager for handling conversation history.

This module provides a unified way to manage conversation histories
using LangChain's memory components.
"""

from typing import Dict, Any, List, Optional
import logging
from langchain_core.messages import HumanMessage, AIMessage, BaseMessage, SystemMessage
from langchain_core.chat_history import BaseChatMessageHistory
from langchain.memory import ConversationBufferMemory, ConversationBufferWindowMemory
from langchain_core.runnables import RunnableConfig
from app.schemas.message_schemas import ConversationHistory, UserMessage, AssistantMessage, Message

logger = logging.getLogger(__name__)


class ConversationMemoryManager:
    """Manages conversation memories for different users using LangChain."""

    def __init__(self, max_history_length: int = 10):
        self.memories: Dict[int, ConversationBufferWindowMemory] = {}
        self.max_history_length = max_history_length
        logger.info(
            f"Initialized ConversationMemoryManager with max_history_length={max_history_length}")

    def get_memory(self, user_id: int) -> ConversationBufferWindowMemory:
        """Get or create memory for a specific user."""
        if user_id not in self.memories:
            logger.debug(f"Creating new memory for user_id={user_id}")
            self.memories[user_id] = ConversationBufferWindowMemory(
                k=self.max_history_length,
                return_messages=True,
                output_key="response"
            )
        return self.memories[user_id]

    def add_user_message(self, user_id: int, content: str) -> None:
        """Add a user message to the conversation history."""
        memory = self.get_memory(user_id)
        content = content[:50] if len(content) > 50 else content
        memory.chat_memory.add_user_message(content)
        logger.debug(
            f"Added user message for user_id={user_id}: {content[:50]}...")

    def add_ai_message(self, user_id: int, content: str) -> None:
        """Add an AI message to the conversation history."""
        memory = self.get_memory(user_id)
        content = content[:50] if len(content) > 50 else content
        memory.chat_memory.add_ai_message(content)
        logger.debug(
            f"Added AI message for user_id={user_id}: {content[:50]}...")

    def get_messages(self, user_id: int) -> List[BaseMessage]:
        """Get all messages for a user."""
        memory = self.get_memory(user_id)
        return memory.chat_memory.messages

    def clear_memory(self, user_id: int) -> None:
        """Clear the conversation history for a user."""
        if user_id in self.memories:
            self.memories[user_id].clear()
            logger.debug(f"Cleared memory for user_id={user_id}")

    def convert_to_chat_history(self, user_id: int) -> ConversationHistory:
        """Convert LangChain memory to app's ConversationHistory format."""
        history = ConversationHistory()
        for message in self.get_messages(user_id):
            content = message.content
            if isinstance(content, list):
                content = str(content)
            if isinstance(message, HumanMessage):
                history.add_message(UserMessage(content=content))
            elif isinstance(message, AIMessage):
                history.add_message(AssistantMessage(content=content))
        return history

    def import_from_chat_history(self, user_id: int, history: ConversationHistory) -> None:
        """Import from app's ConversationHistory to LangChain memory."""
        memory = self.get_memory(user_id)
        memory.clear()

        for msg in history.get_messages():
            if msg.role == "user":
                memory.chat_memory.add_user_message(msg.content)
            elif msg.role == "assistant":
                memory.chat_memory.add_ai_message(msg.content)

        logger.debug(
            f"Imported {len(history.get_messages())} messages for user_id={user_id}")

    def get_chain_config(self, user_id: int) -> RunnableConfig:
        """Get RunnableConfig to be used in LangChain chains."""
        return {"configurable": {"session_id": str(user_id)}}

    def get_langchain_messages(self, user_id: int) -> List[Dict[str, str]]:
        """Get messages in the format expected by OpenAI API."""
        messages = []
        for msg in self.get_messages(user_id):
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
