import React from 'react';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import './EventCard.css';
import { CalendarEvent } from '../types';

interface EventCardProps {
  event: CalendarEvent & {
    occurrence_status?: 'Upcoming' | 'Cancelled' | 'Completed';
  };
  onClick: (event: CalendarEvent) => void;
  onDragStart?: (event: CalendarEvent, e: React.DragEvent) => void;
  style?: React.CSSProperties;
}

const EventCard: React.FC<EventCardProps> = ({ 
  event, 
  onClick, 
  onDragStart,
  style 
}) => {
  const handleDragStart = (e: React.DragEvent) => {
    if (onDragStart) {
      onDragStart(event, e);
      e.dataTransfer.setData('text/plain', ''); // Required for Firefox
    }
  };

  const formatTimeRange = (start: Date, end: Date) => {
    const startTime = format(new Date(start), 'h:mm a');
    const endTime = format(new Date(end), 'h:mm a');
    return `${startTime} - ${endTime}`;
  };

  const getStatusColor = (status?: 'Upcoming' | 'Cancelled' | 'Completed') => {
    switch (status) {
      case 'Cancelled':
        return 'text-red-500';
      case 'Completed':
        return 'text-green-500';
      default:
        return '';
    }
  };

  // Combine inline styles with the provided style prop
  const combinedStyle = {
    ...style,
    borderLeftColor: event.color || undefined
  };

  return (
    <div 
      className={cn(
        "event-card",
        event.occurrence_status === 'Cancelled' && "opacity-50",
        event.is_occurrence && "is-occurrence"
      )}
      draggable={!!onDragStart}
      onDragStart={handleDragStart}
      onClick={() => onClick(event)}
      style={combinedStyle}
    >
      <div className="event-title">
        {event.title || 'Untitled'}
        {event.occurrence_status && (
          <span className={cn("ml-2 text-xs", getStatusColor(event.occurrence_status))}>
            ({event.occurrence_status})
          </span>
        )}
      </div>
      <div className="event-time">
        {formatTimeRange(event.start_time, event.end_time)}
        {event.location && ` • ${event.location}`}
      </div>
    </div>
  );
};

export default EventCard;