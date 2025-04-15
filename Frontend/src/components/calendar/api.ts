import axios from 'axios';
import { 
  CalendarEvent, 
  CreateCalendarEventRequest, 
  UpdateCalendarEventRequest, 
  CalendarEventResponse 
} from './types';
import { getApiUrls } from '@/config';

const { GO_API_URL } = getApiUrls();
export const API_BASE_URL = `${GO_API_URL}/calendar`;

export const fetchEvents = async (startTime: Date, endTime: Date): Promise<CalendarEvent[]> => {
  const response = await axios.get<{ events: CalendarEvent[]; total: number }>(`${API_BASE_URL}/events`, {
    params: {
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
    }
  });
  return response.data.events;
};

export const createEvent = async (data: CreateCalendarEventRequest): Promise<CalendarEventResponse> => {
  const response = await axios.post<CalendarEventResponse>(`${API_BASE_URL}/events`, data);
  return response.data;
};

export const updateEvent = async (eventId: string, data: UpdateCalendarEventRequest): Promise<CalendarEventResponse> => {
  const response = await axios.put<CalendarEventResponse>(`${API_BASE_URL}/events/${eventId}`, data);
  return response.data;
};

export const updateOccurrenceById = async (
  occurrenceId: string, 
  updates: UpdateCalendarEventRequest
): Promise<void> => {
  await axios.put(`${API_BASE_URL}/events/occurrences/${occurrenceId}`, updates);
};

export const deleteEvent = async (eventId: string): Promise<void> => {
  const response = await axios.delete(`${API_BASE_URL}/events/${eventId}`);
  return response.data;
};


