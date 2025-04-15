import { useState, useCallback, useEffect, useMemo } from 'react'
import { ReactFlow, Node, Edge, Connection, addEdge, OnNodesChange, OnEdgesChange, OnConnect, applyNodeChanges, applyEdgeChanges, useReactFlow, Controls, Background, BackgroundVariant, Panel, ReactFlowProvider, NodeTypes } from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import CanvasSidebar from './CanvasSidebar'
import { cn } from '@/lib/utils'
import { useCanvas, useUpdateNode, useUpdateEdge, useCreateEdge, useDeleteEdge, useCreateNode, useDeleteNode, useCanvases, useUpdateCanvas } from '../hooks'
import { NodeData, EdgeData } from '../types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus, LayoutPanelLeft } from 'lucide-react'

// Import node components with any type to avoid declaration file errors
import CustomResizerNode from './nodes/CustomResizerNode'
import DefaultNode from './nodes/DefaultNode'

// Define node types for React Flow
const nodeTypes: NodeTypes = {
  resizable: CustomResizerNode,
  default: DefaultNode,
}

const nodeDefaults = {
  style: {
    background: 'hsl(var(--card))',
    color: 'hsl(var(--card-foreground))',
    border: '1px solid hsl(var(--border))',
    borderRadius: 'var(--radius)',
    padding: '10px',
    minWidth: 150,
    minHeight: 50,
  },
}

function Flow({ canvasId }: { canvasId: string }) {
  const { canvas, loading } = useCanvas(canvasId)
  const { updateNode } = useUpdateNode()
  const { updateEdge } = useUpdateEdge()
  const { createEdge } = useCreateEdge()
  const { deleteEdge } = useDeleteEdge()
  const { createNode } = useCreateNode()
  const { deleteNode } = useDeleteNode()
  const { updateCanvas } = useUpdateCanvas()
  const reactFlowInstance = useReactFlow()
  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [title, setTitle] = useState(canvas?.title || '')
  const [nodeType, setNodeType] = useState<'resizable' | 'default'>('resizable')

  // Transform nodes and edges from backend format to React Flow format
  const transformedNodes = useMemo(() => {
    if (!canvas?.nodes) return []
    return canvas.nodes.map(node => ({
      ...node,
      data: node.data ? JSON.parse(node.data as unknown as string) : { label: 'Node' },
      style: node.style ? JSON.parse(node.style as unknown as string) : nodeDefaults.style
    }))
  }, [canvas?.nodes])

  const transformedEdges = useMemo(() => {
    if (!canvas?.edges) return []
    return canvas.edges.map(edge => ({
      ...edge,
      data: edge.data ? JSON.parse(edge.data as unknown as string) : {},
      style: edge.style ? JSON.parse(edge.style as unknown as string) : { stroke: 'hsl(var(--border))' }
    }))
  }, [canvas?.edges])

  const [nodes, setNodes] = useState<Node<NodeData>[]>(transformedNodes)
  const [edges, setEdges] = useState<Edge<EdgeData>[]>(transformedEdges)

  // Update local state when canvas data changes
  useEffect(() => {
    if (canvas) {
      setNodes(transformedNodes)
      setEdges(transformedEdges)
      setTitle(canvas.title)
    }
  }, [canvas, transformedNodes, transformedEdges])

  const handleTitleUpdate = async () => {
    if (canvas && title !== canvas.title) {
      await updateCanvas(canvasId, {
        title,
        description: canvas.description,
        tags: canvas.tags
      })
    }
    setIsEditingTitle(false)
  }

  // Update nodes in the backend when they change
  const onNodesChange: OnNodesChange = useCallback(async (changes) => {
    const updatedNodes = applyNodeChanges(changes, nodes) as Node<NodeData>[]
    setNodes(updatedNodes)

    // Handle node position updates, resize, and deletion
    for (const change of changes) {
      if (change.type === 'position' && !change.dragging && change.id) {
        await updateNode(change.id, {
          canvasId,
          position: change.position,
        })
      } else if (change.type === 'dimensions' && change.id && change.dimensions) {
        // Handle node resize - only for resizable nodes
        const node = nodes.find(n => n.id === change.id)
        if (node && node.style && node.type === 'resizable') {
          const updatedStyle = {
            ...node.style,
            width: change.dimensions.width,
            height: change.dimensions.height,
          }
          await updateNode(change.id, {
            canvasId,
            style: JSON.stringify(updatedStyle),
          })
        }
      } else if (change.type === 'remove' && change.id) {
        await deleteNode(change.id)
      }
    }
  }, [canvasId, updateNode, deleteNode, nodes])

  // Handle edge updates
  const onEdgesChange: OnEdgesChange = useCallback(async (changes) => {
    const updatedEdges = applyEdgeChanges(changes, edges) as Edge<EdgeData>[]
    setEdges(updatedEdges)

    // Handle edge deletion
    for (const change of changes) {
      if (change.type === 'remove' && change.id) {
        await deleteEdge(change.id)
      }
    }
  }, [deleteEdge, edges])

  // Handle new connections
  const onConnect: OnConnect = useCallback(async (connection: Connection) => {
    const newEdge = {
      source: connection.source || '',
      target: connection.target || '',
      canvasId,
      data: JSON.stringify({ label: '' }),
      style: JSON.stringify({ stroke: 'hsl(var(--border))' }),
    }

    const edge = await createEdge(newEdge)
    if (edge) {
      const transformedEdge = {
        ...edge,
        data: edge.data ? JSON.parse(edge.data as unknown as string) : {},
        style: edge.style ? JSON.parse(edge.style as unknown as string) : { stroke: 'hsl(var(--border))' }
      }
      setEdges((eds) => addEdge(transformedEdge as Edge<EdgeData>, eds))
    }
  }, [canvasId, createEdge])

  // Handle new node creation
  const onAddNode = useCallback(async () => {
    const { x: viewX, y: viewY, zoom } = reactFlowInstance.getViewport()
    const position = {
      x: Math.round((-viewX + window.innerWidth / 3) / zoom),
      y: Math.round((-viewY + window.innerHeight / 3) / zoom),
    }

    const newNode = {
      canvasId,
      type: nodeType,
      position,
      data: JSON.stringify({ label: 'New Node' }),
      style: JSON.stringify(nodeDefaults.style),
    }

    const node = await createNode(newNode)
    if (node) {
      const transformedNode = {
        ...node,
        data: node.data ? JSON.parse(node.data as unknown as string) : { label: 'New Node' },
        style: node.style ? JSON.parse(node.style as unknown as string) : nodeDefaults.style
      }
      setNodes((nds) => [...nds, transformedNode as Node<NodeData>])
    }
  }, [reactFlowInstance, canvasId, createNode, nodeType])

  // Toggle node type
  const toggleNodeType = useCallback(() => {
    setNodeType(prev => prev === 'resizable' ? 'default' : 'resizable')
  }, [])

  // Add updateNodeLabel handler
  const updateNodeLabel = useCallback(async (nodeId: string, newLabel: string) => {
    const node = nodes.find(n => n.id === nodeId)
    if (node) {
      const updatedData = { ...node.data, label: newLabel }
      await updateNode(nodeId, {
        canvasId,
        data: JSON.stringify(updatedData),
      })
      setNodes(nds => 
        nds.map(n => {
          if (n.id === nodeId) {
            return { ...n, data: updatedData }
          }
          return n
        })
      )
    }
  }, [nodes, updateNode, canvasId])

  // Update nodeTypes to include updateNodeLabel for both node types
  const nodeTypesWithProps = useMemo(() => ({
    resizable: (props: any) => <CustomResizerNode {...props} updateNodeLabel={updateNodeLabel} />,
    default: (props: any) => <DefaultNode {...props} updateNodeLabel={updateNodeLabel} />
  }), [updateNodeLabel])

  if (loading) {
    return <div className="flex items-center justify-center h-full">Loading canvas...</div>
  }

  return (
    <div className="w-full h-full relative">
      <Panel position="top-center" className="bg-background/50 backdrop-blur-sm p-2 rounded-lg shadow-sm">
        {isEditingTitle ? (
          <div className="flex items-center gap-2">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onBlur={handleTitleUpdate}
              onKeyDown={(e) => e.key === 'Enter' && handleTitleUpdate()}
              className="h-8"
              autoFocus
            />
          </div>
        ) : (
          <div
            className="text-lg font-medium cursor-pointer hover:text-primary transition-colors px-2"
            onClick={() => setIsEditingTitle(true)}
          >
            {title || 'Untitled Canvas'}
          </div>
        )}
      </Panel>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypesWithProps}
        fitView
        className="bg-background"
        defaultEdgeOptions={{
          style: { stroke: 'hsl(var(--border))' },
          type: 'default',
        }}
        proOptions={{ hideAttribution: true }}
      >
        <Panel position="top-right" className="flex gap-2">
          <Button 
            size="sm" 
            variant="outline" 
            onClick={toggleNodeType} 
            className="bg-background/50 backdrop-blur-sm"
          >
            <LayoutPanelLeft className="h-4 w-4 mr-2" />
            Node Type: {nodeType === 'resizable' ? 'Custom' : 'Default'}
          </Button>
          <Button 
            size="sm" 
            variant="outline" 
            onClick={onAddNode} 
            className="bg-background/50 backdrop-blur-sm"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Node
          </Button>
        </Panel>
        <Controls className="custom-controls" />
        <Background gap={12} size={1} />
      </ReactFlow>
    </div>
  )
}

export default function Canvas() {
  const { canvases, loading } = useCanvases()
  const [selectedCanvasId, setSelectedCanvasId] = useState<string | null>(null)
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  // Automatically select the first canvas when the component loads
  useEffect(() => {
    if (canvases?.length > 0 && !selectedCanvasId) {
      setSelectedCanvasId(canvases[0].id)
    }
  }, [canvases])

  if (loading) {
    return <div className="flex items-center justify-center h-full">Loading canvases...</div>
  }

  return (
    <ReactFlowProvider>
      <div className="flex h-full">
        <CanvasSidebar
          selectedCanvasId={selectedCanvasId || ''}
          onCanvasSelect={setSelectedCanvasId}
          isCollapsed={isSidebarCollapsed}
          onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />
        <div className={cn(
          "flex-1 h-full overflow-hidden relative transition-all duration-300",
          isSidebarCollapsed ? "ml-0" : "ml-0"
        )}>
          {selectedCanvasId ? (
            <Flow canvasId={selectedCanvasId} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              Select a canvas to get started
            </div>
          )}
        </div>
      </div>
    </ReactFlowProvider>
  )
}
