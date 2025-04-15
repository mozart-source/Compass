export interface ActivityReportContent {
    activity_score: number;
    key_metrics: {
        tasks_completed: string;
        overdue_tasks: number;
        meetings_attended: number;
        total_meeting_time_minutes: number;
    };
    insights: string[];
} 