from typing import Dict, Any, Optional, List, AsyncIterator, cast
from ai_services.llm.llm_service import LLMService
from orchestration.ai_registry import ai_registry
from core.mcp_state import get_mcp_client
from ai_services.llm.mongodb_memory import get_mongodb_memory
from orchestration.prompts import SYSTEM_PROMPT
from data_layer.cache.ai_cache_manager import AICacheManager
import logging
import json
import time
import asyncio
from ai_services.base.mongo_client import get_mongo_client
from datetime import datetime
from ai_services.rag.rag_service import RAGService

logger = logging.getLogger(__name__)


class AIOrchestrator:
    def __init__(self):
        self.llm_service = LLMService()
        self.logger = logging.getLogger(__name__)
        self._current_model_id: Optional[int] = None
        self.ai_registry = ai_registry
        self.max_history_length = 10
        self.mcp_client = None
        self._init_lock = asyncio.Lock()
        self.rag_service = RAGService()  # Initialize RAG service

        # MongoDB client for direct database operations
        self.mongo_client = get_mongo_client()

        # We'll initialize the MCP client lazily when needed
        self.logger.info(
            "AIOrchestrator initialized with lazy MCP client loading")

    async def _get_mcp_client(self):
        """Get MCP client, with lazy initialization."""
        if self.mcp_client is None:
            async with self._init_lock:
                # Check again in case another task initialized it while we were waiting
                if self.mcp_client is None:
                    self.logger.info("Fetching MCP client from global state")
                    self.mcp_client = get_mcp_client()
                    if self.mcp_client is None:
                        self.logger.warning(
                            "MCP client not available in global state")
        return self.mcp_client

    async def _get_available_tools(self) -> List[Dict[str, Any]]:
        """Fetch available tools from MCP client with caching."""
        try:
            # Try to get from cache first
            cached_tools = await AICacheManager.get_cached_tools()
            if cached_tools:
                self.logger.info("Retrieved tools from cache")
                return cached_tools

            mcp_client = await self._get_mcp_client()
            if not mcp_client:
                self.logger.warning(
                    "Could not get MCP client, returning empty tools list")
                return []

            tools = await mcp_client.get_tools()
            self.logger.info(f"Retrieved {len(tools)} tools from MCP client")

            # Remove any auth-related parameters since we use JWT
            for tool in tools:
                if "input_schema" in tool and "properties" in tool["input_schema"]:
                    auth_params = ["user_id", "auth_token",
                                   "token", "authorization"]
                    for param in auth_params:
                        if param in tool["input_schema"]["properties"]:
                            tool["input_schema"]["properties"].pop(param)
                    if "required" in tool["input_schema"]:
                        tool["input_schema"]["required"] = [
                            r for r in tool["input_schema"]["required"] if r not in auth_params]

            # Cache the tools
            await AICacheManager.set_cached_tools(tools)
            self.logger.info("Cached tools in Redis")

            return tools
        except Exception as e:
            self.logger.error(f"Error getting available tools: {str(e)}")
            return []

    def _format_tools_for_prompt(self, tools: List[Dict[str, Any]]) -> str:
        """Format tools into a string for the system prompt."""
        if not tools:
            return "No tools are currently available."

        tool_strings = []
        for tool in tools:
            # Format tool input schema in a more readable way
            schema_str = json.dumps(tool.get('input_schema', {}), indent=2)
            tool_str = f"- {tool['name']}: {tool.get('description', 'No description')}\n  Arguments: {schema_str}"
            tool_strings.append(tool_str)

        return "\n".join(tool_strings)

    def _extract_tool_calls(self, text: str) -> List[Dict[str, Any]]:
        """Extract tool calls from LLM response, supporting both XML tags and plain JSON format."""
        tool_calls = []

        # Method 1: Try XML-wrapped tool calls first (original format)
        start_tag = "<tool_call>"
        end_tag = "</tool_call>"
        temp_text = text

        while start_tag in temp_text and end_tag in temp_text:
            start = temp_text.find(start_tag) + len(start_tag)
            end = temp_text.find(end_tag)
            if start > -1 and end > -1:
                tool_call_text = temp_text[start:end].strip()
                try:
                    tool_call = json.loads(tool_call_text)
                    # Validate required fields
                    if "name" in tool_call:
                        tool_calls.append(tool_call)
                        self.logger.info(
                            f"Extracted XML-wrapped tool call: {tool_call['name']}")
                    else:
                        self.logger.warning(
                            f"Tool call missing 'name' field: {tool_call_text}")
                except json.JSONDecodeError:
                    self.logger.error(
                        f"Failed to parse XML-wrapped tool call: {tool_call_text}")
                temp_text = temp_text[end + len(end_tag):]
            else:
                break

        # Method 2: If no XML-wrapped tool calls found, try to parse the entire response as JSON
        if not tool_calls:
            self.logger.info(
                "No XML-wrapped tool calls found, trying to parse as plain JSON")
            try:
                # Try to parse the entire response as a single JSON tool call
                tool_call = json.loads(text.strip())
                if isinstance(tool_call, dict) and "name" in tool_call:
                    # Ensure arguments exist
                    if "arguments" not in tool_call:
                        tool_call["arguments"] = {}
                    tool_calls.append(tool_call)
                    self.logger.info(
                        f"Extracted plain JSON tool call: {tool_call['name']}")
                else:
                    self.logger.warning(
                        f"Plain JSON response is not a valid tool call: {text[:200]}...")
            except json.JSONDecodeError:
                # Method 3: Try to find JSON objects within the text using regex
                import re
                self.logger.info(
                    "Failed to parse as plain JSON, trying regex extraction")

                # Pattern to match JSON objects that look like tool calls
                json_pattern = r'\{[^{}]*"name"\s*:\s*"[^"]*"[^{}]*\}'
                matches = re.findall(json_pattern, text)

                for match in matches:
                    try:
                        tool_call = json.loads(match)
                        if isinstance(tool_call, dict) and "name" in tool_call:
                            # Ensure arguments exist
                            if "arguments" not in tool_call:
                                tool_call["arguments"] = {}
                            tool_calls.append(tool_call)
                            self.logger.info(
                                f"Extracted regex-found tool call: {tool_call['name']}")
                    except json.JSONDecodeError:
                        self.logger.warning(
                            f"Failed to parse regex match as JSON: {match}")

                # Method 4: If still no tool calls, check if response contains tool names
                if not tool_calls:
                    self.logger.info(
                        "No JSON tool calls found, checking for tool name patterns")
                    # List of known tool names to look for
                    known_tools = [
                        "get_items", "todos.create", "habits.create", "calendar.getEvents",
                        "calendar.createEvent", "todos.smartUpdate", "notes.get", "notes.create",
                        "notes.rewriteInStyle", "todos.addChecklist", "get_tasks", "create_task"
                    ]

                    for tool_name in known_tools:
                        if tool_name in text.lower():
                            self.logger.info(
                                f"Found tool name '{tool_name}' in response, attempting to construct tool call")
                            # Try to extract arguments if they exist
                            try:
                                # Look for common argument patterns
                                if "habits" in text.lower():
                                    tool_calls.append({
                                        "name": "get_items",
                                        "arguments": {"item_type": "habits"}
                                    })
                                elif "todos" in text.lower():
                                    tool_calls.append({
                                        "name": "get_items",
                                        "arguments": {"item_type": "todos"}
                                    })
                                break
                            except Exception as e:
                                self.logger.error(
                                    f"Error constructing tool call for {tool_name}: {str(e)}")

        self.logger.info(f"Total tool calls extracted: {len(tool_calls)}")
        return tool_calls

    def _make_serializable(self, obj):
        """Convert non-serializable objects to serializable structures recursively."""
        # Base case: object is already a basic type
        if isinstance(obj, (str, int, float, bool, type(None))):
            return obj

        # Handle lists recursively
        if isinstance(obj, list):
            return [self._make_serializable(item) for item in obj]

        # Handle dictionaries recursively
        if isinstance(obj, dict):
            return {k: self._make_serializable(v) for k, v in obj.items()}

        # Handle TextContent or other special objects
        self.logger.warning(
            f"Converting non-serializable content type {type(obj)} to serializable form")

        try:
            # Try to convert to a dictionary if the object has specific attributes
            if hasattr(obj, '__dict__'):
                return {k: self._make_serializable(v) for k, v in obj.__dict__.items() if not k.startswith('_')}

            # Handle TextContent specifically
            if hasattr(obj, 'text'):
                text = obj.text
                # Try to parse JSON
                if isinstance(text, str):
                    try:
                        return json.loads(text)
                    except json.JSONDecodeError:
                        return text
                return text

            # Handle other attribute combinations
            if hasattr(obj, 'data'):
                return self._make_serializable(obj.data)

            if hasattr(obj, 'content'):
                return self._make_serializable(obj.content)

            # Final fallback: convert to string
            return str(obj)

        except Exception as e:
            self.logger.error(
                f"Error converting object to serializable form: {str(e)}")
            return str(obj)

    async def process_request_stream(
        self,
        user_input: str,
        user_id: int,
        domain: Optional[str] = None,
        auth_token: Optional[str] = None,
        client_ip: Optional[str] = None,
        user_agent: Optional[str] = None,
        real_user_id: Optional[str] = None,
        organization_id: Optional[str] = None
    ) -> AsyncIterator[Dict[str, Any]]:
        """Process an AI request with MCP integration and stream the response tokens."""
        try:
            start_time = time.time()
            self.logger.info(
                f"Processing streaming request for user {user_id} in domain {domain or 'default'}")
            self.logger.info(f"Auth token provided: {auth_token is not None}")

            # Ensure user_id is a string for MongoDB
            user_id_str = str(user_id)

            # Create or get conversation in MongoDB
            session_id = f"session_{user_id}"
            conversation = self.mongo_client.get_conversation_by_session(
                session_id)

            if not conversation:
                # Create new conversation
                conversation = self.mongo_client.create_conversation(
                    user_id=user_id_str,
                    session_id=session_id,
                    title=f"Conversation {datetime.utcnow().strftime('%Y-%m-%d %H:%M')}",
                    domain=domain or "default"
                )
                self.logger.info(
                    f"Created new conversation with ID {conversation.id} for streaming")
            else:
                self.logger.info(
                    f"Using existing conversation with ID {conversation.id} for streaming")

            # Store user message in MongoDB
            if conversation and conversation.id:
                self.mongo_client.add_message_to_conversation(
                    conversation_id=conversation.id,
                    role="user",
                    content=user_input,
                    metadata={
                        "domain": domain,
                        "streaming": True,
                        "client_ip": client_ip,
                        "user_agent": user_agent,
                        "real_user_id": real_user_id,
                        "organization_id": organization_id
                    }
                )
                self.logger.info(
                    f"Stored user message for streaming in conversation {conversation.id}")

            # Get conversation history using MongoDB memory - just load, don't write
            mongo_memory = get_mongodb_memory(
                user_id=user_id_str, session_id=session_id)
            messages = mongo_memory.get_langchain_messages()
            self.logger.debug(
                f"Retrieved {len(messages)} conversation history messages")

            # Get tools and format system prompt with caching
            tools = await self._get_available_tools()
            formatted_tools = self._format_tools_for_prompt(tools)

            # Try to get cached system prompt
            enhanced_system_prompt = await AICacheManager.get_cached_system_prompt()
            if not enhanced_system_prompt:
                enhanced_system_prompt = SYSTEM_PROMPT.format(
                    tools=formatted_tools)
                # Cache the formatted system prompt
                await AICacheManager.set_cached_system_prompt(enhanced_system_prompt)
                self.logger.info("Cached formatted system prompt")
            else:
                self.logger.info("Retrieved system prompt from cache")

            # Get relevant context from RAG
            relevant_context = await self.rag_service.get_relevant_context(user_input)
            self.logger.info("Retrieved relevant context from RAG service")

            # Add RAG context to system prompt
            if relevant_context:
                enhanced_system_prompt += f"\n\nRelevant context from knowledge base:\n{relevant_context}"

            # First, get complete response to check for tool calls
            complete_response = await self.llm_service.generate_response(
                prompt=user_input,
                context={
                    "system_prompt": enhanced_system_prompt,
                    "conversation_history": messages
                },
                stream=False
            )

            # Extract text from response
            if not isinstance(complete_response, dict):
                self.logger.error("Unexpected response type from LLM")
                yield {"token": "Error: Unexpected response from AI service", "error": True}
                return

            response_text = complete_response.get("text", "")

            # Check for tool calls
            tool_calls = self._extract_tool_calls(response_text)
            self.logger.info(
                f"Extracted {len(tool_calls)} tool calls from response")

            final_response = ""
            tool_info = None

            # Process tool calls if present, similar to original process_request
            if tool_calls:
                self.logger.info("Processing tool calls before streaming")
                tool_results = []
                last_tool_call = None

                for idx, tool_call in enumerate(tool_calls):
                    try:
                        self.logger.info(
                            f"Processing tool call {idx+1}/{len(tool_calls)}: {tool_call['name']}")

                        # Add authorization if provided
                        if auth_token:
                            self.logger.info(
                                f"Adding auth token to tool call: {tool_call['name']}")
                            if "arguments" not in tool_call:
                                tool_call["arguments"] = {}
                            tool_call["arguments"]["authorization"] = auth_token

                        # Get MCP client safely
                        mcp_client = await self._get_mcp_client()
                        if mcp_client is None:
                            raise ValueError("MCP client is not available")

                        # Execute the tool call with retry logic built into the client
                        result = await mcp_client.call_tool(
                            tool_call["name"],
                            tool_call["arguments"]
                        )

                        # Process result
                        if result.get("status") == "success":
                            self.logger.info(
                                f"Tool call {tool_call['name']} succeeded")
                            content = result.get("content", {})
                            content = self._make_serializable(content)
                            tool_results.append({
                                "tool": tool_call["name"],
                                "result": content
                            })
                            last_tool_call = {
                                "name": tool_call["name"],
                                "arguments": tool_call["arguments"],
                                "success": True
                            }
                        else:
                            self.logger.warning(
                                f"Tool call {tool_call['name']} failed: {result.get('error', 'Unknown error')}")
                            tool_results.append({
                                "tool": tool_call["name"],
                                "error": result.get("error", "Unknown error")
                            })
                            last_tool_call = {
                                "name": tool_call["name"],
                                "arguments": tool_call["arguments"],
                                "success": False
                            }
                    except Exception as e:
                        self.logger.error(
                            f"Tool call execution failed: {str(e)}")
                        tool_results.append({
                            "tool": tool_call["name"],
                            "error": str(e)
                        })
                        last_tool_call = {
                            "name": tool_call["name"],
                            "arguments": tool_call["arguments"],
                            "success": False
                        }

                # Generate streaming response with tool results
                if tool_results:
                    self.logger.info(
                        "Generating streaming response with tool results")
                    prompt_for_llm = f"Based on the user query: {user_input}\nHere are the tool results: {json.dumps(tool_results, indent=2)}\nPlease provide a helpful response."

                    # Now use streaming for final response with tool results
                    stream_generator = await self.llm_service.generate_response(
                        prompt=prompt_for_llm,
                        context={
                            "system_prompt": "Format the tool results in a natural, helpful way for the user."
                        },
                        stream=True,
                        user_id=real_user_id,
                        session_id=session_id,
                        client_ip=client_ip,
                        user_agent=user_agent,
                        organization_id=organization_id
                    )

                    tool_info = last_tool_call

                    async for token in cast(AsyncIterator[str], stream_generator):
                        final_response += token
                        yield {"token": token, "tool_used": tool_info["name"] if tool_info else None}
            else:
                self.logger.info(
                    "No tool calls detected, streaming direct response")

                # Start streaming response
                stream_generator = await self.llm_service.generate_response(
                    prompt=user_input,
                    context={
                        "system_prompt": enhanced_system_prompt,
                        "conversation_history": messages
                    },
                    stream=True,
                    user_id=real_user_id,
                    session_id=session_id,
                    client_ip=client_ip,
                    user_agent=user_agent,
                    organization_id=organization_id
                )

                # Stream tokens
                async for token in cast(AsyncIterator[str], stream_generator):
                    final_response += token
                    yield {"token": token}

            # Store only assistant message in MongoDB (user message already stored above)
            if final_response and conversation and conversation.id:
                self.mongo_client.add_message_to_conversation(
                    conversation_id=conversation.id,
                    role="assistant",
                    content=final_response,
                    metadata={
                        "tool_used": tool_info["name"] if tool_info else None,
                        "streaming": True,
                        "client_ip": client_ip,
                        "user_agent": user_agent,
                        "real_user_id": real_user_id,
                        "organization_id": organization_id
                    }
                )
                self.logger.info(
                    f"Stored streaming assistant response in conversation {conversation.id}")

            execution_time = time.time() - start_time
            self.logger.info(
                f"Streaming request processed in {execution_time:.2f} seconds")

            # Log model usage in MongoDB
            try:
                if hasattr(self.llm_service, "model_name"):
                    self.mongo_client.model_usage_repo.log_usage(
                        model_id="1",  # Default model ID
                        model_name=self.llm_service.model_name,
                        request_type="streaming_request",
                        tokens_in=len(user_input),  # Approximation
                        # Approximation
                        tokens_out=len(
                            final_response) if final_response else 0,
                        latency_ms=int(execution_time * 1000),
                        success=True,
                        user_id=user_id_str,
                        session_id=session_id
                    )
                    self.logger.info(
                        f"Logged model usage for streaming request in MongoDB")
            except Exception as e:
                self.logger.error(
                    f"Failed to log model usage for streaming: {str(e)}")

            # Send completion message with tool info if applicable
            yield {
                "token": "",
                "complete": True,
                "tool_used": tool_info["name"] if tool_info else None,
                "tool_success": tool_info["success"] if tool_info else None
            }

        except Exception as e:
            self.logger.error(
                f"Error in process_request_stream: {str(e)}", exc_info=True)
            yield {"token": f"I'm sorry, but I encountered an error: {str(e)}", "error": True}

    async def log_model_usage(self, **kwargs):
        """Log model usage to MongoDB."""
        try:
            if hasattr(self.llm_service, "model_name"):
                self.mongo_client.model_usage_repo.log_usage(**kwargs)
                self.logger.info("Logged model usage in MongoDB")
        except Exception as e:
            self.logger.error(f"Failed to log model usage: {str(e)}")
