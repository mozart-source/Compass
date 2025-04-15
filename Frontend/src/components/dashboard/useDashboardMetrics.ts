import { useQuery } from "@tanstack/react-query";
import { useWebSocket, QUERY_KEYS, DashboardMetrics } from "@/contexts/websocket-provider";
import { getApiUrls } from "@/config";

const fetchDashboardMetrics = async (): Promise<DashboardMetrics> => {
  const token = localStorage.getItem("token");
  
  if (!token) {
    throw new Error("No authentication token found");
  }

  const { PYTHON_API_URL } = getApiUrls();

  try {
    const response = await fetch(`${PYTHON_API_URL}/dashboard/metrics`, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  } catch (error) {
    console.error("Error fetching dashboard metrics:", error);
    throw error;
  }
};

export const useDashboardMetrics = () => {
  const { requestRefresh, isConnected } = useWebSocket();

  const query = useQuery<DashboardMetrics, Error>({
    queryKey: QUERY_KEYS.DASHBOARD_METRICS,
    queryFn: fetchDashboardMetrics,
    staleTime: 30000,  // Consider data fresh for 30 seconds
    gcTime: 300000,    // Keep cached data for 5 minutes
    refetchOnWindowFocus: false, // Disable refetch on window focus since we have WebSocket
    refetchOnMount: true,
    refetchInterval: false, 
    refetchIntervalInBackground: false,
    networkMode: "always",
    retry: (failureCount, error) => {
      // Don't retry auth errors
      if (error.message.includes("401") || error.message.includes("403")) {
        return false;
      }
      
      // Retry other errors max 2 times
      return failureCount < 2;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 5000), // Exponential backoff with 5s cap
    
    // Enable refetch on reconnect only when WebSocket isn't connected
    refetchOnReconnect: !isConnected,
  });

  return {
    ...query,
    requestRefresh, // Expose WebSocket refresh function
    isConnected,    // Expose connection status
  };
};
