export interface HabitsReportContent {
    overall_score: number;
    key_metrics: {
        overall_completion_rate: number;
        total_habits: number;
        completed_habits: number;
        longest_streak: number;
        average_streak: number;
    };
    insights: string[];
    habit_recommendations: string[];
} 