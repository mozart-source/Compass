import { useEffect, useState, useCallback } from "react";
import { useDashboardMetrics } from "@/components/dashboard/useDashboardMetrics";
import { useWebSocket } from "@/contexts/websocket-provider";
import { getApiUrls } from "@/config";

export interface FocusSettings {
  daily_target_seconds: number;
  weekly_target_seconds?: number;
  streak_target_days?: number;
}

export interface FocusData {
  total_focus_seconds: number;
  streak: number;
  longest_streak: number;
  sessions: number;
  daily_target_seconds: number;
  daily_breakdown: Array<{ day: string; minutes: number }>;
}

export const useFocusSettings = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<FocusData | null>(null);
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

  // Listen for WebSocket focus-specific events
  useEffect(() => {
    // Set up an event listener for custom focus events from the WebSocket provider
    const handleFocusUpdate = (e: CustomEvent) => {
      if (e.detail?.type === 'focus_data' && e.detail?.data?.focus) {
        // Direct update from WebSocket with focus data
        setData(e.detail.data.focus as unknown as FocusData);
        setIsRefreshing(false);
      } else if (e.detail?.type === 'focus_stats' && e.detail?.data) {
        // Direct update from stats event
        const focusStats = e.detail.data as unknown as FocusData;
        // Only update if we have valid stats data
        if (focusStats.total_focus_seconds !== undefined) {
          setData(focusStats);
          setIsRefreshing(false);
        }
      } else if (['focus_session_started', 'focus_session_stopped'].includes(e.detail?.type)) {
        // Session state changed - refresh and also set a loading state
        setIsRefreshing(true);
        
        // This is just to show immediate feedback in the UI 
        // that something has changed, while we wait for the data refresh
        if (data && e.detail?.type === 'focus_session_started') {
          // Optimistically update the UI to show an active session
          setData({
            ...data,
            sessions: data.sessions + 1
          });
        }
        
        // Request a real refresh via WebSocket
        if (sendMessage && isConnected) {
          sendMessage({
            type: "refresh_focus"
          });
        } else {
          // Fallback to full refresh if WebSocket is not available
          requestRefresh();
        }
      }
    };

    // Add event listener for websocket focus events
    window.addEventListener('websocket_focus_event', handleFocusUpdate as EventListener);

    // Cleanup
    return () => {
      window.removeEventListener('websocket_focus_event', handleFocusUpdate as EventListener);
    };
  }, [sendMessage, isConnected, requestRefresh, data]);

  useEffect(() => {
    setLoading(metricsLoading);
    
    if (metricsError && metricsErrorData) {
      setError(metricsErrorData instanceof Error ? metricsErrorData : new Error("Failed to load focus data"));
    } else {
      setError(null);
    }
    
    if (metricsData?.focus) {
      setData(metricsData.focus as unknown as FocusData);
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
  
  // Function to refresh only focus data via WebSocket
  const refreshFocus = useCallback(() => {
    setIsRefreshing(true);
    
    if (sendMessage && isConnected) {
      sendMessage({
        type: "refresh_focus"
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

  // Function to update focus settings
  const updateSettings = useCallback(async (settings: Partial<FocusSettings>) => {
    setIsRefreshing(true);
    try {
      const token = localStorage.getItem("token");
      
      if (!token) {
        throw new Error("No authentication token found");
      }
      
      const { PYTHON_API_URL } = getApiUrls();
      const response = await fetch(`${PYTHON_API_URL}/focus/settings`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify(settings),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to update settings: ${response.statusText}`);
      }
      
      // Refresh data after successful update
      refreshFocus();
      return await response.json();
    } catch (error) {
      console.error("Error updating focus settings:", error);
      setIsRefreshing(false);
      throw error;
    }
  }, [refreshFocus]);

  return { 
    data, 
    loading: loading || isRefreshing, 
    error, 
    refreshData, 
    refreshFocus,
    updateSettings,
    isConnected
  };
};

export default useFocusSettings; 