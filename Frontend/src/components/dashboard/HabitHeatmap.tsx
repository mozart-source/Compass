import { useEffect, useState } from "react";
import { useHabitHeatmap } from "@/hooks/useHabitHeatmap";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import CustomHeatmapCard from "./CustomHeatmapCard";

interface HeatmapProps {
  className?: string;
}

export default function HabitHeatmap({ className }: HeatmapProps) {
  const { data, loading, error, refreshHeatmap } = useHabitHeatmap();
  const [activeData, setActiveData] = useState(data);
  const [currentDate] = useState(new Date());

  // Define a constant for horizontal gap size
  const X_GAP = "8px";

  useEffect(() => {
    setActiveData(data);
  }, [data]);

  // Determine color intensity based on completion count
  const getColorClass = (count: number) => {
    if (count === 0) return "bg-zinc-800 hover:bg-zinc-700"; // No completions
    if (count === 1) return "bg-blue-900 hover:bg-blue-800"; // 1 completion
    if (count <= 3) return "bg-blue-700 hover:bg-blue-600"; // 2-3 completions
    if (count <= 5) return "bg-blue-500 hover:bg-blue-400"; // 4-5 completions
    return "bg-blue-400 hover:bg-blue-300"; // 6+ completions
  };

  // Generate dates for the selected period
  const generateDates = () => {
    const dates: string[] = [];

    // For month view, use the month from currentDate
    const startOfMonth = new Date(
      currentDate.getFullYear(),
      currentDate.getMonth(),
      1
    );
    const endOfMonth = new Date(
      currentDate.getFullYear(),
      currentDate.getMonth() + 1,
      0
    );
    const daysInMonth = endOfMonth.getDate();

    // Generate all days in the current month
    for (let i = 0; i < daysInMonth; i++) {
      const date = new Date(startOfMonth);
      date.setDate(startOfMonth.getDate() + i);

      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, "0");
      const day = String(date.getDate()).padStart(2, "0");
      const formattedDate = `${year}-${month}-${day}`;
      dates.push(formattedDate);
    }

    return dates;
  };

  // Calculate calendar view with weekday alignment
  const calculateCalendarData = () => {
    const dates = generateDates();
    const weeks: string[][] = [];

    // For month view, need to align with the weekdays
    let currentWeek: string[] = [];

    // Fill in empty cells at the beginning
    const firstDate = new Date(dates[0]);
    const firstDayOfWeek = firstDate.getDay(); // 0 for Sunday, 1 for Monday, etc.

    // Add empty strings for days before the start date
    for (let i = 0; i < firstDayOfWeek; i++) {
      currentWeek.push("");
    }

    // Populate the dates
    dates.forEach((date) => {
      currentWeek.push(date);

      if (currentWeek.length === 7) {
        weeks.push([...currentWeek]);
        currentWeek = [];
      }
    });

    // Add the last partial week if needed
    if (currentWeek.length > 0) {
      while (currentWeek.length < 7) {
        currentWeek.push("");
      }
      weeks.push([...currentWeek]);
    }

    return weeks;
  };

  const weeks = calculateCalendarData();

  // Function to handle refresh button click
  const handleRefresh = () => {
    refreshHeatmap();
  };

  return (
    <CustomHeatmapCard
      className={cn("w-[230px] rounded-3xl p-5 shadow-lg", className)}
    >
      <div className="flex flex-col">
        <div className="flex flex-col mb-4">
          <span className="text-xs text-zinc-500 mt-1">This month</span>
          <div className="flex justify-between items-center">
            <span className="text-xl font-medium text-zinc-200">
              Habit's Activity
            </span>
            {loading && (
              <span className="ml-2 text-sm text-zinc-500">(Loading...)</span>
            )}
            {error && (
              <span className="ml-2 text-sm text-red-500">
                Error loading data
              </span>
            )}
            <button
              onClick={handleRefresh}
              className="text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
              aria-label="Refresh heatmap data"
            >
              Refresh
            </button>
          </div>
        </div>
        <div className="flex items-center justify-center">
          <div className="w-full">
            <div className="grid grid-cols-7" style={{ columnGap: X_GAP }}>
              {["S", "M", "T", "W", "T", "F", "S"].map((day, index) => (
                <div
                  key={index}
                  className="w-5 text-center text-xs text-zinc-500"
                >
                  {day}
                </div>
              ))}
            </div>
            <div className="mt-2">
              {weeks.map((week, weekIndex) => (
                <div
                  key={weekIndex}
                  className="grid grid-cols-7"
                  style={{ columnGap: X_GAP, marginBottom: "6px" }}
                >
                  {week.map((date, dateIndex) => {
                    const completions = date ? activeData[date] || 0 : 0;

                    return (
                      <TooltipProvider key={dateIndex}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div
                              className={cn(
                                "h-5 w-5 rounded-md flex items-center justify-center transition-colors cursor-pointer",
                                date
                                  ? getColorClass(completions)
                                  : "bg-transparent"
                              )}
                            />
                          </TooltipTrigger>
                          {date && (
                            <TooltipContent side="top">
                              <div className="text-xs">
                                <p className="font-medium">
                                  {new Date(date).toLocaleDateString("en-US", {
                                    weekday: "long",
                                    year: "numeric",
                                    month: "long",
                                    day: "numeric",
                                  })}
                                </p>
                                <p>
                                  {completions}{" "}
                                  {completions === 1
                                    ? "completion"
                                    : "completions"}
                                </p>
                              </div>
                            </TooltipContent>
                          )}
                        </Tooltip>
                      </TooltipProvider>
                    );
                  })}
                </div>
              ))}
            </div>
            <div className="mt-6 flex justify-center">
              <div className="flex items-center gap-1.5 text-xs text-zinc-500">
                <span>Less</span>
                <div className="h-3 w-3 rounded-sm bg-zinc-800" />
                <div className="h-3 w-3 rounded-sm bg-blue-900" />
                <div className="h-3 w-3 rounded-sm bg-blue-700" />
                <div className="h-3 w-3 rounded-sm bg-blue-500" />
                <div className="h-3 w-3 rounded-sm bg-blue-400" />
                <span>More</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </CustomHeatmapCard>
  );
}
