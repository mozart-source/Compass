const mongoose = require('mongoose');
const { Schema, ObjectId } = mongoose;

const CanvasNodeSchema = new Schema({
  canvasId: { type: ObjectId, ref: 'Canvas', required: true, index: true },
  type: { type: String, default: 'default', trim: true },
  data: { type: Object, default: {} },
  position: {
    x: { type: Number, required: true },
    y: { type: Number, required: true }
  },
  style: { type: Object, default: {} },
  label: { type: String, trim: true },
  color: { type: String, trim: true },
  userId: { type: String, required: true, index: true },
  isDeleted: { type: Boolean, default: false, index: true }
}, { timestamps: true });

const CanvasEdgeSchema = new Schema({
  canvasId: { type: ObjectId, ref: 'Canvas', required: true, index: true },
  source: { type: ObjectId, ref: 'CanvasNode', required: true },
  target: { type: ObjectId, ref: 'CanvasNode', required: true },
  type: { type: String, default: 'default', trim: true },
  data: { type: Object, default: {} },
  label: { type: String, trim: true },
  style: { type: Object, default: {} },
  userId: { type: String, required: true, index: true },
  isDeleted: { type: Boolean, default: false, index: true }
}, { timestamps: true });

const CanvasSchema = new Schema({
  userId: { type: String, required: true, index: true },
  title: { type: String, required: true, trim: true, maxlength: 200 },
  description: { type: String, trim: true, maxlength: 1000 },
  nodes: [{ type: ObjectId, ref: 'CanvasNode' }],
  edges: [{ type: ObjectId, ref: 'CanvasEdge' }],
  tags: [{ type: String, trim: true, maxlength: 50, index: true }],
  isDeleted: { type: Boolean, default: false, index: true },
  sharedWith: [{ type: String, index: true }],
  permissions: [{
    userId: { type: String, required: true },
    level: { type: String, enum: ['view', 'edit', 'comment'], default: 'view' }
  }]
}, { timestamps: true });

CanvasSchema.index({ userId: 1, createdAt: -1 });
CanvasSchema.index({ userId: 1, tags: 1, createdAt: -1 });
CanvasSchema.index({ userId: 1, isDeleted: 1 });

module.exports = {
  Canvas: mongoose.model('Canvas', CanvasSchema),
  CanvasNode: mongoose.model('CanvasNode', CanvasNodeSchema),
  CanvasEdge: mongoose.model('CanvasEdge', CanvasEdgeSchema)
}; 