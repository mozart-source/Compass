"use client"

import * as React from "react";
import { useState, useRef, useEffect } from "react";
import { Eye, EyeOff, Shield, KeyRound, X } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

interface ProfileEditModalProps {
  onClose: () => void;
}

const ProfileEditModal: React.FC<ProfileEditModalProps> = ({ onClose }) => {
  const { user, updateUser, mfaStatus, setupMFA, verifyMFA, disableMFA, isLoadingUser } = useAuth();
  const [isClosing, setIsClosing] = useState(false);
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [currentPasswordVisible, setCurrentPasswordVisible] = useState(false);
  const [verificationCode, setVerificationCode] = useState(["", "", "", "", "", ""]);
  const codeInputsRef = useRef<(HTMLInputElement | null)[]>([]);
  const [showMfaSetup, setShowMfaSetup] = useState(false);
  const [disablePassword, setDisablePassword] = useState("");
  const [formData, setFormData] = useState({
    first_name: "",
    last_name: "",
    email: "",
    phone_number: "",
    username: "",
    timezone: "UTC",
    locale: "en-US"
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { id, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [id]: value
    }));
  };

  const togglePasswordVisibility = () => setPasswordVisible(!passwordVisible);
  const toggleCurrentPasswordVisibility = () => setCurrentPasswordVisible(!currentPasswordVisible);

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(() => onClose(), 150);
  };

  const handleMfaToggle = async (checked: boolean) => {
    if (checked && !mfaStatus.data?.enabled) {
      try {
        const response = await setupMFA.mutateAsync();
        if (response.qr_code_base64) {
          setShowMfaSetup(true);
        }
      } catch (error) {
        toast.error("Failed to setup MFA");
      }
    } else if (!checked && mfaStatus.data?.enabled) {
      try {
        await disableMFA.mutateAsync(disablePassword);
        toast.success("MFA disabled successfully");
        setShowMfaSetup(false);
      } catch (error) {
        toast.error("Failed to disable MFA");
      }
    }
  };

  const handleVerifyMfa = async () => {
    const code = verificationCode.join("");
    if (code.length === 6) {
      try {
        await verifyMFA.mutateAsync(code);
        toast.success("MFA enabled successfully");
        setShowMfaSetup(false);
      } catch (error) {
        toast.error("Invalid verification code");
      }
    }
  };

  const handleCodeChange = (index: number, value: string) => {
    // Only allow digits
    if (!/^\d*$/.test(value)) return;
    
    const newCode = [...verificationCode];
    newCode[index] = value;
    setVerificationCode(newCode);
    
    // Auto-focus next input if a digit was entered
    if (value && index < 5) {
      codeInputsRef.current[index + 1]?.focus();
    }
  };

  const handleCodeKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    // Handle backspace to go to previous input
    if (e.key === "Backspace" && !verificationCode[index] && index > 0) {
      codeInputsRef.current[index - 1]?.focus();
    }
    
    // Handle left/right arrow keys
    if (e.key === "ArrowLeft" && index > 0) {
      codeInputsRef.current[index - 1]?.focus();
    }
    if (e.key === "ArrowRight" && index < 5) {
      codeInputsRef.current[index + 1]?.focus();
    }
  };

  const handleCodePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData("text");
    const digits = pastedData.replace(/\D/g, "").substring(0, 6).split("");
    
    const newCode = [...verificationCode];
    digits.forEach((digit, index) => {
      if (index < 6) newCode[index] = digit;
    });
    
    setVerificationCode(newCode);
    
    // Focus last input or the next empty one
    if (digits.length < 6) {
      codeInputsRef.current[digits.length]?.focus();
    }
  };

  const onSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    try {
      await updateUser.mutateAsync(formData);
      toast.success("Profile updated successfully");
      handleClose();
    } catch (error) {
      toast.error("Failed to update profile");
    }
  };

  // Update form data when user data changes and is not loading
  React.useEffect(() => {
    if (user && !isLoadingUser) {
      setFormData({
        first_name: user.first_name || "",
        last_name: user.last_name || "",
        email: user.email || "",
        phone_number: user.phone_number || "",
        username: user.username || "",
        timezone: user.timezone || "UTC",
        locale: user.locale || "en-US"
      });
    }
  }, [user, isLoadingUser]);

  // Show loading state
  if (isLoadingUser) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-[#1a1a1a] rounded-lg shadow-xl w-full max-w-[500px] p-6 relative">
          <div className="space-y-4">
            <div className="h-8 w-48 animate-pulse bg-muted rounded" />
            <div className="space-y-2">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="h-10 animate-pulse bg-muted rounded" />
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`fixed inset-0 bg-black/50 flex items-center justify-center z-50 ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}>
      <div className={`bg-[#1a1a1a] rounded-lg shadow-xl w-full max-w-[500px] p-6 relative ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}>
        <button
          onClick={handleClose}
          className="absolute right-4 top-4 text-[#a0a0a0] hover:text-[#e5e5e5]"
        >
          <X className="w-5 h-5" />
        </button>

        <h2 className="text-xl font-semibold text-[#e5e5e5] mb-6">Profile Settings</h2>

        <Tabs defaultValue="profile" className="mt-2">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="profile">Profile</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
          </TabsList>

          <TabsContent value="profile" className="space-y-4 py-4">
            <div className="flex items-start gap-8 mb-8">
              <div className="flex flex-col items-center gap-2 min-w-[120px]">
                <Avatar className="h-24 w-24 border-2 border-border">
                  <AvatarImage src={user?.avatar_url || ""} alt={`${user?.first_name} ${user?.last_name}`} />
                  <AvatarFallback className="text-lg">{user?.first_name?.[0]}{user?.last_name?.[0]}</AvatarFallback>
                </Avatar>
                <span className="text-sm mt-1.5 text-blue-500 cursor-pointer hover:underline w-full text-center">
                  Change Avatar
                </span>
              </div>
              <div className="space-y-4 flex-1">
                <div className="space-y-2">
                  <Label htmlFor="first_name">First Name</Label>
                  <Input
                    id="first_name"
                    value={formData.first_name}
                    onChange={handleInputChange}
                    className="w-full"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="last_name">Last Name</Label>
                  <Input
                    id="last_name"
                    value={formData.last_name}
                    onChange={handleInputChange}
                    className="w-full"
                  />
                </div>
              </div>
            </div>

            <form onSubmit={onSubmit} className="space-y-4">
              <div className="grid gap-4">
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={formData.username}
                    onChange={handleInputChange}
                    className="w-full"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    value={formData.email}
                    onChange={handleInputChange}
                    className="w-full"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="phone_number">Phone Number</Label>
                  <Input
                    id="phone_number"
                    type="tel"
                    value={formData.phone_number}
                    onChange={handleInputChange}
                    className="w-full"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="timezone">Timezone</Label>
                    <Input
                      id="timezone"
                      value={formData.timezone}
                      onChange={handleInputChange}
                      className="w-full"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="locale">Locale</Label>
                    <Input
                      id="locale"
                      value={formData.locale}
                      onChange={handleInputChange}
                      className="w-full"
                    />
                  </div>
                </div>
              </div>
            </form>
          </TabsContent>

          <TabsContent value="security" className="space-y-4 py-4">
            {showMfaSetup ? (
              <div className="space-y-4">
                <div className="flex flex-col items-center space-y-4 p-4 border rounded-lg bg-muted/30">
                  <Shield className="h-12 w-12 text-primary" />
                  <div className="text-center">
                    <h3 className="text-lg font-medium">Set up Two-Factor Authentication</h3>
                    <p className="text-sm text-muted-foreground mt-1">
                      Scan the QR code with your authenticator app and enter the verification code below.
                    </p>
                  </div>
                  <div className="bg-white p-4 rounded-lg">
                    {setupMFA.data?.qr_code_base64 && (
                      <img 
                        src={setupMFA.data.qr_code_base64} 
                        alt="MFA QR Code"
                        className="h-40 w-40"
                      />
                    )}
                  </div>
                  <div className="w-full space-y-2">
                    <Label htmlFor="verification-code">Verification Code</Label>
                    <div className="flex justify-between gap-2">
                      {verificationCode.map((digit, index) => (
                        <Input
                          key={index}
                          ref={el => codeInputsRef.current[index] = el}
                          className="w-12 h-12 text-center text-xl"
                          maxLength={1}
                          value={digit}
                          onChange={(e) => handleCodeChange(index, e.target.value)}
                          onKeyDown={(e) => handleCodeKeyDown(index, e)}
                          onPaste={index === 0 ? handleCodePaste : undefined}
                        />
                      ))}
                    </div>
                  </div>
                  <div className="flex space-x-2 w-full">
                    <Button 
                      variant="outline" 
                      className="flex-1"
                      onClick={() => setShowMfaSetup(false)}
                    >
                      Cancel
                    </Button>
                    <Button 
                      className="flex-1"
                      onClick={handleVerifyMfa}
                      disabled={verificationCode.some(digit => !digit) || verifyMFA.isPending}
                    >
                      {verifyMFA.isPending ? 'Verifying...' : 'Verify'}
                    </Button>
                  </div>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <div className="flex items-center">
                      <KeyRound className="mr-2 h-4 w-4" />
                      <Label className="text-base">Change Password</Label>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Update your password to keep your account secure.
                    </p>
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="space-y-2">
                    <Label htmlFor="current-password">Current Password</Label>
                    <div className="relative">
                      <Input
                        id="current-password"
                        type={currentPasswordVisible ? "text" : "password"}
                        className="pr-10"
                        value={disablePassword}
                        onChange={(e) => setDisablePassword(e.target.value)}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full px-3"
                        onClick={toggleCurrentPasswordVisibility}
                      >
                        {currentPasswordVisible ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="new-password">New Password</Label>
                    <div className="relative">
                      <Input
                        id="new-password"
                        type={passwordVisible ? "text" : "password"}
                        className="pr-10"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full px-3"
                        onClick={togglePasswordVisibility}
                      >
                        {passwordVisible ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>
                </div>

                <Separator className="my-4" />

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <div className="flex items-center">
                      <Shield className="mr-2 h-4 w-4" />
                      <Label className="text-base">Two-Factor Authentication</Label>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      Add an extra layer of security to your account.
                    </p>
                  </div>
                  <Switch
                    checked={mfaStatus.data?.enabled || false}
                    onCheckedChange={handleMfaToggle}
                    disabled={setupMFA.isPending || disableMFA.isPending}
                  />
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>

        <div className="flex justify-end gap-3 mt-6">
          <Button type="button" variant="outline" onClick={handleClose}>
            Cancel
          </Button>
          <Button 
            onClick={onSubmit}
            disabled={updateUser.isPending}
          >
            {updateUser.isPending ? 'Saving...' : 'Save changes'}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ProfileEditModal;