"use client";
import React from "react";
import { useDashboardMetrics } from "./useDashboardMetrics";

type Event = {
  id: string;
  title: string;
  startTime: string;
  endTime: string;
  date: Date;
  isPurple?: boolean;
  isCompleted?: boolean;
};

type Day = {
  date: Date;
  dayOfWeek: string;
  dayOfMonth: number;
  events: Event[];
};

const CalendarWidget: React.FC = () => {
  const {
    data: metricsData,
    isLoading,
    isConnected,
    requestRefresh,
  } = useDashboardMetrics();

  // Get current month name
  const currentMonth = new Date().toLocaleString("default", {
    month: "long",
  });

  // Format time helper function
  const formatTime = (date: Date) => {
    let hours = date.getHours();
    const minutes = date.getMinutes().toString().padStart(2, "0");
    const ampm = hours >= 12 ? "PM" : "AM";
    hours = hours % 12;
    hours = hours ? hours : 12; // the hour '0' should be '12'
    return `${hours}:${minutes} ${ampm}`;
  };

  // Process timeline data into days and events
  const days = React.useMemo(() => {
    // Initialize with current day even if no events
    const today = new Date();
    const result: Day[] = [
      {
        date: today,
        dayOfWeek: today
          .toLocaleString("default", { weekday: "short" })
          .substring(0, 2),
        dayOfMonth: today.getDate(),
        events: [],
      },
    ];

    // If no timeline data, return just today with no events
    if (!metricsData?.daily_timeline?.length) {
      return result;
    }

    // Filter timeline to only include calendar events
    const calendarEvents = metricsData.daily_timeline.filter(
      (item) => item.type === "event"
    );

    // Only include upcoming events (today and future)
    const now = new Date();
    now.setHours(0, 0, 0, 0); // Set to start of day for proper comparison

    const upcomingEvents = calendarEvents.filter((item) => {
      const eventDate = new Date(item.start_time);
      return eventDate >= now;
    });

    // Sort by date (soonest first) and limit to 4 events
    upcomingEvents.sort((a, b) => {
      return (
        new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
      );
    });

    const limitedEvents = upcomingEvents.slice(0, 4);

    // Group events by day
    const eventsByDay = new Map<string, Event[]>();

    limitedEvents.forEach((item) => {
      const startTime = new Date(item.start_time);
      const endTime = item.end_time
        ? new Date(item.end_time)
        : new Date(startTime);
      const dateStr = startTime.toDateString();

      const event: Event = {
        id: item.id,
        title: item.title,
        startTime: formatTime(startTime),
        endTime: formatTime(endTime),
        date: startTime,
        isPurple: item.is_completed,
        isCompleted: item.is_completed,
      };

      if (!eventsByDay.has(dateStr)) {
        eventsByDay.set(dateStr, []);
      }
      eventsByDay.get(dateStr)?.push(event);
    });

    // Create day objects
    const daysWithEvents: Day[] = [];
    eventsByDay.forEach((events, dateStr) => {
      const date = new Date(dateStr);
      daysWithEvents.push({
        date,
        dayOfWeek: date
          .toLocaleString("default", { weekday: "short" })
          .substring(0, 2),
        dayOfMonth: date.getDate(),
        events,
      });
    });

    // Sort days chronologically
    daysWithEvents.sort((a, b) => a.date.getTime() - b.date.getTime());

    // If no days with events, just return today
    if (daysWithEvents.length === 0) {
      return result;
    }

    return daysWithEvents; // Show all days with events (max 4 events total)
  }, [metricsData?.daily_timeline]);

  const handleAddEvent = () => {
    // Function to add a new event
    console.log("Add event clicked");
  };

  const handleEventClick = (event: Event) => {
    // Function to handle event click
    console.log("Event clicked:", event);
  };

  const handleKeyDown = (e: React.KeyboardEvent, callback: () => void) => {
    if (e.key === "Enter" || e.key === " ") {
      callback();
    }
  };

  // Handle refresh button click
  const handleRefresh = () => {
    requestRefresh();
  };

  return (
    <div className="bg-[#18191b] border rounded-3xl p-5 w-[350px] text-white shadow-lg">
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center">
          <h2 className="text-2xl font-semibold">{currentMonth}</h2>
          {/* Connection indicator */}
          <div
            className={`ml-2 w-2 h-2 rounded-full ${
              isConnected ? "bg-green-500" : "bg-gray-400"
            }`}
            title={isConnected ? "Connected" : "Disconnected"}
          />
          {isLoading && (
            <span className="ml-2 text-sm text-zinc-500">(Loading...)</span>
          )}
        </div>
        <div className="flex items-center">
          <button
            onClick={handleRefresh}
            className="mr-2 text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
            aria-label="Refresh calendar data"
          >
            Refresh
          </button>
          <button
            aria-label="Add event"
            className="bg-[#303030] w-8 h-8 rounded-full flex items-center justify-center hover:bg-[#404040] transition-colors"
            onClick={handleAddEvent}
            onKeyDown={(e) => handleKeyDown(e, handleAddEvent)}
            tabIndex={0}
          >
            <span className="text-xl relative top-[-3px]">+</span>
          </button>
        </div>
      </div>

      <div className="space-y-4">
        {days.map((day) => (
          <div
            key={day.dayOfMonth + "-" + day.date.getMonth()}
            className="flex"
          >
            <div className="flex flex-col items-center mr-4 w-10">
              <span className="text-xs text-gray-400">{day.dayOfWeek}</span>
              <span className="text-lg font-medium">{day.dayOfMonth}</span>
            </div>

            <div className="flex-1">
              {isLoading ? (
                <div className="bg-[#252525] rounded-lg p-3 text-gray-400 text-sm">
                  Loading events...
                </div>
              ) : day.events.length === 0 ? (
                <div className="bg-[#252525] rounded-lg p-3 text-gray-400 text-sm">
                  Nothing Scheduled
                </div>
              ) : (
                <div className="space-y-2">
                  {day.events.map((event) => (
                    <div
                      key={event.id}
                      className={`rounded-lg p-3 cursor-pointer ${
                        event.isPurple ? "bg-[#1d4ed8]" : "bg-[#252525]"
                      }`}
                      onClick={() => handleEventClick(event)}
                      onKeyDown={(e) =>
                        handleKeyDown(e, () => handleEventClick(event))
                      }
                      tabIndex={0}
                      aria-label={`${event.title} from ${event.startTime} to ${
                        event.endTime
                      }${event.isCompleted ? ", completed" : ""}`}
                    >
                      <div className="text-sm font-medium">{event.title}</div>
                      <div className="text-xs opacity-80">
                        {event.startTime} - {event.endTime}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CalendarWidget;
