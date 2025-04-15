import { DashboardReportContent } from './dashboard';
import { ActivityReportContent } from './activity';
import { ProductivityReportContent } from './productivity';
import { HabitsReportContent } from './habits';
import { TaskReportContent } from './task';
import { SummaryReportContent } from './summary';

export type ReportType = "productivity" | "activity" | "dashboard" | "habits" | "task" | "summary";

export interface CreateReportPayload {
  title: string;
  type: ReportType;
  time_range: {
    start_date: string;
    end_date: string;
  };
}

export interface CreateReportResponse {
  report_id: string;
}

export interface ReportGenerationUpdate {
  status: "in_progress" | "completed" | "failed";
  progress: number;
  message: string;
  report_id: string;
}

export interface ParsedReportSection {
    title: string;
    content: string;
    type: string;
}

export interface TimeRange {
    start_date: string;
    end_date: string;
}

export interface ReportTextContent {
    text: string;
}

export interface ParsedReportContent {
    summary: string;
    content: DashboardReportContent | ActivityReportContent | ProductivityReportContent | HabitsReportContent | TaskReportContent | SummaryReportContent;
    sections: ParsedReportSection[];
}

export interface Report {
    id: string;
    title: string;
    type: ReportType;
    status: 'completed' | 'in_progress' | 'failed';
    content: ReportTextContent;
    user_id: string;
    created_at: string;
    updated_at: string;
    completed_at: string | null;
    parameters: object;
    time_range: TimeRange;
    custom_prompt: string | null;
    summary: string | null;
    sections: ParsedReportSection[];
    error: string | null;
    parsedContent?: ParsedReportContent;
}

export type { DashboardReportContent, ActivityReportContent, ProductivityReportContent, HabitsReportContent, TaskReportContent, SummaryReportContent }; 