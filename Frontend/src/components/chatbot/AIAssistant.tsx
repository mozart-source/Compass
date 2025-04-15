import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Brain, FileText, Users, Send, Trash, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useState, useRef, useEffect } from "react";
import { cn } from "@/lib/utils";
import { useChat } from "@/components/chatbot/chat-context";
import "./AIAssistant.css";
import Reports from "@/components/AiReports/Reports";

interface AIAssistantProps {
  view?: "chat" | "reports" | "agents";
}

export default function AIAssistant({ view = "chat" }: AIAssistantProps) {
  const [input, setInput] = useState("");
  const chatContentRef = useRef<HTMLDivElement>(null);
  const [isClearingMessages, setIsClearingMessages] = useState(false);
  const [activeTab, setActiveTab] = useState(view);

  // Use the shared chat context
  const { messages, isLoading, streamingText, sendMessage, clearMessages } =
    useChat();

  const scrollToBottom = () => {
    if (chatContentRef.current) {
      const container = chatContentRef.current;
      container.scrollTop = container.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, streamingText]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    await sendMessage(input);
    setInput("");
  };

  // Handle trash button click - properly await the async clearMessages function
  const handleClearMessages = async () => {
    if (isClearingMessages) return;

    try {
      setIsClearingMessages(true);
      await clearMessages();
    } catch (error) {
      console.error("Error clearing messages:", error);
    } finally {
      setIsClearingMessages(false);
    }
  };

  const formatTimestamp = (date: Date) => {
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) {
      return "Just now";
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60);
      return `${minutes} minute${minutes > 1 ? "s" : ""} ago`;
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600);
      return `${hours} hour${hours > 1 ? "s" : ""} ago`;
    } else {
      return date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
  };

  // Format markdown-style text to HTML
  const formatMessageContent = (content: string) => {
    // Process line breaks first
    let formattedContent = content.replace(/\n/g, "<br/>");

    // Process bold text: **text**
    formattedContent = formattedContent.replace(
      /\*\*(.*?)\*\*/g,
      "<strong>$1</strong>"
    );

    // Process italic text: *text*
    formattedContent = formattedContent.replace(/\*(.*?)\*/g, "<em>$1</em>");

    // Process bullet points
    formattedContent = formattedContent.replace(/- (.*?)(<br\/>|$)/g, "• $1$2");

    // Process todo list items: "- [ ]" and "- [x]"
    formattedContent = formattedContent.replace(
      /• \[ \] (.*?)(<br\/>|$)/g,
      "<div>☐ $1</div>$2"
    );
    formattedContent = formattedContent.replace(
      /• \[x\] (.*?)(<br\/>|$)/g,
      "<div>✓ $1</div>$2"
    );

    // Process numbered lists (1. 2. 3. etc) while preserving the original numbers
    formattedContent = formattedContent.replace(
      /(\d+)\. (.*?)(<br\/>|$)/g,
      "<div>$1. $2</div>$3"
    );

    // Process headings ## and ###
    formattedContent = formattedContent.replace(
      /#{3} (.*?)(<br\/>|$)/g,
      "<h3>$1</h3>$2"
    );
    formattedContent = formattedContent.replace(
      /#{2} (.*?)(<br\/>|$)/g,
      "<h2>$1</h2>$2"
    );

    // Process code blocks with ```
    formattedContent = formattedContent.replace(
      /```(.*?)```/gs,
      "<pre>$1</pre>"
    );

    // Highlight todo references with their titles
    formattedContent = formattedContent.replace(
      /"([^"]+)"(\s+todo)?/gi,
      "<span>$1</span>"
    );

    return formattedContent;
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-6 h-[calc(100vh-32px)] overflow-hidden">
      <p className="text-xs uppercase text-muted-foreground tracking-wider">
        Chatbot
      </p>
      <Tabs
        defaultValue={view}
        onValueChange={(value) => setActiveTab(value as typeof view)}
        className="flex-1 flex flex-col min-h-0"
      >
        <div className="flex justify-start">
          <div>
            <h1 className="text-2xl font-bold tracking-tight leading-none">
              IRIS - Your personal assistant
            </h1>
          </div>
          <div className="ml-auto -mt-2">
            <TabsList>
              <TabsTrigger value="chat">Chat</TabsTrigger>
              <TabsTrigger value="reports">Reports & Insights</TabsTrigger>
              <TabsTrigger value="agents">Agent Management</TabsTrigger>
            </TabsList>
          </div>
        </div>

        <div className="flex-1 flex flex-col min-h-0 mt-4">
          <TabsContent
            value="chat"
            className={cn(
              "flex-1 flex flex-col min-h-0",
              activeTab !== "chat" && "hidden"
            )}
          >
            <div className="ai-chat-container">
              <div className="ai-chat-header">
                <div className="ai-chat-title">
                  <Eye className="h-5 w-5 text-muted-foreground" />
                  <span>AI Assistant - Iris</span>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleClearMessages}
                  disabled={isClearingMessages || isLoading}
                  className="h-8 w-8 p-0"
                  title="Clear conversation"
                  aria-label="Clear conversation"
                >
                  {isClearingMessages ? (
                    <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
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
                    <Trash className="h-4 w-4" />
                  )}
                </Button>
              </div>

              <div className="ai-chat-content" ref={chatContentRef}>
                {messages.map((message) => (
                  <div
                    key={message.id}
                    className={cn(
                      "ai-message-row",
                      message.sender === "user" ? "user" : "assistant"
                    )}
                  >
                    <div
                      className={cn(
                        "ai-message-bubble",
                        message.sender === "user" ? "user" : "assistant"
                      )}
                      style={{ transition: "none" }}
                    >
                      {message.sender === "assistant" && (
                        <div className={cn("ai-avatar", "assistant")}>
                          <Eye className="h-4 w-4" />
                        </div>
                      )}
                      <div className="ai-message-content">
                        <div className="ai-message-name">
                          {message.sender === "user"
                            ? "You"
                            : "AI Assistant - Iris"}
                          <span className="ai-message-timestamp">
                            {formatTimestamp(message.timestamp)}
                          </span>
                        </div>
                        <div
                          className="ai-message-text"
                          dangerouslySetInnerHTML={{
                            __html: formatMessageContent(message.text),
                          }}
                        />
                      </div>
                    </div>
                  </div>
                ))}

                {/* Streaming message */}
                {streamingText &&
                  !messages.some((m) => m.text === streamingText) && (
                    <div className="ai-message-row assistant">
                      <div className="ai-message-bubble assistant">
                        <div className="ai-avatar assistant">
                          <Eye className="h-4 w-4" />
                        </div>
                        <div className="ai-message-content">
                          <div className="ai-message-name">AI Assistant</div>
                          <div
                            className="ai-message-text"
                            dangerouslySetInnerHTML={{
                              __html: formatMessageContent(streamingText),
                            }}
                          />
                        </div>
                      </div>
                    </div>
                  )}

                {isLoading && !streamingText && (
                  <div className="ai-message-row assistant">
                    <div className="ai-message-bubble assistant">
                      <div className="ai-avatar assistant">
                        <Eye className="h-4 w-4" />
                      </div>
                      <div className="ai-message-content">
                        <div className="ai-message-name">AI Assistant</div>
                        <div className="ai-typing-indicator">
                          <div className="ai-typing-dot"></div>
                          <div className="ai-typing-dot"></div>
                          <div className="ai-typing-dot"></div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="ai-chat-footer">
                <form onSubmit={handleSubmit} className="ai-input-form">
                  <Input
                    placeholder={
                      isLoading
                        ? "AI is thinking..."
                        : "Type your message here..."
                    }
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    className="ai-chat-input"
                    disabled={isLoading || isClearingMessages}
                  />
                  <Button
                    type="submit"
                    className="ai-send-button"
                    disabled={isLoading || isClearingMessages || !input.trim()}
                  >
                    {isLoading ? (
                      <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
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
                      <>
                        <Send className="h-4 w-4" />
                        <span>Send</span>
                      </>
                    )}
                  </Button>
                </form>
              </div>
            </div>
          </TabsContent>

          <TabsContent
            value="reports"
            className={cn(
              "flex-1 min-h-0",
              activeTab !== "reports" && "hidden"
            )}
          >
            <Reports />
          </TabsContent>

          <TabsContent
            value="agents"
            className={cn("space-y-4", activeTab !== "agents" && "hidden")}
          >
            <Card>
              <CardHeader>
                <CardTitle>AI Agent Management</CardTitle>
                <CardDescription>
                  Configure and monitor your AI agents
                </CardDescription>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Task Assistant
                    </CardTitle>
                    <Brain className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">Active</div>
                    <p className="text-xs text-muted-foreground">
                      Managing 3 tasks
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Research Agent
                    </CardTitle>
                    <FileText className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">Standby</div>
                    <p className="text-xs text-muted-foreground">
                      Ready for queries
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Meeting Assistant
                    </CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">Active</div>
                    <p className="text-xs text-muted-foreground">
                      Monitoring 1 meeting
                    </p>
                  </CardContent>
                </Card>
              </CardContent>
            </Card>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}
