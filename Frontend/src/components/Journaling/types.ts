export type Mood = 'HAPPY' | 'SAD' | 'ANGRY' | 'NEUTRAL' | 'EXCITED' | 'ANXIOUS' | 'TIRED' | 'GRATEFUL';

export enum JournalSortField {
  DATE = 'DATE',
  CREATED_AT = 'CREATED_AT',
  UPDATED_AT = 'UPDATED_AT',
  TITLE = 'TITLE',
  WORD_COUNT = 'WORD_COUNT',
  MOOD = 'MOOD'
}

export enum JournalSortOrder {
  ASC = 'ASC',
  DESC = 'DESC'
}

export interface Journal {
  id: string
  userId: string
  title: string
  content: string
  date: string
  mood?: Mood
  tags: string[]
  aiPromptUsed?: string | null
  aiGenerated: boolean
  archived: boolean
  wordCount: number
  createdAt: string
  updatedAt: string
  isDeleted: boolean
}

export interface JournalInput {
  title: string
  content: string
  date: string
  mood?: Mood
  tags?: string[]
  aiPromptUsed?: string
  aiGenerated?: boolean
  archived?: boolean
}

export interface JournalFilter {
  wordCountMin?: number
  wordCountMax?: number
  aiGenerated?: boolean
  tags?: string[]
  mood?: Mood
  archived?: boolean
  dateFrom?: string
  dateTo?: string
}

export interface PageInfo {
  totalItems: number
  currentPage: number
  totalPages: number
}

export interface ApiResponse<T> {
  success: boolean
  message: string
  data: T
  errors?: Array<{
    message: string
    field?: string
    code: string
  }> | null
}

export interface JournalListResponse {
  success: boolean
  message: string
  data: Journal[]
  pageInfo: PageInfo
  errors?: Array<{
    message: string
    field?: string
    code: string
  }> | null
}

export interface JournalPageProps {
  journal: Journal
  onSave: (updates: Partial<Journal>) => void
  onDelete?: () => void
}

export interface JournalSidebarProps {
  journals: Journal[]
  selectedJournalId: string | null
  onJournalSelect: (id: string) => void
  onCreateJournal: () => void
  isCollapsed: boolean
  onToggleCollapse: () => void
}
