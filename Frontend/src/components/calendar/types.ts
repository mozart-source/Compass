export type EventType = 'None' | 'Task' | 'Meeting' | 'Todo' | 'Holiday' | 'Reminder';

export type RecurrenceType = 'None' | 'Daily' | 'Weekly' | 'Biweekly' | 'Monthly' | 'Yearly' | 'Custom';

export type OccurrenceStatus = 'Upcoming' | 'Cancelled' | 'Completed';

export type NotificationMethod = 'Email' | 'Push' | 'SMS';

export type Transparency = 'opaque' | 'transparent';

export interface CalendarEvent {
  id: string;
  user_id: string;
  title: string;
  description: string;
  event_type: EventType;
  start_time: Date;
  end_time: Date;
  is_all_day: boolean;
  location?: string;
  color?: string;
  transparency: Transparency;
  priority?: 'High' | 'Medium' | 'Low';
  status?: 'Upcoming' | 'In Progress' | 'Completed' | 'Cancelled' | 'Blocked' | 'Under Review' | 'Deferred';
  is_recurring?: boolean;
  recurrence?: RecurrenceType;
  recurrence_end_date?: Date;
  created_at: Date;
  updated_at: Date;
  recurrence_rules?: RecurrenceRule[];
  occurrences?: OccurrenceResponse[];
  exceptions?: EventException[];
  reminders?: EventReminder[];
  is_occurrence?: boolean;
  original_event_id?: string;
  occurrence_id?: string;
  occurrence_status?: OccurrenceStatus;
}

export interface CreateEventData {
  title: string;
  description: string;
  start_time: Date;
  end_time: Date;
  is_all_day?: boolean;
  location?: string;
  color?: string;
  transparency?: 'opaque' | 'transparent';
  priority?: 'High' | 'Medium' | 'Low';
  status?: 'Upcoming' | 'In Progress' | 'Completed' | 'Cancelled' | 'Blocked' | 'Under Review' | 'Deferred';
  is_recurring?: boolean;
  recurrence?: 'None' | 'Daily' | 'Weekly' | 'Biweekly' | 'Monthly' | 'Yearly' | 'Weekdays' | 'Custom';
  recurrence_end_date?: Date;
}

export interface RecurrenceRule {
  id: string;
  event_id: string;
  freq: RecurrenceType;
  interval: number;
  by_day?: string[];
  by_month?: number[];
  by_month_day?: number[];
  count?: number;
  until?: Date;
  created_at: Date;
  updated_at: Date;
}

export interface OccurrenceResponse {
  id: string;
  event_id: string;
  occurrence_time: Date;
  status: OccurrenceStatus;
  created_at: Date;
  updated_at: Date;
  title?: string;
  description?: string;
  location?: string;
  color?: string;
  transparency?: Transparency;
  end_time?: Date;
}

export interface EventException {
  id: string;
  event_id: string;
  original_time: Date;
  is_deleted: boolean;
  override_start_time?: Date;
  override_end_time?: Date;
  override_title?: string;
  override_description?: string;
  override_location?: string;
  override_color?: string;
  override_transparency?: Transparency;
  created_at: Date;
  updated_at: Date;
}

export interface EventReminder {
  id: string;
  event_id: string;
  minutes_before: number;
  method: NotificationMethod;
  created_at: Date;
  updated_at: Date;
}

export interface CalendarViewProps {
  date: Date;
  events: CalendarEvent[];
  onEventClick: (event: CalendarEvent) => void;
}

export interface CalendarHeaderProps {
  view: 'day' | 'threeDays' | 'week' | 'month';
  onViewChange: (view: 'day' | 'threeDays' | 'week' | 'month') => void;
  currentDate: Date;
  onNavigate: (direction: 'prev' | 'next' | 'today') => void;
}

export interface EventFormProps {
  event?: CalendarEvent;
  onClose: () => void;
  onSubmit: (event: CalendarEvent) => void;
  onDelete?: (eventId: string) => void;
}

export type CalendarView = 'sync' | 'schedule' | 'notes';

// Request/Response DTOs
export interface CreateCalendarEventRequest {
  title: string;
  description: string;
  event_type: EventType;
  start_time: Date;
  end_time: Date;
  is_all_day?: boolean;
  location?: string;
  color?: string;
  transparency?: Transparency;
  priority?: 'High' | 'Medium' | 'Low';
  status?: 'Upcoming' | 'In Progress' | 'Completed' | 'Cancelled' | 'Blocked' | 'Under Review' | 'Deferred';
  is_recurring?: boolean;
  recurrence?: RecurrenceType;
  recurrence_end_date?: Date;
  recurrence_rule?: CreateRecurrenceRuleRequest;
  reminders?: CreateEventReminderRequest[];
}

export interface CreateRecurrenceRuleRequest {
  freq: RecurrenceType;
  interval: number;
  by_day?: string[];
  by_month?: number[];
  by_month_day?: number[];
  count?: number;
  until?: Date;
}

export interface CreateEventReminderRequest {
  minutes_before: number;
  method: NotificationMethod;
}

export interface UpdateCalendarEventRequest {
  title?: string;
  description?: string;
  event_type?: EventType;
  start_time?: Date;
  end_time?: Date;
  is_all_day?: boolean;
  location?: string;
  color?: string;
  transparency?: Transparency;
  priority?: 'High' | 'Medium' | 'Low';
  status?: 'Upcoming' | 'In Progress' | 'Completed' | 'Cancelled' | 'Blocked' | 'Under Review' | 'Deferred';
  is_recurring?: boolean;
  recurrence?: RecurrenceType;
  recurrence_end_date?: Date;
  preserve_date_sequence?: boolean;
}

export interface CalendarEventResponse {
  event: CalendarEvent;
  occurrences?: OccurrenceResponse[];
}
