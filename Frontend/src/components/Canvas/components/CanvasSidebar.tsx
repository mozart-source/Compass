import { useState } from 'react'
import { useCanvases, useCreateCanvas } from '../hooks'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'

interface CanvasSidebarProps {
  selectedCanvasId: string;
  onCanvasSelect: (id: string) => void;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

interface NewCanvasForm {
  title: string;
  description: string;
}

export default function CanvasSidebar({ 
  selectedCanvasId, 
  onCanvasSelect, 
  isCollapsed, 
  onToggleCollapse 
}: CanvasSidebarProps) {
  const { canvases } = useCanvases()
  const { createCanvas } = useCreateCanvas()
  const [isOpen, setIsOpen] = useState(false)
  const [newCanvas, setNewCanvas] = useState<NewCanvasForm>({
    title: '',
    description: ''
  })

  const handleCreateCanvas = async () => {
    if (!newCanvas.title.trim()) return

    const canvas = await createCanvas({
      title: newCanvas.title,
      description: newCanvas.description,
      tags: []
    })

    if (canvas) {
      setNewCanvas({ title: '', description: '' })
      setIsOpen(false)
      onCanvasSelect(canvas.id)
    }
  }

  return (
    <div className={cn(
      "border-r border-border h-full transition-all duration-300 flex flex-col",
      isCollapsed ? "w-12" : "w-64"
    )}>
      {/* Header with collapse toggle and add button */}
      <div className="flex h-12 border-b border-border">
        <Button 
          variant="ghost" 
          size="icon" 
          onClick={onToggleCollapse}
          className="h-12 rounded-none"
        >
          {isCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </Button>
        
        {!isCollapsed && (
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button variant="ghost" size="sm" className="flex-1 h-12 rounded-none">
                <Plus className="h-4 w-4 mr-2" />
                New Canvas
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Canvas</DialogTitle>
              </DialogHeader>
              <div className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Input
                    placeholder="Canvas Title"
                    value={newCanvas.title}
                    onChange={(e) => setNewCanvas(prev => ({ ...prev, title: e.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <Textarea
                    placeholder="Canvas Description (optional)"
                    value={newCanvas.description}
                    onChange={(e) => setNewCanvas(prev => ({ ...prev, description: e.target.value }))}
                  />
                </div>
                <Button onClick={handleCreateCanvas} className="w-full">
                  Create Canvas
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* List of canvases */}
      {!isCollapsed && (
        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {canvases?.map((canvas) => (
            <div
              key={canvas.id}
              className={cn(
                "p-2 cursor-pointer rounded transition-colors",
                selectedCanvasId === canvas.id 
                  ? "bg-accent text-accent-foreground" 
                  : "hover:bg-accent/50"
              )}
              onClick={() => onCanvasSelect(canvas.id)}
            >
              <div className="font-medium">{canvas.title}</div>
              {canvas.description && (
                <div className="text-sm text-muted-foreground truncate">
                  {canvas.description}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
} 