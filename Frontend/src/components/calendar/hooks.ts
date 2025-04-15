import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { 
  CalendarEvent, 
  CreateCalendarEventRequest, 
  UpdateCalendarEventRequest, 
  CalendarEventResponse
} from './types';
import { User } from '@/hooks/useAuth';
import { fetchEvents, createEvent, updateEvent, deleteEvent, updateOccurrenceById } from './api';
import { startOfMonth, endOfMonth, startOfDay, endOfDay, startOfWeek, addDays, endOfWeek } from 'date-fns';

// Calendar hooks
export const useEvents = (
  user: User | undefined, 
  date: Date = new Date(),
  viewType: 'day' | 'threeDays' | 'week' | 'month' = 'month'
) => {
  const startTime = viewType === 'day' ? startOfDay(date) : 
                   viewType === 'threeDays' ? startOfDay(date) :
                   viewType === 'week' ? startOfDay(date) :
                   startOfMonth(date);
                   
  const endTime = viewType === 'day' ? endOfDay(date) : 
                 viewType === 'threeDays' ? endOfDay(addDays(date, 2)) :
                 viewType === 'week' ? endOfDay(addDays(date, 6)) :
                 endOfMonth(date);

  return useQuery<CalendarEvent[]>({
    queryKey: ['events', user?.id, startTime, endTime, viewType],
    queryFn: () => user ? fetchEvents(startTime, endTime) : Promise.resolve([]),
    enabled: !!user,  
    gcTime: 0,
    staleTime: 0,
    refetchOnMount: true
  });
};

// Specific hook for day view
export const useDayEvents = (user: User | undefined, date: Date) => {
  return useEvents(user, date, 'day');
};

export const useThreeDayEvents = (user: User | undefined, date: Date) => {
  return useEvents(user, date, 'threeDays');
};

export const useWeekEvents = (user: User | undefined, date: Date) => {
  return useEvents(user, date, 'week');
};

export const useMonthEvents = (user: User | undefined, date: Date) => {
  return useEvents(user, date, 'month');
};

export const useCreateEvent = () => {
  const queryClient = useQueryClient();
  return useMutation<CalendarEventResponse, Error, CreateCalendarEventRequest>({
    mutationFn: (data: CreateCalendarEventRequest) => createEvent(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
};

export const useUpdateEvent = () => {
  const queryClient = useQueryClient();
  return useMutation<CalendarEventResponse, Error, { eventId: string; event: UpdateCalendarEventRequest }>({
    mutationFn: ({ eventId, event }) => updateEvent(eventId, event),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
};

export const useDeleteEvent = () => {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: deleteEvent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
};

export const useUpdateOccurrenceById = () => {
  const queryClient = useQueryClient();
  return useMutation<void, Error, { occurrenceId: string; updates: UpdateCalendarEventRequest }>({
    mutationFn: ({ occurrenceId, updates }) => updateOccurrenceById(occurrenceId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
};

