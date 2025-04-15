from typing import Dict, Any, List, Optional
import logging
import json

from ai_services.agents.base_agent import BaseAgent, BaseIOSchema
from pydantic import Field, BaseModel

from atomic_agents.lib.components.system_prompt_generator import SystemPromptGenerator


class TodoAgentInputSchema(BaseIOSchema):
    """Input schema for TodoAgent."""
    todo_id: str = Field(..., description="ID of the todo")
    user_id: str = Field(..., description="ID of the user")


class TodoAgentOutputSchema(BaseIOSchema):
    """Output schema for TodoAgent."""
    options: List[Dict[str, Any]
                  ] = Field(..., description="List of AI options for the todo")


class TodoAgent(BaseAgent):
    """
    Agent for handling todo-related AI operations.
    Follows Atomic Agents pattern for entity-specific agents.
    """

    def __init__(self):
        super().__init__()

        # Create specialized system prompt for todo agent
        self.system_prompt_generator = SystemPromptGenerator(
            background=[
                "You are IRIS, an AI assistant for the COMPASS productivity app.",
                "You specialize in helping users manage their todos effectively."
            ],
            steps=[
                "Analyze the todo to understand its context, priority, and deadline.",
                "Identify ways you can help the user with this todo."
            ],
            output_instructions=[
                "Provide practical, actionable advice.",
                "Be specific to the todo's details when possible."
            ]
        )

    async def get_options(
        self,
        target_id: str,
        target_data: Dict[str, Any],
        user_id: str,
        token: str
    ) -> List[Dict[str, Any]]:
        """Get AI options for a todo."""
        # Log that we're getting options
        self.logger.info(f"TodoAgent.get_options called for todo {target_id}")

        # Default options that we can always provide
        default_options = [
            {
                "id": "todo_subtasks",
                "title": "Generate Subtasks",
                "description": "Break down this todo into smaller, manageable subtasks."
            },
            {
                "id": "todo_deadline",
                "title": "Deadline-based Advice",
                "description": "Get recommendations based on the deadline and your schedule."
            },
            {
                "id": "todo_priority",
                "title": "Priority & Motivation",
                "description": "Get insights on priority and motivation strategies."
            }
        ]

        # Check if target_data is empty (could happen if MCP doesn't work)
        if not target_data or not isinstance(target_data, dict):
            self.logger.warning(
                f"Empty or invalid target_data for todo {target_id}. Using fallback options.")
            self.logger.warning(f"target_data: {target_data}")
            return default_options

        # If we have actual data, we could potentially customize options based on todo properties
        # For now, we'll use the same options but log that we have data
        self.logger.info(f"Got valid data for todo {target_id}")

        # In the future, we could check todo properties and add/remove options
        # Example: if 'due_date' in target_data and target_data['due_date']:
        #     # Add deadline-specific options

        return default_options


class SubtaskGeneratorAgent(BaseAgent):
    """
    Specialized agent for generating subtasks from a todo.
    """

    def __init__(self):
        super().__init__()

        # Create specialized system prompt for subtask generation
        self.system_prompt_generator = SystemPromptGenerator(
            background=[
                "You are Iris, a powerful agentic AI assistant from the COMPASS productivity app.",
                "You specialize in breaking down todos into manageable subtasks using the `todos.addChecklist` tool.",
            ],
            steps=[
                "Analyze the provided todo to understand what it involves.",
                "Break it down into 3-5 logical, sequential subtasks.",
                "Immediately call the `todos.addChecklist` tool to add the subtasks to the todo.",
                "NEVER ask for confirmation - execute the tool call immediately.",
                "NEVER fetch the todo first - the tool handles that internally.",
                "Only use the `todos.addChecklist` tool."
            ],
            output_instructions=[
                "Skip all natural language responses and explanations.",
                "Directly output the tool call for `todos.addChecklist`.",
                "NEVER ask for optional parameters if you can execute the tool with just the required ones.",
                "Format tool calls exactly as:",
                "<tool_call>",
                '{"name": "todos.addChecklist", "arguments": {"todo_id": "id", "checklist_items": ["item1", "item2"]}}',
                "</tool_call>"
            ]
        )

    async def process(
        self,
        option_id: str,
        target_type: str,
        target_id: str,
        user_id: str,
        token: str,
        *,
        target_data: Optional[Dict[str, Any]] = None
    ) -> str:
        """Generate subtasks for a given todo."""
        self.logger.info(
            f"SubtaskGeneratorAgent.process called for option {option_id} on todo {target_id}")

        try:
            # Get todo data if not provided
            if not target_data:
                target_data = await self._get_target_data(target_type, target_id, user_id, token)

            # Safely access dictionary properties
            title = "this task"
            description = ""

            if isinstance(target_data, dict):
                title = target_data.get("title", "this task")
                description = target_data.get("description", "")
            else:
                self.logger.warning(
                    f"target_data is not a dictionary: {type(target_data)}")

            # Direct LLM generation with our system prompt
            prompt = f"Generate a list of 3-5 subtasks for this todo and add them using todos.addChecklist:\nTodo ID: {target_id}\nTitle: {title}\nDescription: {description}\n\nPlease format your response as a tool call to todos.addChecklist with the checklist items."

            # Use our run method which directly calls the LLM service
            result = await self.run(
                {"prompt": prompt},
                user_id
            )

            if result["status"] == "success":
                # Extract tool calls from the response
                tool_calls = self._extract_tool_calls(result["response"])

                if not tool_calls:
                    self.logger.warning("No tool calls found in LLM response")
                    return "I was unable to generate subtasks. Please try again."

                # Process each tool call
                for tool_call in tool_calls:
                    try:
                        if tool_call["name"] != "todos.addChecklist":
                            continue

                        # Get MCP client
                        mcp_client = await self._get_mcp_client()
                        if not mcp_client:
                            raise ValueError("MCP client is not available")

                        # Get arguments and inject authorization token
                        tool_args = tool_call.get("arguments", {})
                        tool_args["authorization"] = f"Bearer {token}"

                        # Execute the tool call with retry logic
                        tool_result = await mcp_client.call_tool(
                            tool_call["name"],
                            tool_args
                        )

                        # Process the result
                        if tool_result.get("status") == "success":
                            self.logger.info(
                                f"Successfully added checklist items to todo {target_id}")
                            content = tool_result.get("content", {})
                            return f"Successfully added {len(tool_args.get('checklist_items', []))} subtasks to your todo."
                        else:
                            error_msg = tool_result.get(
                                "error", "Unknown error")
                            self.logger.error(f"Tool call failed: {error_msg}")
                            return f"Failed to add subtasks: {error_msg}"

                    except Exception as e:
                        self.logger.error(
                            f"Error executing tool call: {str(e)}")
                        return f"Error adding subtasks: {str(e)}"

                return "No valid subtasks were generated. Please try again."
            else:
                # Fall back to direct LLM call with tool instruction
                fallback_result = await self._generate_response_with_tools(
                    f"Add 3-5 subtasks to this todo using todos.addChecklist:\nTodo ID: {target_id}\nTitle: {title}\nDescription: {description}",
                    user_id,
                    {"temperature": 0.7, "top_p": 0.9},
                    token
                )

                # Process the fallback result the same way
                tool_calls = self._extract_tool_calls(fallback_result)
                if not tool_calls:
                    return "I was unable to generate subtasks. Please try again."

                # Process tool calls from fallback
                for tool_call in tool_calls:
                    if tool_call["name"] == "todos.addChecklist":
                        try:
                            mcp_client = await self._get_mcp_client()
                            if not mcp_client:
                                raise ValueError("MCP client is not available")

                            tool_args = tool_call.get("arguments", {})
                            tool_args["authorization"] = f"Bearer {token}"

                            tool_result = await mcp_client.call_tool(
                                tool_call["name"],
                                tool_args
                            )

                            if tool_result.get("status") == "success":
                                return f"Successfully added {len(tool_args.get('checklist_items', []))} subtasks to your todo."
                            else:
                                return f"Failed to add subtasks: {tool_result.get('error', 'Unknown error')}"
                        except Exception as e:
                            self.logger.error(
                                f"Error in fallback tool execution: {str(e)}")
                            return f"Error adding subtasks: {str(e)}"

                return "No valid subtasks were generated. Please try again."

        except Exception as e:
            self.logger.error(
                f"Error in SubtaskGeneratorAgent.process: {str(e)}", exc_info=True)
            return f"Sorry, I encountered an error while generating subtasks: {str(e)}"

    def _extract_tool_calls(self, text: str) -> List[Dict[str, Any]]:
        """Extract tool calls from LLM response."""
        tool_calls = []
        start_tag = "<tool_call>"
        end_tag = "</tool_call>"

        while start_tag in text and end_tag in text:
            start = text.find(start_tag) + len(start_tag)
            end = text.find(end_tag)
            if start > -1 and end > -1:
                tool_call_text = text[start:end].strip()
                try:
                    tool_call = json.loads(tool_call_text)
                    if "name" in tool_call:
                        tool_calls.append(tool_call)
                    else:
                        self.logger.warning(
                            f"Tool call missing 'name' field: {tool_call_text}")
                except json.JSONDecodeError:
                    self.logger.error(
                        f"Failed to parse tool call: {tool_call_text}")
                text = text[end + len(end_tag):]
            else:
                break

        return tool_calls

    async def _get_mcp_client(self):
        """Get MCP client from global state."""
        from core.mcp_state import get_mcp_client
        return get_mcp_client()


class DeadlineAdvisorAgent(BaseAgent):
    """
    Specialized agent for providing deadline-based advice.
    """

    def __init__(self):
        super().__init__()

        # Create a specialized system prompt for expert deadline advice.
        self.system_prompt_generator = SystemPromptGenerator(
            background=[
                "You are IRIS, an expert AI productivity coach within the COMPASS app.",
                "Your mission is to provide concise, actionable advice to help users meet their deadlines."
            ],
            steps=[
                "1. Assess Urgency & Importance: Evaluate the todo's due date against its priority.",
                "2. Strategize: Formulate a core, actionable strategy. e.g., tackle it now, schedule it, or break it down.",
                "3. Synthesize: Distill your analysis into a brief, powerful recommendation."
            ],
            output_instructions=[
                "Provide sharp, specific, and actionable advice in a supportive but direct tone.",
                "Focus on a single, powerful strategy for the user to implement.",
                "Your response **must be under 100 words**.",
                "Directly address the user. Avoid generic phrases like 'It is important to...'."
            ]
        )

    async def process(
        self,
        option_id: str,
        target_type: str,
        target_id: str,
        user_id: str,
        token: str,
        *,
        target_data: Optional[Dict[str, Any]] = None
    ) -> str:
        """Provide deadline-based advice for a todo."""
        self.logger.info(
            f"DeadlineAdvisorAgent.process called for option {option_id} on todo {target_id}")

        try:
            # Get todo data if not provided
            if not target_data:
                target_data = await self._get_target_data(target_type, target_id, user_id, token)

            # Safely access dictionary properties
            title = "this task"
            due_date = "unknown"
            priority = "medium"
            status = "pending"

            if isinstance(target_data, dict):
                title = target_data.get("title", "this task")
                due_date = target_data.get("due_date", "unknown")
                priority = target_data.get("priority", "medium")
                status = target_data.get("status", "pending")
            else:
                self.logger.warning(
                    f"target_data is not a dictionary: {type(target_data)}")

            # Create prompt that encourages tool use
            prompt = f"""
Provide deadline advice for the following task. DO NOT EVER USE TOOLS.

Todo: {title}
Due date: {due_date}
Priority: {priority}
Status: {status}

Please provide specific recommendations on how to approach this task based on its deadline. Consider time management strategies, scheduling tips, and how to prioritize it among other tasks.
"""

            # Generate response with model parameters for better advice
            model_params = {
                "temperature": 0.5,
                "top_p": 0.8
            }
            return await self._generate_response_with_tools(prompt, user_id, model_params, token)
        except Exception as e:
            self.logger.error(
                f"Error in DeadlineAdvisorAgent.process: {str(e)}", exc_info=True)
            return f"Sorry, I encountered an error while generating deadline advice: {str(e)}"


class PriorityOptimizerAgent(BaseAgent):
    """
    Specialized agent for optimizing task priority.
    """

    def __init__(self):
        super().__init__()

        # Create specialized system prompt for priority optimization
        self.system_prompt_generator = SystemPromptGenerator(
            background=[
                "You are IRIS, an AI assistant for the COMPASS productivity app.",
                "You specialize in optimizing task priorities and providing motivation."
            ],
            steps=[
                "Analyze the todo's priority, description, and due date.",
                "Evaluate if the current priority setting is appropriate.",
                "Develop motivation strategies specific to this task."
            ],
            output_instructions=[
                "Provide insights on whether the priority is appropriate.",
                "Offer specific motivation strategies for completing this task.",
                "Keep your response under 150 words and make it actionable."
            ]
        )

    async def process(
        self,
        option_id: str,
        target_type: str,
        target_id: str,
        user_id: str,
        token: str,
        *,
        target_data: Optional[Dict[str, Any]] = None
    ) -> str:
        """Optimize the priority of a todo."""
        self.logger.info(
            f"PriorityOptimizerAgent.process called for option {option_id} on todo {target_id}")

        try:
            # Get todo data if not provided
            if not target_data:
                target_data = await self._get_target_data(target_type, target_id, user_id, token)

            # Safely access dictionary properties
            title = "this task"
            priority = "medium"
            description = ""
            due_date = "unknown"

            if isinstance(target_data, dict):
                title = target_data.get("title", "this task")
                priority = target_data.get("priority", "medium")
                description = target_data.get("description", "")
                due_date = target_data.get("due_date", "unknown")
            else:
                self.logger.warning(
                    f"target_data is not a dictionary: {type(target_data)}")

            # Create prompt that encourages tool use
            prompt = f"""
You are IRIS, an AI assistant for the COMPASS productivity app.
I need priority and motivation advice for this todo. You can use tools to understand the context of my other work.

Todo: {title}
Description: {description}
Current priority: {priority}
Due date: {due_date}

Please provide insights on whether this priority is appropriate, and offer specific motivation strategies for completing this task. Keep your response under 150 words and make it actionable.
"""

            # Generate response with model parameters for better motivation advice
            model_params = {
                "temperature": 0.6,
                "top_p": 0.85
            }
            return await self._generate_response_with_tools(prompt, user_id, model_params, token)
        except Exception as e:
            self.logger.error(
                f"Error in PriorityOptimizerAgent.process: {str(e)}", exc_info=True)
            return f"Sorry, I encountered an error while generating priority advice: {str(e)}"
