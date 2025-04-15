from datetime import datetime
from data_layer.models.system_metric_model import SystemMetric
from .base_repo import BaseMongoRepository
from typing import Optional, List, Dict, Any
from pymongo import ASCENDING


class SystemMetricRepository(BaseMongoRepository[SystemMetric]):
    def __init__(self):
        super().__init__(SystemMetric)

    def find_by_user(self, user_id: str) -> List[SystemMetric]:
        return self.find_many({"user_id": user_id})

    def find_by_type_and_range(self, user_id: str, metric_type: str, start: datetime, end: datetime) -> List[SystemMetric]:
        return self.find_many({
            "user_id": user_id,
            "metric_type": metric_type,
            "timestamp": {"$gte": start, "$lte": end}
        })

    def create_metric(self, metric: SystemMetric) -> str:
        return self.insert(metric)

    def aggregate_metrics(self, user_id: str, period: str = "daily", metric_type: Optional[str] = None, start: Optional[datetime] = None, end: Optional[datetime] = None) -> List[Dict[str, Any]]:
        """
        Aggregate metrics for a user by period (daily/weekly/monthly) and metric_type.
        Returns a list of dicts: [{period, metric_type, sum, avg, min, max, count}]
        """
        pipeline = []
        match = {"user_id": user_id}
        if metric_type:
            match["metric_type"] = metric_type
        if start or end:
            match["timestamp"] = {}
            if start:
                match["timestamp"]["$gte"] = start
            if end:
                match["timestamp"]["$lte"] = end
        pipeline.append({"$match": match})

        # Group by period and metric_type
        if period == "daily":
            group_id = {"date": {"$dateToString": {
                "format": "%Y-%m-%d", "date": "$timestamp"}}, "metric_type": "$metric_type"}
        elif period == "weekly":
            group_id = {"week": {"$isoWeek": "$timestamp"}, "year": {
                "$isoWeekYear": "$timestamp"}, "metric_type": "$metric_type"}
        elif period == "monthly":
            group_id = {"month": {"$month": "$timestamp"}, "year": {
                "$year": "$timestamp"}, "metric_type": "$metric_type"}
        else:
            group_id = {"metric_type": "$metric_type"}

        pipeline.append({
            "$group": {
                "_id": group_id,
                "sum": {"$sum": "$value"},
                "avg": {"$avg": "$value"},
                "min": {"$min": "$value"},
                "max": {"$max": "$value"},
                "count": {"$sum": 1}
            }
        })
        pipeline.append({"$sort": {"_id": ASCENDING}})
        collection = self.get_collection()
        results = list(collection.aggregate(pipeline))
        # Post-process for easier consumption
        for r in results:
            period_info = r.pop("_id")
            if isinstance(period_info, dict):
                if "date" in period_info:
                    r["period"] = period_info["date"]
                elif "week" in period_info and "year" in period_info:
                    r["period"] = f"{period_info['year']}-W{period_info['week']}"
                elif "month" in period_info and "year" in period_info:
                    r["period"] = f"{period_info['year']}-{period_info['month']:02d}"
                else:
                    r["period"] = "all"
                r["metric_type"] = period_info.get("metric_type", "all")
            else:
                r["period"] = "all"
                r["metric_type"] = "all"
        return results
