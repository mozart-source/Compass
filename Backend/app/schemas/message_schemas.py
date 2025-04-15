from typing import List, Optional
from pydantic import BaseModel


class Message(BaseModel):
    content: str
    role: str = "user"


class UserMessage(Message):
    role: str = "user"


class AssistantMessage(Message):
    role: str = "assistant"


class ConversationHistory:
    def __init__(self):
        self.messages: List[Message] = []
        self.max_messages = 10

    def add_message(self, message: Message) -> None:
        self.messages.append(message)
        if len(self.messages) > self.max_messages * 2:
            self.messages = self.messages[-self.max_messages * 2:]

    def get_messages(self) -> List[Message]:
        return self.messages
