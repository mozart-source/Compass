import React, { useRef, useEffect, useState } from "react";
import { Position } from "@/components/chatbot/types";
import { useTheme } from "@/contexts/theme-provider";
import { useNavigate } from "react-router-dom";
import { useChat } from "@/components/chatbot/chat-context";

interface ChatWindowProps {
  inputText: string;
  setInputText: React.Dispatch<React.SetStateAction<string>>;
  toggleChat: () => void;
  isFullPage: boolean;
  toggleFullPage: () => void;
  position: Position;
  setPosition: React.Dispatch<React.SetStateAction<Position>>;
  isClosing: boolean;
  isOpening: boolean;
  onClose: () => void;
}

const ChatWindow: React.FC<ChatWindowProps> = ({
  inputText,
  setInputText,
  toggleChat,
  isFullPage,
  toggleFullPage,
  position,
  setPosition,
  isClosing,
  isOpening,
  onClose,
}) => {
  const chatWindowRef = useRef<HTMLDivElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { theme } = useTheme();
  const isDarkTheme = theme === "dark";
  const [hasInitializedAnimation, setHasInitializedAnimation] = useState(false);
  const navigate = useNavigate();

  // Using shared chat context
  const { messages, isLoading, streamingText, sendMessage } = useChat();

  // Initialize animation state
  useEffect(() => {
    // Set a very short timeout to trigger the animation
    if (!hasInitializedAnimation) {
      // Force an initial state for the animation
      setTimeout(() => {
        setHasInitializedAnimation(true);
      }, 10);
    }
  }, [hasInitializedAnimation]);

  // Scroll to bottom of messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, streamingText]);

  // Simple drag functionality
  const [isDragging, setIsDragging] = React.useState(false);
  const [dragStart, setDragStart] = React.useState({ x: 0, y: 0 });

  const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isFullPage) return;
    setIsDragging(true);
    setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!isDragging || !chatWindowRef.current) return;

    // Calculate new position
    const newX = e.clientX - dragStart.x;
    const newY = e.clientY - dragStart.y;

    setPosition({
      x: newX,
      y: newY,
    });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  useEffect(() => {
    if (isDragging) {
      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
    }

    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDragging, dragStart]);

  // Determine the current animation state
  const getAnimationStyles = () => {
    if (isClosing) {
      return {
        opacity: 0,
        scale: 0.95,
      };
    } else if (isOpening || !hasInitializedAnimation) {
      return {
        opacity: hasInitializedAnimation ? 1 : 0,
        scale: hasInitializedAnimation ? 1 : 0.95,
      };
    } else {
      return {
        opacity: 1,
        scale: 1,
      };
    }
  };

  const { opacity, scale } = getAnimationStyles();

  // Handle navigation to AI Assistant page
  const handleExpandToAIAssistant = () => {
    // Close the chat window
    onClose();
    // Navigate to the AI Assistant page
    navigate("/ai");
  };

  // Handle sending a message
  const handleSendMessage = async () => {
    if (!inputText.trim() || isLoading) return;

    const messageToSend = inputText.trim();
    setInputText(""); // Clear input immediately for better UX

    await sendMessage(messageToSend);
  };

  return (
    <div
      ref={chatWindowRef}
      style={{
        position: "fixed",
        top: isFullPage ? 0 : undefined,
        left: isFullPage ? 0 : undefined,
        right: isFullPage ? 0 : undefined,
        bottom: isFullPage ? 0 : undefined,
        width: isFullPage ? "100%" : "350px",
        height: isFullPage ? "100%" : "500px",
        transform: isFullPage
          ? "none"
          : `translate(${position.x}px, ${position.y}px) scale(${scale})`,
        borderRadius: isFullPage ? "0" : "12px",
        overflow: "hidden",
        zIndex: 60,
        boxShadow:
          "0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)",
        transition: isDragging
          ? "width 0.2s ease, height 0.2s ease, border-radius 0.2s ease, opacity 0.3s ease"
          : "width 0.2s ease, height 0.2s ease, border-radius 0.2s ease, opacity 0.3s ease, transform 0.3s ease",
        opacity: opacity,
        willChange: "transform",
        touchAction: "none",
        userSelect: "none",
      }}
      className={`${!isFullPage && "bottom-6 right-6"} ${
        isDarkTheme ? "bg-[#202020] text-[#e5e5e5]" : "bg-white text-gray-800"
      } flex flex-col border-b-[2px] border-[#E7E7E7]`}
    >
      {/* Chat header */}
      <div
        className={`${
          isDarkTheme ? "bg-[#202020]" : "bg-white"
        } shadow-sm p-4 flex justify-between items-center`}
        onMouseDown={handleMouseDown}
        style={{ cursor: isFullPage ? "default" : "move" }}
      >
        <h1
          className={`text-lg font-medium ${
            isDarkTheme ? "text-[#e5e5e5]" : "text-gray-800"
          }`}
        >
          AI Assistant - Iris
        </h1>
        <div className="flex space-x-2">
          <button
            onClick={handleExpandToAIAssistant}
            className={`p-1.5 rounded-md ${
              isDarkTheme ? "hover:bg-[#3b3b3b]" : "hover:bg-gray-100"
            } transition-colors`}
            aria-label="Open full AI Assistant"
            tabIndex={0}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4 text-[#E7E7E7]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
              />
            </svg>
          </button>
          <button
            onClick={toggleChat}
            className={`p-1.5 rounded-md ${
              isDarkTheme ? "hover:bg-[#3b3b3b]" : "hover:bg-gray-100"
            } transition-colors`}
            aria-label="Close chat"
            tabIndex={0}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4 text-[#E7E7E7]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      </div>

      {/* Messages container */}
      <div
        className={`flex-1 overflow-y-auto p-4 space-y-3 ${
          isDarkTheme ? "bg-[#202020]" : "bg-white"
        }`}
      >
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${
              message.sender === "user" ? "justify-end" : "justify-start"
            }`}
          >
            <div
              className={`max-w-[70%] rounded-lg p-2.5 ${
                message.sender === "user"
                  ? "bg-white-500 text-white"
                  : isDarkTheme
                  ? "bg-[#3b3b3b] text-[#e5e5e5]"
                  : "bg-gray-100 text-gray-800"
              }`}
            >
              <p className="text-sm">{message.text}</p>
              <span className="text-xs opacity-75 mt-1 inline-block">
                {message.timestamp.toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
          </div>
        ))}

        {/* Streaming message */}
        {streamingText && !messages.some((m) => m.text === streamingText) && (
          <div className="flex justify-start">
            <div
              className={`max-w-[70%] rounded-lg p-2.5 ${
                isDarkTheme
                  ? "bg-[#3b3b3b] text-[#e5e5e5]"
                  : "bg-gray-100 text-gray-800"
              }`}
            >
              <p className="text-sm">{streamingText}</p>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input area */}
      <div
        className={`${
          isDarkTheme ? "bg-[#202020]" : "bg-white"
        } p-3 shadow-sm border-t ${
          isDarkTheme ? "border-[#3b3b3b]" : "border-gray-200"
        }`}
      >
        <div className="flex space-x-2 rounded-[15px]">
          <input
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            onKeyPress={(e) =>
              e.key === "Enter" && !isLoading && handleSendMessage()
            }
            placeholder={
              isLoading ? "AI is thinking..." : "Type your message..."
            }
            disabled={isLoading}
            className={`flex-1 p-2 text-sm border ${
              isDarkTheme
                ? "bg-[#262626] border-[#3b3b3b] text-white placeholder-gray-500 focus:border-white-400"
                : "bg-white border-gray-300 text-gray-800 focus:border-white-500"
            } rounded-[15px] focus:outline-none ${
              isLoading ? "opacity-50" : ""
            }`}
            aria-label="Message input"
            tabIndex={0}
          />
          <button
            onClick={handleSendMessage}
            disabled={isLoading}
            className={`px-3 py-2 bg-white-500 text-white text-sm rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-white-300 ${
              isLoading ? "opacity-50 cursor-not-allowed" : "hover:bg-white-600"
            }`}
            aria-label="Send message"
            tabIndex={0}
          >
            {isLoading ? (
              <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
            ) : (
              "Send"
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChatWindow;
