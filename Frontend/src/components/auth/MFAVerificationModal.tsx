"use client";

import * as React from "react";
import { useState, useRef } from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface MFAVerificationModalProps {
  onVerify: (code: string) => Promise<void>;
  onClose: () => void;
  isLoading: boolean;
}

const MFAVerificationModal: React.FC<MFAVerificationModalProps> = ({
  onVerify,
  onClose,
  isLoading
}) => {
  const [isClosing, setIsClosing] = useState(false);
  const [verificationCode, setVerificationCode] = useState(["", "", "", "", "", ""]);
  const codeInputsRef = useRef<(HTMLInputElement | null)[]>([]);

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(() => onClose(), 150);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const code = verificationCode.join("");
    if (code.length === 6) {
      await onVerify(code);
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

  return (
    <div className={`fixed inset-0 bg-black/50 flex items-center justify-center z-50 ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}>
      <div className={`bg-[#1a1a1a] rounded-lg shadow-xl w-full max-w-[400px] p-6 relative ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}>
        <button
          onClick={handleClose}
          className="absolute right-4 top-4 text-[#a0a0a0] hover:text-[#e5e5e5]"
        >
          <X className="w-5 h-5" />
        </button>

        <h2 className="text-xl font-semibold text-[#e5e5e5] mb-2">Two-Factor Authentication</h2>
        <p className="text-sm text-muted-foreground mb-6">
          Please enter the verification code from your authenticator app.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="verification-code-0">Verification Code</Label>
            <div className="flex justify-between gap-2">
              {verificationCode.map((digit, index) => (
                <Input
                  key={index}
                  id={`verification-code-${index}`}
                  ref={el => codeInputsRef.current[index] = el}
                  className="w-12 h-12 text-center text-xl"
                  maxLength={1}
                  autoFocus={index === 0}
                  value={digit}
                  onChange={(e) => handleCodeChange(index, e.target.value)}
                  onKeyDown={(e) => handleCodeKeyDown(index, e)}
                  onPaste={index === 0 ? handleCodePaste : undefined}
                />
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-3 mt-6">
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button 
              type="submit"
              disabled={verificationCode.some(digit => !digit) || isLoading}
            >
              {isLoading ? 'Verifying...' : 'Verify'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default MFAVerificationModal; 