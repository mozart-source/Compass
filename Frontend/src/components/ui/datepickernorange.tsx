"use client"

import * as React from "react"
import { format } from "date-fns"
import { CalendarIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { useTranslation } from 'react-i18next'

interface CalendarFormProps extends React.HTMLAttributes<HTMLDivElement> {
  selected: Date | null
  onSelect: (date: Date | null) => void
  darkMode?: boolean
}

export function CalendarForm({
  className,
  selected,
  onSelect,
  darkMode = false,
}: CalendarFormProps) {
  const { i18n } = useTranslation();
  
  const handleSelect = (date: Date | null) => {
    if (date) {
      // Set the time to noon to avoid timezone issues
      const adjustedDate = new Date(date.setHours(12, 0, 0, 0));
      onSelect(adjustedDate);
    } else {
      onSelect(null);
    }
  };

  return (
    <div className={className}>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant={"outline"}
            className={cn(
              "w-full h-[42px] justify-start text-left font-normal",
              !selected && "text-muted-foreground",
              darkMode
                ? "bg-[#1c1c1e] border-[#3c3c3e] text-white hover:bg-[#2c2c2e] hover:text-white focus:border-[#0A84FF] focus:ring-[#0A84FF]"
                : "bg-white border-gray-300 text-gray-900 hover:bg-gray-50 hover:text-gray-900 focus:border-blue-500 focus:ring-blue-500"
            )}
          >
            <CalendarIcon className={cn(
              "h-4 w-4 opacity-50",
              darkMode ? "text-[#86868b]" : "text-gray-700",
              i18n.dir() === 'rtl' ? 'ml-0' : 'mr-0'
            )} />
            {selected ? (
              format(selected, "dd/MM/yyyy")
            ) : (
              <span>Pick a date</span>
            )}
          </Button>
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
            selected={selected}
            onSelect={handleSelect}
            initialFocus
            darkMode={darkMode}
          />
        </PopoverContent>
      </Popover>
    </div>
  )
}
