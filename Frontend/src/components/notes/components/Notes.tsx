import { useState, useEffect, useCallback, useRef } from 'react'
import NoteSidebar from './NoteSidebar'
import NotePage from './NotePage'
import { cn } from '@/lib/utils'
import { useNotes } from '@/components/notes/hooks'
import { Note } from '@/components/notes/types'
import debounce from 'lodash/debounce'

export default function Notes() {
  const { 
    notes, 
    loading, 
    error, 
    createNote, 
    updateNote, 
    deleteNote 
  } = useNotes()
  const [selectedNoteId, setSelectedNoteId] = useState<string | null>(null)
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [localNotes, setLocalNotes] = useState<Note[]>([])
  
  // Keep track of local changes before they're synced
  const localChangesRef = useRef<{[key: string]: Partial<Note>}>({})
  
  // Create a ref for the save function to be used outside debounce
  const saveNoteRef = useRef(async (noteId: string, updates: Partial<Note>) => {
    try {
      await updateNote(noteId, updates)
      // Clear local changes after successful save
      const { [noteId]: _, ...rest } = localChangesRef.current
      localChangesRef.current = rest
    } catch (error) {
      console.error('Error updating note:', error)
      // Revert optimistic update on error
      setLocalNotes(notes)
    }
  })

  // Initialize localNotes with the fetched notes
  useEffect(() => {
    if (notes.length > 0) {
      setLocalNotes(notes)
    }
  }, [notes])

  // Select first note by default when notes are loaded
  useEffect(() => {
    if (notes.length > 0 && !selectedNoteId) {
      setSelectedNoteId(notes[0].id)
    }
  }, [notes, selectedNoteId])

  const selectedNote = localNotes.find(note => note.id === selectedNoteId)

  // Create debounced save function only once
  const debouncedSave = useRef(
    debounce(async (noteId: string, updates: Partial<Note>) => {
      await saveNoteRef.current(noteId, updates)
    }, 3000)
  ).current

  const handleSaveNote = useCallback((noteId: string, updates: Partial<Note>) => {
    // Update local state immediately
    setLocalNotes(prevNotes => 
      prevNotes.map(note => 
        note.id === noteId
          ? { ...note, ...updates, updatedAt: new Date().toISOString() }
          : note
      )
    )

    // Track local changes
    localChangesRef.current[noteId] = {
      ...(localChangesRef.current[noteId] || {}),
      ...updates
    }

    // Debounce the API call
    debouncedSave(noteId, localChangesRef.current[noteId])
  }, [debouncedSave])

  // Save any pending changes before switching notes
  const savePendingChanges = useCallback(async (currentNoteId: string) => {
    if (localChangesRef.current[currentNoteId]) {
      // Cancel any pending debounced saves
      debouncedSave.cancel()
      // Save immediately
      await saveNoteRef.current(currentNoteId, localChangesRef.current[currentNoteId])
    }
  }, [debouncedSave])

  const handleNoteSelect = useCallback(async (noteId: string) => {
    if (noteId !== selectedNoteId && selectedNoteId) {
      // Save any pending changes from the current note before switching
      await savePendingChanges(selectedNoteId)
    }
    setSelectedNoteId(noteId)
  }, [selectedNoteId, savePendingChanges])

  // Save pending changes before unmounting
  useEffect(() => {
    return () => {
      if (selectedNoteId && localChangesRef.current[selectedNoteId]) {
        debouncedSave.cancel()
        saveNoteRef.current(selectedNoteId, localChangesRef.current[selectedNoteId])
      }
    }
  }, [selectedNoteId])

  const handleCreateNote = async () => {
    try {
      // Save any pending changes before creating new note
      if (selectedNoteId) {
        await savePendingChanges(selectedNoteId)
      }
      
      const newNote = await createNote({
        title: 'Untitled Note',
        content: '<p></p>',
        tags: [],
        favorited: false
      })
      setSelectedNoteId(newNote.id)
    } catch (error) {
      console.error('Error creating note:', error)
    }
  }

  const handleDeleteNote = async (noteId: string) => {
    try {
      const currentIndex = localNotes.findIndex(note => note.id === noteId)
      const nextNote = localNotes.length > 1
        ? localNotes[currentIndex === localNotes.length - 1 ? currentIndex - 1 : currentIndex + 1]
        : null

      // Cancel any pending saves for the note being deleted
      debouncedSave.cancel()
      
      setLocalNotes(prevNotes => prevNotes.filter(note => note.id !== noteId))
      await deleteNote(noteId)
      
      // Clear any pending changes for the deleted note
      const { [noteId]: _, ...rest } = localChangesRef.current
      localChangesRef.current = rest

      if (nextNote) {
        setSelectedNoteId(nextNote.id)
      } else {
        setSelectedNoteId(null)
      }
    } catch (error) {
      console.error('Error deleting note:', error)
      setLocalNotes(notes)
    }
  }

  useEffect(() => {
    if (localNotes.length > 0 && !localNotes.some(note => note.id === selectedNoteId)) {
      setSelectedNoteId(localNotes[0].id)
    } else if (localNotes.length === 0) {
      setSelectedNoteId(null)
    }
  }, [localNotes, selectedNoteId])

  if (loading) return <div className="flex h-full items-center justify-center">Loading...</div>
  if (error) return <div className="flex h-full items-center justify-center text-red-500">Error: {error.message}</div>

  return (
    <div className="flex h-full relative overflow-hidden">
      <NoteSidebar
        notes={localNotes}
        selectedNoteId={selectedNoteId}
        onNoteSelect={handleNoteSelect}
        onCreateNote={handleCreateNote}
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        loading={loading}
      />
      
      <div className={cn(
        "flex-1 h-full overflow-hidden relative transition-all duration-300",
        isSidebarCollapsed ? "ml-0" : "ml-0"
      )}>
        <div className="h-full overflow-auto">
          {selectedNote ? (
            <NotePage
              key={selectedNote.id}
              {...selectedNote}
              onSave={(updates) => handleSaveNote(selectedNote.id, updates)}
              onDelete={() => handleDeleteNote(selectedNote.id)}
            />
          ) : (
            <div className="flex h-full items-center justify-center text-white/70">
              Select a note or create a new one
            </div>
          )}
        </div>
      </div>
    </div>
  )
}