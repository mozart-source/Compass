import axios from 'axios';
import { getApiUrls } from '@/config';

// Get API URLs using the configuration system
const { GO_API_URL } = getApiUrls();

// Types
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  email: string;
  username: string;
  password: string;
  first_name: string;
  last_name: string;
  phone_number?: string;
  timezone?: string;
  locale?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  session: SessionResponse;
  expires_at: string;
}

export interface User {
  id: string;
  email: string;
  username: string;
  is_active: boolean;
  is_superuser: boolean;
  created_at: string;
  updated_at: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  avatar_url?: string;
  bio?: string;
  timezone: string;
  locale: string;
  mfa_enabled: boolean;
  failed_login_attempts: number;
  force_password_change: boolean;
  max_sessions: number;
  deleted_at?: string;
  last_login?: string;
  account_locked_until?: string;
  organization_id?: string;
  notification_preferences?: Record<string, any>;
  workspace_settings?: Record<string, any>;
  allowed_ip_ranges?: string[];
}

export interface SessionResponse {
  id: string;
  device_info: string;
  ip_address: string;
  last_activity: string;
  expires_at: string;
}

// MFA Types
export interface MFASetupResponse {
  secret: string;
  qr_code_base64: string;
  otp_auth_url: string;
  backup_codes?: string[];
}

export interface MFAStatusResponse {
  enabled: boolean;
}

export interface VerifyMFARequest {
  code: string;
}

export interface DisableMFARequest {
  password: string;
}

export interface MFALoginResponse {
  mfa_required: boolean;
  user_id: string;
  message: string;
  ttl: number;
}

export interface ValidateMFARequest {
  user_id: string;
  code: string;
}

// API client
const authApi = {
  register: async (credentials: RegisterCredentials): Promise<User> => {
    const response = await axios.post(`${GO_API_URL}/users/register`, credentials, {
      headers: {
        'Content-Type': 'application/json',
      },
    });

    return response.data.user;
  },

  login: async (credentials: LoginCredentials): Promise<AuthResponse | MFALoginResponse> => {
    const response = await axios.post(`${GO_API_URL}/users/login`, {
      email: credentials.email,
      password: credentials.password
    }, {
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const data = response.data;
    
    // If MFA is required, return the MFA response
    if (data.mfa_required) {
      return data;
    }

    // Otherwise, handle normal login
    const token = data.token || data.access_token;
    
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      localStorage.setItem('token', token);
    } else {
      console.error('No token received in auth response');
    }

    return data;
  },

  validateMFA: async (request: ValidateMFARequest): Promise<AuthResponse> => {
    const response = await axios.post(`${GO_API_URL}/auth/mfa/validate`, request);
    const data = response.data;
    
    const token = data.token || data.access_token;
    
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      localStorage.setItem('token', token);
    }

    return data;
  },

  getMe: async (): Promise<User> => {
    const response = await axios.get(`${GO_API_URL}/users/profile`);
    return response.data.user;
  },

  logout: async (): Promise<void> => {
    try {
      await axios.post(`${GO_API_URL}/users/logout`, null);
      localStorage.removeItem('token');
      delete axios.defaults.headers.common['Authorization'];
    } catch (error) {
      console.error('Logout error:', error);
      // Still clear token even if logout request fails
      localStorage.removeItem('token');
      delete axios.defaults.headers.common['Authorization'];
    }
  },

  updateUser: async (userData: {
    first_name?: string;
    last_name?: string;
    email?: string;
  }): Promise<User> => {
    const response = await axios.put(`${GO_API_URL}/users/profile`, userData);
    return response.data.user;
  },

  // MFA Methods
  setupMFA: async (): Promise<MFASetupResponse> => {
    const response = await axios.post(`${GO_API_URL}/users/mfa/setup`);
    return response.data;
  },

  verifyMFA: async (code: string): Promise<void> => {
    await axios.post(`${GO_API_URL}/users/mfa/verify`, { code });
  },

  disableMFA: async (password: string): Promise<void> => {
    await axios.post(`${GO_API_URL}/users/mfa/disable`, { password });
  },

  getMFAStatus: async (): Promise<MFAStatusResponse> => {
    const response = await axios.get(`${GO_API_URL}/users/mfa/status`);
    return response.data;
  },
};

export default authApi; 