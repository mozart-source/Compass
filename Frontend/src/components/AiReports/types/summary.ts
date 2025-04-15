export interface SummaryReportContent {
    overall_score: number;
    key_metrics: {
        "Active Days": number;
        "Activity Trend": string;
        "Average Productivity Score": string;
        "Average Daily Focus Time": string;
        "Task Completion Rate": string;
        "Tasks Completed": number;
        "Habit Completion Rate": string;
        "Meeting Time": string;
        "Project Completion Rate": string;
        "Workflows Executed": string;
    };
    achievements: string[];
    areas_for_improvement: string[];
    recommendations: string[];
} 