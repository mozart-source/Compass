const CanvasService = require('../../../domain/canvas/canvasService');
const {
  CanvasResponseType,
  CanvasNodeResponseType,
  CanvasEdgeResponseType,
  CanvasInput,
  CanvasNodeInput,
  CanvasEdgeInput,
  getSelectedFields
} = require('../schemas/canvasTypes');
const { GraphQLID } = require('graphql');

module.exports = {
  createCanvas: {
    type: CanvasResponseType,
    args: { input: { type: CanvasInput } },
    resolve: async (parent, { input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.createCanvas(input, selectedFields, userId);
      return { success: true, data };
    }
  },
  updateCanvas: {
    type: CanvasResponseType,
    args: { id: { type: GraphQLID }, input: { type: CanvasInput } },
    resolve: async (parent, { id, input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.updateCanvas(id, input, selectedFields, userId);
      return { success: true, data };
    }
  },
  deleteCanvas: {
    type: CanvasResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.deleteCanvas(id, selectedFields, userId);
      return { success: true, data };
    }
  },
  createCanvasNode: {
    type: CanvasNodeResponseType,
    args: { input: { type: CanvasNodeInput } },
    resolve: async (parent, { input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.createNode(input, selectedFields, userId);
      return { success: true, data };
    }
  },
  updateCanvasNode: {
    type: CanvasNodeResponseType,
    args: { id: { type: GraphQLID }, input: { type: CanvasNodeInput } },
    resolve: async (parent, { id, input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.updateNode(id, input, selectedFields, userId);
      return { success: true, data };
    }
  },
  deleteCanvasNode: {
    type: CanvasNodeResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.deleteNode(id, selectedFields, userId);
      return { success: true, data };
    }
  },
  createCanvasEdge: {
    type: CanvasEdgeResponseType,
    args: { input: { type: CanvasEdgeInput } },
    resolve: async (parent, { input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.createEdge(input, selectedFields, userId);
      return { success: true, data };
    }
  },
  updateCanvasEdge: {
    type: CanvasEdgeResponseType,
    args: { id: { type: GraphQLID }, input: { type: CanvasEdgeInput } },
    resolve: async (parent, { id, input }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.updateEdge(id, input, selectedFields, userId);
      return { success: true, data };
    }
  },
  deleteCanvasEdge: {
    type: CanvasEdgeResponseType,
    args: { id: { type: GraphQLID } },
    resolve: async (parent, { id }, context, info) => {
      const userId = context.user?.id;
      const selectedFields = getSelectedFields(info);
      const data = await CanvasService.deleteEdge(id, selectedFields, userId);
      return { success: true, data };
    }
  }
}; 