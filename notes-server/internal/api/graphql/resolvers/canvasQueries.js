const CanvasService = require('../../../domain/canvas/canvasService');
const {
  CanvasResponseType,
  CanvasListResponseType,
  CanvasNodeResponseType,
  CanvasEdgeResponseType,
  getSelectedFields
} = require('../schemas/canvasTypes');
const { GraphQLID, GraphQLList } = require('graphql');
const { Canvas } = require('../../../domain/canvas/model');

module.exports = {
  getCanvas: {
    type: CanvasResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.getCanvas(id, selectedFields, userId);
      return { success: true, data };
    }
  },
  getCanvasNode: {
    type: CanvasNodeResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.getNode(id, selectedFields, userId);
      return { success: true, data };
    }
  },
  getCanvasEdge: {
    type: CanvasEdgeResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.getEdge(id, selectedFields, userId);
      return { success: true, data };
    }
  },
  // List queries (basic, can be extended with filters/pagination)
  listCanvases: {
    type: CanvasListResponseType,
    resolve: async (parent, args, context, info) => {
      const userId = context.user?.id;
      const canvases = await Canvas.find({ userId, isDeleted: false })
        .populate({
          path: 'nodes',
          match: { isDeleted: false },
          select: 'type data position style label'
        })
        .populate({
          path: 'edges',
          match: { isDeleted: false },
          select: 'source target type data style label'
        })
        .lean();
      return { success: true, data: canvases };
    }
  }
}; 