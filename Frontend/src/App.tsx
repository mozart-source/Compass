import React, { useCallback, useEffect } from "react";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";
import {
  DndContext,
  DragEndEvent,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { client } from "./components/notes/apollo-client";
import { AnimatePresence } from "framer-motion";
import PageTransition from "./components/layout/PageTransition";
import AuthTransition from "./components/layout/AuthTransition";
import Dashboard from "./components/dashboard/Dashboard";
import HealthDashboard from "./components/health/HealthDashboard";
import Workflow from "./components/workflow/components/Workflow";
import WorkflowDetailPage from "./components/workflow/components/WorkflowDetail";
import { Tasks } from "./components/todo/Components/TodoParentPage";
import Calendar from "./components/calendar/components/Calendar";
import AIAssistant from "@/components/chatbot/AIAssistant";
import FocusMode from "./components/productivity/FocusMode";
import FileManager from "./components/files/FileManager";
import Notes from "./components/notes/components/Notes";
import Canvas from "./components/Canvas/components/Canvas";
import { ThemeProvider } from "./contexts/theme-provider";
import { WebSocketProvider } from "./contexts/websocket-provider";
import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { SidebarInset } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/toaster";
import TitleBar from "@/components/layout/TitleBar";
import { Login } from "./components/auth/Login";
import { Signup } from "./components/auth/Signup";
import OAuthCallback from "./pages/auth/callback";
import { useAuth } from "@/hooks/useAuth";
import { useQueryClient } from "@tanstack/react-query";
import { QueryClientProvider } from "@tanstack/react-query";
import Chatbot from "./components/chatbot/Chatbot";
import CommandPage from "@/pages/command";
import Journaling from "./components/Journaling/components/Journaling";
import { useDragStore } from "@/dragStore";
import ErrorBoundary from "./components/debug/ErrorBoundary";
import EnvironmentDebug from "./components/debug/EnvironmentDebug";

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuth();
  console.log(isAuthenticated);
  return isAuthenticated ? children : <Navigate to="/login" replace />;
};

function App() {
  const queryClient = useQueryClient();
  const location = useLocation();

  // Add sensors configuration
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        delay: 100, // Adjust as needed (i think 100ms is good)
        tolerance: 5, // Allow 5px of movement during delay
      },
    })
  );

  useEffect(() => {
    const handleGlobalClick = (event: MouseEvent) => {
      if ((event.target as HTMLElement).closest("[data-no-dismiss]")) {
        return;
      }
      useDragStore.getState().setChatbotAttachedTo(null);
      useDragStore.getState().setAttachmentPosition(null);
      useDragStore.getState().setLastDroppedId(null);
    };

    document.addEventListener("mousedown", handleGlobalClick);
    return () => document.removeEventListener("mousedown", handleGlobalClick);
  }, []);

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;

    if (active.id === "chatbot-bubble") {
      if (over) {
        useDragStore.getState().setChatbotAttachedTo(over.id as string);
      } else {
        // If dropped outside any droppable area, detach it.
        useDragStore.getState().setChatbotAttachedTo(null);
        useDragStore.getState().setAttachmentPosition(null);
      }
    }

    if (active && over) {
      // This is used to show the buttons next to the todo item
      useDragStore.getState().setLastDroppedId(over.id as string);
      console.log(`Dragged: ${active.id} → Dropped on: ${over.id}`);
    }
  }, []);

  // Check if we're on the command page
  const isCommandPage = window.location.hash === "#command";

  // If it's the command page, render it directly without authentication
  if (isCommandPage) {
    return (
      <ErrorBoundary>
        <ThemeProvider defaultTheme="dark" storageKey="aiwa-theme">
          <CommandPage />
        </ThemeProvider>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <ApolloProvider client={client}>
        <ThemeProvider defaultTheme="dark" storageKey="aiwa-theme">
          <QueryClientProvider client={queryClient}>
            <WebSocketProvider>
              <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
                <div className="h-screen flex flex-col overflow-hidden">
                  <AnimatePresence mode="wait">
                    <Routes location={location} key={location.pathname}>
                      <Route
                        path="/login"
                        element={
                          <AuthTransition>
                            <Login />
                          </AuthTransition>
                        }
                      />
                      <Route
                        path="/signup"
                        element={
                          <AuthTransition>
                            <Signup />
                          </AuthTransition>
                        }
                      />
                      <Route
                        path="/auth/callback"
                        element={
                          <AuthTransition>
                            <OAuthCallback />
                          </AuthTransition>
                        }
                      />
                      <Route path="/debug" element={<EnvironmentDebug />} />
                      <Route
                        path="/*"
                        element={
                          <ProtectedRoute>
                            <ErrorBoundary>
                              <SidebarProvider>
                                <AppSidebar />
                                <SidebarInset className="flex flex-col">
                                  <TitleBar darkMode={false} />
                                  <main className="flex-1 h-full main-content">
                                    <Routes>
                                      <Route
                                        index
                                        element={
                                          <Navigate to="/login" replace />
                                        }
                                      />
                                      <Route
                                        path="dashboard"
                                        element={
                                          <PageTransition>
                                            <Dashboard />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="Todos&Habits"
                                        element={
                                          <PageTransition>
                                            <Tasks />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="calendar"
                                        element={
                                          <PageTransition>
                                            <Calendar />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="ai"
                                        element={
                                          <PageTransition>
                                            <AIAssistant />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="focus"
                                        element={
                                          <PageTransition>
                                            <FocusMode />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="files"
                                        element={
                                          <PageTransition>
                                            <FileManager />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="health"
                                        element={
                                          <PageTransition>
                                            <HealthDashboard />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="workflow"
                                        element={
                                          <PageTransition>
                                            <Workflow />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="workflow/:id"
                                        element={
                                          <PageTransition>
                                            <WorkflowDetailPage />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="notes"
                                        element={
                                          <PageTransition>
                                            <Notes />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="canvas"
                                        element={
                                          <PageTransition>
                                            <Canvas />
                                          </PageTransition>
                                        }
                                      />
                                      <Route
                                        path="journaling"
                                        element={
                                          <PageTransition>
                                            <Journaling />
                                          </PageTransition>
                                        }
                                      />
                                    </Routes>
                                  </main>
                                  <Chatbot />
                                </SidebarInset>
                              </SidebarProvider>
                            </ErrorBoundary>
                          </ProtectedRoute>
                        }
                      />
                    </Routes>
                  </AnimatePresence>
                  <Toaster />
                </div>
              </DndContext>
            </WebSocketProvider>
          </QueryClientProvider>
        </ThemeProvider>
      </ApolloProvider>
    </ErrorBoundary>
  );
}

export default App;
