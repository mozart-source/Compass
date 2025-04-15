export interface Entity {
  type: 'idea' | 'tasks' | 'person' | 'todos'
  refId: string
}

export interface Permission {
  userId: string
  level: 'view' | 'edit' | 'comment'
}

export interface Note {
  id: string
  userId: string
  title: string
  content: string
  linksOut?: string[]
  linksIn?: string[]
  entities?: Entity[]
  tags: string[]
  isDeleted?: boolean
  favorited: boolean
  icon?: string
  sharedWith?: string[]
  permissions?: Permission[]
  updatedAt: string
  createdAt?: string
}

export interface PageInfo {
  totalPages: number
  totalItems: number
  currentPage: number
  limit?: number
}

export interface UseNotesResult {
  notes: Note[]
  loading: boolean
  error: any
  createNote: (input: Partial<Note>) => Promise<Note>
  updateNote: (id: string, input: Partial<Note>) => Promise<Note>
  deleteNote: (id: string) => Promise<void>
  pageInfo?: PageInfo
}

export interface NoteSidebarProps {
  notes: Note[]
  onNoteSelect: (noteId: string) => void
  onCreateNote: () => void
  isCollapsed: boolean
  onToggleCollapse: () => void
  loading?: boolean
  selectedNoteId?: string | null
}

export interface NotePageProps extends Note {
  onSave?: (note: Partial<Note>) => void
  onDelete?: () => void
}
