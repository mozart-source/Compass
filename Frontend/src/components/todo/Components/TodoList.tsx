import React, { useState, useRef, useEffect } from "react";
import { createPortal } from "react-dom";
import {
  Plus,
  X,
  MoreVertical,
  CalendarFold,
  Repeat,
  Check,
  ArrowLeft,
  CalendarSync,
  CalendarCheck,
  CalendarClock,
  ChevronDown,
  ListTodo,
  Clock,
  Zap,
} from "lucide-react";
import { useDroppable } from "@dnd-kit/core";
import { useDragStore } from "@/dragStore";
import PriorityIndicator from "./PriorityIndicator";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../../ui/dropdown-menu";
import { Button } from "../../ui/button";
import { Input } from "../../ui/input";
import { Progress } from "../../ui/progress";
import Checkbox from "../../ui/checkbox";
import TodoForm from "./TodoForm";
import { Badge } from "../../ui/badge";
import cn from "classnames";
import { useTheme } from "@/contexts/theme-provider";
import { useQuery } from "@tanstack/react-query";
import authApi, { User } from "@/api/auth";
import { Habit } from "@/components/todo/types-habit";
import {
  Todo,
  TodoFormData,
  TodoStatus,
  TodoPriority,
} from "@/components/todo/types-todo";
import {
  useCreateTodo,
  useUpdateTodo,
  useDeleteTodo,
  useToggleTodoStatus,
  useHabits,
  useCreateHabit,
  useToggleHabit,
  useDeleteHabit,
  useUpdateHabit,
  useTodoLists,
  useCreateTodoList,
  useDeleteTodoList,
} from "../hooks";
import { Separator } from "@/components/ui/separator";
import { motion, AnimatePresence } from "framer-motion";
import { useWebSocket } from "@/contexts/websocket-provider";

type TodoFilterType = "log" | "thisWeek" | "today" | "done";

const AttachedActionsBox: React.FC<{
  todo: Todo;
  position: { x: number; y: number; side: "left" | "right" };
  onClose: () => void;
  onSubtasksGenerated: () => void;
}> = ({ todo, position, onClose, onSubtasksGenerated }) => {
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [requestTimeout, setRequestTimeout] = useState<NodeJS.Timeout | null>(
    null
  );
  const [selectedAction, setSelectedAction] = useState<string | null>(null);

  const { sendMessage, isConnected } = useWebSocket();

  const ACTION_BOX_WIDTH = 245; // w-64
  const RESULT_BOX_WIDTH = 245; // w-64
  const GAP = 16;
  const CHATBOT_ICON_WIDTH = 38; // w-12

  const isRightSide = position.side === "right";

  const actionBoxLeft = isRightSide
    ? position.x + CHATBOT_ICON_WIDTH + GAP
    : position.x - ACTION_BOX_WIDTH - GAP;

  const resultBoxLeft = isRightSide
    ? actionBoxLeft + ACTION_BOX_WIDTH + GAP
    : actionBoxLeft - RESULT_BOX_WIDTH - GAP;

  // Map action buttons to option IDs from the TodoAgent
  const actionButtons = [
    { label: "Subtask Generation", icon: ListTodo, optionId: "todo_subtasks" },
    { label: "Deadline-based Advice", icon: Clock, optionId: "todo_deadline" },
    {
      label: "Priority & Motivation Boost",
      icon: Zap,
      optionId: "todo_priority",
    },
  ];

  // Handle AI option processing
  const handleOptionSelect = (optionId: string, label: string) => {
    setSelectedAction(label);
    setProcessing(true);
    setError(null);
    setResult(null);

    // Set a timeout for the request
    const timeout = setTimeout(() => {
      setError("Request timed out. Please try again.");
      setProcessing(false);
    }, 30000); // 30 second timeout

    setRequestTimeout(timeout);

    // Enhanced connection validation and retry logic
    if (isConnected && sendMessage) {
      // First, try to send the AI options request to get available actions
      const optionsSuccess = sendMessage({
        type: "ai_options_request",
        target_type: "todo",
        target_id: todo.id,
        target_data: todo,
      });

      if (optionsSuccess !== true) {
        setError("Could not connect to AI service. Please try again later.");
        setProcessing(false);
        clearTimeout(timeout);
        setRequestTimeout(null);
        return;
      }

      // Then send the actual processing request after a brief delay
      setTimeout(() => {
        const processSuccess = sendMessage({
          type: "ai_process_request",
          option_id: optionId,
          target_type: "todo",
          target_id: todo.id,
          target_data: todo,
        });

        if (processSuccess !== true) {
          setError("Could not connect to AI service. Please try again later.");
          setProcessing(false);
          clearTimeout(timeout);
          setRequestTimeout(null);
        }
      }, 100); // Small delay to ensure options are received first
    } else {
      // Enhanced error message with connection details
      const connectionStatus = isConnected ? "connected" : "disconnected";
      const messageStatus = sendMessage ? "available" : "unavailable";
      setError(
        `AI service connection failed (WebSocket: ${connectionStatus}, Messaging: ${messageStatus}). Please refresh the page and try again.`
      );
      setProcessing(false);
      clearTimeout(timeout);
      setRequestTimeout(null);
    }
  };

  // Listen for AI options response from WebSocket
  useEffect(() => {
    const handleAIResponse = (
      e: CustomEvent<{
        type: string;
        data: {
          targetId: string;
          targetType: string;
          optionId?: string;
          error?: string;
          result?: string;
          success?: boolean;
        };
      }>
    ) => {
      if (e.detail?.data?.targetId !== todo.id) return;

      if (requestTimeout) {
        clearTimeout(requestTimeout);
        setRequestTimeout(null);
      }

      if (e.detail?.type === "ai_option_processing") {
        setProcessing(true);
        setError(null);
      }

      if (e.detail?.type === "ai_option_result") {
        if (e.detail?.data?.error || e.detail?.data?.success === false) {
          setError(
            e.detail.data.error || "An error occurred processing the request"
          );
          setProcessing(false);
          return;
        }

        const resultContent = e.detail.data.result || null;
        setResult(
          typeof resultContent === "string"
            ? resultContent
            : JSON.stringify(resultContent, null, 2)
        );
        setProcessing(false);

        if (
          e.detail?.data?.optionId === "todo_subtasks" &&
          e.detail?.data?.success
        ) {
          onSubtasksGenerated();
        }
      }
    };

    window.addEventListener(
      "websocket_ai_event",
      handleAIResponse as EventListener
    );

    return () => {
      window.removeEventListener(
        "websocket_ai_event",
        handleAIResponse as EventListener
      );
    };
  }, [todo.id, requestTimeout, onSubtasksGenerated]);

  return createPortal(
    <>
      {/* Result Box */}
      {(result || processing || error) && (
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.9 }}
          transition={{ duration: 0.2, ease: "easeOut" }}
          style={{
            position: "fixed",
            top: `${position.y}px`,
            left: `${resultBoxLeft}px`,
          }}
          className="bg-card border rounded-lg shadow-xl p-3 space-y-2 z-[9999] w-64"
          data-no-dismiss
        >
          <div className="flex items-center justify-between pb-2 border-b mb-2">
            <h4 className="text-sm font-medium">
              {selectedAction || "AI Response"}
            </h4>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 -mr-2 -mt-2"
              onClick={() => {
                setResult(null);
                setError(null);
                setSelectedAction(null);
              }}
              aria-label="Close Result"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
          {error ? (
            <div className="bg-red-100 border border-red-400 text-red-700 px-3 py-2 rounded relative text-sm">
              <span className="block sm:inline">{error}</span>
            </div>
          ) : processing ? (
            <div className="flex flex-col items-center justify-center py-4">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <p className="text-sm text-muted-foreground mt-2">
                Processing...
              </p>
            </div>
          ) : result ? (
            <div className="bg-muted p-3 rounded-md text-sm whitespace-pre-wrap max-h-[300px] overflow-y-auto">
              {result}
            </div>
          ) : null}
        </motion.div>
      )}

      {/* Actions Box */}
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.9 }}
        transition={{ duration: 0.2, ease: "easeOut" }}
        style={{
          position: "fixed",
          top: `${position.y}px`,
          left: `${actionBoxLeft}px`,
        }}
        className="bg-card border rounded-lg shadow-xl p-3 space-y-2 z-[9999] w-64"
        data-no-dismiss
      >
        <div className="flex items-center justify-between pb-2 border-b mb-2">
          <h4 className="text-sm font-medium">AI Actions</h4>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 -mr-2 -mt-2"
            onClick={onClose}
            aria-label="Close AI Actions"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="flex flex-col gap-1">
          {actionButtons.map((button, i) => (
            <Button
              key={i}
              variant="ghost"
              size="sm"
              className="gap-2 whitespace-nowrap justify-start w-full"
              onClick={() => handleOptionSelect(button.optionId, button.label)}
              disabled={processing}
            >
              {processing && selectedAction === button.label ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary" />
              ) : (
                <button.icon className="h-4 w-4" />
              )}
              <span>{button.label}</span>
            </Button>
          ))}
        </div>
      </motion.div>
    </>,
    document.body
  );
};

interface TodoItemProps {
  todo: Todo;
  onToggle: (todo: Todo) => void;
  onEdit: (todo: Todo) => void;
  onDelete: (id: string) => void;
  onUpdateChecklist: (
    id: string,
    checklist: { items: Array<{ title: string; completed: boolean }> }
  ) => void;
  isDarkMode: boolean;
  columnType: TodoFilterType;
}

const TodoItem: React.FC<TodoItemProps> = ({
  todo,
  onToggle,
  onEdit,
  onDelete,
  onUpdateChecklist,
  isDarkMode,
  columnType,
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id: todo.id,
  });

  const { chatbotAttachedTo, setAttachmentPosition } = useDragStore();
  const todoRef = useRef<HTMLDivElement | null>(null);

  const combinedRef = (node: HTMLDivElement) => {
    todoRef.current = node;
    setNodeRef(node);
  };

  useEffect(() => {
    if (chatbotAttachedTo === todo.id && todoRef.current) {
      const rect = todoRef.current.getBoundingClientRect();
      const chatbotIconWidth = 48; // w-12 from ChatbotIcon.tsx
      const y = rect.top + rect.height / 2 - chatbotIconWidth / 2;

      let x: number;
      const side: "left" | "right" =
        columnType === "log" || columnType === "thisWeek" ? "right" : "left";

      if (side === "right") {
        x = rect.right + 12;
      } else {
        x = rect.left - 60;
      }

      setAttachmentPosition({ x, y, side });

      const handleResizeOrScroll = () => {
        if (todoRef.current) {
          const newRect = todoRef.current.getBoundingClientRect();
          let newX: number;
          if (side === "right") {
            newX = newRect.right + 12;
          } else {
            newX = newRect.left - 60;
          }
          const newY = newRect.top + newRect.height / 2 - chatbotIconWidth / 2;
          setAttachmentPosition({ x: newX, y: newY, side });
        }
      };

      window.addEventListener("resize", handleResizeOrScroll);
      document
        .querySelector(".main-content")
        ?.addEventListener("scroll", handleResizeOrScroll);

      return () => {
        window.removeEventListener("resize", handleResizeOrScroll);
        document
          .querySelector(".main-content")
          ?.removeEventListener("scroll", handleResizeOrScroll);
      };
    }
  }, [chatbotAttachedTo, todo.id, setAttachmentPosition, columnType]);

  return (
    <div className="relative">
      <div
        ref={combinedRef}
        className={cn(
          "group relative rounded-lg border bg-card p-3 transition-all hover:border-primary/50",
          todo.is_completed && "bg-muted",
          isOver && "border-primary bg-primary/10 shadow-lg"
        )}
        data-no-dismiss
      >
        <div className="flex items-start gap-4">
          <Checkbox
            name={`todo-${todo.id}`}
            checked={todo.is_completed}
            onChange={() => onToggle(todo)}
            darkMode={isDarkMode}
            className="mt-0.5"
          />
          <div className="flex-1 space-y-2">
            <div className="flex items-center">
              <span
                className={cn(
                  "text-sm font-medium",
                  todo.is_completed && "line-through text-muted-foreground"
                )}
              >
                {todo.title}
              </span>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <PriorityIndicator
                priority={todo.priority || TodoPriority.MEDIUM}
              />
              {todo.tags &&
                typeof todo.tags === "object" &&
                Object.keys(todo.tags).map((tag) => (
                  <Badge
                    key={tag}
                    variant="default"
                    className="text-xs text-black hover:bg-white/100 transition-colors"
                  >
                    {tag}
                  </Badge>
                ))}
            </div>
            {todo.checklist?.items && todo.checklist.items.length > 0 && (
              <div className="mt-2 space-y-2">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span>Checklist</span>
                  <span>
                    {
                      todo.checklist.items.filter((item) => item.completed)
                        .length
                    }
                    /{todo.checklist.items.length}
                  </span>
                </div>
                <Progress
                  value={
                    (todo.checklist.items.filter((item) => item.completed)
                      .length /
                      todo.checklist.items.length) *
                    100
                  }
                  className="h-1"
                />
                <div className="space-y-1.5">
                  {todo.checklist.items.map((item, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <Checkbox
                        name={`todo-${todo.id}-checklist-${index}`}
                        checked={item.completed}
                        onChange={() => {
                          const newChecklist = {
                            items: [...todo.checklist!.items],
                          };
                          newChecklist.items[index] = {
                            ...item,
                            completed: !item.completed,
                          };
                          onUpdateChecklist(todo.id, newChecklist);
                        }}
                        darkMode={isDarkMode}
                        className="h-3 w-3"
                      />
                      <span
                        className={cn(
                          "text-xs",
                          item.completed && "line-through text-muted-foreground"
                        )}
                      >
                        {item.title}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100"
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-[160px]">
              <DropdownMenuItem onClick={() => onEdit(todo)}>
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDelete(todo.id)}>
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  );
};

const TodoList: React.FC = () => {
  const { theme } = useTheme();
  const isDarkMode = theme === "dark";
  const [editingTodo, setEditingTodo] = useState<Todo | null>(null);
  const [showTodoForm, setShowTodoForm] = useState(false);
  const [isFlipping, setIsFlipping] = useState(false);
  const [newHabit, setNewHabit] = useState("");
  const [showHabitInput, setShowHabitInput] = useState(false);
  const [showHabitTracker, setShowHabitTracker] = useState(false);
  const [editingHabit, setEditingHabit] = useState<Habit | null>(null);

  const {
    chatbotAttachedTo,
    attachmentPosition,
    setChatbotAttachedTo,
    setAttachmentPosition,
  } = useDragStore();

  // New state for managing multiple lists
  const [currentListId, setCurrentListId] = useState<string>("");
  const [showNewListInput, setShowNewListInput] = useState(false);
  const [newListName, setNewListName] = useState("");
  const newListInputRef = useRef<HTMLInputElement>(null);

  // User authentication query
  const { data: user } = useQuery<User>({
    queryKey: ["user"],
    queryFn: async () => {
      const token = localStorage.getItem("token");
      if (!token) throw new Error("No token found");
      return authApi.getMe();
    },
  });

  // Use the custom hooks
  const { data: todoLists = [], refetch: refetchTodoLists } =
    useTodoLists(user);
  const { data: habits = [] } = useHabits(user);
  const createTodoMutation = useCreateTodo();
  const updateTodoMutation = useUpdateTodo();
  const deleteTodoMutation = useDeleteTodo();
  const toggleTodoStatus = useToggleTodoStatus();
  const createHabitMutation = useCreateHabit();
  const toggleHabitMutation = useToggleHabit();
  const deleteHabitMutation = useDeleteHabit();
  const updateHabitMutation = useUpdateHabit();
  const createTodoListMutation = useCreateTodoList();
  const deleteTodoListMutation = useDeleteTodoList();

  // Refetch todos when component mounts
  React.useEffect(() => {
    if (user) {
      refetchTodoLists();
    }
  }, [user]);

  // Set default list ID on initial load
  React.useEffect(() => {
    if (todoLists.length > 0 && !currentListId) {
      const defaultList = todoLists.find((list) => list.is_default);
      setCurrentListId(defaultList?.id || todoLists[0].id);
    }
  }, [todoLists, currentListId]);

  // Get current list and its todos
  const currentList = todoLists.find((list) => list.id === currentListId);
  const todos = currentList?.todos || [];

  const attachedTodo = todos.find((t) => t.id === chatbotAttachedTo);

  const handleEditHabit = (habit: Habit) => {
    setEditingHabit(habit);
    setNewHabit(habit.title);
    setShowHabitInput(true);
  };

  const handleEditTodo = (todo: Todo) => {
    setEditingTodo(todo);
    setShowTodoForm(true);
  };

  const handleTodoFormSubmit = (formData: TodoFormData) => {
    if (!user) return;

    if (editingTodo) {
      const updates: Partial<Todo> = {
        ...formData,
        due_date: formData.due_date?.toISOString() || null,
        reminder_time: formData.reminder_time?.toISOString() || null,
        tags: formData.tags?.reduce((acc, tag) => ({ ...acc, [tag]: {} }), {}),
        checklist: { items: formData.checklist?.items || [] },
        is_completed: editingTodo.is_completed,
      };
      updateTodoMutation.mutate({ id: editingTodo.id, updates });
    } else {
      const newTodo: Omit<
        Todo,
        "id" | "created_at" | "updated_at" | "completed_at"
      > = {
        user_id: user.id,
        list_id: currentListId === "default" ? undefined : currentListId,
        title: formData.title,
        description: formData.description,
        status: TodoStatus.PENDING,
        priority: formData.priority,
        is_recurring: formData.is_recurring,
        due_date: formData.due_date?.toISOString() || null,
        reminder_time: formData.reminder_time?.toISOString() || null,
        tags: formData.tags?.reduce((acc, tag) => ({ ...acc, [tag]: {} }), {}),
        checklist: { items: formData.checklist?.items || [] },
        is_completed: false,
        linked_task_id: null,
        linked_calendar_event_id: null,
        recurrence_pattern: {},
      };
      createTodoMutation.mutate(newTodo);
    }
    setShowTodoForm(false);
    setEditingTodo(null);
  };

  const handleToggleTodo = (todo: Todo) => {
    toggleTodoStatus.mutate(todo);
  };

  const handleDeleteTodo = (id: string) => {
    deleteTodoMutation.mutate(id);
  };

  const renderTodoList = (todos: Todo[], columnType: TodoFilterType) => {
    if (!todos || todos.length === 0) {
      return <div className="text-sm text-muted-foreground">No Todos</div>;
    }

    return (
      <div className="space-y-2">
        {todos.map((todo) => (
          <TodoItem
            key={todo.id}
            todo={todo}
            onToggle={handleToggleTodo}
            onEdit={handleEditTodo}
            onDelete={handleDeleteTodo}
            onUpdateChecklist={(id, checklist) =>
              updateTodoMutation.mutate({ id, updates: { checklist } })
            }
            isDarkMode={isDarkMode}
            columnType={columnType}
          />
        ))}
      </div>
    );
  };

  const handleAddHabit = () => {
    if (!newHabit.trim() || !user) return;

    if (editingHabit) {
      updateHabitMutation.mutate({
        habitId: editingHabit.id,
        title: newHabit.trim(),
      });
    } else {
      createHabitMutation.mutate({
        title: newHabit.trim(),
        description: "",
        start_day: new Date().toISOString(),
        end_day: new Date().toISOString(),
      });
    }

    // Clear input and close form
    setNewHabit("");
    setShowHabitInput(false);
    setEditingHabit(null);
  };

  const toggleHabit = (habitId: string, isCompleted: boolean) => {
    toggleHabitMutation.mutate({ habitId, isCompleted });
  };

  const deleteHabit = (habitId: string) => {
    deleteHabitMutation.mutate(habitId);
  };

  const handleFlipToHabits = () => {
    setIsFlipping(true);
    setShowHabitTracker(true);
    document.body.style.overflow = "hidden";
    setTimeout(() => {
      setIsFlipping(false);
      document.body.style.overflow = "";
    }, 250);
  };

  const handleFlipToLog = () => {
    setShowHabitTracker(false);
    setIsFlipping(true);
    document.body.style.overflow = "hidden";
    setTimeout(() => {
      setIsFlipping(false);
      document.body.style.overflow = "";
    }, 250);
  };

  const filterTodos = (type: TodoFilterType): Todo[] => {
    // Ensure we have an array to work with
    if (!Array.isArray(todos)) {
      return [];
    }

    const filtered = todos.filter((todo) => {
      // Handle completed todos first
      if (todo.is_completed) {
        return type === "done";
      }

      // For non-completed todos, never show in done section
      if (type === "done") {
        return false;
      }

      // If no due date, put it in the log
      if (!todo.due_date) {
        return type === "log";
      }

      // Parse the due date string to a Date object
      const dueDate = new Date(todo.due_date);

      // Get current date at midnight in local timezone
      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

      // Calculate end of today in local timezone
      const endOfToday = new Date(today);
      endOfToday.setHours(23, 59, 59, 999);

      // Calculate end of week in local timezone
      const endOfWeek = new Date(today);
      endOfWeek.setDate(today.getDate() + 7);
      endOfWeek.setHours(23, 59, 59, 999);

      // Test each condition separately using timestamps for reliable comparison
      const isPastDue = dueDate < today;
      const isDueToday = dueDate >= today && dueDate <= endOfToday;
      const isDueThisWeek = dueDate > endOfToday && dueDate <= endOfWeek;
      const isDueLater = dueDate > endOfWeek;

      switch (type) {
        case "today":
          return isDueToday || isPastDue;
        case "thisWeek":
          return isDueThisWeek;
        case "log":
          return isDueLater || !todo.due_date;
        default:
          return false;
      }
    });

    return filtered;
  };

  const renderTodoColumn = (type: TodoFilterType, title: string) => {
    const columnTodos = filterTodos(type);

    if (type === "log" && (showHabitTracker || isFlipping)) {
      return (
        <div className="h-[calc(100vh-190px)] w-full perspective-1000 -mt-4">
          <div
            className={cn(
              "relative h-full w-full [transition:transform_250ms_ease-in-out] [transform-style:preserve-3d]",
              showHabitTracker ? "[transform:rotateY(180deg)]" : ""
            )}
          >
            {/* Front side - Log */}
            <div className="absolute inset-0 w-full h-full rounded-lg border bg-card p-4 flex flex-col [backface-visibility:hidden]">
              <div className="mb-4 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <CalendarFold className="h-4 w-4 text-muted-foreground" />
                  <h3 className="font-medium">Log</h3>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleFlipToHabits()}
                    className="text-muted-foreground hover:text-foreground"
                  >
                    <Repeat className="w-4 h-4" />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setShowTodoForm(true);
                    }}
                    className="text-muted-foreground hover:text-foreground"
                  >
                    <Plus className="w-4 h-4" />
                  </Button>
                </div>
              </div>
              <div className="flex-1 overflow-y-auto">
                {renderTodoList(columnTodos, type)}
              </div>
            </div>

            {/* Back side - Habits */}
            <div className="absolute inset-0 w-full h-full rounded-lg border bg-card p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
              <div className="flex flex-col h-full w-full">
                <div className="mb-4 flex items-center justify-between">
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <div className="flex items-center gap-2">
                        <CalendarSync className="h-4 w-4 text-muted-foreground" />
                        <h3 className="font-medium">Daily Habits</h3>
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {habits.filter((h) => h.is_completed).length}/
                        {habits.length} Done
                      </span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleFlipToLog()}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <ArrowLeft className="h-4 w-4" />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setShowHabitInput(true)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <Plus className="h-4 h-4" />
                    </Button>
                  </div>
                </div>

                {showHabitInput && (
                  <div className="flex gap-2 mb-4">
                    <Input
                      value={newHabit}
                      onChange={(e) => setNewHabit(e.target.value)}
                      placeholder="New habit..."
                      className="flex-1"
                      onKeyDown={(e) => e.key === "Enter" && handleAddHabit()}
                    />
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => {
                        setShowHabitInput(false);
                        setEditingHabit(null);
                        setNewHabit("");
                      }}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleAddHabit}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <Check className="h-4 w-4" />
                    </Button>
                  </div>
                )}

                <div className="flex-1 overflow-y-auto w-full">
                  <div className="space-y-2">
                    {habits.map((habit) => (
                      <div
                        key={habit.id}
                        className={cn(
                          "group relative rounded-lg border bg-card p-4 transition-all hover:border-border/50",
                          habit.is_completed && "bg-muted"
                        )}
                      >
                        <div className="flex items-start gap-4">
                          <Checkbox
                            name={`habit-${habit.id}`}
                            checked={habit.is_completed}
                            onChange={() =>
                              toggleHabit(habit.id, habit.is_completed)
                            }
                            darkMode={isDarkMode}
                            className="mt-0.5"
                          />
                          <div className="flex-1 space-y-2">
                            <div className="flex items-center">
                              <span
                                className={cn(
                                  "text-sm font-medium",
                                  habit.is_completed &&
                                    "line-through text-muted-foreground"
                                )}
                              >
                                {habit.title}
                              </span>
                            </div>
                            {habit.current_streak > 0 && (
                              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                                <Repeat className="w-3 h-3" />
                                <span>{habit.current_streak} day streak</span>
                              </div>
                            )}
                          </div>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button
                                variant="ghost"
                                className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100"
                              >
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                              align="end"
                              className="w-[160px]"
                            >
                              <DropdownMenuItem
                                onClick={() => handleEditHabit(habit)}
                              >
                                Edit
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => deleteHabit(habit.id)}
                              >
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return (
      <div className="h-[calc(100vh-190px)] w-full rounded-lg border bg-card p-4 -mt-4 flex flex-col">
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            {type === "log" && (
              <CalendarFold className="h-4 w-4 text-muted-foreground" />
            )}
            {type === "thisWeek" && (
              <CalendarClock className="h-4 w-4 text-muted-foreground" />
            )}
            {type === "today" && (
              <CalendarCheck className="h-4 w-4 text-muted-foreground" />
            )}
            {type === "done" && (
              <Check className="h-4 w-4 text-muted-foreground" />
            )}
            <h3 className="font-medium">{title}</h3>
          </div>
          <div className="flex gap-2">
            {type === "log" && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => handleFlipToHabits()}
                className="text-muted-foreground hover:text-foreground"
              >
                <Repeat className="w-4 h-4" />
              </Button>
            )}
            {type !== "done" && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  setShowTodoForm(true);
                }}
                className="text-muted-foreground hover:text-foreground"
              >
                <Plus className="w-4 h-4" />
              </Button>
            )}
          </div>
        </div>
        <div className="flex-1 overflow-y-auto">
          {renderTodoList(columnTodos, type)}
        </div>
      </div>
    );
  };

  // Add list selection handler
  const handleListChange = (listId: string) => {
    setCurrentListId(listId);
  };

  // Add new list handler
  const handleAddNewList = () => {
    if (!newListName.trim()) return;

    createTodoListMutation.mutate({
      name: newListName.trim(),
      description: "",
      is_default: false,
    });

    setNewListName("");
    setShowNewListInput(false);
  };

  // Add delete list handler
  const handleDeleteList = (listId: string) => {
    deleteTodoListMutation.mutate(listId);
  };

  if (!user) {
    return <div>Please log in to view todos</div>;
  }

  return (
    <div className="grid grid-cols-4 gap-4 p-6 h-full w-full">
      {/* Todo Label - Main page title */}
      <div className="col-span-4">
        <p className="text-xs uppercase text-muted-foreground tracking-wider mb-4">
          todo
        </p>
        <div className="flex justify-start">
          <h1 className="text-2xl font-bold tracking-tight leading-none mr-2">
            Todos & Habits
          </h1>
          <Separator
            orientation="vertical"
            className="h-5 relative top-[6px] z-[1] mr-1"
          />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                size="sm"
                className="-ml-2 bg-background hover:bg-background relative top-[-1px]"
              >
                {todoLists.find((list) => list.id === currentListId)?.name}
                <ChevronDown className="h-4 w-4 opacity-70" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="center" className="w-[180px]">
              {todoLists.map((list) => (
                <DropdownMenuItem
                  key={list.id}
                  className="flex items-center justify-between group"
                  onClick={() => handleListChange(list.id)}
                >
                  <span
                    className={cn(
                      "flex-1",
                      currentListId === list.id && "font-medium"
                    )}
                  >
                    {list.name}
                  </span>
                  {!list.is_default && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteList(list.id);
                      }}
                      className="h-4 w-4 p-2 opacity-0 group-hover:opacity-100"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </DropdownMenuItem>
              ))}
              <DropdownMenuItem
                className="border-t mt-1 pt-1 cursor-pointer text-foreground font-medium"
                onClick={() => {
                  setShowNewListInput(true);
                  setTimeout(() => {
                    newListInputRef.current?.focus();
                  }, 0);
                }}
              >
                <Plus className="h-4 w-4 mr-2" />
                Create new list
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <div className="flex items-baseline">
            {/* New list input */}
            {showNewListInput && (
              <div className="flex gap-2 items-center ml-1">
                {" "}
                {/* //////Remove the outline and focus styles */}
                <Input
                  ref={newListInputRef}
                  value={newListName}
                  onChange={(e) => setNewListName(e.target.value)}
                  placeholder="List name..."
                  className="w-40 h-8"
                  onKeyDown={(e) => e.key === "Enter" && handleAddNewList()}
                />
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setShowNewListInput(false)}
                  className="h-8 w-8 p-0"
                >
                  <X className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleAddNewList}
                  className="h-8 w-8 p-0"
                >
                  <Check className="h-4 w-4" />
                </Button>
              </div>
            )}
          </div>
          <div className="col-span-4 mb-4 flex items-center ml-auto -mt-1">
            <Button
              variant="outline"
              size="sm"
              className="gap-2"
              onClick={() => {
                setEditingTodo(null);
                setShowTodoForm(true);
              }}
            >
              <Plus className="h-4 w-4" />
              New Todo
            </Button>
          </div>
        </div>
      </div>
      {renderTodoColumn("log", "Log")}
      {renderTodoColumn("thisWeek", "This Week")}
      {renderTodoColumn("today", "Today")}
      {renderTodoColumn("done", "Done")}

      <AnimatePresence>
        {attachedTodo && attachmentPosition && (
          <AttachedActionsBox
            todo={attachedTodo}
            position={attachmentPosition}
            onClose={() => {
              setChatbotAttachedTo(null);
              setAttachmentPosition(null);
            }}
            onSubtasksGenerated={refetchTodoLists}
          />
        )}
      </AnimatePresence>

      {showTodoForm && (
        <TodoForm
          onClose={() => {
            setShowTodoForm(false);
            setEditingTodo(null);
          }}
          user={user}
          todo={editingTodo || undefined}
          onSubmit={handleTodoFormSubmit}
          onDelete={handleDeleteTodo}
          currentListId={currentListId}
          listId={currentListId}
        />
      )}
    </div>
  );
};

export default TodoList;
