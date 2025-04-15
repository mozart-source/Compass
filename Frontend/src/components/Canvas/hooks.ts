import { useQuery, useMutation } from '@apollo/client'
import { Node as ReactFlowNode, Edge as ReactFlowEdge } from '@xyflow/react'
import {
  GET_CANVASES,
  GET_CANVAS,
  CREATE_CANVAS,
  UPDATE_CANVAS,
  DELETE_CANVAS,
  CREATE_NODE,
  UPDATE_NODE,
  DELETE_NODE,
  CREATE_EDGE,
  UPDATE_EDGE,
  DELETE_EDGE,
} from './api'
import { Canvas, CanvasInput, NodeInput, EdgeInput, CanvasesData, CanvasData, GraphQLResponse} from './types'

export const useCanvases = () => {
  const { data, loading, error } = useQuery<CanvasesData>(GET_CANVASES)
  return {
    canvases: data?.listCanvases?.data || [],
    loading,
    error,
  }
}

export const useCanvas = (id: string) => {
  const { data, loading, error } = useQuery<CanvasData>(GET_CANVAS, {
    variables: { id },
    skip: !id,
  })
  return {
    canvas: data?.getCanvas?.data,
    loading,
    error,
  }
}

export const useCreateCanvas = () => {
  const [createCanvas, { loading }] = useMutation<{ createCanvas: GraphQLResponse<Canvas> }>(CREATE_CANVAS, {
    update(cache, { data }) {
      if (!data) return
      const existingData = cache.readQuery<CanvasesData>({ query: GET_CANVASES })
      const existingCanvases = existingData?.listCanvases?.data || []
      
      cache.writeQuery<CanvasesData>({
        query: GET_CANVASES,
        data: {
          listCanvases: {
            success: true,
            data: [...existingCanvases, data.createCanvas.data],
          },
        },
      })
    },
  })

  return {
    createCanvas: async (input: CanvasInput) => {
      const { data } = await createCanvas({ variables: { input } })
      return data?.createCanvas?.data
    },
    loading,
  }
}

export const useUpdateCanvas = () => {
  const [updateCanvas, { loading }] = useMutation<{ updateCanvas: GraphQLResponse<Canvas> }>(UPDATE_CANVAS)

  return {
    updateCanvas: async (id: string, input: Partial<CanvasInput>) => {
      const { data } = await updateCanvas({ variables: { id, input } })
      return data?.updateCanvas?.data
    },
    loading,
  }
}

export const useDeleteCanvas = () => {
  const [deleteCanvas, { loading }] = useMutation<{ deleteCanvas: GraphQLResponse<{ id: string }> }>(DELETE_CANVAS, {
    update(cache, { data }) {
      if (!data) return
      const existingData = cache.readQuery<CanvasesData>({ query: GET_CANVASES })
      const existingCanvases = existingData?.listCanvases?.data || []
      
      cache.writeQuery<CanvasesData>({
        query: GET_CANVASES,
        data: {
          listCanvases: {
            success: true,
            data: existingCanvases.filter((canvas) => canvas.id !== data.deleteCanvas.data.id),
          },
        },
      })
    },
  })

  return {
    deleteCanvas: async (id: string) => {
      const { data } = await deleteCanvas({ variables: { id } })
      return data?.deleteCanvas?.data
    },
    loading,
  }
}

export const useCreateNode = () => {
  const [createNode, { loading }] = useMutation<{ createCanvasNode: GraphQLResponse<ReactFlowNode> }>(CREATE_NODE)

  return {
    createNode: async (input: NodeInput) => {
      const { data } = await createNode({ variables: { input } })
      return data?.createCanvasNode?.data
    },
    loading,
  }
}

export const useUpdateNode = () => {
  const [updateNode, { loading }] = useMutation<{ updateCanvasNode: GraphQLResponse<ReactFlowNode> }>(UPDATE_NODE)

  return {
    updateNode: async (id: string, input: Partial<NodeInput>) => {
      const { data } = await updateNode({ variables: { id, input } })
      return data?.updateCanvasNode?.data
    },
    loading,
  }
}

export const useDeleteNode = () => {
  const [deleteNode, { loading }] = useMutation<{ deleteCanvasNode: GraphQLResponse<{ id: string }> }>(DELETE_NODE)

  return {
    deleteNode: async (id: string) => {
      const { data } = await deleteNode({ variables: { id } })
      return data?.deleteCanvasNode?.data
    },
    loading,
  }
}

export const useCreateEdge = () => {
  const [createEdge, { loading }] = useMutation<{ createCanvasEdge: GraphQLResponse<ReactFlowEdge> }>(CREATE_EDGE)

  return {
    createEdge: async (input: EdgeInput) => {
      const { data } = await createEdge({ variables: { input } })
      return data?.createCanvasEdge?.data
    },
    loading,
  }
}

export const useUpdateEdge = () => {
  const [updateEdge, { loading }] = useMutation<{ updateCanvasEdge: GraphQLResponse<ReactFlowEdge> }>(UPDATE_EDGE)

  return {
    updateEdge: async (id: string, input: Partial<EdgeInput>) => {
      const { data } = await updateEdge({ variables: { id, input } })
      return data?.updateCanvasEdge?.data
    },
    loading,
  }
}

export const useDeleteEdge = () => {
  const [deleteEdge, { loading }] = useMutation<{ deleteCanvasEdge: GraphQLResponse<{ id: string }> }>(DELETE_EDGE)

  return {
    deleteEdge: async (id: string) => {
      const { data } = await deleteEdge({ variables: { id } })
      return data?.deleteCanvasEdge?.data
    },
    loading,
  }
}
