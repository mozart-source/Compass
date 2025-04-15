export interface TaskReportContent {
    task_efficiency_score: number;
    key_metrics: {
        "Task Completion Rate": string;
        "Tasks Completed": string;
        "Overdue Tasks": number;
        "Average Task Completion Time": string;
        "Todo Completion Rate": string;
        "Todos Completed": string;
        "Project Completion Rate": string;
        "Projects Completed": string;
        "Projects On Time": number;
        "Projects Delayed": number;
    };
    insights: string[];
    bottlenecks: string[];
    recommendations: string[];
} 