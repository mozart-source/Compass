import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App1 from "./App";
import "./index.css";

// Move QueryClientProvider to main entry file
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "./lib/queryClient";
import { ChatProvider } from "@/components/chatbot/chat-context";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <ChatProvider>
          <App1 />
        </ChatProvider>
      </QueryClientProvider>
    </BrowserRouter>
  </React.StrictMode>
);
