import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { oauthService, OAuthProvider } from '@/services/auth/oauth';
import { Github, Mail, Loader2 } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

interface OAuthButtonsProps {
  className?: string;
}

export function OAuthButtons({ className }: OAuthButtonsProps) {
  const [providers, setProviders] = useState<OAuthProvider[]>([]);
  const [loadingProvider, setLoadingProvider] = useState<string | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    const loadProviders = async () => {
      try {
        const availableProviders = await oauthService.getProviders();
        setProviders(availableProviders);
      } catch (error) {
        console.error('Failed to load OAuth providers:', error);
        toast({
          title: "Error",
          description: "Failed to load authentication providers",
          variant: "destructive",
        });
      }
    };

    loadProviders();
  }, [toast]);

  const handleOAuthLogin = async (provider: string) => {
    try {
      setLoadingProvider(provider);
      const { auth_url } = await oauthService.initiateLogin(provider);
      window.location.href = auth_url;
    } catch (error) {
      console.error(`Failed to initiate ${provider} login:`, error);
      toast({
        title: "Error",
        description: `Failed to initiate ${provider} login. Please try again.`,
        variant: "destructive",
      });
    } finally {
      setLoadingProvider(null);
    }
  };

  const getProviderIcon = (provider: string) => {
    switch (provider.toLowerCase()) {
      case 'github':
        return <Github className="mr-2 h-4 w-4" />;
      case 'google':
        return <Mail className="mr-2 h-4 w-4" />;
      default:
        return null;
    }
  };

  return (
    <div className={className}>
      {providers.map((provider) => (
        <Button
          key={provider.name}
          variant="outline"
          type="button"
          disabled={loadingProvider !== null}
          className="w-full mb-2"
          onClick={() => handleOAuthLogin(provider.name)}
        >
          {loadingProvider === provider.name ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            getProviderIcon(provider.name)
          )}
          {loadingProvider === provider.name ? 'Connecting...' : `Continue with ${provider.display_name}`}
        </Button>
      ))}
    </div>
  );
} 