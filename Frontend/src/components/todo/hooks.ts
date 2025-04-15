import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Todo, TodoStatus, TodoList, TodoListsResponse, CreateTodoListInput, UpdateTodoListInput } from './types-todo';
import { Habit } from './types-habit';
import { User } from '@/api/auth';
import { fetchTodos, createTodo, updateTodo, deleteTodo, fetchHabits, createHabit, completeHabit, uncompleteHabit, deleteHabit, updateHabit, completeTodo, uncompleteTodo, fetchTodoLists, createTodoList, updateTodoList, deleteTodoList } from './api';

// Todo hooks
export const useTodos = (user: User | undefined) => {
  return useQuery<Todo[]>({
    queryKey: ['todos', user?.id],
    queryFn: () => user ? fetchTodos(user.id) : Promise.resolve([]),
    enabled: !!user,
    gcTime: 0,
    staleTime: 0,
    refetchOnMount: true,
  });
};

export const useCreateTodo = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTodo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

export const useUpdateTodo = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<Todo> }) =>
      updateTodo(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

export const useDeleteTodo = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTodo,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

export const useToggleTodoStatus = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (todo: Todo) =>
      todo.is_completed ? 
        uncompleteTodo(todo.id) : 
        completeTodo(todo.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

// Habit hooks
export const useHabits = (user: User | undefined) => {
  return useQuery<Habit[]>({
    queryKey: ['habits', user?.id],
    queryFn: () => user ? fetchHabits() : Promise.resolve([]),
    enabled: !!user,
    gcTime: 0,
    staleTime: 0,
    refetchOnMount: true
  });
};

export const useCreateHabit = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createHabit,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
    },
  });
};

export const useToggleHabit = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ habitId, isCompleted }: { habitId: string; isCompleted: boolean }) =>
      isCompleted ? uncompleteHabit(habitId) : completeHabit(habitId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
    },
  });
};

export const useDeleteHabit = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteHabit,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
    },
  });
};

interface UpdateHabitData {
  habitId: string;
  title: string;
}

export const useUpdateHabit = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateHabitData) => updateHabit(data.habitId, data.title),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['habits'] });
    },
  });
};

// Todo List hooks
export const useTodoLists = (user: User | undefined) => {
  return useQuery<TodoList[]>({
    queryKey: ['todoLists', user?.id],
    queryFn: fetchTodoLists,
    enabled: !!user,
    gcTime: 0,
    staleTime: 0,
    refetchOnMount: true,
  });
};

export const useCreateTodoList = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createTodoList,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

export const useUpdateTodoList = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTodoListInput }) =>
      updateTodoList(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};

export const useDeleteTodoList = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTodoList,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todoLists'] });
    },
  });
};
