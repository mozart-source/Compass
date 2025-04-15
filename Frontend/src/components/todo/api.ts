import axios from 'axios';
import { Todo, TodosResponse, TodoList, TodoListsResponse, CreateTodoListInput, UpdateTodoListInput } from './types-todo';
import { Habit } from './types-habit';
import { getApiUrls } from '@/config';

const { GO_API_URL } = getApiUrls();
const API_BASE_URL = `${GO_API_URL}`;

// Configure axios defaults
//axios.defaults.headers.common['Accept-Encoding'] = 'gzip, deflate';

// Todo API functions
export const fetchTodos = async (userId: string, listId?: string): Promise<Todo[]> => {
  const url = listId ? 
    `${API_BASE_URL}/todos?list_id=${listId}` :
    `${API_BASE_URL}/todos`;
  const response = await axios.get<{ data: TodosResponse }>(url);
  return response.data.data.todos;
};

export const createTodo = async (newTodo: Omit<Todo, 'id' | 'created_at' | 'updated_at' | 'completed_at'>): Promise<Todo> => {
  const todoData = { ...newTodo };
  if (!todoData.list_id) {
    delete todoData.list_id;
  }

  const response = await axios.post<Todo>(
    `${API_BASE_URL}/todos`,
    todoData
  );
  return response.data;
};

export const updateTodo = async (id: string, updates: Partial<Todo>): Promise<Todo> => {
  const response = await axios.put<Todo>(
    `${API_BASE_URL}/todos/${id}`,
    updates
  );
  return response.data;
};

export const deleteTodo = async (id: string): Promise<void> => {
  await axios.delete(
    `${API_BASE_URL}/todos/${id}`
  );
};

export const completeTodo = async (id: string): Promise<Todo> => {
  const response = await axios.patch<{ data: Todo }>(`${API_BASE_URL}/todos/${id}/complete`);
  return response.data.data;
};

export const uncompleteTodo = async (id: string): Promise<Todo> => {
  const response = await axios.patch<{ data: Todo }>(`${API_BASE_URL}/todos/${id}/uncomplete`);
  return response.data.data;
};

// Habit API functions
export const fetchHabits = async (): Promise<Habit[]> => {
  const response = await axios.get<{ data: { habits: Habit[] } }>(`${API_BASE_URL}/habits`);
  return response.data.data.habits;
};

export const createHabit = async (data: Omit<Habit, 'id' | 'created_at' | 'updated_at' | 'current_streak' | 'longest_streak' | 'is_completed' | 'last_completed_date'>) => {
  return axios.post(`${API_BASE_URL}/habits`, data);
};

export const completeHabit = async (habitId: string) => {
  return axios.post(`${API_BASE_URL}/habits/${habitId}/complete`);
};

export const uncompleteHabit = async (habitId: string) => {
  return axios.post(`${API_BASE_URL}/habits/${habitId}/uncomplete`);
};

export const deleteHabit = async (habitId: string) => {
  return axios.delete(`${API_BASE_URL}/habits/${habitId}`);
};

export const updateHabit = async (habitId: string, title: string) => {
  return axios.put(
    `${API_BASE_URL}/habits/${habitId}`,
    { title }
  );
};

export const fetchHeatmapData = async (period: 'week' | 'month' | 'year' = 'year') => {
  const response = await axios.get(`${API_BASE_URL}/habits/heatmap`, {
    params: { period }
  });
  return response.data.data.data;
};

export const fetchTodoLists = async (): Promise<TodoList[]> => {
  const response = await axios.get<{ data: TodoListsResponse }>(`${API_BASE_URL}/todo-lists`);
  return response.data.data.lists;
};

export const createTodoList = async (data: CreateTodoListInput): Promise<TodoList> => {
  const response = await axios.post<{ data: TodoList }>(`${API_BASE_URL}/todo-lists`, data);
  return response.data.data;
};

export const updateTodoList = async (id: string, data: UpdateTodoListInput): Promise<TodoList> => {
  const response = await axios.put<{ data: TodoList }>(`${API_BASE_URL}/todo-lists/${id}`, data);
  return response.data.data;
};

export const deleteTodoList = async (id: string): Promise<void> => {
  await axios.delete(`${API_BASE_URL}/todo-lists/${id}`);
};
