import React from 'react';
import { format } from "date-fns";
import { Calendar as CalendarIcon } from "lucide-react";
import { cn } from "../../lib/utils";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";
import { Calendar } from "./calendar";
import { useTranslation } from 'react-i18next'

interface CompactDatePickerProps {
  value: Date;
  onChange: (date: Date) => void;
  darkMode?: boolean;
  className?: string;
}

export function CompactDatePicker({
  value,
  onChange,
  darkMode = false,
  className
}: CompactDatePickerProps) {
  const { i18n } = useTranslation();

  const handleDateSelect = (date: Date | undefined) => {
    if (date) {
      // Create a new date at noon UTC to avoid timezone issues
      const adjustedDate = new Date(Date.UTC(
        date.getFullYear(),
        date.getMonth(),
        date.getDate(),
        12, 0, 0, 0
      ));
      onChange(adjustedDate);
    }
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          className={cn(
            "inline-flex items-center justify-start rounded-md border text-sm transition-colors",
            "px-2 py-1 h-8",
            darkMode 
              ? "bg-[#3c3c3e] border-[#3c3c3e] text-white hover:bg-[#4c4c4e]" 
              : "bg-white border-gray-300 text-gray-900 hover:bg-gray-50",
            className
          )}
        >
          <CalendarIcon className={cn(
            "h-4 w-4 opacity-50",
            i18n.dir() === 'rtl' ? 'ml-2' : 'mr-2'
           )} />
          <span>{format(value, "dd/MM/yyyy")}</span>
        </button>
      </PopoverTrigger>
      <PopoverContent 
        className={cn(
          "w-auto p-0",
          darkMode && "bg-[#2c2c2e] border-[#3c3c3e]"
        )}
        align="start"
      >
        <Calendar
          mode="single"
          selected={value}
          onSelect={handleDateSelect}
          initialFocus
          darkMode={darkMode}
        />
      </PopoverContent>
    </Popover>
  );
}
