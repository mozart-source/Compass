const {
  GraphQLObjectType,
  GraphQLString,
  GraphQLID,
  GraphQLList,
  GraphQLBoolean,
  GraphQLInputObjectType,
  GraphQLInt,
  GraphQLFloat
} = require('graphql');
const { createResponseType } = require('./responseTypes');
const { PermissionLevelEnum, PermissionInput, PermissionType } = require('./permissionTypes');

const NodePositionInput = new GraphQLInputObjectType({
  name: 'NodePositionInput',
  fields: {
    x: { type: GraphQLFloat },
    y: { type: GraphQLFloat }
  }
});

const CanvasNodeInput = new GraphQLInputObjectType({
  name: 'CanvasNodeInput',
  fields: {
    canvasId: { type: GraphQLID },
    type: { type: GraphQLString },
    data: { type: GraphQLString }, // JSON as string
    position: { type: NodePositionInput },
    style: { type: GraphQLString }, // JSON as string
    label: { type: GraphQLString },
    color: { type: GraphQLString }
  }
});

const CanvasEdgeInput = new GraphQLInputObjectType({
  name: 'CanvasEdgeInput',
  fields: {
    canvasId: { type: GraphQLID },
    source: { type: GraphQLID },
    target: { type: GraphQLID },
    type: { type: GraphQLString },
    data: { type: GraphQLString }, // JSON as string
    label: { type: GraphQLString },
    style: { type: GraphQLString }, // JSON as string
  }
});

const CanvasInput = new GraphQLInputObjectType({
  name: 'CanvasInput',
  fields: {
    title: { type: GraphQLString },
    description: { type: GraphQLString },
    tags: { type: new GraphQLList(GraphQLString) },
    sharedWith: { type: new GraphQLList(GraphQLID) },
    permissions: { type: new GraphQLList(PermissionInput) }
  }
});

const CanvasNodeType = new GraphQLObjectType({
  name: 'CanvasNode',
  fields: () => ({
    id: { type: GraphQLID, resolve: (parent) => parent._id ? parent._id.toString() : null },
    canvasId: { type: GraphQLID },
    type: { type: GraphQLString },
    data: { type: GraphQLString }, // JSON as string
    position: {
      type: new GraphQLObjectType({
        name: 'NodePosition',
        fields: {
          x: { type: GraphQLFloat },
          y: { type: GraphQLFloat }
        }
      }),
      resolve: (parent) => parent.position
    },
    style: { type: GraphQLString }, // JSON as string
    label: { type: GraphQLString },
    color: { type: GraphQLString },
    userId: { type: GraphQLID },
    isDeleted: { type: GraphQLBoolean },
    createdAt: { type: GraphQLString },
    updatedAt: { type: GraphQLString }
  })
});

const CanvasEdgeType = new GraphQLObjectType({
  name: 'CanvasEdge',
  fields: () => ({
    id: { type: GraphQLID, resolve: (parent) => parent._id ? parent._id.toString() : null },
    canvasId: { type: GraphQLID },
    source: { type: GraphQLID },
    target: { type: GraphQLID },
    type: { type: GraphQLString },
    data: { type: GraphQLString }, // JSON as string
    label: { type: GraphQLString },
    style: { type: GraphQLString }, // JSON as string
    userId: { type: GraphQLID },
    isDeleted: { type: GraphQLBoolean },
    createdAt: { type: GraphQLString },
    updatedAt: { type: GraphQLString }
  })
});

const CanvasType = new GraphQLObjectType({
  name: 'Canvas',
  fields: () => ({
    id: { type: GraphQLID, resolve: (parent) => parent._id ? parent._id.toString() : null },
    userId: { type: GraphQLID },
    title: { type: GraphQLString },
    description: { type: GraphQLString },
    nodes: { type: new GraphQLList(CanvasNodeType) },
    edges: { type: new GraphQLList(CanvasEdgeType) },
    tags: { type: new GraphQLList(GraphQLString) },
    isDeleted: { type: GraphQLBoolean },
    sharedWith: { type: new GraphQLList(GraphQLID) },
    permissions: { type: new GraphQLList(PermissionType) },
    createdAt: { type: GraphQLString },
    updatedAt: { type: GraphQLString }
  })
});

const CanvasResponseType = createResponseType(CanvasType, 'Canvas');
const CanvasListResponseType = createResponseType(new GraphQLList(CanvasType), 'CanvasList');
const CanvasNodeResponseType = createResponseType(CanvasNodeType, 'CanvasNode');
const CanvasEdgeResponseType = createResponseType(CanvasEdgeType, 'CanvasEdge');

// Helper to get selected fields
const getSelectedFields = (info) => {
  try {
    const selections = info.fieldNodes[0].selectionSet.selections;
    const dataSelection = selections.find(selection => selection.name.value === 'data');
    if (dataSelection && dataSelection.selectionSet) {
      return dataSelection.selectionSet.selections.map(sel => sel.name.value).join(' ');
    }
    return 'title description tags createdAt updatedAt userId';
  } catch (e) {
    return 'title description tags createdAt updatedAt userId';
  }
};

// --- Subscription Fields for Canvas ---
const canvasSubscriptionFields = {
  canvasCreated: {
    type: CanvasResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas is created.'
  },
  canvasUpdated: {
    type: CanvasResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas is updated.'
  },
  canvasDeleted: {
    type: CanvasResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas is deleted.'
  },
  canvasNodeCreated: {
    type: CanvasNodeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas node is created.'
  },
  canvasNodeUpdated: {
    type: CanvasNodeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas node is updated.'
  },
  canvasNodeDeleted: {
    type: CanvasNodeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas node is deleted.'
  },
  canvasEdgeCreated: {
    type: CanvasEdgeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas edge is created.'
  },
  canvasEdgeUpdated: {
    type: CanvasEdgeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas edge is updated.'
  },
  canvasEdgeDeleted: {
    type: CanvasEdgeResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a canvas edge is deleted.'
  }
};

module.exports = {
  CanvasType,
  CanvasResponseType,
  CanvasListResponseType,
  CanvasInput,
  CanvasNodeType,
  CanvasNodeInput,
  CanvasNodeResponseType,
  CanvasEdgeType,
  CanvasEdgeInput,
  CanvasEdgeResponseType,
  getSelectedFields,
  canvasSubscriptionFields
}; 