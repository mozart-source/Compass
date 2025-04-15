import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import React from 'react';
import authApi, { User, LoginCredentials, AuthResponse, MFASetupResponse, MFAStatusResponse } from '@/api/auth';
import axios from 'axios';

// Re-export types for convenience
export type { User, LoginCredentials, AuthResponse, MFASetupResponse, MFAStatusResponse };

// This is a custom hook that combines React Query with auth functionality
export function useAuth() {
  const queryClient = useQueryClient();
  const [token, setToken] = React.useState<string | null>(localStorage.getItem('token'));

  // Set up axios defaults when token changes
  React.useEffect(() => {
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } else {
      delete axios.defaults.headers.common['Authorization'];
    }
  }, [token]);

  // Query for current user
  const { data: user, isLoading: isLoadingUser } = useQuery({
    queryKey: ['user'],
    queryFn: authApi.getMe,
    enabled: !!token,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * (2 ** attemptIndex), 30000),
    staleTime: 1000 * 60 * 60 * 24,
    gcTime: 1000 * 60 * 60 * 24 * 7,
  });

  // Clear token and queries when component unmounts or token becomes invalid
  React.useEffect(() => {
    const cleanup = () => {
      localStorage.removeItem('token');
      setToken(null);
      queryClient.clear();
    };

    const subscription = queryClient.getQueryCache().subscribe((event) => {
      const error = event?.query?.state?.error as AxiosError<{ message: string }>;
      if (error?.response?.status === 401 && error?.response?.data?.message !== 'Network Error') {
        cleanup();
      }
    });

    // Also watch for token removal
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'token' && !e.newValue) {
        cleanup();
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => {
      subscription();
      window.removeEventListener('storage', handleStorageChange);
    };
  }, [queryClient]);

  // Login mutation
  const login = useMutation({
    mutationFn: (credentials: LoginCredentials) => authApi.login(credentials),
    onSuccess: async (data) => {
      if ('token' in data && data.token) {
        localStorage.setItem('token', data.token);
        setToken(data.token);
        await queryClient.invalidateQueries({ queryKey: ['user'] });
      }
    },
  });

  // Logout mutation
  const logout = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      localStorage.removeItem('token');
      setToken(null);
      queryClient.clear();
    },
  });

  const updateUser = useMutation({
    mutationFn: authApi.updateUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
  });

  // MFA Queries and Mutations
  const mfaStatus = useQuery({
    queryKey: ['mfa-status'],
    queryFn: authApi.getMFAStatus,
    enabled: !!token,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * (2 ** attemptIndex), 30000),
    staleTime: 1000 * 60 * 60 * 24,
    gcTime: 1000 * 60 * 60 * 24 * 7,
  });

  const setupMFA = useMutation({
    mutationFn: authApi.setupMFA,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] });
    },
  });

  const verifyMFA = useMutation({
    mutationFn: authApi.verifyMFA,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] });
    },
  });

  const disableMFA = useMutation({
    mutationFn: authApi.disableMFA,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mfa-status'] });
    },
  });

  return {
    user,
    login,
    logout,
    updateUser,
    isAuthenticated: !!token,
    isLoadingUser,
    queryClient,
    mfaStatus,
    setupMFA,
    verifyMFA,
    disableMFA,
  };
} 