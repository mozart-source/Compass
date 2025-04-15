import axios from "axios";
import { CreateReportPayload, CreateReportResponse, Report } from "./types";
import { getApiUrls } from "@/config";

export const createReport = async (payload: CreateReportPayload): Promise<CreateReportResponse> => {
    const { PYTHON_API_URL } = getApiUrls();
    const response = await axios.post(`${PYTHON_API_URL}/reports`, payload, {
        headers: {
            'X-Organization-ID': undefined
        }
    });
    return response.data;
};

export const getReport = async (reportId: string): Promise<Report> => {
    const { PYTHON_API_URL } = getApiUrls();
    const response = await axios.get<Report>(`${PYTHON_API_URL}/reports/${reportId}`, {
        headers: {
            'X-Organization-ID': undefined
        }
    });
    const report = response.data;
    if (report.content && report.content.text) {
        try {
            // The content text is a JSON string wrapped in ```json ... ```, so we need to strip that before parsing.
            const sanitizedJsonString = report.content.text.replace(/^```json\n|```$/g, '');
            report.parsedContent = JSON.parse(sanitizedJsonString);
        } catch (e) {
            console.error("Failed to parse report content:", e);
        }
    }
    return report;
}; 