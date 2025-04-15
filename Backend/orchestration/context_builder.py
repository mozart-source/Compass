from typing import Dict, Any, Optional, List
from orchestration.ai_registry import ai_registry
import logging

logger = logging.getLogger(__name__)


class ContextBuilder:
    def __init__(self, db_session):
        self.db = db_session

    async def get_full_context(self, user_id: int, domains: Optional[List[str]] = None) -> Dict[str, Any]:
        """
        Retrieve and merge data from specified domains or all registered domains.

        Args:
            user_id: The user's ID
            domains: Optional list of specific domains to fetch context from

        Returns:
            Dict containing merged context from all relevant domains
        """
        context = {}
        target_domains = domains if domains else ai_registry.handler_mapping.keys()

        for domain in target_domains:
            try:
                repository_class = ai_registry.get_repository(domain)
                repository = repository_class(self.db)

                # Fetch and enrich context
                domain_context = await repository.get_context(user_id)

                # Add user_id to context for handlers
                if isinstance(domain_context, dict):
                    domain_context["user_id"] = user_id

                handler = ai_registry.get_handler(domain, self.db)
                if handler:
                    try:
                        domain_context = await handler.enrich_context(domain_context)
                    except Exception as e:
                        logger.error(
                            f"Error enriching context for domain {domain}: {str(e)}")

                context[domain] = domain_context
            except Exception as e:
                logger.error(
                    f"Error fetching context for domain {domain}: {str(e)}")
                context[domain] = {"error": str(e)}

        return context

    async def get_user_profile(self, user_id: int) -> Dict[str, Any]:
        """Get user profile data for context enrichment"""
        # TODO: Implement user profile fetching
        return {}

    async def merge_contexts(self, contexts: Dict[str, Dict[str, Any]]) -> Dict[str, Any]:
        """
        Merge multiple domain contexts into a single context.
        This could be enhanced with ranking or filtering based on relevance.
        """
        merged = {}
        for domain, context in contexts.items():
            merged[domain] = context
        return merged
