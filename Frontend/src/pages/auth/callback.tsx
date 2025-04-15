import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useToast } from '@/components/ui/use-toast';
import { oauthService } from '@/services/auth/oauth';
import { Loader2 } from 'lucide-react';
import { AxiosError } from 'axios';

export default function OAuthCallback() {
  const navigate = useNavigate();
  const location = useLocation();
  const { toast } = useToast();

  useEffect(() => {
    const handleCallback = async () => {
      // Parse query parameters from the URL
      const params = new URLSearchParams(location.search);
      const code = params.get('code');
      const state = params.get('state');
      const provider = params.get('provider');

      if (!code || !state || !provider) {
        toast({
          title: "Error",
          description: "Invalid callback parameters",
          variant: "destructive",
        });
        navigate('/login');
        return;
      }

      try {
        const response = await oauthService.handleCallback(provider, code, state);
        
        // Store the token and user data
        localStorage.setItem('token', response.token);
        localStorage.setItem('expires_at', response.expires_at.toString());
        localStorage.setItem('user', JSON.stringify({
          ...response.user,
          firstName: response.user.first_name,
          lastName: response.user.last_name,
          phoneNumber: response.user.phone_number,
          avatarUrl: response.user.avatar_url,
          isActive: response.user.is_active,
          createdAt: response.user.created_at,
          updatedAt: response.user.updated_at
        }));
        
        // Show success message
        toast({
          title: "Success",
          description: "Successfully logged in",
        });

        // Redirect to dashboard
        navigate('/dashboard');
      } catch (error) {
        const axiosError = error as AxiosError<{ detail: string }>;
        console.error('OAuth callback error:', error);
        toast({
          title: "Error",
          description: axiosError.response?.data?.detail || "Failed to complete authentication",
          variant: "destructive",
        });
        navigate('/login');
      }
    };

    handleCallback();
  }, [navigate, location, toast]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
        <h2 className="text-2xl font-semibold mb-2">Completing login...</h2>
        <p className="text-muted-foreground">Please wait while we verify your credentials</p>
      </div>
    </div>
  );
} 