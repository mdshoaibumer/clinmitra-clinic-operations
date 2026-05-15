import { format, parse } from "date-fns"
import { CalendarIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { useState } from "react"

interface DatePickerProps {
  /** Date value in YYYY-MM-DD string format */
  value: string
  /** Callback with YYYY-MM-DD string */
  onChange: (date: string) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function DatePicker({
  value,
  onChange,
  placeholder = "Pick a date",
  className,
  disabled,
}: DatePickerProps) {
  const [open, setOpen] = useState(false)

  const dateValue = value
    ? parse(value, "yyyy-MM-dd", new Date())
    : undefined

  const handleSelect = (day: Date | undefined) => {
    if (day) {
      onChange(format(day, "yyyy-MM-dd"))
      setOpen(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "w-full justify-start text-left font-normal h-10",
            !value && "text-muted-foreground",
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {value ? format(dateValue!, "dd MMM yyyy") : placeholder}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={dateValue}
          onSelect={handleSelect}
          initialFocus
        />
      </PopoverContent>
    </Popover>
  )
}

interface MonthPickerProps {
  /** Value in YYYY-MM format */
  value: string
  onChange: (value: string) => void
  className?: string
}

export function MonthPicker({ value, onChange, className }: MonthPickerProps) {
  const [open, setOpen] = useState(false)

  const year = value ? parseInt(value.split("-")[0]) : new Date().getFullYear()
  const month = value ? parseInt(value.split("-")[1]) - 1 : new Date().getMonth()

  const months = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December",
  ]

  const [selectedYear, setSelectedYear] = useState(year)

  const handleSelect = (monthIndex: number) => {
    const m = String(monthIndex + 1).padStart(2, "0")
    onChange(`${selectedYear}-${m}`)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "w-full justify-start text-left font-normal h-10",
            !value && "text-muted-foreground",
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {value ? `${months[month]} ${year}` : "Pick a month"}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-4" align="start">
        <div className="flex items-center justify-between mb-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSelectedYear(selectedYear - 1)}
          >
            ←
          </Button>
          <span className="text-sm font-medium">{selectedYear}</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSelectedYear(selectedYear + 1)}
          >
            →
          </Button>
        </div>
        <div className="grid grid-cols-3 gap-2">
          {months.map((name, i) => (
            <Button
              key={name}
              variant={i === month && selectedYear === year ? "default" : "ghost"}
              size="sm"
              className="text-xs"
              onClick={() => handleSelect(i)}
            >
              {name.substring(0, 3)}
            </Button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
