import axios from 'axios';
import { getApiUrls } from '@/config';

// Get API URLs using the configuration system
const { GO_API_URL } = getApiUrls();

export interface OAuthProvider {
  name: string;
  display_name: string;
  scopes: string[];
}

export interface OAuthLoginResponse {
  auth_url: string;
  state: string;
}

export interface OAuthCallbackResponse {
  token: string;
  expires_at: number;
  user: {
    id: string;
    email: string;
    username: string;
    first_name: string;
    last_name: string;
    phone_number: string;
    avatar_url: string;
    bio: string;
    timezone: string;
    locale: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
  };
}

export const oauthService = {
  // Get available OAuth providers
  getProviders: async (): Promise<OAuthProvider[]> => {
    const response = await axios.get(`${GO_API_URL}/auth/oauth/providers`);
    return response.data.providers;
  },

  // Initiate OAuth login
  initiateLogin: async (provider: string): Promise<OAuthLoginResponse> => {
    const response = await axios.post(`${GO_API_URL}/auth/oauth/login`, {
      provider
    });
    return response.data;
  },

  // Handle OAuth callback
  handleCallback: async (provider: string, code: string, state: string): Promise<OAuthCallbackResponse> => {
    const response = await axios.post(`${GO_API_URL}/auth/oauth/callback`, {
      provider,
      code,
      state
    });
    return response.data;
  }
}; 