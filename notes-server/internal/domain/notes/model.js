const mongoose = require('mongoose');
const { Schema, ObjectId } = mongoose;
const { updateBidirectionalLinks, handleCascadingDelete } = require('./linkService');

const NotePageSchema = new Schema({
  userId: { 
    type: String, // UUID from Go backend
    required: [true, 'User ID is required'], 
    index: true 
  },
  title: { 
    type: String, 
    required: [true, 'Title is required'],
    trim: true,
    maxlength: [200, 'Title cannot be more than 200 characters']
  },
  content: { 
    type: String, 
    default: '',
    trim: true,
    maxlength: [10000, 'Content cannot be more than 10000 characters']
  },
  linksOut: [{ 
    type: ObjectId, 
    ref: 'NotePage',
    validate: {
      validator: function(v) {
        return v.toString() !== this._id.toString();
      },
      message: 'Cannot link to self'
    }
  }],
  linksIn: [{ 
    type: ObjectId, 
    ref: 'NotePage' 
  }],
  entities: [{
    type: { 
      type: String, 
      enum: ['idea', 'tasks', 'person', 'todos'], 
      required: true 
    },
    refId: { type: ObjectId }
  }],
  template: { type: ObjectId, ref: 'Template' },
  tags: [{ 
    type: String, 
    index: true,
    trim: true,
    maxlength: [50, 'Tag cannot be more than 50 characters']
  }],
  isDeleted: { 
    type: Boolean, 
    default: false,
    index: true
  },
  favorited: { 
    type: Boolean, 
    default: false,
    index: true
  },
  icon: { 
    type: String,
    trim: true,
    maxlength: [50, 'Icon name or emoji too long']
  },
  sharedWith: [{
    type: String, // UUID from Go backend
    index: true
  }],
  permissions: [{
    userId: { type: String, required: true }, // UUID from Go backend
    level: { type: String, enum: ['view', 'edit', 'comment'], default: 'view' }
  }]
}, {
  timestamps: true,
  toJSON: { virtuals: true },
  toObject: { virtuals: true }
});

// Virtual for linked notes count
NotePageSchema.virtual('linkedNotesCount').get(function() {
  return this.linksOut.length + this.linksIn.length;
});

// Full-text search index with weights
NotePageSchema.index(
  { title: 'text', content: 'text' },
  { 
    weights: { 
      title: 10, 
      content: 5 
    },
    name: 'text_search'
  }
);

// Compound indexes for common queries
NotePageSchema.index({ userId: 1, createdAt: -1 });
NotePageSchema.index({ userId: 1, favorited: 1, createdAt: -1 });
NotePageSchema.index({ userId: 1, tags: 1, createdAt: -1 });
NotePageSchema.index({ userId: 1, isDeleted: 1 });

// Pre-save middleware to ensure unique tags
NotePageSchema.pre('save', function(next) {
  if (this.tags) {
    this.tags = [...new Set(this.tags)];
  }
  next();
});

// Store old linksOut before saving
NotePageSchema.pre('save', function(next) {
  if (this.isModified('linksOut')) {
    this._oldLinksOut = this.linksOut;
  }
  next();
});

// Maintain bi-directional links after saving
NotePageSchema.post('save', async function(doc, next) {
  try {
    if (doc.isModified('linksOut')) {
      const oldLinksOut = doc._oldLinksOut || [];
      await updateBidirectionalLinks(doc._id, oldLinksOut, doc.linksOut);
    }
    next();
  } catch (error) {
    next(error);
  }
});

// Handle cascading deletes
NotePageSchema.post('save', async function(doc, next) {
  try {
    if (doc.isDeleted) {
      await handleCascadingDelete(doc._id);
    }
    next();
  } catch (error) {
    next(error);
  }
});

// Method to check if note is linked to another note
NotePageSchema.methods.isLinkedTo = function(noteId) {
  return this.linksOut.includes(noteId) || this.linksIn.includes(noteId);
};

// Static method to find notes by tag
NotePageSchema.statics.findByTag = function(tag) {
  return this.find({ tags: tag, isDeleted: false });
};

module.exports = mongoose.model('NotePage', NotePageSchema); 