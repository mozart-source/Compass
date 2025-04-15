"use client";

import * as React from "react";
import { useNavigate } from "react-router-dom";
import {
  Eye,
  EyeOff,
  BarChart2,
  Calendar,
  CheckSquare,
  Focus,
  Globe,
  Sparkles,
  Command,
} from "lucide-react";
import vectorCompass from "@/components/vector-compass.svg";
import { Button } from "@/components/ui/button";
import Checkbox from "../ui/checkbox";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/use-toast";
import InfiniteScroll from "@/components/ui/infinitescrolling";
import TitleBar from "@/components/layout/AuthTitleBar";
import { useTheme } from "@/contexts/theme-provider";
import { useMutation } from "@tanstack/react-query";
import authApi, { LoginCredentials, MFALoginResponse } from "@/api/auth";
import MFAVerificationModal from "./MFAVerificationModal";
import { OAuthButtons } from "./OAuthButtons";

interface UserAuthFormProps extends React.HTMLAttributes<HTMLDivElement> {
  onLogin?: () => void;
}

const FeatureCard = ({
  icon: Icon,
  title,
  description,
}: {
  icon: any;
  title: string;
  description: string;
}) => (
  <div className="p-6 rounded-2xl backdrop-blur-sm border border-border/50 bg-card/50 shadow-lg transition-all duration-300 hover:bg-card/80">
    <div className="flex items-start gap-4">
      <div className="p-2 rounded-lg bg-background/20">
        <Icon className="w-6 h-6" />
      </div>
      <div>
        <h3 className="text-lg font-semibold mb-1">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  </div>
);

export function Login({ className, onLogin, ...props }: UserAuthFormProps) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { theme } = useTheme();
  const isDarkMode = theme === "dark";
  const [isLoading, setIsLoading] = React.useState<boolean>(false);
  const [showPassword, setShowPassword] = React.useState<boolean>(false);
  const [rememberMe, setRememberMe] = React.useState<boolean>(false);
  const [formData, setFormData] = React.useState({
    email: "",
    password: "",
  });
  const [mfaData, setMfaData] = React.useState<MFALoginResponse | null>(null);

  const items = [
    {
      content: (
        <FeatureCard
          icon={Sparkles}
          title="AI-Powered Insights"
          description="Get intelligent recommendations and insights powered by advanced AI"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={BarChart2}
          title="Advanced Analytics"
          description="Track your progress with detailed analytics and visualizations"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={Calendar}
          title="Smart Scheduling"
          description="Optimize your time with AI-driven calendar management"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={CheckSquare}
          title="Task Management"
          description="Stay organized with intelligent task prioritization"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={Focus}
          title="Focus Mode"
          description="Enhance productivity with distraction-free work sessions"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={Globe}
          title="Global Sync"
          description="Access your workspace from anywhere, anytime"
        />
      ),
    },
    {
      content: (
        <FeatureCard
          icon={Command}
          title="Command Center"
          description="Control everything from a powerful command palette"
        />
      ),
    },
  ];

  const loginMutation = useMutation({
    mutationFn: (credentials: LoginCredentials) => authApi.login(credentials),
    onSuccess: (data) => {
      if ('mfa_required' in data) {
        // Handle MFA required
        setMfaData(data as MFALoginResponse);
      } else {
        // Normal login success
        navigate("/dashboard", { replace: true });
        toast({
          title: "Success",
          description: "You have successfully logged in.",
          duration: 1200,
        });
      }
    },
    onError: (error: any) => {
      toast({
        title: "Error",
        description: error.response?.data?.detail || "Login failed",
        variant: "destructive",
        duration: 1200,
      });
    },
  });

  const verifyMFAMutation = useMutation({
    mutationFn: (code: string) => {
      if (!mfaData) throw new Error("MFA data not found");
      return authApi.validateMFA({
        user_id: mfaData.user_id,
        code
      });
    },
    onSuccess: (data) => {
      setMfaData(null);
      navigate("/dashboard", { replace: true });
      toast({
        title: "Success",
        description: "You have successfully logged in.",
        duration: 1200,
      });
    },
    onError: (error: any) => {
      toast({
        title: "Error",
        description: error.response?.data?.detail || "Invalid verification code",
        variant: "destructive",
        duration: 1200,
      });
    },
  });

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    loginMutation.mutate(formData);
  }

  const handleMFAVerify = async (code: string) => {
    verifyMFAMutation.mutate(code);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background relative overflow-hidden">
      {/* Background pattern */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]" />

      <div className="relative container flex gap-12 items-center justify-center p-8">
        {/* Left side - Infinite Scroll */}
        <div className="flex-1 left-8 h-[600px] w-[440px]">
          <InfiniteScroll
            items={items}
            isTilted={true}
            tiltDirection="left"
            autoplay={true}
            autoplaySpeed={0.2}
            autoplayDirection="down"
            pauseOnHover={true}
            width="440px"
          />
        </div>

        {/* Right side - Login Form */}
        <div className="w-[440px] mr-8">
          <TitleBar darkMode={isDarkMode} />
          <Card className="backdrop-blur-sm bg-card/50 transition-all duration-300 hover:shadow-lg">
            <CardHeader className="space-y-1">
              <div className="flex items-center justify-center mb-4">
                <div className="p-2 rounded-xl bg-background/20">
                  <img src={vectorCompass} alt="Logo" className="h-8 w-8" />
                </div>
                <span className="text-2xl font-bold">Compass</span>
              </div>
              <CardTitle className="text-2xl text-center">
                Welcome back
              </CardTitle>
              <CardDescription className="text-center">
                Please enter your details to sign in.
              </CardDescription>
            </CardHeader>
            <form onSubmit={onSubmit}>
              <CardContent className="grid gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="email">Email/Username</Label>
                  <Input
                    id="email"
                    type="text"
                    placeholder="Enter your email or username"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        email: e.target.value,
                      }))
                    }
                    disabled={loginMutation.isPending}
                    className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="password">Password</Label>
                  <div className="relative">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="Enter your password"
                      value={formData.password}
                      onChange={(e) =>
                        setFormData((prev) => ({
                          ...prev,
                          password: e.target.value,
                        }))
                      }
                      disabled={loginMutation.isPending}
                      className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="absolute right-1 top-0 h-full px-3 py-2 hover:bg-transparent"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4 text-muted-foreground" />
                      ) : (
                        <Eye className="h-4 w-4 text-muted-foreground" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="flex items-center justify-between mt-2">
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      name="remember"
                      checked={rememberMe}
                      onChange={(e) => setRememberMe(e.target.checked)}
                      darkMode={isDarkMode}
                    />
                    <label
                      htmlFor="remember"
                      className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                      Remember for 30 days
                    </label>
                  </div>
                  <a
                    href="#"
                    className="text-sm text-muted-foreground hover:text-foreground"
                  >
                    Forgot password?
                  </a>
                </div>
              </CardContent>
              <CardFooter className="flex flex-col gap-4">
                <Button
                  className="w-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                  type="submit"
                  disabled={loginMutation.isPending}
                >
                  {loginMutation.isPending ? (
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
                  ) : null}
                  Sign in
                </Button>

                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-background px-2 text-muted-foreground">
                      Or continue with
                    </span>
                  </div>
                <OAuthButtons className="w-full"/>

                <div className="text-center text-sm">
                  Don't have an account?{" "}
                  <a
                    href="/signup"
                    className="text-primary hover:underline"
                    onClick={(e) => {
                      e.preventDefault();
                      navigate("/signup");
                    }}
                  >
                    Create account
                  </a>
                </div>
              </CardFooter>
            </form>
          </Card>
        </div>
      </div>

      {mfaData && (
        <MFAVerificationModal
          onVerify={handleMFAVerify}
          onClose={() => setMfaData(null)}
          isLoading={verifyMFAMutation.isPending}
        />
      )}
    </div>
  );
}
