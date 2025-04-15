export interface ProductivityReportContent {
    productivity_score: number;
    key_metrics: {
        average_productivity_score: string;
        average_daily_focus_time_hours: string;
        task_completion_rate: string;
        tasks_completed: string;
        meeting_time_minutes: number;
        number_of_meetings: number;
    };
    insights: string[];
    areas_for_improvement: string[];
} 