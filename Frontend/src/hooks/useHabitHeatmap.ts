import { useEffect, useState, useCallback } from "react";
import { useDashboardMetrics } from "@/components/dashboard/useDashboardMetrics";
import { useWebSocket } from "@/contexts/websocket-provider";

export type HeatmapPeriod = "week" | "month" | "year";
export type HeatmapData = Record<string, number>;

export const useHabitHeatmap = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<HeatmapData>({});
  const [isRefreshing, setIsRefreshing] = useState(false);
  
  // Get data from dashboard metrics
  const { 
    data: metricsData, 
    isLoading: metricsLoading, 
    isError: metricsError,
    error: metricsErrorData,
    requestRefresh
  } = useDashboardMetrics();
  
  // Get WebSocket connection for more targeted refresh
  const { sendMessage, isConnected } = useWebSocket();

  useEffect(() => {
    setLoading(metricsLoading);
    
    if (metricsError && metricsErrorData) {
      setError(metricsErrorData instanceof Error ? metricsErrorData : new Error("Failed to load heatmap data"));
    } else {
      setError(null);
    }
    
    if (metricsData?.habit_heatmap) {
      setData(metricsData.habit_heatmap);
      // If we were refreshing, mark as done
      if (isRefreshing) {
        setIsRefreshing(false);
      }
    }
  }, [metricsData, metricsLoading, metricsError, metricsErrorData, isRefreshing]);

  // Function to refresh all dashboard metrics
  const refreshData = useCallback(() => {
    requestRefresh();
    setIsRefreshing(true);
  }, [requestRefresh]);
  
  // Function to refresh only heatmap data via WebSocket
  const refreshHeatmap = useCallback(() => {
    setIsRefreshing(true);
    
    if (sendMessage && isConnected) {
      sendMessage({
        type: "refresh_heatmap"
      });
    } else {
      // Fallback to full refresh if WebSocket is not available
      requestRefresh();
    }
    
    // Set a timeout to clear refreshing state if no response
    setTimeout(() => {
      if (isRefreshing) {
        setIsRefreshing(false);
      }
    }, 5000);
  }, [sendMessage, isConnected, requestRefresh, isRefreshing]);

  return { 
    data, 
    loading: loading || isRefreshing, 
    error, 
    refreshData, 
    refreshHeatmap,
    isConnected
  };
};

export default useHabitHeatmap; 