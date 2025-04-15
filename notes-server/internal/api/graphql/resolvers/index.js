const noteQueries = require('./noteQueries');
const noteMutations = require('./noteMutations');
const journalQueries = require('./journalQueries');
const journalMutations = require('./journalMutations');
const canvasQueries = require('./canvasQueries');
const canvasMutations = require('./canvasMutations');
const { withFilter } = require('graphql-subscriptions');
const { pubsub } = require('../../../infrastructure/cache/pubsub');

module.exports = {
  Query: {
    ...noteQueries,
    ...journalQueries,
    ...canvasQueries
  },
  Mutation: {
    ...noteMutations,
    ...journalMutations,
    ...canvasMutations
  },
  Subscription: {
    notePageCreated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('NOTE_PAGE_CREATED'),
        (payload, variables, context) => {
          return payload.notePageCreated.data.userId === variables.userId;
        }
      )
    },
    notePageUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('NOTE_PAGE_UPDATED'),
        (payload, variables, context) => {
          return payload.notePageUpdated.data.userId === variables.userId;
        }
      )
    },
    notePageDeleted: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('NOTE_PAGE_DELETED'),
        (payload, variables, context) => {
          return payload.notePageDeleted.data.userId === variables.userId;
        }
      )
    },
    journalCreated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('JOURNAL_CREATED'),
        (payload, variables, context) => {
          return payload.journalCreated.data.userId === variables.userId;
        }
      )
    },
    journalUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('JOURNAL_UPDATED'),
        (payload, variables, context) => {
          return payload.journalUpdated.data.userId === variables.userId;
        }
      )
    },
    journalDeleted: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('JOURNAL_DELETED'),
        (payload, variables, context) => {
          return payload.journalDeleted.data.userId === variables.userId;
        }
      )
    },
    canvasCreated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_CREATED'),
        (payload, variables, context) => {
          return payload.canvasCreated && payload.userId === variables.userId;
        }
      )
    },
    canvasUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_UPDATED'),
        (payload, variables, context) => {
          return payload.canvasUpdated && payload.userId === variables.userId;
        }
      )
    },
    canvasDeleted: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_DELETED'),
        (payload, variables, context) => {
          return payload.canvasDeleted && payload.userId === variables.userId;
        }
      )
    },
    canvasNodeCreated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_NODE_CREATED'),
        (payload, variables, context) => {
          return payload.canvasNodeCreated && payload.userId === variables.userId;
        }
      )
    },
    canvasNodeUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_NODE_UPDATED'),
        (payload, variables, context) => {
          return payload.canvasNodeUpdated && payload.userId === variables.userId;
        }
      )
    },
    canvasNodeDeleted: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_NODE_DELETED'),
        (payload, variables, context) => {
          return payload.canvasNodeDeleted && payload.userId === variables.userId;
        }
      )
    },
    canvasEdgeCreated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_EDGE_CREATED'),
        (payload, variables, context) => {
          return payload.canvasEdgeCreated && payload.userId === variables.userId;
        }
      )
    },
    canvasEdgeUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_EDGE_UPDATED'),
        (payload, variables, context) => {
          return payload.canvasEdgeUpdated && payload.userId === variables.userId;
        }
      )
    },
    canvasEdgeDeleted: {
      subscribe: withFilter(
        () => pubsub.asyncIterator('CANVAS_EDGE_DELETED'),
        (payload, variables, context) => {
          return payload.canvasEdgeDeleted && payload.userId === variables.userId;
        }
      )
    }
  }
};