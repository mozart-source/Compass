from typing import Dict, List, Any, Optional, Tuple
from difflib import SequenceMatcher
from datetime import datetime, timedelta
import re
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)


class ReferenceResolver:
    def __init__(self):
        self.time_patterns = {
            "yesterday": timedelta(days=1),
            "last week": timedelta(days=7),
            "last month": timedelta(days=30),
            "today": timedelta(days=0),
            "this week": timedelta(days=7),
            "this month": timedelta(days=30),
        }

    def _calculate_similarity(self, a: str, b: str) -> float:
        """Calculate string similarity using SequenceMatcher."""
        return SequenceMatcher(None, a.lower(), b.lower()).ratio()

    def _extract_temporal_reference(self, text: str) -> Optional[Tuple[datetime, datetime]]:
        """Extract time references from text and convert to datetime ranges."""
        now = datetime.now()

        for pattern, delta in self.time_patterns.items():
            if pattern in text.lower():
                if "last" in pattern:
                    end = now - delta
                    start = end - delta
                else:
                    start = now - delta
                    end = now
                return start, end

        return None

    async def resolve_reference(
        self,
        reference: str,
        context: Dict[str, Any],
        similarity_threshold: float = 0.6
    ) -> Dict[str, Any]:
        """
        Resolve fuzzy references to items in the context.

        Args:
            reference: The reference text to resolve
            context: The full context containing items from all domains
            similarity_threshold: Minimum similarity score to consider a match

        Returns:
            Dict containing resolved items and confidence scores
        """
        matches = []
        temporal_range = self._extract_temporal_reference(reference)

        # Process each domain in the context
        for domain, domain_data in context.items():
            if not isinstance(domain_data, (list, dict)):
                continue

            items = domain_data if isinstance(
                domain_data, list) else [domain_data]

            for item in items:
                if not isinstance(item, dict):
                    continue

                # Calculate similarity scores for relevant fields
                title_score = self._calculate_similarity(
                    reference, item.get('title', ''))
                desc_score = self._calculate_similarity(
                    reference, item.get('description', ''))

                # Get the best score between title and description
                best_score = max(title_score, desc_score)

                # Check temporal match if temporal reference exists
                temporal_match = False
                if temporal_range and 'created_at' in item:
                    created_at = datetime.fromisoformat(item['created_at'])
                    start_time, end_time = temporal_range
                    temporal_match = start_time <= created_at <= end_time

                # Consider it a match if either similarity is high or temporal reference matches
                if best_score >= similarity_threshold or temporal_match:
                    matches.append({
                        'item': item,
                        'domain': domain,
                        'confidence': best_score,
                        'temporal_match': temporal_match,
                        'match_type': 'temporal' if temporal_match else 'fuzzy'
                    })

        # Sort matches by confidence score
        matches.sort(key=lambda x: x['confidence'], reverse=True)

        return {
            'matches': matches,
            'total_matches': len(matches),
            'best_match': matches[0] if matches else None,
            'has_temporal_reference': temporal_range is not None
        }
