from data_layer.models.goal_model import Goal
from .base_repo import BaseMongoRepository
from typing import Optional, List


class GoalRepository(BaseMongoRepository[Goal]):
    def __init__(self):
        super().__init__(Goal)

    def find_by_user(self, user_id: str) -> List[Goal]:
        return self.find_many({"user_id": user_id})


    def create_goal(self, goal: Goal) -> str:
        return self.insert(goal)

    def update_goal(self, goal_id: str, data: dict) -> Optional[Goal]:
        return self.update(goal_id, data)

    def delete_goal(self, goal_id: str) -> bool:
        return self.delete(goal_id)
