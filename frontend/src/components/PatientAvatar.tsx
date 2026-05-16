import { cn } from '@/lib/utils'

interface PatientAvatarProps {
  name: string
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const COLORS = [
  'bg-blue-100 text-blue-700',
  'bg-green-100 text-green-700',
  'bg-purple-100 text-purple-700',
  'bg-orange-100 text-orange-700',
  'bg-pink-100 text-pink-700',
  'bg-teal-100 text-teal-700',
  'bg-indigo-100 text-indigo-700',
  'bg-rose-100 text-rose-700',
]

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/)
  if (parts.length >= 2) {
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

function getColorIndex(name: string): number {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return Math.abs(hash) % COLORS.length
}

export default function PatientAvatar({ name, size = 'md', className }: PatientAvatarProps) {
  const initials = getInitials(name)
  const color = COLORS[getColorIndex(name)]

  const sizeClasses = {
    sm: 'h-7 w-7 text-xs',
    md: 'h-9 w-9 text-sm',
    lg: 'h-12 w-12 text-lg',
  }

  return (
    <div
      className={cn(
        "rounded-full flex items-center justify-center font-semibold flex-shrink-0",
        sizeClasses[size],
        color,
        className
      )}
      title={name}
    >
      {initials}
    </div>
  )
}
