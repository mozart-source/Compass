const NotePage = require('./model');
const { updateBidirectionalLinks, validateLinks } = require('./linkService');
const { ValidationError, NotFoundError, DatabaseError } = require('../../../pkg/utils/errorHandler');
const { logger } = require('../../../pkg/utils/logger');
const RedisService = require('../../infrastructure/cache/redisService');
const redisConfig = require('../../infrastructure/cache/config');
const { dashboardEvents } = require('../../infrastructure/cache/dashboardEvents');

const redisClient = new RedisService(redisConfig);

class NoteService {
  /**
   * Create a new note page
   * @param {Object} input - Note data
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Created note
   */
  async createNote(input, selectedFields = '', currentUserId = null) {
    logger.debug('Creating note page', { input });
    // Always set userId from currentUserId, ignore input.userId
    const userId = currentUserId || input.userId;
    if (!userId) {
      throw new ValidationError('User ID is required', 'userId');
    }
    if (!input.title?.trim()) {
      throw new ValidationError('Title is required', 'title');
    }
    // Validate links if provided
    if (input.linksOut && input.linksOut.length > 0) {
      try {
        await validateLinks(input.linksOut);
      } catch (error) {
        throw new ValidationError(error.message, 'linksOut');
      }
    }
    // Always set userId from trusted backend
    const note = new NotePage({ ...input, userId });
    await note.save();
    logger.info('Note page created', { noteId: note._id });
    const savedNote = await NotePage.findById(note._id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!savedNote || !savedNote.userId) {
      throw new Error('Saved note is missing userId or does not exist');
    }

    // Cache the new note using unified API
    await redisClient.setEntity(
      'note',
      note._id.toString(),
      savedNote,
      [userId, ...(Array.isArray(input.tags) ? input.tags : [])],
      userId
    );

    // Invalidate user's note list cache
    await redisClient.invalidateByPattern(`user:${userId}:notes:*`);

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(userId, note._id.toString(), null, {
      action: 'create',
      entityType: 'note'
    });

    logger.info('Note page creation completed', {
      noteId: note._id,
      userId: userId
    });

    return savedNote;
  }

  /**
   * Update an existing note page
   * @param {string} id - Note ID
   * @param {Object} input - Updated note data
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Updated note
   */
  async updateNote(id, input, selectedFields = '', currentUserId = null) {
    logger.debug('Updating note page', { noteId: id, input });
    if (!id) {
      throw new ValidationError('Note ID is required', 'id');
    }
    const note = await NotePage.findOne({ _id: id, isDeleted: false });
    if (!note) {
      throw new NotFoundError('Note');
    }
    if (currentUserId && note.userId.toString() !== currentUserId && !(note.sharedWith || []).includes(currentUserId)) {
      throw new ValidationError('You do not have access to update this note');
    }
    if (input.title !== undefined && !input.title?.trim()) {
      throw new ValidationError('Title is required', 'title');
    }
    if (input.linksOut?.length > 0) {
      await validateLinks(input.linksOut);
    }
    await redisClient.invalidateByTags([
      note.userId.toString(),
      ...(Array.isArray(note.tags) ? note.tags : [])
    ]);
    await redisClient.invalidateByPattern(redisClient.generateKey('note', id));
    // Always set userId from currentUserId
    Object.assign(note, { ...input, userId: currentUserId });
    await note.save();
    logger.info('Note page updated', { noteId: id });

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(note.userId.toString(), id, null, {
      action: 'update',
      entityType: 'note'
    });
    const updatedNote = await NotePage.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!updatedNote || !updatedNote.userId) {
      throw new Error('Updated note is missing userId or does not exist');
    }
    await redisClient.invalidateByTags([
      updatedNote.userId.toString(),
      ...(Array.isArray(updatedNote.tags) ? updatedNote.tags : [])
    ]);
    await redisClient.setEntity(
      'note',
      id.toString(),
      updatedNote,
      [updatedNote.userId, ...(Array.isArray(updatedNote.tags) ? updatedNote.tags : [])],
      updatedNote.userId
    );
    await redisClient.invalidateByPattern(`user:${updatedNote.userId}:notes:*`);
    logger.info('Note page update completed', { noteId: id, userId: note.userId });
    return updatedNote;
  }

  /**
   * Delete (soft delete) a note page
   * @param {string} id - Note ID
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Deleted note
   */
  async deleteNote(id, selectedFields = '', currentUserId = null) {
    logger.debug('Deleting note page', { noteId: id });

    if (!id) {
      throw new ValidationError('Note ID is required', 'id');
    }

    const note = await NotePage.findOne({ _id: id, isDeleted: false });
    if (!note) {
      throw new NotFoundError('Note');
    }

    if (currentUserId && note.userId.toString() !== currentUserId && !(note.sharedWith || []).includes(currentUserId)) {
      throw new ValidationError('You do not have access to delete this note');
    }

    // Invalidate by tags for note before delete
    await redisClient.invalidateByTags([
      note.userId.toString(),
      ...(Array.isArray(note.tags) ? note.tags : [])
    ]);

    // Soft delete
    note.isDeleted = true;
    await note.save();
    logger.info('Note page marked as deleted', { noteId: id });

    const deletedNote = await NotePage.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!deletedNote || !deletedNote.userId) {
      throw new Error('Deleted note is missing userId or does not exist');
    }

    // Invalidate user's note list cache
    await redisClient.invalidateByPattern(`user:${note.userId}:notes:*`);

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(note.userId.toString(), id, null, {
      action: 'delete',
      entityType: 'note'
    });

    logger.info('Note page deletion completed', {
      noteId: id,
      userId: note.userId
    });

    return deletedNote;
  }

  /**
   * Toggle favorite status of a note
   * @param {string} id - Note ID
   * @param {boolean} favorited - Favorite status
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Updated note
   */
  async toggleFavorite(id, favorited, selectedFields = '', currentUserId = null) {
    logger.debug('Toggling favorite status', { noteId: id, favorited });

    if (!id) {
      throw new ValidationError('Note ID is required', 'id');
    }

    if (typeof favorited !== 'boolean') {
      throw new ValidationError('Favorite status is required', 'favorited');
    }

    const note = await NotePage.findOne({ _id: id, isDeleted: false });
    if (!note) {
      throw new NotFoundError('Note');
    }

    if (currentUserId && note.userId.toString() !== currentUserId && !(note.sharedWith || []).includes(currentUserId)) {
      throw new ValidationError('You do not have access to favorite this note');
    }

    note.favorited = favorited;
    await note.save();

    // Update cache
    try {
      await redisClient.setEntity(
        'note',
        id.toString(),
        note.toObject(),
        [note.userId, ...(Array.isArray(note.tags) ? note.tags : [])],
        note.userId
      );
      await redisClient.invalidateByPattern(`user:${note.userId}:notes:*`);
    } catch (cacheError) {
      logger.warn('Failed to update note cache', {
        error: cacheError.message,
        noteId: id
      });
    }

    logger.info('Note favorite status updated', { noteId: id, favorited });

    const updatedNote = await NotePage.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!updatedNote || !updatedNote.userId) {
      throw new Error('Updated note is missing userId or does not exist');
    }

    return updatedNote;
  }

  /**
   * Get a note by ID (with cache)
   * @param {string} id - Note ID
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Note
   */
  async getNote(id, selectedFields = '', currentUserId = null) {
    logger.debug('Getting note by ID', { noteId: id });
    if (!id) {
      throw new ValidationError('Note ID is required', 'id');
    }
    const note = await NotePage.findOne({ _id: id, isDeleted: false })
      .select(selectedFields)
      .lean();
    if (!note) {
      throw new NotFoundError('Note');
    }
    if (currentUserId && note.userId.toString() !== currentUserId && !(note.sharedWith || []).includes(currentUserId)) {
      throw new ValidationError('You do not have access to this note');
    }
    // Try to get from cache first
    const cachedNote = await redisClient.getEntity('note', id);
    if (cachedNote) {
      logger.debug('Note retrieved from cache', { noteId: id });
      return cachedNote;
    }
    // Cache the note
    await redisClient.setEntity(
      'note',
      id.toString(),
      note,
      [note.userId, ...(Array.isArray(note.tags) ? note.tags : [])],
      note.userId
    );
    logger.debug('Note cached', { noteId: id });
    return note;
  }

  /**
   * Get notes with pagination, filtering, and search (with cache)
   * @param {Object} params - Pagination, filtering, and search parameters
   * @param {string} selectedFields - Fields to select
   * @returns {Object} Notes and pagination information
   */
  async getNotes({ userId, search, page = 1, limit = 10, sortField = 'createdAt', sortOrder = -1, filter = {} }, selectedFields = '') {
    logger.debug('Getting notes', { userId, page, limit, sortField, sortOrder, filter, search });
    // Always require userId from argument, never from input
    if (!userId) {
      throw new ValidationError('User ID is required', 'userId');
    }
    const cacheKey = redisClient.generateListKey(userId, 'notes', { page, limit, sortField, sortOrder, filter, search });
    const cachedResult = await redisClient.getList(cacheKey);
    if (cachedResult) {
      logger.debug('Notes retrieved from cache', { userId, page, limit });
      return cachedResult;
    }
    const skip = (page - 1) * limit;
    const query = { userId, isDeleted: false };
    // Apply filters if provided
    if (filter) {
      if (filter.tags && filter.tags.length > 0) {
        query.tags = { $in: filter.tags };
      }
      if (typeof filter.favorited === 'boolean') {
        query.favorited = filter.favorited;
      }
      if (filter.createdAfter || filter.createdBefore) {
        query.createdAt = {};
        if (filter.createdAfter) {
          query.createdAt.$gte = new Date(filter.createdAfter);
        }
        if (filter.createdBefore) {
          query.createdAt.$lte = new Date(filter.createdBefore);
        }
      }
    }
    const sortOptions = { [sortField]: sortOrder };
    let notes;
    let totalItems;
    if (search?.query) {
      logger.debug('Performing text search', { query: search.query });
      query.$text = {
        $search: search.query,
        $caseSensitive: false,
        $diacriticSensitive: false
      };
      notes = await NotePage.find(query)
        .select(selectedFields || 'title content tags favorited icon createdAt updatedAt userId')
        .sort({ score: { $meta: 'textScore' }, ...sortOptions })
        .skip(skip)
        .limit(limit)
        .lean();
    } else {
      notes = await NotePage.find(query)
        .select(selectedFields || 'title content tags favorited icon createdAt updatedAt userId')
        .sort(sortOptions)
        .skip(skip)
        .limit(limit)
        .lean();
    }
    totalItems = await NotePage.countDocuments(query);
    const result = {
      success: true,
      message: 'Notes retrieved successfully',
      data: notes,
      pageInfo: {
        totalItems,
        currentPage: page,
        totalPages: Math.ceil(totalItems / limit)
      }
    };
    // Collect all tags from the result set for robust invalidation
    const allTags = Array.from(new Set(notes.flatMap(n => n.tags || [])));
    await redisClient.setList(cacheKey, result, [userId, ...allTags], userId);
    logger.debug('Notes cached', { userId, page, limit, totalItems });
    logger.info('Notes retrieved successfully', { userId, page, limit, totalItems, hasSearch: !!search?.query });
    return result;
  }
}

// Collaboration methods
NoteService.prototype.shareNote = async function (noteId, userId, level = 'view', currentUserId, selectedFields = '') {
  const note = await NotePage.findById(noteId);
  if (!note) throw new NotFoundError('Note');
  if (note.userId.toString() !== currentUserId) throw new ValidationError('Only the owner can share this note');
  if (!note.sharedWith.includes(userId)) note.sharedWith.push(userId);
  const existingPerm = note.permissions.find(p => p.userId === userId);
  if (existingPerm) {
    existingPerm.level = level;
  } else {
    note.permissions.push({ userId, level });
  }
  await note.save();
  return await NotePage.findById(noteId).select(selectedFields).lean();
};

NoteService.prototype.unshareNote = async function (noteId, userId, currentUserId, selectedFields = '') {
  const note = await NotePage.findById(noteId);
  if (!note) throw new NotFoundError('Note');
  if (note.userId.toString() !== currentUserId) throw new ValidationError('Only the owner can unshare this note');
  note.sharedWith = note.sharedWith.filter(id => id !== userId);
  note.permissions = note.permissions.filter(p => p.userId !== userId);
  await note.save();
  return await NotePage.findById(noteId).select(selectedFields).lean();
};

NoteService.prototype.getNotesSharedWithUser = async function (userId, { page = 1, limit = 10, filter = {} } = {}, selectedFields = '') {
  const skip = (page - 1) * limit;
  const query = { sharedWith: userId, isDeleted: false };
  if (filter.tags && filter.tags.length > 0) query.tags = { $in: filter.tags };
  if (typeof filter.favorited === 'boolean') query.favorited = filter.favorited;
  if (filter.createdAfter || filter.createdBefore) {
    query.createdAt = {};
    if (filter.createdAfter) query.createdAt.$gte = new Date(filter.createdAfter);
    if (filter.createdBefore) query.createdAt.$lte = new Date(filter.createdBefore);
  }
  const notes = await NotePage.find(query)
    .select(selectedFields)
    .skip(skip)
    .limit(limit)
    .lean();
  const totalItems = await NotePage.countDocuments(query);
  return {
    success: true,
    message: 'Notes shared with you retrieved successfully',
    data: notes,
    pageInfo: {
      totalItems,
      currentPage: page,
      totalPages: Math.ceil(totalItems / limit)
    }
  };
};

module.exports = new NoteService();