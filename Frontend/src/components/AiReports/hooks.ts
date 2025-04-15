import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import { createReport, getReport } from "./api";
import { CreateReportPayload, ReportGenerationUpdate, Report } from "./types";
import { User } from '@/api/auth';
import { useEffect, useState, useRef } from "react";
import { getApiUrls } from "@/config";

export const useCreateReport = (user: User | undefined) => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (payload: CreateReportPayload) => user ? createReport(payload) : Promise.reject("No user"),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['reports'] });
        },
    });
};

export const useGetReport = (reportId: string | null) => {
    return useQuery<Report, Error>({
        queryKey: ['report', reportId],
        queryFn: () => getReport(reportId!),
        enabled: !!reportId,
    });
};

export const useReportGeneration = (reportId: string | null) => {
    const [progress, setProgress] = useState<ReportGenerationUpdate | null>(null);
    const [isGenerating, setIsGenerating] = useState(false);
    const [isGenerationComplete, setIsGenerationComplete] = useState(false);
    const socketRef = useRef<WebSocket | null>(null);
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        if (!reportId || !isGenerating) {
            return;
        }

        const token = localStorage.getItem("token");
        if (!token) {
            console.error("No token found in local storage");
            setIsGenerating(false);
            return;
        }

        const { WS_PYTHON_URL } = getApiUrls();
        // Convert to WebSocket URL properly - WS_PYTHON_URL already includes /api/v1
        const wsUrl = `${WS_PYTHON_URL.replace(/^http/, 'ws')}/ws/reports/${reportId}?token=${token}`;
        
        console.log('Connecting to AI Reports WebSocket:', wsUrl);
        socketRef.current = new WebSocket(wsUrl);

        socketRef.current.onopen = () => {
            console.log("WebSocket connected");
            socketRef.current?.send(JSON.stringify({ action: "generate" }));
        };

        socketRef.current.onmessage = (event) => {
            const data: ReportGenerationUpdate = JSON.parse(event.data);
            setProgress(data);
            if (data.status === 'completed') {
                timeoutRef.current = setTimeout(() => {
                    setIsGenerationComplete(true);
                    setIsGenerating(false);
                    socketRef.current?.close();
                }, 1000); // 1-second delay
            } else if (data.status === 'failed') {
                setIsGenerating(false);
                socketRef.current?.close();
            }
        };

        socketRef.current.onerror = (error) => {
            console.error("WebSocket error:", error);
            setIsGenerating(false);
        };

        socketRef.current.onclose = () => {
            console.log("WebSocket disconnected");
            setIsGenerating(false);
        };

        return () => {
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }
            socketRef.current?.close();
        };
    }, [reportId, isGenerating]);

    const startGeneration = () => {
        if (reportId) {
            setIsGenerating(true);
            setIsGenerationComplete(false);
            setProgress(null);
        }
    };

    return { progress, startGeneration, isGenerating, isGenerationComplete };
}; 