export enum TodoPriority {
  HIGH = "high",
  MEDIUM = "medium",
  LOW = "low"
}

export enum TodoStatus {
  PENDING = "pending",
  IN_PROGRESS = "in_progress",
  ARCHIVED = "archived"
}

export interface ChecklistItem {
  title: string;
  completed: boolean;
}

export interface Checklist {
  items: ChecklistItem[];
}

export interface Todo {
  id: string;
  user_id: string;
  list_id?: string;
  title: string;
  description?: string;
  status: TodoStatus;
  priority: TodoPriority;
  due_date?: string | null;
  reminder_time?: string | null;
  is_recurring: boolean;
  recurrence_pattern?: Record<string, any> | null;
  tags?: Record<string, any> | null;
  checklist?: Checklist | null;
  linked_task_id?: string | null;
  linked_calendar_event_id?: string | null;
  is_completed: boolean;
  completed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface TodoFormData {
  title: string;
  description?: string;
  due_date?: Date;
  priority: TodoPriority;
  reminder_time?: Date;
  is_recurring: boolean;
  tags?: string[];
  checklist?: Checklist;
}

export interface TodosResponse {
  todos: Todo[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface TodoList {
  id: string;
  name: string;
  description: string;
  is_default: boolean;
  user_id: string;
  created_at: string;
  updated_at: string;
  todos: Todo[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface TodoListsResponse {
  lists: TodoList[];
}

export interface CreateTodoListInput {
  name: string;
  description: string;
  is_default: boolean;
}

export interface UpdateTodoListInput {
  name?: string;
  description?: string;
  is_default?: boolean;
}