import { QueryClient } from '@tanstack/react-query'

// This file creates and configures the central QueryClient instance
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1, // Only retry failed queries once
      staleTime: 5 * 60 * 1000, // Consider data stale after 5 minutes
    },
  },
}) 