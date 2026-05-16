import { useMemo } from 'react'
import { formatTime, cn } from '@/lib/utils'
import type { Appointment } from '@/types/models'

interface WeekViewProps {
  appointments: Appointment[]
  weekStart: string // YYYY-MM-DD
  onComplete: (id: string) => void
  onCancel: (id: string) => void
}

const HOURS = Array.from({ length: 12 }, (_, i) => i + 8) // 8 AM to 7 PM

function getWeekDays(startDate: string): { date: string; dayName: string; dayNum: number; isToday: boolean }[] {
  const start = new Date(startDate)
  const today = new Date().toISOString().split('T')[0]
  const days = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(start)
    d.setDate(d.getDate() + i)
    const dateStr = d.toISOString().split('T')[0]
    days.push({
      date: dateStr,
      dayName: d.toLocaleDateString('en-IN', { weekday: 'short' }),
      dayNum: d.getDate(),
      isToday: dateStr === today,
    })
  }
  return days
}

export default function WeekView({ appointments, weekStart, onComplete, onCancel }: WeekViewProps) {
  const days = useMemo(() => getWeekDays(weekStart), [weekStart])

  const appointmentsByDay = useMemo(() => {
    const map: Record<string, Appointment[]> = {}
    days.forEach(d => { map[d.date] = [] })
    appointments.forEach(apt => {
      const date = apt.appointmentDate?.split('T')[0] || apt.appointmentDate
      if (map[date]) {
        map[date].push(apt)
      }
    })
    // Sort each day by start time
    Object.keys(map).forEach(k => {
      map[k].sort((a, b) => a.startTime.localeCompare(b.startTime))
    })
    return map
  }, [appointments, days])

  return (
    <div className="border rounded-lg overflow-hidden bg-white">
      {/* Header */}
      <div className="grid grid-cols-[60px_repeat(7,1fr)] border-b bg-muted/30">
        <div className="p-2 text-xs text-muted-foreground text-center border-r">Time</div>
        {days.map(day => (
          <div
            key={day.date}
            className={cn(
              "p-2 text-center border-r last:border-r-0",
              day.isToday && "bg-primary/5"
            )}
          >
            <p className="text-xs text-muted-foreground">{day.dayName}</p>
            <p className={cn(
              "text-sm font-bold",
              day.isToday && "text-primary"
            )}>
              {day.dayNum}
            </p>
          </div>
        ))}
      </div>

      {/* Time Grid */}
      <div className="grid grid-cols-[60px_repeat(7,1fr)] max-h-[500px] overflow-y-auto">
        {HOURS.map(hour => (
          <div key={hour} className="contents">
            <div className="p-1 text-xs text-muted-foreground text-right pr-2 border-r border-b h-16 flex items-start justify-end pt-1">
              {hour > 12 ? `${hour - 12} PM` : hour === 12 ? '12 PM' : `${hour} AM`}
            </div>
            {days.map(day => {
              const dayAppts = appointmentsByDay[day.date]?.filter(apt => {
                const h = parseInt(apt.startTime.split(':')[0])
                return h === hour
              }) || []

              return (
                <div
                  key={`${day.date}-${hour}`}
                  className={cn(
                    "border-r border-b last:border-r-0 p-0.5 h-16 overflow-hidden",
                    day.isToday && "bg-primary/5"
                  )}
                >
                  {dayAppts.map(apt => (
                    <div
                      key={apt.id}
                      className={cn(
                        "text-[10px] leading-tight rounded px-1 py-0.5 mb-0.5 cursor-default group relative",
                        apt.status === 'completed' && "bg-green-100 text-green-800",
                        apt.status === 'cancelled' && "bg-red-100 text-red-800 line-through",
                        apt.status === 'scheduled' && "bg-blue-100 text-blue-800",
                        apt.status === 'no_show' && "bg-gray-100 text-gray-800"
                      )}
                      title={`${apt.patient?.name || 'Patient'} - ${apt.purpose || 'General'} (${formatTime(apt.startTime)})`}
                    >
                      <p className="font-medium truncate">{apt.patient?.name || 'Patient'}</p>
                      <p className="truncate">{formatTime(apt.startTime)}</p>

                      {/* Hover actions for scheduled appointments */}
                      {apt.status === 'scheduled' && (
                        <div className="absolute top-0 right-0 hidden group-hover:flex gap-0.5 p-0.5 bg-white rounded shadow-sm border">
                          <button
                            onClick={() => onComplete(apt.id)}
                            className="text-green-600 hover:text-green-800 p-0.5"
                            title="Mark complete"
                          >
                            ✓
                          </button>
                          <button
                            onClick={() => onCancel(apt.id)}
                            className="text-red-500 hover:text-red-700 p-0.5"
                            title="Cancel"
                          >
                            ✕
                          </button>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
