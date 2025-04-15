const mongoose = require('mongoose');
const { Schema, ObjectId } = mongoose;

const JournalSchema = new Schema({
  userId: {
    type: String,
    required: [true, 'User ID is required'],
    index: true
  },
  title: {
    type: String,
    required: [true, 'Title is required'],
    trim: true,
    maxlength: [200, 'Title cannot be more than 200 characters']
  },
  date: {
    type: Date,
    required: [true, 'Date is required'],
    index: true
  },
  content: {
    type: String,
    default: '',
    trim: true,
    maxlength: [10000, 'Content cannot be more than 10000 characters']
  },
  mood: {
    type: String,
    enum: ['happy', 'sad', 'angry', 'neutral', 'excited', 'anxious', 'tired', 'grateful'],
    default: null
  },
  tags: [{
    type: String,
    index: true,
    trim: true,
    maxlength: [50, 'Tag cannot be more than 50 characters']
  }],
  aiPromptUsed: {
    type: String,
    default: null
  },
  aiGenerated: {
    type: Boolean,
    default: false
  },
  archived: {
    type: Boolean,
    default: false,
    index: true
  },
  wordCount: {
    type: Number,
    default: 0
  },
  isDeleted: {
    type: Boolean,
    default: false,
    index: true
  }
}, {
  timestamps: true,
  toJSON: { virtuals: true },
  toObject: { virtuals: true }
});

// Full-text search index with weights
JournalSchema.index(
  { title: 'text', content: 'text' },
  {
    weights: {
      title: 10,
      content: 5
    },
    name: 'journal_text_search'
  }
);

// Compound indexes for common queries
JournalSchema.index({ userId: 1, date: -1 });
JournalSchema.index({ userId: 1, archived: 1, date: -1 });
JournalSchema.index({ userId: 1, tags: 1, date: -1 });
JournalSchema.index({ userId: 1, mood: 1, date: -1 });
JournalSchema.index({ userId: 1, isDeleted: 1 });

// Pre-save middleware to ensure unique tags
JournalSchema.pre('save', function (next) {
  if (this.tags) {
    this.tags = [...new Set(this.tags)];
  }
  next();
});

// Pre-save middleware to calculate word count
JournalSchema.pre('save', function (next) {
  if (this.isModified('content')) {
    this.wordCount = this.content.trim().split(/\s+/).length;
  }
  next();
});

//instance method to check if journal is archived
JournalSchema.methods.isArchived = function () {
  return this.archived;
};

// Add static method to find journals by date range
JournalSchema.statics.findByDateRange = function (startDate, endDate, userId) {
  return this.find({
    userId,
    date: {
      $gte: startDate,
      $lte: endDate
    },
    archived: false,
    isDeleted: false
  });
};

// Static method to find journals by tag
JournalSchema.statics.findByTag = function (tag) {
  return this.find({ tags: tag, archived: false, isDeleted: false });
};

// Static method to find journals by mood
JournalSchema.statics.findByMood = function (mood) {
  return this.find({ mood, archived: false, isDeleted: false });
};

// Static method to get mood summary for dashboard
JournalSchema.statics.getMoodSummary = async function (userId) {
  const thirtyDaysAgo = new Date();
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

  const moodCounts = await this.aggregate([
    {
      $match: {
        userId: userId,
        date: { $gte: thirtyDaysAgo },
        mood: { $ne: null },
        isDeleted: false
      }
    },
    { $group: { _id: "$mood", count: { $sum: 1 } } },
    { $sort: { count: -1 } }
  ]);

  return JSON.stringify(moodCounts);
};

module.exports = mongoose.model('Journal', JournalSchema);