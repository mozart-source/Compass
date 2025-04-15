import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Mood } from '../types'

const moodEmojis: Record<Mood, string> = {
  HAPPY: '😊',
  SAD: '😢',
  ANGRY: '😠',
  NEUTRAL: '😐',
  EXCITED: '🤩',
  ANXIOUS: '😰',
  TIRED: '😴',
  GRATEFUL: '🙏'
}

interface MoodSelectorProps {
  isOpen: boolean
  onClose: () => void
  onSelect: (mood: Mood) => void
  currentMood?: Mood
}

export default function MoodSelector({ isOpen, onClose, onSelect, currentMood }: MoodSelectorProps) {
  const moods = Object.keys(moodEmojis) as Mood[]

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="bg-[#1a1a1a] sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-center">How are you feeling today?</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-4 gap-4 py-4">
          {moods.map((mood) => (
            <Button
              key={mood}
              variant="ghost"
              className={`flex flex-col items-center p-8 hover:bg-accent rounded-lg transition-colors ${
                currentMood === mood ? 'bg-accent' : ''
              }`}
              onClick={() => {
                onSelect(mood)
                onClose()
              }}
            >
              <span className="text-2xl">{moodEmojis[mood]}</span>
              <span className="text-sm capitalize">{mood.toLowerCase()}</span>
            </Button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export { moodEmojis } 