from typing import Dict, List, Optional, Any
from datetime import datetime, date, timedelta
from crewai import Agent, Task, Crew
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field
from Backend.services.daily_habits_service import DailyHabitService
import json
import re
from sqlalchemy.ext.asyncio import AsyncSession

logger = get_logger(__name__)


class HabitAgent:
    def __init__(self, db_session=None):
        self.agent = Agent(
            role="Habit Creation Specialist",
            goal="Create habits from natural language instructions",
            backstory="""I am an expert in habit formation and tracking. I help users 
                        efficiently create and manage their habits by understanding natural 
                        language and extracting structured information.""",
            tools=[
                Tool(
                    name="create_habit",
                    func=self.create_habit,
                    description="Creates a new habit from natural language"
                ),
                Tool(
                    name="edit_habit",
                    func=self.edit_habit,
                    description="Processes a natural language request to edit a habit, identifying which habit to edit and how"
                )
            ],
            verbose=True,
            allow_delegation=True  # Enable task delegation
        )
        
        self.ai_service = LLMService()
        self.db = db_session
        
        # Initialize services instead of repositories
        if db_session:
            # For DailyHabitService, we need to provide a DailyHabitRepository instance
            from Backend.data_layer.repositories.daily_habits_repository import DailyHabitRepository
            habit_repo = DailyHabitRepository(db_session)
            self.habit_service = DailyHabitService(habit_repo)
        else:
            self.habit_service = None
    
    async def create_habit(self, description: str, user_id: int = None) -> Dict[str, Any]:
        """Create a new habit from natural language description."""
        try:
            # Set current_time once at the beginning
            current_time = datetime.utcnow()
            
            # Use LLM to extract structured data
            habit_analysis = await self.ai_service.generate_response(
                prompt=f"""Extract structured habit information from this description:
Description: {description}

Extract and format as JSON with ONLY these fields:
- habit_name: A clear, concise name for the habit
- description: Full description with details
- start_day: Start date in YYYY-MM-DD format
- end_day: Optional end date in YYYY-MM-DD format""",
                context={
                    "system_message": "You are a habit extraction AI. Extract structured habit information from natural language, using only the supported fields."
                }
            )
            
            try:
                # Parse the JSON response
                response_text = habit_analysis.get("text", "")
                
                # Extract JSON from markdown if needed
                if "```json" in response_text:
                    json_match = re.search(r'```json\n(.*?)\n```', response_text, re.DOTALL)
                    if json_match:
                        response_text = json_match.group(1)
                
                habit_data = json.loads(response_text)
                
                # Convert date strings to date objects
                if "start_day" in habit_data and habit_data["start_day"]:
                    try:
                        start_day = datetime.strptime(habit_data["start_day"], "%Y-%m-%d").date()
                        # Ensure start_day is not in the past
                        if start_day < current_time.date():
                            logger.warning(f"Start day {start_day} is in the past, using current date")
                            habit_data["start_day"] = current_time.date()
                        else:
                            habit_data["start_day"] = start_day
                    except ValueError:
                        logger.warning(f"Invalid start_day format: {habit_data['start_day']}, using current date")
                        habit_data["start_day"] = current_time.date()
                else:
                    habit_data["start_day"] = current_time.date()
                
                if "end_day" in habit_data and habit_data["end_day"]:
                    try:
                        end_day = datetime.strptime(habit_data["end_day"], "%Y-%m-%d").date()
                        # Ensure end_day is after start_day
                        if end_day < habit_data["start_day"]:
                            # Set end_day to 30 days after start_day
                            logger.warning(f"End day {end_day} is before start day, adjusting to 30 days later")
                            habit_data["end_day"] = habit_data["start_day"] + timedelta(days=30)
                        else:
                            habit_data["end_day"] = end_day
                    except ValueError:
                        logger.warning(f"Invalid end_day format: {habit_data['end_day']}")
                        habit_data.pop("end_day", None)
                
                # Add metadata - use actual datetime objects, not strings
                habit_data["created_at"] = current_time
                habit_data["updated_at"] = current_time
                habit_data["current_streak"] = 0
                habit_data["longest_streak"] = 0
                habit_data["is_completed"] = False
                habit_data["user_id"] = user_id

                # Save to database if service is available
                if self.habit_service is not None:
                    db_habit = await self.habit_service.create_habit(**habit_data)
                    if db_habit:
                        habit_data["id"] = db_habit.id if hasattr(db_habit, 'id') else None
                
                # Ensure all values are JSON serializable
                # Convert datetime objects to strings
                for key, value in list(habit_data.items()):
                    if isinstance(value, datetime):
                        habit_data[key] = value.isoformat()
                    elif isinstance(value, date):  # Handle date objects
                        habit_data[key] = value.isoformat()
                    elif isinstance(value, (list, tuple)):
                        # Handle lists that might contain non-serializable items
                        habit_data[key] = [
                            item.isoformat() if hasattr(item, "isoformat") else 
                            item 
                            for item in value
                        ]
                
                return {
                    "status": "success",
                    "message": f"Habit '{habit_data['habit_name']}' created successfully",
                    "habit": habit_data
                }
            except Exception as e:
                logger.error(f"Error parsing habit data: {str(e)}")
                return {
                    "status": "error",
                    "message": f"Could not parse habit information: {str(e)}",
                    "raw_response": habit_analysis.get("text", "")
                }
        except Exception as e:
            logger.error(f"Habit creation failed: {str(e)}")
            return {"status": "error", "message": f"Habit creation failed: {str(e)}"} 

    async def edit_habit(self, description: str, user_id: int = None, previous_messages: List[Dict[str, str]] = None) -> Dict[str, Any]:
        """Edit an existing habit based on natural language description."""
        try:
            current_time = datetime.utcnow()
            
            # Prepare context for identifying which habit to edit
            context_prompt = ""
            
            # Add previous conversation messages for context if available
            if previous_messages and len(previous_messages) > 0:
                context_prompt += "Previous conversation:\n"
                # Add up to 5 most recent messages for context
                for msg in previous_messages[-5:]:
                    sender = msg.get("sender", "unknown")
                    text = msg.get("text", "")
                    context_prompt += f"{sender.capitalize()}: {text}\n"
                context_prompt += "\n"
            
            # Get user's habits for context if habit_service is available
            habits_context = ""
            if self.habit_service is not None:
                try:
                    # Get all habits for this user
                    user_habits = await self.habit_service.get_user_habits(user_id=user_id)
                    if user_habits:
                        habits_context += "User's current habits:\n"
                        for i, habit in enumerate(user_habits, 1):
                            # Format each habit with key information
                            # Handle both dict and object formats
                            if isinstance(habit, dict):
                                habit_name = habit.get("habit_name")
                                habit_description = habit.get("description")
                                habit_start_day = habit.get("start_day", "No start date")
                                habit_end_day = habit.get("end_day", "No end date")
                                habit_id = habit.get("id")
                                habit_current_streak = habit.get("current_streak", 0)
                                habit_longest_streak = habit.get("longest_streak", 0)
                                habit_is_completed = habit.get("is_completed", False)
                            else:
                                # Original object access
                                habit_name = habit.habit_name
                                habit_description = habit.description
                                habit_start_day = habit.start_day.strftime("%Y-%m-%d") if habit.start_day else "No start date"
                                habit_end_day = habit.end_day.strftime("%Y-%m-%d") if habit.end_day else "No end date"
                                habit_id = habit.id
                                habit_current_streak = habit.current_streak
                                habit_longest_streak = habit.longest_streak
                                habit_is_completed = habit.is_completed
                            
                            habits_context += f"{i}. ID: {habit_id} - Name: '{habit_name}' - Description: '{habit_description}' - "
                            habits_context += f"Start: {habit_start_day} - End: {habit_end_day} - "
                            habits_context += f"Current Streak: {habit_current_streak} - Longest Streak: {habit_longest_streak} - "
                            habits_context += f"Completed: {'Yes' if habit_is_completed else 'No'}\n"
                except Exception as e:
                    logger.error(f"Error fetching user habits: {str(e)}")
                    habits_context = "Error fetching user habits\n"
            
            # Use LLM to identify which habit to edit and what changes to make
            edit_analysis = await self.ai_service.generate_response(
                prompt=f"""Identify which habit to edit and what changes to make based on this user request:

{context_prompt}
{habits_context}

User request: {description}

IMPORTANT INSTRUCTIONS FOR HABIT EDITING:
1. When a user wants to change the habit frequency (e.g., from daily to weekly), update the "habit_name" field
2. Pay careful attention to the "habit_name" when the user wants to rename or change the activity's frequency
3. Pay careful attention to verbs, frequencies, and activities in the habit name
4. The primary purpose of "description" is for additional details only, not to rename the habit

EXAMPLE:
- If habit_name is "Exercise for 30 minutes Daily" and user says "change my exercise habit from daily to weekly"
- You MUST update habit_name to "Exercise for 30 minutes Weekly"
- NOT just update the description

Return a STRICT JSON object with the following structure (no additional text):
{{
    "habit_id": number,
    "changes": {{
        "habit_name"?: string,
        "description"?: string,
        "start_day"?: "YYYY-MM-DD",
        "end_day"?: "YYYY-MM-DD",
        "is_completed"?: boolean
    }},
    "reason": string
}}

Note: Only include fields in "changes" that need to be updated.""",
                context={
                    "system_message": "You are a habit management AI. Generate valid JSON for habit updates. When users want to change the habit activity or frequency, update the habit_name field, not just the description. Pay special attention to frequency changes (daily, weekly, monthly) in the habit name."
                }
            )
            
            try:
                # Parse the JSON response
                response_text = edit_analysis.get("text", "")
                logger.debug(f"Raw LLM response:\n{response_text}")
                
                # Clean and parse JSON
                response_text = (response_text
                    .strip()                     # Remove leading/trailing whitespace
                    .replace('\n', ' ')          # Remove newlines
                    .replace('\r', '')           # Remove carriage returns
                )
                
                # Remove any markdown code block markers if present
                response_text = response_text.replace('```json', '').replace('```', '')
                
                logger.debug(f"After cleanup:\n{response_text}")
                try:
                    edit_data = json.loads(response_text)
                except json.JSONDecodeError as e:
                    logger.error(f"JSON parse error at position {e.pos}. Near text: {response_text[max(0, e.pos-50):min(len(response_text), e.pos+50)]}")
                    raise
                
                # Ensure we have the required habit_id
                if "habit_id" not in edit_data or not edit_data["habit_id"]:
                    return {
                        "status": "error",
                        "message": "Could not identify which habit to edit",
                        "raw_response": response_text
                    }
                
                # Ensure habit_id is an integer
                try:
                    # Convert string ID to integer if needed
                    habit_id = int(str(edit_data["habit_id"]).strip())
                    edit_data["habit_id"] = habit_id
                except (ValueError, TypeError):
                    return {
                        "status": "error", 
                        "message": f"Invalid habit ID format: {edit_data['habit_id']}. Expected a numeric ID.",
                        "raw_response": response_text
                    }
                
                # Ensure we have changes to apply
                if "changes" not in edit_data or not edit_data["changes"]:
                    return {
                        "status": "error",
                        "message": "No changes specified",
                        "raw_response": response_text
                    }
                
                # Prepare update data
                update_data = edit_data["changes"]
                update_data["updated_at"] = current_time
                
                # Log the full update data for debugging
                logger.info(f"Habit update requested - ID: {edit_data['habit_id']}, Changes: {update_data}")
                if "habit_name" in update_data:
                    logger.info(f"Updating habit_name to: {update_data['habit_name']}")
                
                # Convert date strings to date objects
                if "start_day" in update_data and update_data["start_day"]:
                    try:
                        start_day = datetime.strptime(update_data["start_day"], "%Y-%m-%d").date()
                        update_data["start_day"] = start_day
                    except ValueError:
                        logger.warning(f"Invalid start_day format: {update_data['start_day']}")
                        update_data.pop("start_day", None)
                
                if "end_day" in update_data and update_data["end_day"]:
                    try:
                        end_day = datetime.strptime(update_data["end_day"], "%Y-%m-%d").date()
                        update_data["end_day"] = end_day
                    except ValueError:
                        logger.warning(f"Invalid end_day format: {update_data['end_day']}")
                        update_data.pop("end_day", None)
                
                # Apply the update if service is available
                if self.habit_service is not None:
                    habit_id = edit_data["habit_id"]
                    updated_habit = await self.habit_service.update_habit(habit_id, user_id, **update_data)
                    
                    if updated_habit:
                        # Convert the updated habit to a serializable format
                        habit_dict = {
                            "id": updated_habit["id"] if isinstance(updated_habit, dict) else updated_habit.id,
                            "habit_name": updated_habit["habit_name"] if isinstance(updated_habit, dict) else updated_habit.habit_name,
                            "description": updated_habit["description"] if isinstance(updated_habit, dict) else updated_habit.description,
                            "start_day": updated_habit["start_day"] if isinstance(updated_habit, dict) else updated_habit.start_day.isoformat(),
                            "end_day": updated_habit["end_day"] if isinstance(updated_habit, dict) else (updated_habit.end_day.isoformat() if updated_habit.end_day else None),
                            "current_streak": updated_habit["current_streak"] if isinstance(updated_habit, dict) else updated_habit.current_streak,
                            "longest_streak": updated_habit["longest_streak"] if isinstance(updated_habit, dict) else updated_habit.longest_streak,
                            "is_completed": updated_habit["is_completed"] if isinstance(updated_habit, dict) else updated_habit.is_completed,
                            "user_id": updated_habit["user_id"] if isinstance(updated_habit, dict) else updated_habit.user_id,
                            "created_at": updated_habit["created_at"] if isinstance(updated_habit, dict) else updated_habit.created_at.isoformat(),
                            "updated_at": updated_habit["updated_at"] if isinstance(updated_habit, dict) else updated_habit.updated_at.isoformat()
                        }
                        
                        completion_status = ""
                        if "is_completed" in update_data:
                            completion_status = f" and marked as {'completed' if update_data['is_completed'] else 'not completed'}"
                        
                        # Build a more descriptive success message
                        changes_description = []
                        if "habit_name" in update_data:
                            changes_description.append(f"name changed to '{update_data['habit_name']}'")
                        if "description" in update_data:
                            changes_description.append("description updated")
                        if "start_day" in update_data:
                            changes_description.append(f"start date set to {update_data['start_day'].isoformat()}")
                        if "end_day" in update_data:
                            changes_description.append(f"end date set to {update_data['end_day'].isoformat()}")
                        
                        changes_text = ""
                        if changes_description:
                            changes_text = f" with {', '.join(changes_description)}"
                        
                        return {
                            "status": "success",
                            "message": f"Habit '{habit_dict['habit_name']}' updated successfully{changes_text}{completion_status}",
                            "habit": habit_dict,
                            "changes_applied": list(update_data.keys())
                        }
                    else:
                        return {
                            "status": "error",
                            "message": f"Habit with ID {habit_id} not found or could not be updated",
                            "habit_id": habit_id
                        }
                else:
                    # For testing without database
                    return {
                        "status": "success",
                        "message": "Habit would be updated (no database connection)",
                        "habit_id": edit_data["habit_id"],
                        "changes": update_data,
                        "reason": edit_data.get("reason", "")
                    }
                
            except Exception as e:
                logger.error(f"Error parsing edit data: {str(e)}")
                return {
                    "status": "error",
                    "message": f"Could not parse habit edit information: {str(e)}",
                    "raw_response": response_text
                }
                
        except Exception as e:
            logger.error(f"Habit edit failed: {str(e)}")
            return {"status": "error", "message": f"Habit edit failed: {str(e)}"} 