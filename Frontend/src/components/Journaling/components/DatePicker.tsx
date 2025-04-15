import * as React from "react"
import { Calendar } from "@/components/ui/calendar"
import { Button } from "@/components/ui/button"
import { Calendar as CalendarIcon } from "lucide-react"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

export interface DatePickerProps {
  date?: Date
  onSelect: (date: Date | undefined) => void
  highlightedDates?: Date[]
}

export function DatePicker({ date, onSelect, highlightedDates }: DatePickerProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          className={date ? "font-semibold" : ""}
        >
          <CalendarIcon className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="end">
        <Calendar
          mode="single"
          selected={date}
          onSelect={onSelect}
          modifiers={{ highlighted: highlightedDates || [] }}
          modifiersStyles={{
            highlighted: { backgroundColor: "var(--primary)" }
          }}
        />
      </PopoverContent>
    </Popover>
  )
} 