import { Node as ReactFlowNode, Edge as ReactFlowEdge } from '@xyflow/react'

export interface NodeData {
  label: string
  [key: string]: any
}

export interface EdgeData {
  label: string
  [key: string]: any
}

export interface Canvas {
  id: string
  title: string
  description: string
  tags: string[]
  nodes: ReactFlowNode<NodeData>[]
  edges: ReactFlowEdge<EdgeData>[]
  updatedAt: string
  createdAt: string
}

export interface CanvasInput {
  title: string
  description: string
  tags: string[]
}

export interface NodeInput {
  canvasId: string
  type?: string
  data: string
  position: { x: number; y: number }
  style?: string
}

export interface EdgeInput {
  canvasId: string
  source: string
  target: string
  type?: string
  data?: string
  style?: string
}

export interface CanvasSidebarProps {
  selectedCanvasId: string | null
  onCanvasSelect: (id: string) => void
  isCollapsed: boolean
  onToggleCollapse: () => void
}

export interface GraphQLResponse<T> {
  success: boolean
  data: T
}

export interface CanvasesData {
  listCanvases: GraphQLResponse<Canvas[]>
}

export interface CanvasData {
  getCanvas: GraphQLResponse<Canvas>
}
