import { useEditor, EditorContent, type Editor } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import TextAlign from '@tiptap/extension-text-align'
import Typography from '@tiptap/extension-typography'
import Highlight from '@tiptap/extension-highlight'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import Underline from '@tiptap/extension-underline'
import Placeholder from '@tiptap/extension-placeholder'
import { cn } from '@/lib/utils'
import EditorBubbleMenu from './EditorBubbleMenu'
import './task-list.css'
import { useEffect, useState } from 'react'

interface TiptapEditorProps {
  content: string
  onChange?: (content: string) => void
  editable?: boolean
  className?: string
  placeholder?: string
}

const TiptapEditor = ({ 
  content, 
  onChange, 
  editable = true, 
  className,
  placeholder = 'Start typing or paste content...'
}: TiptapEditorProps) => {
  const [isFocused, setIsFocused] = useState(false)
  const [isInitialized, setIsInitialized] = useState(false)

  const editor = useEditor({
    extensions: [
      StarterKit,
      Image,
      TaskList.configure({
        HTMLAttributes: {
          class: 'not-prose pl-0',
        },
      }),
      TaskItem.configure({
        nested: true,
        HTMLAttributes: {
          class: 'flex gap-2 items-start my-1',
        },
      }),
      TextAlign.configure({
        types: ['heading', 'paragraph'],
      }),
      Typography,
      Highlight.configure({ 
        multicolor: true,
        HTMLAttributes: {
          class: 'transition-colors duration-300',
        }
      }),
      Subscript,
      Superscript,
      Underline,
      Placeholder.configure({
        placeholder,
        emptyEditorClass: 'is-editor-empty',
      }),
    ],
    content,
    editable,
    onUpdate: ({ editor }) => {
      onChange?.(editor.getHTML())
    },
    onFocus: () => {
      setIsFocused(true)
    },
    onBlur: () => {
      setIsFocused(false)
    },
    editorProps: {
      attributes: {
        autocomplete: 'off',
        autocorrect: 'off',
        autocapitalize: 'off',
        'aria-label': 'Main content area, start typing to enter text.',
        class: 'outline-none min-h-[300px] transition-all duration-300',
      },
    },
  })

  // Add subtle entrance animation when editor is first loaded
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsInitialized(true)
    }, 100)
    return () => clearTimeout(timer)
  }, [])

  if (!editor) {
    return null
  }

  return (
    <div className={cn(
      "relative transition-all duration-300 ease-in-out", 
      isInitialized ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2",
      className
    )}>
      <EditorBubbleMenu editor={editor} />
      <div className={cn(
        'prose-headings:text-white prose-p:text-white prose-strong:text-white prose-em:text-white prose-li:text-white',
        'prose-h1:text-3xl prose-h1:font-bold prose-h1:mb-6',
        'prose-h2:text-2xl prose-h2:font-semibold prose-h2:mb-4',
        'prose-h3:text-xl prose-h3:font-medium prose-h3:mb-3',
        'prose-p:mb-4 prose-p:text-white prose-p:text-base',
        'prose-ul:mb-4 prose-li:mb-2 prose-li:text-white prose-li:text-base',
        'prose-blockquote:border-l-4 prose-blockquote:border-white/50 prose-blockquote:pl-4 prose-blockquote:italic prose-blockquote:text-white',
        'prose-code:text-white prose-code:bg-gray-800/50 prose-code:rounded prose-code:px-1',
        '[&_*]:transition-colors [&_*]:duration-200',
        '[&_p]:!text-white [&_li]:!text-white [&_div]:!text-white',
        '[&_ul[data-type="taskList"]_li_div_p]:!text-white',
        '[&_ul[data-type="taskList"]_li_div_p]:!text-base',
        '[&_ul[data-type="taskList"]_li_div]:text-base',
        '[&_hr]:my-4 [&_hr]:border-t',
        'p-4 rounded-md',
        // Placeholder styling
        '[&_.is-editor-empty]:before:text-gray-400 [&_.is-editor-empty]:before:content-[attr(data-placeholder)] [&_.is-editor-empty]:before:float-left [&_.is-editor-empty]:before:pointer-events-none'
      )}>
        <EditorContent editor={editor} />
      </div>
    </div>
  )
}

export default TiptapEditor 