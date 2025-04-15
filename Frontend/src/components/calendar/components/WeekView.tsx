import React, { useEffect } from 'react';
import { format, isSameDay, addDays } from 'date-fns';
import { cn } from '@/lib/utils';
import './WeekView.css';
import EventCard from './EventCard';
import { CalendarEvent } from '../types';
import { useWeekEvents, useUpdateEvent } from '@/components/calendar/hooks';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuth } from '@/hooks/useAuth';

interface WeekViewProps {
  date: Date;
  onEventClick: (event: CalendarEvent) => void;
  darkMode?: boolean;
}

const WeekView: React.FC<WeekViewProps> = ({ date, onEventClick, darkMode }) => {
  const [draggingEvent, setDraggingEvent] = React.useState<CalendarEvent | null>(null);
  const [currentTime, setCurrentTime] = React.useState(new Date());
  const { user } = useAuth();

  const { 
    data: events = [], 
    isLoading, 
    isError,
    error,
    refetch 
  } = useWeekEvents(user, date);
  
  const updateEventMutation = useUpdateEvent();

  useEffect(() => {
    const updateTimeIndicator = () => {
      setCurrentTime(new Date());
    };

    updateTimeIndicator();
    const interval = setInterval(updateTimeIndicator, 60000); // Update every minute

    return () => clearInterval(interval);
  }, []);

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

  const days = [date, addDays(date, 1), addDays(date, 2), addDays(date, 3), addDays(date, 4), addDays(date, 5), addDays(date, 6)];
  const timeSlots = Array.from({ length: 24 }, (_, i) => i);

  const handleDragStart = (event: CalendarEvent, e: React.DragEvent) => {
    setDraggingEvent(event);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = async (hour: number, e: React.DragEvent) => {
    e.preventDefault();
    if (!draggingEvent) return;

    const rect = (e.target as HTMLElement).getBoundingClientRect();
    const minutes = Math.floor(((e.clientY - rect.top) / rect.height) * 60);
    
    const newStart = new Date(draggingEvent.start_time);
    newStart.setHours(hour);
    newStart.setMinutes(minutes);

    const duration = draggingEvent.end_time.getTime() - draggingEvent.start_time.getTime();
    const newEnd = new Date(newStart.getTime() + duration);

    try {
      await updateEventMutation.mutateAsync({
        eventId: draggingEvent.id,
        event: {
          ...draggingEvent,
          start_time: newStart,
          end_time: newEnd,
        }
      });
    } catch (error) {
      console.error('Failed to update event:', error);
    }

    setDraggingEvent(null);
  };

  const getDurationInMinutes = (start: Date, end: Date): number => {
    return (end.getTime() - start.getTime()) / (1000 * 60);
  };

  if (isLoading) {
    return (
      <div className="week-view">
        <div className="week-container">
          <div className="days-header">
            <div className="time-label-header"></div>
            {Array(7).fill(null).map((_, i) => (
              <div key={i} className="day-header">
                <Skeleton className="h-11 w-20" />
              </div>
            ))}
          </div>
          <div className="time-slots">
            {Array(24).fill(null).map((_, hour) => (
              <div key={hour} className="time-row">
                <div className="time-label">
                  <Skeleton className="h-4 w-16" />
                </div>
                <div className="days-content">
                  {Array(7).fill(null).map((_, i) => (
                    <div key={i} className="day-column">
                      {Math.random() > 0.8 && (
                        <Skeleton 
                          className="absolute w-[calc(100%-8px)] rounded-md" 
                          style={{
                            height: `${Math.floor(Math.random() * 100 + 30)}px`,
                            top: `${Math.floor(Math.random() * 45)}px`
                          }}
                        />
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
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
            darkMode ? "bg-gray-800 text-white hover:bg-gray-700" : "bg-white text-gray-900 hover:bg-gray-50"
          )}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className={cn("week-view", darkMode && "dark")}>
      <div className="week-container">
        <div className="days-header">
          <div className="time-label-header"></div>
          {days.map(day => (
            <div 
              key={day.toISOString()} 
              className={cn(
                "day-header",
                isSameDay(day, new Date()) && "current-day"
              )}
            >
              <div className="day-name">{format(day, 'EEE')}</div>
              <div className="day-date">{format(day, 'MMM d')}</div>
            </div>
          ))}
        </div>
        <div className="time-slots">
          {timeSlots.map(hour => (
            <div key={hour} className="time-row">
              <div className="time-label">
                {format(new Date().setHours(hour, 0), 'h:mm a')}
              </div>
              <div className="days-content">
                {days.map(day => (
                  <div
                    key={day.toISOString()}
                    className={cn(
                      "day-column",
                      isSameDay(day, currentTime) && hour === currentTime.getHours() && "has-current-time"
                    )}
                    onDragOver={handleDragOver}
                    onDrop={(e) => handleDrop(hour, e)}
                  >
                    {isSameDay(day, currentTime) && hour === currentTime.getHours() && (
                      <div 
                        className="current-time-indicator"
                        style={{
                          top: `${currentTime.getMinutes()}px`,
                        }}
                      />
                    )}
                    {expandedEvents
                      .filter(event => {
                        const eventStart = new Date(event.start_time);
                        const eventHour = eventStart.getHours();
                        return eventHour === hour && isSameDay(eventStart, day);
                      })
                      .map((event: CalendarEvent) => (
                        <EventCard
                          key={event.id}
                          event={event}
                          onClick={onEventClick}
                          onDragStart={handleDragStart}
                          style={{
                            height: `${getDurationInMinutes(new Date(event.start_time), new Date(event.end_time))}px`,
                            top: `${new Date(event.start_time).getMinutes()}px`
                          }}
                        />
                      ))}
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

export default WeekView;