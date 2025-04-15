from typing import Dict, Any, List, Optional, AsyncGenerator, Union, cast
import logging
import json
import asyncio
from pydantic import BaseModel

from ai_services.llm.llm_service import LLMService

logger = logging.getLogger(__name__)

class GitHubModelAdapter:
    """Adapter to make GitHub-hosted models compatible with Atomic Agents framework."""
    
    def __init__(self, llm_service: LLMService):
        self.llm_service = llm_service
        
    async def chat_completions_create(self, 
                                     messages: List[Dict[str, Any]], 
                                     model: Optional[str] = None, 
                                     stream: bool = False,
                                     **kwargs) -> Union[Dict[str, Any], AsyncGenerator[Dict[str, Any], None]]:
        """
        Adapts the GitHub-hosted model API to match the interface expected by Atomic Agents.
        
        Args:
            messages: List of messages in OpenAI format
            model: Model name (ignored, using GitHub model)
            stream: Whether to stream the response
            **kwargs: Additional arguments
            
        Returns:
            Response dict or async generator for streaming
        """
        logger.info(f"GitHub Model Adapter: Processing request with {len(messages)} messages")
        
        # Extract the last user message
        last_user_message = None
        system_content = ""
        
        for msg in messages:
            if msg["role"] == "user":
                last_user_message = msg["content"]
            elif msg["role"] == "system":
                system_content = msg["content"]
        
        if not last_user_message:
            last_user_message = "Hello"
            logger.warning("No user message found in the messages list, using default")
        
        # Create context with system message if available
        context = {}
        if system_content:
            context["system_prompt"] = system_content
        
        if stream:
            return self._stream_response(last_user_message, context, **kwargs)
        else:
            return await self._complete_response(last_user_message, context, **kwargs)
    
    async def _complete_response(self, prompt: str, context: Dict[str, Any], **kwargs) -> Dict[str, Any]:
        """Generate a complete response from the LLM service."""
        try:
            # Call the existing LLM service
            response = await self.llm_service.generate_response(
                prompt=prompt,
                context=context,
                stream=False,
                user_id=kwargs.get("user", "default_user")
            )
            
            # Convert to OpenAI-compatible format
            if isinstance(response, dict) and "text" in response:
                return {
                    "choices": [
                        {
                            "message": {
                                "role": "assistant",
                                "content": response["text"]
                            },
                            "index": 0,
                            "finish_reason": "stop"
                        }
                    ],
                    "id": "github-model-response",
                    "model": "github-hosted-model",
                    "created": response.get("timestamp", 0),
                    "object": "chat.completion"
                }
            else:
                logger.error(f"Unexpected response format: {response}")
                # Return a fallback response
                return {
                    "choices": [
                        {
                            "message": {
                                "role": "assistant",
                                "content": str(response) if response else "Error: Empty response"
                            },
                            "index": 0,
                            "finish_reason": "stop"
                        }
                    ],
                    "id": "github-model-response",
                    "model": "github-hosted-model",
                    "created": 0,
                    "object": "chat.completion"
                }
        except Exception as e:
            logger.error(f"Error in GitHub model adapter: {str(e)}", exc_info=True)
            # Return an error response
            return {
                "choices": [
                    {
                        "message": {
                            "role": "assistant",
                            "content": f"Error: {str(e)}"
                        },
                        "index": 0,
                        "finish_reason": "error"
                    }
                ],
                "id": "github-model-error",
                "model": "github-hosted-model",
                "created": 0,
                "object": "chat.completion"
            }
    
    async def _stream_response(self, prompt: str, context: Dict[str, Any], **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        """Stream the response from the LLM service."""
        try:
            # Use the existing LLM service with streaming
            response_stream = await self.llm_service.generate_response(
                prompt=prompt,
                context=context,
                stream=True,
                user_id=kwargs.get("user", "default_user")
            )
            
            # If the response is a generator, yield properly formatted chunks
            if isinstance(response_stream, AsyncGenerator):
                async for chunk in response_stream:
                    if isinstance(chunk, dict) and "text" in chunk:
                        chunk_text = chunk["text"]
                        yield {
                            "choices": [
                                {
                                    "delta": {
                                        "role": "assistant",
                                        "content": chunk_text
                                    },
                                    "index": 0,
                                    "finish_reason": None
                                }
                            ],
                            "id": "github-model-stream",
                            "model": "github-hosted-model",
                            "created": chunk.get("timestamp", 0),
                            "object": "chat.completion.chunk"
                        }
                
                # Send a final chunk indicating completion
                yield {
                    "choices": [
                        {
                            "delta": {},
                            "index": 0,
                            "finish_reason": "stop"
                        }
                    ],
                    "id": "github-model-stream",
                    "model": "github-hosted-model",
                    "created": 0,
                    "object": "chat.completion.chunk"
                }
            else:
                # Handle case where a full response is returned instead of a stream
                if isinstance(response_stream, dict) and "text" in response_stream:
                    yield {
                        "choices": [
                            {
                                "delta": {
                                    "role": "assistant",
                                    "content": response_stream["text"]
                                },
                                "index": 0,
                                "finish_reason": None
                            }
                        ],
                        "id": "github-model-stream",
                        "model": "github-hosted-model",
                        "created": response_stream.get("timestamp", 0),
                        "object": "chat.completion.chunk"
                    }
                    
                    # Send a final chunk
                    yield {
                        "choices": [
                            {
                                "delta": {},
                                "index": 0,
                                "finish_reason": "stop"
                            }
                        ],
                        "id": "github-model-stream",
                        "model": "github-hosted-model",
                        "created": 0,
                        "object": "chat.completion.chunk"
                    }
                else:
                    # Fallback for unexpected response type
                    yield {
                        "choices": [
                            {
                                "delta": {
                                    "role": "assistant",
                                    "content": str(response_stream)
                                },
                                "index": 0,
                                "finish_reason": "stop"
                            }
                        ],
                        "id": "github-model-stream",
                        "model": "github-hosted-model",
                        "created": 0,
                        "object": "chat.completion.chunk"
                    }
        except Exception as e:
            logger.error(f"Error in GitHub model streaming: {str(e)}", exc_info=True)
            # Yield an error response
            yield {
                "choices": [
                    {
                        "delta": {
                            "role": "assistant",
                            "content": f"Error: {str(e)}"
                        },
                        "index": 0,
                        "finish_reason": "error"
                    }
                ],
                "id": "github-model-error",
                "model": "github-hosted-model",
                "created": 0,
                "object": "chat.completion.chunk"
            } 