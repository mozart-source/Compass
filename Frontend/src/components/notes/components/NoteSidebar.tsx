import { useState, useEffect } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Search, Plus, Star, Tag as TagIcon, SortAsc, Filter, ChevronLeft } from 'lucide-react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"
import { NoteSidebarProps } from '@/components/notes/types'

export default function NoteSidebar({ 
  notes, 
  onNoteSelect, 
  onCreateNote,
  isCollapsed,
  onToggleCollapse,
  loading = false,
  selectedNoteId
}: NoteSidebarProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [sortBy, setSortBy] = useState<'title' | 'updated'>('updated')
  const [showFavorited, setShowFavorited] = useState(false)
  const [marqueePosition, setMarqueePosition] = useState(0)

  // Marquee animation effect
  useEffect(() => {
    if (!isCollapsed) return;
    
    const intervalId = setInterval(() => {
      setMarqueePosition(prev => (prev - 1) % 100);
    }, 50);
    
    return () => clearInterval(intervalId);
  }, [isCollapsed]);

  // Get unique tags from all notes
  const allTags = Array.from(new Set(notes.flatMap(note => note.tags)))

  // Filter and sort notes
  const filteredNotes = notes
    .filter(note => {
      const matchesSearch = searchQuery ? 
        (note.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
         note.content?.toLowerCase().includes(searchQuery.toLowerCase())) : true
      const matchesTags = selectedTags.length === 0 || 
                         selectedTags.every(tag => note.tags?.includes(tag))
      const matchesFavorited = !showFavorited || note.favorited
      return matchesSearch && matchesTags && matchesFavorited
    })
    .sort((a, b) => {
      if (sortBy === 'title') {
        return a.title.localeCompare(b.title)
      }
      return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
    })

  const handleNoteClick = (noteId: string) => {
    console.log('Note clicked:', noteId) // Debug log
    onNoteSelect(noteId)
  }

  return (
    <div 
      className={cn(
        "border-r relative flex flex-col transition-all duration-300",
        isCollapsed ? "w-[45px]" : "w-80"
      )}
    >
      <div className={cn(
        "h-full flex flex-col transition-opacity duration-300",
        isCollapsed ? "opacity-0" : "opacity-100"
      )}>
        <div className="p-4 border-b">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold">Notes</h2>
            <div className="flex items-center gap-2">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <SortAsc className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => setSortBy('title')}>
                    Sort by Title
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setSortBy('updated')}>
                    Sort by Last Updated
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <Filter className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <div className="p-2">
                    <div className="space-y-1">
                      <h4 className="text-sm font-medium">Filter by Tags</h4>
                      <div className="flex flex-wrap gap-1">
                        {allTags.map(tag => (
                          <Badge
                            key={tag}
                            variant={selectedTags.includes(tag) ? "default" : "outline"}
                            className="cursor-pointer"
                            onClick={() => {
                              setSelectedTags(prev =>
                                prev.includes(tag)
                                  ? prev.filter(t => t !== tag)
                                  : [...prev, tag]
                              )
                            }}
                          >
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    </div>
                    <div className="mt-4">
                      <Button
                        variant="outline"
                        size="sm"
                        className="w-full"
                        onClick={() => setShowFavorited(!showFavorited)}
                      >
                        <Star className={`h-4 w-4 mr-2 ${showFavorited ? 'text-yellow-500' : ''}`} />
                        {showFavorited ? 'Show All' : 'Show Favorited'}
                      </Button>
                    </div>
                  </div>
                </DropdownMenuContent>
              </DropdownMenu>
              <Button onClick={onCreateNote} size="icon" variant="ghost">
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search notes..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>
        </div>
        <div className="flex-1 overflow-auto p-4">
          {loading ? (
            <div className="text-center py-8 text-muted-foreground">
              Loading notes...
            </div>
          ) : filteredNotes.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No notes found
            </div>
          ) : (
            <div className="space-y-2">
              {filteredNotes.map(note => (
                <div
                  key={note.id}
                  className={cn(
                    "group relative flex flex-col gap-2 p-3 rounded-lg border cursor-pointer",
                    "hover:bg-accent hover:border-accent-foreground/20",
                    "transition-colors duration-200",
                    selectedNoteId === note.id && "bg-accent border-accent-foreground/20"
                  )}
                  onClick={() => handleNoteClick(note.id)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      handleNoteClick(note.id)
                    }
                  }}
                >
                  {/* Title and Favorite Icon */}
                  <div className="flex items-center justify-between">
                    <h3 className="font-medium line-clamp-1 flex-1">{note.title}</h3>
                    {note.favorited && (
                      <Star className="h-4 w-4 text-yellow-500 flex-shrink-0" />
                    )}
                  </div>

                  {/* Preview Text */}
                  <p className="text-sm text-muted-foreground line-clamp-2 min-h-[2.5rem]">
                    {note.content.replace(/<[^>]*>/g, '')}
                  </p>

                  {/* Tags and Date */}
                  <div className="flex items-center justify-between gap-2 mt-1">
                    {note.tags.length > 0 && (
                      <div className="flex items-center gap-1 flex-1 min-w-0">
                        <TagIcon className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                        <div className="flex gap-1 overflow-x-auto no-scrollbar">
                          {note.tags.map(tag => (
                            <Badge 
                              key={tag} 
                              variant="secondary" 
                              className="text-xs px-1 py-0 whitespace-nowrap"
                            >
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                    <span className="text-xs text-muted-foreground whitespace-nowrap flex-shrink-0">
                      {new Date(note.updatedAt).toLocaleDateString()}
                    </span>
                  </div>

                  {/* Hover Effect Gradient Border */}
                  <div className={cn(
                    "absolute inset-0 rounded-lg opacity-0 group-hover:opacity-100",
                    "transition-opacity duration-200",
                    "pointer-events-none",
                    "border-2 border-accent-foreground/20"
                  )} />
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Collapsed state marquee animation */}
      {isCollapsed && (
        <div className="absolute inset-y-0 left-0 w-full overflow-hidden px-2 select-none flex items-center justify-center">
          {/* Fade effect at the top */}
          <div className="absolute top-0 left-0 right-0 h-16 z-10 pointer-events-none" 
               style={{ 
                 background: 'linear-gradient(to bottom, hsl(var(--background)), transparent)'
               }} 
          />
          
          <div className="relative h-full w-full">
            {/* First copy of text */}
            <div 
              className="text-sm text-muted-foreground [writing-mode:vertical-lr] rotate-180 mx-auto whitespace-nowrap marquee-animation"
            >
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; 
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
            </div>
            
            {/* Second copy of text, delayed to create seamless effect */}
            <div 
              className="text-sm text-muted-foreground [writing-mode:vertical-lr] rotate-180 mx-auto whitespace-nowrap marquee-animation"
              style={{ animationDelay: '15s' }}
            >
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; 
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
              Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp; Notes sidebar collapsed &nbsp;&nbsp;
            </div>
          </div>
        </div>
      )}

      {/*collapse button */}
      <Button
        variant="ghost"
        size="icon"
        className={cn(
          "absolute -right-3 top-10 z-10 h-6 w-6 rounded-full border bg-background shadow-md",
          "hover:bg-accent hover:text-accent-foreground",
          "transition-transform duration-200",
          isCollapsed && "rotate-180"
        )}
        onClick={onToggleCollapse}
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>
    </div>
  )
} 