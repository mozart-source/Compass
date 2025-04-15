import React, { useState, useEffect, useRef } from "react";
import { useTheme } from "@/contexts/theme-provider";
import { useDraggable } from "@dnd-kit/core";
import {
  motion,
  AnimatePresence,
  useMotionValue,
  useSpring,
} from "framer-motion";
import { useDragStore } from "@/dragStore";
import IrisIcon from "@/components/IRIS.svg";
import IrisHat from "@/components/IRISHAT.svg";

interface ChatbotIconProps {
  toggleChat: () => void;
  isChatOpen: boolean;
}

const ChatbotIcon: React.FC<ChatbotIconProps> = ({
  toggleChat,
  isChatOpen,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const [showTooltip, setShowTooltip] = useState(false);
  const { theme } = useTheme();
  const isDarkTheme = theme === "dark";

  const { attachmentPosition, chatbotAttachedTo } = useDragStore();
  const iconContainerRef = useRef<HTMLDivElement | null>(null);

  // Motion values for smooth animation
  const x = useMotionValue(0);
  const y = useMotionValue(0);

  // Spring configuration for smooth movement
  const springConfig = { damping: 20, stiffness: 400, mass: 0.5 };
  const springX = useSpring(x, springConfig);
  const springY = useSpring(y, springConfig);

  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: "chatbot-bubble",
    });

  const combinedRef = (node: HTMLDivElement) => {
    setNodeRef(node);
    iconContainerRef.current = node;
  };

  // Update motion values when transform changes
  useEffect(() => {
    if (transform) {
      x.set(transform.x);
      y.set(transform.y);
    } else if (
      attachmentPosition &&
      chatbotAttachedTo &&
      iconContainerRef.current
    ) {
      const { x: attachX, y: attachY } = attachmentPosition;

      // Temporarily remove transform to get the icon's base position
      const originalTransform = iconContainerRef.current.style.transform;
      iconContainerRef.current.style.transform = "none";
      const iconRect = iconContainerRef.current.getBoundingClientRect();
      iconContainerRef.current.style.transform = originalTransform;

      const newX = attachX - iconRect.left;
      const newY = attachY - iconRect.top;

      x.set(newX);
      y.set(newY);
    } else {
      // Smoothly animate back to original position
      x.set(0);
      y.set(0);
    }
  }, [transform, x, y, attachmentPosition, chatbotAttachedTo]);

  useEffect(() => {
    let timeoutId: NodeJS.Timeout;
    if (isHovered && !isChatOpen && !isDragging) {
      timeoutId = setTimeout(() => setShowTooltip(true), 500);
    } else {
      setShowTooltip(false);
    }
    return () => clearTimeout(timeoutId);
  }, [isHovered, isChatOpen, isDragging]);

  return (
    <motion.div
      className="fixed bottom-6 right-6 z-50"
      ref={combinedRef}
      style={{
        x: springX,
        y: springY,
        cursor: isDragging ? "grabbing" : "grab",
      }}
      data-no-dismiss
    >
      <motion.button
        onClick={!isDragging ? toggleChat : undefined}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
        className={`
          w-12 h-12 bg-[#2a2a2a] text-[#e5e5e5] rounded-full shadow-md 
          hover:shadow-lg hover:bg-[#3b3b3b] transition-all 
          flex items-center justify-center focus:outline-none 
          border-b-[2px] border-[#E7E7E7]
          ${isDragging ? "scale-95 cursor-grabbing" : "cursor-grab"}
        `}
        aria-label="Open chat"
        whileHover={{ scale: 1.05 }}
        {...listeners}
        {...attributes}
      >
        <div>
          <img
            src={IrisIcon}
            alt="Logo"
            className="pointer-events-none h-8 w-8"
          />
          <img
            src={IrisHat}
            alt="Iris Hat"
            className="pointer-events-none absolute -top-9 -left-1 size-20"
          />
        </div>
      </motion.button>

      <AnimatePresence>
        {showTooltip && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10 }}
            transition={{ duration: 0.2 }}
            className={`absolute bottom-16 right-0 ${
              isDarkTheme
                ? "bg-[#2a2a2a] text-[#e5e5e5]"
                : "bg-white text-gray-800"
            } px-3 py-1.5 rounded-md shadow-md text-sm whitespace-nowrap`}
          >
            {isDragging ? "Release to drop" : "Hold for 1 second to drag"}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
};

export default ChatbotIcon;
