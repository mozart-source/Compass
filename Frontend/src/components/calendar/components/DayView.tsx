import React, { useState, useEffect } from 'react';
import { format, isSameDay } from 'date-fns';
import './DayView.css';
import { cn } from '@/lib/utils';
import EventCard from './EventCard';
import { CalendarEvent } from '../types';
import { useDayEvents, useUpdateEvent } from '@/components/calendar/hooks';
import { Skeleton } from '@/components/ui/skeleton';
import { useAuth } from '@/hooks/useAuth';

interface DayViewProps {
  date: Date;
  onEventClick: (event: CalendarEvent) => void;
  onEventDrop?: (event: CalendarEvent, hour: number, minutes: number) => void;
  darkMode: boolean;
}

const DayView: React.FC<DayViewProps> = ({ date, onEventClick, onEventDrop, darkMode }) => {
  const [draggingEvent, setDraggingEvent] = React.useState<CalendarEvent | null>(null);
  const [currentTime, setCurrentTime] = useState<Date>(new Date());
  const { user } = useAuth();
  
  const { 
    data: events = [], 
    isLoading, 
    isError,
    error,
    refetch 
  } = useDayEvents(user, date);
  
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

  const todayEvents = expandedEvents.filter(event => isSameDay(new Date(event.start_time), date));
  const sortedEvents = todayEvents.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
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

  if (isLoading) {
    return <DayViewSkeleton />;
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
    <div className="day-view">
      <div className="day-container">
        <div className="day-header">
          <div className="time-label-header"></div>
          <div className={cn(
            "date-header",
            isSameDay(date, new Date()) && "current-day"
          )}>
            <div className="day-name">{format(date, 'EEEE')}</div>
            <div className="day-date">{format(date, 'd MMMM, yyyy')}</div>
          </div>
        </div>
        <div className="time-slots">
          {timeSlots.map(hour => (
            <div key={hour} className="time-slot">
              <div className="time-label">
                {format(new Date().setHours(hour, 0), 'h:mm a')}
              </div>
              <div className="time-content">
                <div
                  className={cn(
                    "day-column",
                    hour === currentTime.getHours() && isSameDay(date, currentTime) && "has-current-time"
                  )}
                  onDragOver={handleDragOver}
                  onDrop={(e) => handleDrop(hour, e)}
                >
                  {isSameDay(date, currentTime) && hour === currentTime.getHours() && (
                    <div 
                      className="current-time-indicator"
                      style={{
                        top: `${currentTime.getMinutes()}px`,
                      }}
                    />
                  )}
                  {sortedEvents
                    .filter(event => {
                      const eventStart = new Date(event.start_time);
                      return eventStart.getHours() === hour && isSameDay(eventStart, date);
                    })
                    .map(event => (
                      <EventCard
                        key={event.id}
                        event={event}
                        onClick={onEventClick}
                        onDragStart={handleDragStart}
                        style={{
                          top: `${new Date(event.start_time).getMinutes()}px`,
                          height: `${getDurationInMinutes(new Date(event.start_time), new Date(event.end_time))}px`,
                        }}
                      />
                    ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

const getDurationInMinutes = (start: Date, end: Date): number => {
  return (end.getTime() - start.getTime()) / (1000 * 60);
};

const DayViewSkeleton = () => {
  const timeSlots = Array(24).fill(null);

  return (
    <div className="day-view">
      <div className="day-container">
        <div className="day-header">
          <div className="time-label-header"></div>
          <div className="date-header">
            <Skeleton className="h-11 w-32" />
          </div>
        </div>
        <div className="time-slots">
          {timeSlots.map((_, hour) => (
            <div key={hour} className="time-slot">
              <div className="time-label">
                <Skeleton className="h-4 w-16" />
              </div>
              <div className="time-content">
                <div className="day-column">
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
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default DayView;