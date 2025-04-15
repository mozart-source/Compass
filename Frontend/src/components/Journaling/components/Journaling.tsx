import { useState, useEffect, useCallback, useRef } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Trash2, Tag as TagIcon } from 'lucide-react'
import TiptapEditor from '@/components/notes/components/TiptapEditor'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import debounce from 'lodash/debounce'
import { JournalInput, Mood } from '../types'
import MoodSelector, { moodEmojis } from './MoodSelector'
import { DatePicker } from './DatePicker'
import '../styles/journal.css'
import { useCreateJournal, useUpdateJournal, useDeleteJournal, useJournalsByDateRange } from '../hooks'
import { toast } from '@/components/ui/use-toast'
import { GET_JOURNALS_BY_DATE_RANGE } from '../api'

// Add helper function for date parsing
const parseDate = (date: string | number): Date => {
  if (typeof date === 'string' && !isNaN(Number(date))) {
    return new Date(Number(date));
  }
  if (typeof date === 'number') {
    return new Date(date);
  }
  const parsedDate = Date.parse(date);
  return isNaN(parsedDate) ? new Date() : new Date(parsedDate);
};

// Add helper function for word count
const countWords = (text: string): number => {
  // Remove HTML tags and trim whitespace
  const cleanText = text.replace(/<[^>]*>/g, ' ').trim();
  // Split by whitespace and filter out empty strings
  return cleanText.split(/\s+/).filter(word => word.length > 0).length;
};

// Helper functions for date handling
const startOfDay = (date: Date) => {
  const d = new Date(date)
  d.setHours(0, 0, 0, 0)
  return d
}

const endOfDay = (date: Date) => {
  const d = new Date(date)
  d.setHours(23, 59, 59, 999)
  return d
}

export default function Journaling() {
  const [selectedJournalId, setSelectedJournalId] = useState<string | null>(null)
  const [newTag, setNewTag] = useState('')
  const [isMoodSelectorOpen, setIsMoodSelectorOpen] = useState(false)
  const [selectedDate, setSelectedDate] = useState<Date | undefined>(undefined)
  const [isCreating, setIsCreating] = useState(false)
  const [localTitle, setLocalTitle] = useState('')
  const [localWordCount, setLocalWordCount] = useState(0)
  
  // Set today's date when component mounts
  useEffect(() => {
    if (!selectedDate) {
      setSelectedDate(new Date());
    }
  }, []);

  // Add query for specific date range
  const { data: dateRangeResponse, loading: isLoadingDateRange } = useJournalsByDateRange(
    selectedDate ? startOfDay(selectedDate).toISOString() : '',
    selectedDate ? endOfDay(selectedDate).toISOString() : ''
  )
  
  // Mutations
  const [createJournal] = useCreateJournal()
  const [updateJournal] = useUpdateJournal()
  const [deleteJournal] = useDeleteJournal()
  
  const selectedJournal = dateRangeResponse?.journalsByDateRange?.data?.[0]

  // Get all dates that have journals for the current month
  const currentMonth = selectedDate || new Date()
  const firstDayOfMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth(), 1)
  const lastDayOfMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 0)
  
  const { data: monthJournalsResponse } = useJournalsByDateRange(
    startOfDay(firstDayOfMonth).toISOString(),
    endOfDay(lastDayOfMonth).toISOString()
  )

  // Update journalDates to use the month range response
  const journalDates = monthJournalsResponse?.journalsByDateRange?.data?.map(journal => parseDate(journal.date)) || []

  // Update handleDateSelection to properly handle navigation and fetching
  const handleDateSelection = useCallback((date: Date | undefined) => {
    if (!date) return;
    setSelectedDate(date);
  }, []);

  // Effect to handle journal selection when date range response updates
  useEffect(() => {
    if (!selectedDate || isLoadingDateRange) return;

    const journalForDate = dateRangeResponse?.journalsByDateRange?.data?.[0];
    if (journalForDate) {
      setSelectedJournalId(journalForDate.id);
    } else {
      setSelectedJournalId(null);
    }
  }, [dateRangeResponse, selectedDate, isLoadingDateRange]);

  const handleCreateJournal = async (date?: Date) => {
    if (isCreating) return;
    
    try {
      setIsCreating(true);
      const journalDate = date || new Date();
      const formattedDate = journalDate.toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      });

      const input: JournalInput = {
        title: formattedDate,
        content: '<p></p>',
        date: journalDate.toISOString(),
        tags: [],
        mood: 'NEUTRAL',
        aiGenerated: false,
        archived: false
      }
      
      const { data: response } = await createJournal({ 
        variables: { input },
        refetchQueries: [
          {
            query: GET_JOURNALS_BY_DATE_RANGE,
            variables: {
              startDate: selectedDate ? startOfDay(selectedDate).toISOString() : '',
              endDate: selectedDate ? endOfDay(selectedDate).toISOString() : ''
            }
          },
          {
            query: GET_JOURNALS_BY_DATE_RANGE,
            variables: {
              startDate: startOfDay(firstDayOfMonth).toISOString(),
              endDate: endOfDay(lastDayOfMonth).toISOString()
            }
          }
        ]
      })
      if (response?.createJournal.success && response?.createJournal.data) {
        setSelectedJournalId(response.createJournal.data.id)
        setSelectedDate(new Date(journalDate));
      } else {
        throw new Error(response?.createJournal.message || 'Failed to create journal')
      }
    } catch (error) {
      console.error('Error creating journal:', error)
      toast({
        title: "Error",
        description: "Failed to create new journal",
        variant: "destructive"
      })
    } finally {
      setIsCreating(false);
    }
  }

  // Update localTitle when selectedJournal changes
  useEffect(() => {
    if (selectedJournal?.title) {
      setLocalTitle(selectedJournal.title)
      setLocalWordCount(selectedJournal.wordCount || countWords(selectedJournal.content))
    }
  }, [selectedJournal?.title, selectedJournal?.content, selectedJournal?.wordCount])

  // Create a ref for the save function to be used outside debounce
  const saveJournalRef = useRef(async (journalId: string, updates: Partial<JournalInput>) => {
    try {
      // Only send the fields that are being updated
      const input: Partial<JournalInput> = {}

      // Only include fields that are explicitly provided in updates
      if (updates.title !== undefined) input.title = updates.title
      if (updates.content !== undefined) input.content = updates.content
      if (updates.date !== undefined) input.date = updates.date
      if (updates.mood !== undefined) input.mood = updates.mood
      if (updates.tags !== undefined) input.tags = updates.tags
      if (updates.aiGenerated !== undefined) input.aiGenerated = updates.aiGenerated
      if (updates.archived !== undefined) input.archived = updates.archived
      if (updates.aiPromptUsed !== undefined) input.aiPromptUsed = updates.aiPromptUsed

      // If title, content, or date are missing but required for this update,
      // get them from the current journal
      if (updates.title === '' || updates.content === '' || updates.date === '') {
        if (!selectedJournal) throw new Error('No journal selected')
        if (updates.title === '') input.title = selectedJournal.title
        if (updates.content === '') input.content = selectedJournal.content
        if (updates.date === '') input.date = selectedJournal.date
      }

      await updateJournal({ variables: { id: journalId, input } })
    } catch (error) {
      console.error('Error updating journal:', error)
      toast({
        title: "Error",
        description: "Failed to save journal changes",
        variant: "destructive"
      })
    }
  })

  // Create debounced save function only for content and title updates
  const debouncedSave = useRef(
    debounce(async (journalId: string, updates: Partial<JournalInput>) => {
      await saveJournalRef.current(journalId, updates)
    }, 3000)
  ).current

  const handleSaveJournal = useCallback((journalId: string, updates: Partial<JournalInput>) => {
    // If updating mood or tags, save immediately
    if (updates.mood !== undefined || updates.tags !== undefined) {
      saveJournalRef.current(journalId, updates)
    } else {
      // For other updates (title, content), use debounce
      debouncedSave(journalId, updates)
    }
  }, [debouncedSave])

  const handleDeleteJournal = async (journalId: string) => {
    try {
      const currentIndex = journalDates.findIndex(date => date.toISOString() === journalId)
      const nextJournal = journalDates.length > 1
        ? journalDates[currentIndex === journalDates.length - 1 ? currentIndex - 1 : currentIndex + 1]
        : null

      debouncedSave.cancel()
      
      const { data: response } = await deleteJournal({ variables: { id: journalId } })
      if (response?.deleteJournal.success) {
        if (nextJournal) {
          setSelectedDate(nextJournal)
        } else {
          setSelectedDate(undefined)
        }
      } else {
        throw new Error(response?.deleteJournal.message || 'Failed to delete journal')
      }
    } catch (error) {
      console.error('Error deleting journal:', error)
      toast({
        title: "Error",
        description: "Failed to delete journal",
        variant: "destructive"
      })
    }
  }

  const handleAddTag = useCallback((journalId: string, tag: string) => {
    if (!selectedJournal || !tag || selectedJournal.tags.includes(tag)) return
    
    const newTags = [...selectedJournal.tags, tag]
    handleSaveJournal(journalId, { tags: newTags })
    setNewTag('')
  }, [selectedJournal])

  const handleRemoveTag = useCallback((journalId: string, tag: string) => {
    if (!selectedJournal) return
    
    const newTags = selectedJournal.tags.filter(t => t !== tag)
    handleSaveJournal(journalId, { tags: newTags })
  }, [selectedJournal])

  const handleMoodSelect = useCallback((journalId: string, mood: Mood) => {
    handleSaveJournal(journalId, { mood })
  }, [handleSaveJournal])

  if (isLoadingDateRange) {
    return (
      <div className="flex h-full items-center justify-center text-white/70">
        Loading journals...
      </div>
    )
  }

  if (!selectedJournal && selectedDate) {
    return (
      <div className="flex h-full relative overflow-hidden">
        <div className="fixed right-6 top-1/2 -translate-y-1/2 flex flex-col gap-3 z-10">
          <DatePicker
            date={selectedDate}
            onSelect={handleDateSelection}
            highlightedDates={journalDates}
          />
        </div>
        <div className="flex-1 h-full flex items-center justify-center text-white/70">
          <div className="text-center">
            <p className="mb-4">No journal entry for {selectedDate.toLocaleDateString('en-US', {
              weekday: 'long',
              year: 'numeric',
              month: 'long',
              day: 'numeric'
            })}</p>
            <Button 
              onClick={() => handleCreateJournal(selectedDate)}
              disabled={isCreating}
            >
              {isCreating ? 'Creating...' : 'Create Journal Entry'}
            </Button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full relative overflow-hidden">
      <div className="flex-1 h-full overflow-hidden relative">
        <div className="h-full overflow-auto">
          {selectedJournal && (
            <div className="max-w-3xl mx-auto px-4 py-16">
              {/* Action buttons - vertically centered */}
              <div className="fixed right-6 top-1/2 -translate-y-1/2 flex flex-col gap-3 z-10">
                <DatePicker
                  date={selectedDate || (selectedJournal ? selectedDate : undefined)}
                  onSelect={handleDateSelection}
                  highlightedDates={journalDates}
                />
                
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setIsMoodSelectorOpen(true)}
                  className="relative group"
                >
                  <span className="text-lg">
                    {selectedJournal.mood ? moodEmojis[selectedJournal.mood] : '😐'}
                  </span>
                </Button>
                
                {/* Add Tag Button */}
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="icon">
                      <TagIcon className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <div className="p-2">
                      <Input
                        value={newTag}
                        onChange={(e) => setNewTag(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            handleAddTag(selectedJournal.id, newTag)
                          }
                        }}
                        placeholder="Add tag..."
                        className="text-xs text-white mb-2"
                      />
                      <Button 
                        size="sm" 
                        className="w-full"
                        onClick={() => handleAddTag(selectedJournal.id, newTag)}
                      >
                        Add
                      </Button>
                    </div>
                  </DropdownMenuContent>
                </DropdownMenu>
                
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => handleDeleteJournal(selectedJournal.id)}
                  className="text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>

              <MoodSelector
                isOpen={isMoodSelectorOpen}
                onClose={() => setIsMoodSelectorOpen(false)}
                onSelect={(mood) => handleMoodSelect(selectedJournal.id, mood)}
                currentMood={selectedJournal.mood}
              />

              <div className="mb-6 relative text-center">
                <input
                  type="text"
                  value={localTitle}
                  onChange={(e) => {
                    const newTitle = e.target.value;
                    setLocalTitle(newTitle);
                    handleSaveJournal(selectedJournal.id, { title: newTitle });
                  }}
                  maxLength={201}
                  className="text-3xl font-bold border-0 bg-transparent text-white text-center focus-visible:ring-0 focus-visible:ring-offset-0"
                  placeholder="Enter title..."
                  aria-label="Journal title"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.currentTarget.blur();
                    }
                  }}
                />
                <div className="mt-2 text-sm text-muted-foreground flex items-center justify-center gap-2">
                  <span>
                    {selectedDate?.toLocaleDateString('en-US', {
                      weekday: 'long',
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric'
                    })}
                  </span>
                </div>
              </div>

              <TiptapEditor
                key={selectedJournal.id}
                content={selectedJournal.content}
                onChange={(newContent) => {
                  setLocalWordCount(countWords(newContent));
                  handleSaveJournal(selectedJournal.id, { content: newContent });
                }}
                editable={true}
                className="min-h-[500px] bg-transparent border-0 -mt-4"
              />

              {selectedJournal.tags.length > 0 && (
                <div className="mt-8 mb-12 flex flex-wrap gap-2 justify-center">
                  {selectedJournal.tags.map((tag) => (
                    <Badge 
                      key={tag} 
                      variant="secondary" 
                      className="flex items-center gap-1 px-2 py-1"
                    >
                      {tag}
                      <button
                        onClick={() => handleRemoveTag(selectedJournal.id, tag)}
                        className="ml-1 hover:text-destructive"
                      >
                        ×
                      </button>
                    </Badge>
                  ))}
                </div>
              )}

              <div className="mt-8 text-sm text-muted-foreground text-center">
                {localWordCount} words
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}