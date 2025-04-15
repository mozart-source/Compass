import React, { useState } from "react";
import ChatbotIcon from "@/components/chatbot/ChatbotIcon";
import ChatWindow from "@/components/chatbot/ChatWindow";
import { Position } from "@/components/chatbot/types";
import { useChat } from "@/components/chatbot/chat-context";

const Chatbot: React.FC = () => {
  const { messages } = useChat();
  const [inputText, setInputText] = useState("");
  const [isChatOpen, setIsChatOpen] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const [isOpening, setIsOpening] = useState(false);
  const [isFullPage, setIsFullPage] = useState(false);
  const [position, setPosition] = useState<Position>({ x: 0, y: 0 });

  const toggleChat = () => {
    if (isChatOpen) {
      // Start closing animation
      setIsClosing(true);
      // Wait for animation to complete before removing from DOM
      setTimeout(() => {
        setIsChatOpen(false);
        setIsClosing(false);
      }, 300); // Match this with the CSS transition duration
    } else {
      // Start opening animation
      setIsOpening(true);
      setIsChatOpen(true);
      // Reset opening state after animation completes
      setTimeout(() => {
        setIsOpening(false);
      }, 300);
    }
  };

  // This function is kept for compatibility but will be replaced by navigation to AI Assistant
  const toggleFullPage = () => {
    setIsFullPage(!isFullPage);
  };

  return (
    <>
      <ChatbotIcon toggleChat={toggleChat} isChatOpen={isChatOpen} />
      {(isChatOpen || isClosing) && (
        <ChatWindow
          inputText={inputText}
          setInputText={setInputText}
          toggleChat={toggleChat}
          isFullPage={isFullPage}
          toggleFullPage={toggleFullPage}
          position={position}
          setPosition={setPosition}
          isClosing={isClosing}
          isOpening={isOpening}
          onClose={toggleChat}
        />
      )}
    </>
  );
};

export default Chatbot;
