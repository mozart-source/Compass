import React from 'react';
import { ChevronLeft, ChevronRight, Filter, Plus } from 'lucide-react';
import { format, isToday, isYesterday, isTomorrow } from 'date-fns';
import './CalendarHeader.css';

interface CalendarHeaderProps {
  currentDate: Date;
  onPrevious: () => void;
  onNext: () => void;
  onCreateEvent: () => void;
  view: string;
  onViewChange: (view: string) => void;
}

const CalendarHeader: React.FC<CalendarHeaderProps> = ({
  currentDate,
  onPrevious,
  onNext,
  onCreateEvent,
  view,
  onViewChange,
}) => {
  const views = [
    { label: 'D', value: 'day' },
    { label: '3D', value: '3day' },
    { label: 'W', value: 'week' },
    { label: 'M', value: 'month' },
  ];

  const getDateDisplay = (date: Date) => {
    if (isToday(date)) return 'Today';
    if (isTomorrow(date)) return 'Tomorrow';
    if (isYesterday(date)) return 'Yesterday';
    return format(date, 'MMMM d, yyyy');
  };

  const dateDisplay = getDateDisplay(currentDate);

  return (
    <div className="h-12 flex items-center justify-between px-6 bg-white border-b border-gray-200">      
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-3">
          <button
            onClick={onPrevious}
            className="p-1.5 rounded-lg hover:bg-gray-100 flex items-center justify-center"
          >
            <ChevronLeft className="w-5 h-5 text-gray-600" />
          </button>
          <span className="text-sm font-medium text-gray-900 min-w-[120px] text-center">
            {dateDisplay}
          </span>
          <button
            onClick={onNext}
            className="p-1.5 rounded-lg hover:bg-gray-100 flex items-center justify-center"
          >
            <ChevronRight className="w-5 h-5 text-gray-600" />
          </button>
        </div>

        <div className="view-buttons flex items-center">
          {views.map((v) => (
            <button
              key={v.value}
              data-view={v.value}
              onClick={() => onViewChange(v.value)}
              className={`view-button flex items-center justify-center ${view === v.value ? 'active' : ''}`}
            >
              {v.label}
            </button>
          ))}
          <div className="view-selector" />
        </div>
      </div>

      <div className="flex items-center gap-3">
        <button className="p-1.5 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg flex items-center justify-center">
          <Filter className="w-5 h-5" />
        </button>
        <button
          onClick={onCreateEvent}
          className="create-event-btn px-4 py-2 text-sm font-medium text-white bg-[#1f1f21] hover:bg-[#2f2f31] rounded-lg shadow-sm flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Create Event
        </button>
      </div>
    </div>
  );
};

export default CalendarHeader;