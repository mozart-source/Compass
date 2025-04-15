from typing import Dict, Any


class BaseHandler:
    """
    Base class for domain-specific context handlers.
    """

    def __init__(self, db_session):
        self.db = db_session

    async def enrich_context(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        This method should be overridden by specific handlers.
        """
        raise NotImplementedError(
            "enrich_context must be implemented by the subclass.")
