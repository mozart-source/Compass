const { Canvas, CanvasNode, CanvasEdge } = require('./model');
const { ValidationError, NotFoundError, DatabaseError } = require('../../../pkg/utils/errorHandler');
const { logger } = require('../../../pkg/utils/logger');
const RedisService = require('../../infrastructure/cache/redisService');
const redisConfig = require('../../infrastructure/cache/config');
const { pubsub } = require('../../infrastructure/cache/pubsub');

const redisClient = new RedisService(redisConfig);

class CanvasService {
  // Canvas CRUD
  async createCanvas(input, selectedFields = '', currentUserId = null) {
    logger.debug('Creating canvas', { input });
    const userId = currentUserId || input.userId;
    if (!userId) throw new ValidationError('User ID is required', 'userId');
    if (!input.title?.trim()) throw new ValidationError('Title is required', 'title');
    const canvas = new Canvas({ ...input, userId });
    await canvas.save();
    const saved = await Canvas.findById(canvas._id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvas', canvas._id.toString(), saved, [userId], userId);
    // Publish event
    pubsub.publish('CANVAS_CREATED', { canvasCreated: { success: true, data: saved }, userId });
    return saved;
  }
  async updateCanvas(id, input, selectedFields = '', currentUserId = null) {
    logger.debug('Updating canvas', { id, input });
    if (!id) throw new ValidationError('Canvas ID is required', 'id');
    const canvas = await Canvas.findOne({ _id: id, isDeleted: false });
    if (!canvas) throw new NotFoundError('Canvas');
    if (currentUserId && canvas.userId !== currentUserId) throw new ValidationError('No access');
    Object.assign(canvas, input);
    await canvas.save();
    const updated = await Canvas.findById(id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvas', id.toString(), updated, [canvas.userId], canvas.userId);
    // Publish event
    pubsub.publish('CANVAS_UPDATED', { canvasUpdated: { success: true, data: updated }, userId: canvas.userId });
    return updated;
  }
  async deleteCanvas(id, selectedFields = '', currentUserId = null) {
    logger.debug('Deleting canvas', { id });
    if (!id) throw new ValidationError('Canvas ID is required', 'id');
    const canvas = await Canvas.findOne({ _id: id, isDeleted: false });
    if (!canvas) throw new NotFoundError('Canvas');
    if (currentUserId && canvas.userId !== currentUserId) throw new ValidationError('No access');
    canvas.isDeleted = true;
    await canvas.save();
    await redisClient.invalidateByPattern(`user:${canvas.userId}:canvas:*`);
    const deleted = await Canvas.findById(id).select(selectedFields || '').lean();
    // Publish event
    pubsub.publish('CANVAS_DELETED', { canvasDeleted: { success: true, data: deleted }, userId: canvas.userId });
    return deleted;
  }
  async getCanvas(id, selectedFields = '', currentUserId = null) {
    if (!id) throw new ValidationError('Canvas ID is required', 'id');
    const canvas = await Canvas.findOne({ _id: id, isDeleted: false })
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
      .select(selectedFields || '')
      .lean();
    
    if (!canvas) throw new NotFoundError('Canvas');
    // if (currentUserId && canvas.userId !== currentUserId) throw new ValidationError('No access');
    return canvas;
  }
  // Node CRUD
  async createNode(input, selectedFields = '', currentUserId = null) {
    logger.debug('Creating canvas node', { input });
    const userId = currentUserId || input.userId;
    if (!userId) throw new ValidationError('User ID is required', 'userId');
    if (!input.canvasId) throw new ValidationError('Canvas ID is required', 'canvasId');
    if (!input.position || typeof input.position.x !== 'number' || typeof input.position.y !== 'number') {
      throw new ValidationError('Position (x, y) is required', 'position');
    }

    // Create the node
    const node = new CanvasNode({ ...input, userId });
    await node.save();

    // Update the canvas to include this node
    await Canvas.findByIdAndUpdate(
      input.canvasId,
      { $push: { nodes: node._id } },
      { new: true }
    );

    const saved = await CanvasNode.findById(node._id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvasNode', node._id.toString(), saved, [userId], userId);
    
    // Publish event
    pubsub.publish('CANVAS_NODE_CREATED', { canvasNodeCreated: { success: true, data: saved }, userId });
    return saved;
  }
  async updateNode(id, input, selectedFields = '', currentUserId = null) {
    logger.debug('Updating canvas node', { id, input });
    if (!id) throw new ValidationError('Node ID is required', 'id');
    const node = await CanvasNode.findOne({ _id: id, isDeleted: false });
    if (!node) throw new NotFoundError('CanvasNode');
    if (currentUserId && node.userId !== currentUserId) throw new ValidationError('No access');
    if (input.position && (typeof input.position.x !== 'number' || typeof input.position.y !== 'number')) {
      throw new ValidationError('Position (x, y) must be numbers', 'position');
    }
    Object.assign(node, input);
    await node.save();
    const updated = await CanvasNode.findById(id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvasNode', id.toString(), updated, [node.userId], node.userId);
    // Publish event
    pubsub.publish('CANVAS_NODE_UPDATED', { canvasNodeUpdated: { success: true, data: updated }, userId: node.userId });
    return updated;
  }
  async deleteNode(id, selectedFields = '', currentUserId = null) {
    logger.debug('Deleting canvas node', { id });
    if (!id) throw new ValidationError('Node ID is required', 'id');
    const node = await CanvasNode.findOne({ _id: id, isDeleted: false });
    if (!node) throw new NotFoundError('CanvasNode');
    if (currentUserId && node.userId !== currentUserId) throw new ValidationError('No access');
    node.isDeleted = true;
    await node.save();
    await redisClient.invalidateByPattern(`user:${node.userId}:canvasNode:*`);
    const deleted = await CanvasNode.findById(id).select(selectedFields || '').lean();
    // Publish event
    pubsub.publish('CANVAS_NODE_DELETED', { canvasNodeDeleted: { success: true, data: deleted }, userId: node.userId });
    return deleted;
  }
  async getNode(id, selectedFields = '', currentUserId = null) {
    if (!id) throw new ValidationError('Node ID is required', 'id');
    const node = await CanvasNode.findOne({ _id: id, isDeleted: false }).select(selectedFields || '').lean();
    if (!node) throw new NotFoundError('CanvasNode');
    if (currentUserId && node.userId !== currentUserId) throw new ValidationError('No access');
    return node;
  }
  // Edge CRUD
  async createEdge(input, selectedFields = '', currentUserId = null) {
    logger.debug('Creating canvas edge', { input });
    const userId = currentUserId || input.userId;
    if (!userId) throw new ValidationError('User ID is required', 'userId');
    if (!input.canvasId) throw new ValidationError('Canvas ID is required', 'canvasId');
    if (!input.source || !input.target) throw new ValidationError('Source and target required', 'source/target');
    
    // Create the edge
    const edge = new CanvasEdge({ ...input, userId });
    await edge.save();

    // Update the canvas to include this edge
    await Canvas.findByIdAndUpdate(
      input.canvasId,
      { $push: { edges: edge._id } },
      { new: true }
    );

    const saved = await CanvasEdge.findById(edge._id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvasEdge', edge._id.toString(), saved, [userId], userId);
    // Publish event
    pubsub.publish('CANVAS_EDGE_CREATED', { canvasEdgeCreated: { success: true, data: saved }, userId });
    return saved;
  }
  async updateEdge(id, input, selectedFields = '', currentUserId = null) {
    logger.debug('Updating canvas edge', { id, input });
    if (!id) throw new ValidationError('Edge ID is required', 'id');
    const edge = await CanvasEdge.findOne({ _id: id, isDeleted: false });
    if (!edge) throw new NotFoundError('CanvasEdge');
    if (currentUserId && edge.userId !== currentUserId) throw new ValidationError('No access');
    Object.assign(edge, input);
    await edge.save();
    const updated = await CanvasEdge.findById(id).select(selectedFields || '').lean();
    await redisClient.setEntity('canvasEdge', id.toString(), updated, [edge.userId], edge.userId);
    // Publish event
    pubsub.publish('CANVAS_EDGE_UPDATED', { canvasEdgeUpdated: { success: true, data: updated }, userId: edge.userId });
    return updated;
  }
  async deleteEdge(id, selectedFields = '', currentUserId = null) {
    logger.debug('Deleting canvas edge', { id });
    if (!id) throw new ValidationError('Edge ID is required', 'id');
    const edge = await CanvasEdge.findOne({ _id: id, isDeleted: false });
    if (!edge) throw new NotFoundError('CanvasEdge');
    if (currentUserId && edge.userId !== currentUserId) throw new ValidationError('No access');
    edge.isDeleted = true;
    await edge.save();
    await redisClient.invalidateByPattern(`user:${edge.userId}:canvasEdge:*`);
    const deleted = await CanvasEdge.findById(id).select(selectedFields || '').lean();
    // Publish event
    pubsub.publish('CANVAS_EDGE_DELETED', { canvasEdgeDeleted: { success: true, data: deleted }, userId: edge.userId });
    return deleted;
  }
  async getEdge(id, selectedFields = '', currentUserId = null) {
    if (!id) throw new ValidationError('Edge ID is required', 'id');
    const edge = await CanvasEdge.findOne({ _id: id, isDeleted: false }).select(selectedFields || '').lean();
    if (!edge) throw new NotFoundError('CanvasEdge');
    if (currentUserId && edge.userId !== currentUserId) throw new ValidationError('No access');
    return edge;
  }
}

module.exports = new CanvasService(); 