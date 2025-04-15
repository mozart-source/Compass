import React from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameMonth, isToday, startOfWeek, endOfWeek, isSameDay } from 'date-fns';
import { cn } from '@/lib/utils';
import './MonthView.css';
import { CalendarEvent } from '../types';
import { useMonthEvents, useUpdateEvent } from '@/components/calendar/hooks';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuth } from '@/hooks/useAuth';

interface MonthViewProps {
  date: Date;
  onEventClick: (event: CalendarEvent) => void;
  darkMode?: boolean;
}

const MonthView: React.FC<MonthViewProps> = ({ date, onEventClick, darkMode }) => {
  const { user } = useAuth();
  const monthStart = startOfMonth(date);
  const monthEnd = endOfMonth(date);
  const calendarStart = startOfWeek(monthStart);
  const calendarEnd = endOfWeek(monthEnd);
  const days = eachDayOfInterval({ start: calendarStart, end: calendarEnd });

  const { 
    data: events = [], 
    isLoading, 
    isError,
    error,
    refetch 
  } = useMonthEvents(user, date);


    // Expand recurring events into virtual events
    const expandedEvents = events.flatMap(event => {
      if (!event.occurrences || event.occurrences.length === 0) {
        return [event];
      }
  
      return event.occurrences.map(occurrence => {
        const occurrenceStart = new Date(occurrence.occurrence_time);
        let occurrenceEnd;
        
        // Use the overridden end_time if available, otherwise calculate based on original duration
        if (occurrence.end_time) {
          occurrenceEnd = new Date(occurrence.end_time);
        } else {
          const duration = new Date(event.end_time).getTime() - new Date(event.start_time).getTime();
          occurrenceEnd = new Date(occurrenceStart.getTime() + duration);
        }
  
        // Create a new event object with occurrence overrides
        return {
          ...event,
          id: `${event.id}-${occurrence.id}`,
          start_time: occurrenceStart,
          end_time: occurrenceEnd,
          // Override event properties if they exist in the occurrence
          title: occurrence.title || event.title,
          description: occurrence.description || event.description,
          location: occurrence.location || event.location,
          color: occurrence.color || event.color,
          transparency: occurrence.transparency || event.transparency,
          occurrence_id: occurrence.id,
          occurrence_status: occurrence.status,
          is_occurrence: true,
          original_event_id: event.id
        };
      });
    });

  if (isLoading) {
    return <div className="month-view"><Skeleton className="w-full h-full" /></div>;
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center h-full p-4">
        <div className={cn(
          "p-4 mb-4 rounded-md",
          darkMode ? "bg-red-900/20 text-red-200" : "bg-red-50 text-red-500"
        )}>
          {error instanceof Error ? error.message : 'Failed to load events'}
        </div>
        <button
          onClick={() => refetch()}
          className={cn(
            "px-4 py-2 rounded-md",
            darkMode 
              ? "bg-gray-700 hover:bg-gray-600 text-white" 
              : "bg-blue-500 hover:bg-blue-600 text-white"
          )}
        >
          Try Again
        </button>
      </div>
    );
  }

  return (
    <div className={cn("month-view", darkMode && "dark")}>
      <div className="month-container">
        <div className="month-grid">
          {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
            <div key={day} className="month-weekday-header">
              {day}
            </div>
          ))}

          {days.map(day => (
            <div 
              key={day.toISOString()}
              className={cn(
                "month-day-cell",
                !isSameMonth(day, date) && "month-other-day",
                isToday(day) && "month-today",
                darkMode && "dark"
              )}
            >
              <div className="month-day-header">
                <span className="month-day-number">{format(day, 'd')}</span>
              </div>
              <div className="month-day-events">
                {expandedEvents
                  .filter(event => {
                    const eventStart = new Date(event.start_time);
                    return isSameDay(eventStart, day);
                  })
                  .map((event: CalendarEvent) => (
                    <div 
                      key={event.id}
                      className={cn(
                        "month-event-pill",
                        event.is_occurrence && "month-event-occurrence",
                        darkMode && "dark"
                      )}
                      onClick={() => onEventClick(event)}
                      title={`${event.title}${event.location ? ` - ${event.location}` : ''}`}
                      style={{
                        '--occurrence-color': event.color,
                        '--occurrence-color-dark': event.color,
                        borderLeftColor: event.color
                      } as React.CSSProperties}
                    >
                      <div className="month-event-time">
                        {format(new Date(event.start_time), 'h:mm a')}
                      </div>
                      <div className="month-event-title">
                        {event.title}
                        {event.location && (
                          <span className="month-event-location"> • {event.location}</span>
                        )}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default MonthView;