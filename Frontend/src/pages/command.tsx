import React, { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { PencilLine } from "lucide-react";
import { getApiUrls } from "@/config";

const CommandPage: React.FC = () => {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const resetState = () => {
    setContent("");
    setIsSubmitting(false);
  };

  const createNote = async (noteData: {
    title: string;
    content: string;
    tags: string[];
    favorited: boolean;
  }) => {
    try {
      const { NOTES_API_URL } = getApiUrls();
      const response = await fetch(`${NOTES_API_URL}/graphql`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
        body: JSON.stringify({
          query: `
            mutation CreateNote($input: NotePageInput!) {
              createNotePage(input: $input) {
                success
                message
                data {
                  id
                  title
                  content
                }
              }
            }
          `,
          variables: {
            input: noteData,
          },
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();

      if (result.errors) {
        throw new Error(result.errors[0].message);
      }

      return result.data.createNotePage;
    } catch (error) {
      console.error("Error creating note:", error);
      throw error;
    }
  };

  useEffect(() => {
    const handleKeyDown = async (e: KeyboardEvent) => {
      if (e.key === "Escape" || e.key === "`") {
        resetState();
        window.electron?.close();
      } else if (
        e.key === "Enter" &&
        !e.shiftKey &&
        content.trim() &&
        !isSubmitting
      ) {
        e.preventDefault();
        setIsSubmitting(true);

        try {
          await createNote({
            title: "Quick Note",
            content: content.trim(),
            tags: ["quick-note"],
            favorited: false,
          });
          resetState();
          window.electron?.close();
        } catch (error) {
          console.error("Failed to create quick note:", error);
          setIsSubmitting(false);
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [content, isSubmitting]);

  // Reset state when window is shown
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        resetState();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", handleVisibilityChange);
  }, []);

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.95 }}
      transition={{
        type: "spring",
        stiffness: 300,
        damping: 30,
      }}
      className="w-[600px] bg-zinc-900/90 rounded-md shadow-2xl border border-zinc-700/50 backdrop-blur-xl"
    >
      <div className="p-4">
        <div className="relative">
          <PencilLine className="absolute left-3 top-2.5 h-5 w-5 text-zinc-400" />
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            disabled={isSubmitting}
            className="w-full bg-zinc-800/50 text-white pl-10 pr-4 py-2 rounded-md border border-zinc-700/50 focus:outline-none focus:border-blue-500/50 focus:ring-1 focus:ring-blue-500/50 resize-none h-24 disabled:opacity-50"
            placeholder={
              isSubmitting
                ? "Creating note..."
                : "Type your quick note... (Press Enter to save, Shift+Enter for new line)"
            }
            autoFocus
          />
        </div>
        <div className="mt-2 text-xs text-zinc-500 px-2">
          Press Enter to save • Esc to close
        </div>
      </div>
    </motion.div>
  );
};

export default function StandaloneCommandPage() {
  return (
    <div className="dark">
      <CommandPage />
    </div>
  );
}
