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
  Command
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
import authApi, { RegisterCredentials } from "@/api/auth";
import { OAuthButtons } from './OAuthButtons';

interface UserAuthFormProps extends React.HTMLAttributes<HTMLDivElement> {
  onSignup?: () => void;
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

export function Signup({ className, onSignup, ...props }: UserAuthFormProps) {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { theme } = useTheme();
  const isDarkMode = theme === "dark";
  const [showPassword, setShowPassword] = React.useState<boolean>(false);
  const [showConfirmPassword, setShowConfirmPassword] = React.useState<boolean>(false);
  const [agreeToTerms, setAgreeToTerms] = React.useState<boolean>(false);
  const [formData, setFormData] = React.useState({
    email: "",
    username: "",
    password: "",
    confirmPassword: "",
    first_name: "",
    last_name: "",
    phone_number: "",
  });
  const [formErrors, setFormErrors] = React.useState<Record<string, string>>({});

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

  const registerMutation = useMutation({
    mutationFn: (credentials: RegisterCredentials) => authApi.register(credentials),
    onSuccess: (data) => {
      toast({
        title: "Success",
        description: "Your account has been created. Please log in.",
        duration: 1200,
      });
      navigate("/login", { replace: true });
    },
    onError: (error: any) => {
      toast({
        title: "Error",
        description: error.response?.data?.detail || "Registration failed",
        variant: "destructive",
        duration: 1200,
      });
    },
  });

  const validateForm = () => {
    const errors: Record<string, string> = {};
    
    if (!formData.email) {
      errors.email = "Email is required";
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      errors.email = "Email is invalid";
    }
    
    if (!formData.username) {
      errors.username = "Username is required";
    }
    
    if (!formData.first_name) {
      errors.first_name = "First name is required";
    }
    
    if (!formData.last_name) {
      errors.last_name = "Last name is required";
    }
    
    if (!formData.password) {
      errors.password = "Password is required";
    } else if (formData.password.length < 8) {
      errors.password = "Password must be at least 8 characters";
    }
    
    if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = "Passwords do not match";
    }
    
    if (!agreeToTerms) {
      errors.terms = "You must agree to the terms and conditions";
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    const credentials: RegisterCredentials = {
      email: formData.email,
      username: formData.username,
      password: formData.password,
      first_name: formData.first_name,
      last_name: formData.last_name,
      phone_number: formData.phone_number,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      locale: navigator.language,
    };
    
    registerMutation.mutate(credentials);
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background relative overflow-hidden">
      {/* Background pattern */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]" />

      <div className="relative container flex items-center justify-center p-8">
        {/* Right side - Signup Form */}
        <div className="w-[800px]">
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
                Create an account
              </CardTitle>
              <CardDescription className="text-center">
                Enter your details to get started with Compass
              </CardDescription>
            </CardHeader>
            <form onSubmit={onSubmit}>
              <CardContent className="grid gap-4">
                <OAuthButtons/>
                
                <div className="relative mb-4 ">
                  <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-background px-2 text-muted-foreground">
                      Or continue with email
                    </span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="first_name">First Name</Label>
                        <Input
                          id="first_name"
                          type="text"
                          placeholder="e.g. Sarah"
                          value={formData.first_name}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              first_name: e.target.value,
                            }))
                          }
                          disabled={registerMutation.isPending}
                          className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                        />
                        {formErrors.first_name && (
                          <p className="text-xs text-destructive">{formErrors.first_name}</p>
                        )}
                      </div>
                      <div className="grid gap-2">
                        <Label htmlFor="last_name">Last Name</Label>
                        <Input
                          id="last_name"
                          type="text"
                          placeholder="e.g. Connor"
                          value={formData.last_name}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              last_name: e.target.value,
                            }))
                          }
                          disabled={registerMutation.isPending}
                          className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                        />
                        {formErrors.last_name && (
                          <p className="text-xs text-destructive">{formErrors.last_name}</p>
                        )}
                      </div>
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="username">Username</Label>
                      <Input
                        id="username"
                        type="text"
                        placeholder="e.g. sarah.c"
                        value={formData.username}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            username: e.target.value,
                          }))
                        }
                        disabled={registerMutation.isPending}
                        className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                      />
                      {formErrors.username && (
                        <p className="text-xs text-destructive">{formErrors.username}</p>
                      )}
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="password">Password</Label>
                      <div className="relative">
                        <Input
                          id="password"
                          type={showPassword ? "text" : "password"}
                          placeholder="Create a password"
                          value={formData.password}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              password: e.target.value,
                            }))
                          }
                          disabled={registerMutation.isPending}
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
                      {formErrors.password && (
                        <p className="text-xs text-destructive">{formErrors.password}</p>
                      )}
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="grid gap-2">
                      <Label htmlFor="phone_number">Phone Number (Optional)</Label>
                      <Input
                        id="phone_number"
                        type="tel"
                        placeholder="e.g. +20 123 456 7891"
                        value={formData.phone_number}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            phone_number: e.target.value,
                          }))
                        }
                        disabled={registerMutation.isPending}
                        className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder="you@compass.com"
                        value={formData.email}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            email: e.target.value,
                          }))
                        }
                        disabled={registerMutation.isPending}
                        className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                      />
                      {formErrors.email && (
                        <p className="text-xs text-destructive">{formErrors.email}</p>
                      )}
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="confirmPassword">Confirm Password</Label>
                      <div className="relative">
                        <Input
                          id="confirmPassword"
                          type={showConfirmPassword ? "text" : "password"}
                          placeholder="Confirm your password"
                          value={formData.confirmPassword}
                          onChange={(e) =>
                            setFormData((prev) => ({
                              ...prev,
                              confirmPassword: e.target.value,
                            }))
                          }
                          disabled={registerMutation.isPending}
                          className="focus:outline-none focus:ring-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="absolute right-1 top-0 h-full px-3 py-2 hover:bg-transparent"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                        >
                          {showConfirmPassword ? (
                            <EyeOff className="h-4 w-4 text-muted-foreground" />
                          ) : (
                            <Eye className="h-4 w-4 text-muted-foreground" />
                          )}
                        </Button>
                      </div>
                      {formErrors.confirmPassword && (
                        <p className="text-xs text-destructive">{formErrors.confirmPassword}</p>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-2 mt-2">
                  <Checkbox
                    name="terms"
                    checked={agreeToTerms}
                    onChange={(e) => setAgreeToTerms(e.target.checked)}
                    darkMode={isDarkMode}
                  />
                  <label
                    htmlFor="terms"
                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                  >
                    I agree to the <a href="#" className="text-primary hover:underline">Terms of Service</a> and <a href="#" className="text-primary hover:underline">Privacy Policy</a>
                  </label>
                </div>
                {formErrors.terms && (
                  <p className="text-xs text-destructive">{formErrors.terms}</p>
                )}
              </CardContent>
              <CardFooter className="flex flex-col gap-4">
                <Button
                  className="w-full bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                  type="submit"
                  disabled={registerMutation.isPending}
                >
                  {registerMutation.isPending ? (
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
                  ) : null}
                  Create Account
                </Button>
                <div className="text-center text-sm">
                  Already have an account?{" "}
                  <a
                    href="/login"
                    className="text-primary hover:underline"
                    onClick={(e) => {
                      e.preventDefault();
                      navigate("/login");
                    }}
                  >
                    Sign in
                  </a>
                </div>
              </CardFooter>
            </form>
          </Card>
        </div>
      </div>
    </div>
  );
} 