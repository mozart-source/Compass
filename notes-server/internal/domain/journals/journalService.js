const Journal = require('./model');
const { ValidationError, NotFoundError, DatabaseError } = require('../../../pkg/utils/errorHandler');
const { logger } = require('../../../pkg/utils/logger');
const RedisService = require('../../infrastructure/cache/redisService');
const redisConfig = require('../../infrastructure/cache/config');
const { dashboardEvents } = require('../../infrastructure/cache/dashboardEvents');

const redisClient = new RedisService(redisConfig);

class JournalService {
  /**
   * Create a new journal entry
   * @param {Object} input - Journal data
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Created journal
   */
  async createJournal(input, selectedFields = '', currentUserId = null) {
    logger.debug('Creating journal page', { input });
    // Always set userId from currentUserId, ignore input.userId
    const userId = currentUserId || input.userId;
    if (!userId) {
      throw new ValidationError('User ID is required', 'userId');
    }
    if (!input.title?.trim()) {
      throw new ValidationError('Title is required', 'title');
    }
    if (!input.date) {
      throw new ValidationError('Date is required', 'date');
    }
    // Always set userId from trusted backend
    const journal = new Journal({ ...input, userId });
    await journal.save();
    logger.info('Journal entry created', { journalId: journal._id });
    const savedJournal = await Journal.findById(journal._id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!savedJournal || !savedJournal.userId) {
      throw new Error('Saved journal is missing userId or does not exist');
    }

    // Cache the new journal using unified API
    await redisClient.setEntity(
      'journal',
      journal._id.toString(),
      savedJournal,
      [userId, ...(Array.isArray(input.tags) ? input.tags : [])],
      userId
    );

    // Invalidate user's journal list cache
    await redisClient.invalidateByPattern(`user:${userId}:journals:*`);

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(userId, journal._id.toString(), null, {
      action: 'create',
      entityType: 'journal'
    });

    logger.info('Journal entry creation completed', {
      journalId: journal._id,
      userId: userId
    });

    return savedJournal;
  }

  /**
   * Update an existing journal entry
   * @param {string} id - Journal ID
   * @param {Object} input - Updated journal data
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Updated journal
   */
  async updateJournal(id, input, selectedFields = '', currentUserId = null) {
    logger.debug('Updating journal entry', { journalId: id, input });
    if (!id) {
      throw new ValidationError('Journal ID is required', 'id');
    }
    const journal = await Journal.findOne({ _id: id, isDeleted: false });
    if (!journal) {
      throw new NotFoundError('Journal');
    }
    if (currentUserId && journal.userId.toString() !== currentUserId) {
      throw new ValidationError('You do not have access to update this journal');
    }

    // Only validate title if it's being updated and is empty
    if (input.title !== undefined && !input.title?.trim()) {
      throw new ValidationError('Title is required', 'title');
    }
    if (input.date !== undefined && !input.date) {
      throw new ValidationError('Date is required', 'date');
    }

    // If title is not provided in input, keep the existing title
    if (input.title === undefined) {
      delete input.title;
    }

    await redisClient.invalidateByTags([
      journal.userId.toString(),
      ...(Array.isArray(journal.tags) ? journal.tags : [])
    ]);
    await redisClient.invalidateByPattern(redisClient.generateKey('journal', id));

    // Only update the fields that are provided in input
    // Preserve userId and other required fields
    const updatedFields = { ...input };
    delete updatedFields.userId; // Remove userId from input if present
    Object.assign(journal, updatedFields);

    await journal.save();
    logger.info('Journal page updated', { journalId: id });

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(journal.userId.toString(), id, null, {
      action: 'update',
      entityType: 'journal'
    });
    const updatedJournal = await Journal.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!updatedJournal || !updatedJournal.userId) {
      throw new Error('Updated journal is missing userId or does not exist');
    }
    await redisClient.invalidateByTags([
      updatedJournal.userId.toString(),
      ...(Array.isArray(updatedJournal.tags) ? updatedJournal.tags : [])
    ]);
    await redisClient.setEntity(
      'journal',
      id.toString(),
      updatedJournal,
      [updatedJournal.userId, ...(Array.isArray(updatedJournal.tags) ? updatedJournal.tags : [])],
      updatedJournal.userId
    );
    await redisClient.invalidateByPattern(`user:${updatedJournal.userId}:journals:*`);  
    logger.info('Journal page update completed', { journalId: id, userId: journal.userId });
    return updatedJournal;
  }

  /**
   * Permanently delete a journal entry
   * @param {string} id - Journal ID
   * @returns {Object} Deleted journal
   */
  async deleteJournal(id, selectedFields = '') {
    logger.debug('Deleting journal entry', { journalId: id });

    if (!id) {
      throw new ValidationError('Journal ID is required', 'id');
    }

    // Fetch the old journal before deleting
    const journal = await Journal.findOne({ _id: id, isDeleted: false });
    if (!journal) {
      throw new NotFoundError('Journal');
    }

    // Soft delete
    journal.isDeleted = true;
    await journal.save();

    // Invalidate by tags for journal before delete
    await redisClient.invalidateByTags([
      journal.userId.toString(),
      ...(Array.isArray(journal.tags) ? journal.tags : [])
    ]);
    await redisClient.invalidateByPattern(redisClient.generateKey('journal', id));

    const deletedJournal = await Journal.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!deletedJournal || !deletedJournal.userId) {
      throw new Error('Deleted journal is missing userId or does not exist');
    }

    // Invalidate user's journal list cache
    await redisClient.invalidateByPattern(`user:${journal.userId}:journals:*`);

    // Publish dashboard event for metrics update
    await dashboardEvents.publishMetricsUpdate(journal.userId.toString(), id, null, {
      action: 'delete',
      entityType: 'journal'
    });

    logger.info('Journal entry deletion completed', {
      journalId: id,
      userId: journal.userId
    });

    return deletedJournal;
  }

  /**
   * Archive a journal entry
   * @param {string} id - Journal ID
   * @returns {Object} Archived journal
   */
  async archiveJournal(id, selectedFields = '') {
    logger.debug('Archiving journal entry', { journalId: id });

    if (!id) {
      throw new ValidationError('Journal ID is required', 'id');
    }

    const journal = await Journal.findById(id);
    if (!journal) {
      throw new NotFoundError('Journal');
    }

    journal.archived = true;
    await journal.save();
    logger.info('Journal entry archived', { journalId: id });

    const archivedJournal = await Journal.findById(id)
      .select(`${selectedFields} userId tags`)
      .lean();
    if (!archivedJournal || !archivedJournal.userId) {
      throw new Error('Archived journal is missing userId or does not exist');
    }

    // Update the journal cache instead of invalidating it
    await redisClient.setEntity(
      'journal',
      id.toString(),
      archivedJournal,
      [journal.userId.toString(), ...(Array.isArray(journal.tags) ? journal.tags : [])],
      journal.userId
    );

    // Only invalidate the user's journal list caches
    await redisClient.invalidateByPattern(`user:${journal.userId}:journals:*`);

    logger.info('Journal entry archival completed', {
      journalId: id,
      userId: journal.userId
    });

    return archivedJournal;
  }

  /**
   * Get journal entries by date range
   * @param {Date} startDate - Start date
   * @param {Date} endDate - End date
   * @param {string} userId - User ID
   * @returns {Array} Journal entries
   */
  async getJournalsByDateRange(startDate, endDate, userId, selectedFields = '') {
    logger.debug('Getting journals by date range', { startDate, endDate, userId });
    // Always require userId from argument, never from input
    if (!userId) {
      throw new ValidationError('User ID is required', 'userId');
    }

    const cacheKey = redisClient.generateListKey(userId, 'journals:dateRange', { startDate, endDate, selectedFields });
    const cachedResult = await redisClient.getList(cacheKey);
    if (cachedResult) {
      logger.debug('Journals by date range retrieved from cache', { userId, startDate, endDate });
      return cachedResult;
    }

    const journals = await Journal.find({
      date: {
        $gte: startDate,
        $lte: endDate
      },
      userId,
      archived: false,
      isDeleted: false
    })
      .select(selectedFields)
      .lean();

    logger.info('Retrieved journals by date range', {
      userId,
      count: journals.length,
      startDate,
      endDate
    });

    // Cache the result
    const allTags = Array.from(new Set(journals.flatMap(j => j.tags || [])));
    await redisClient.setList(cacheKey, journals, [userId, ...allTags], userId);

    return journals;
  }

  /**
   * Get a single journal by ID
   * @param {string} id - Journal ID
   * @param {string} selectedFields - Fields to select
   * @param {string} currentUserId - User ID
   * @returns {Object} Journal entry
   */
  async getJournal(id, selectedFields = '', currentUserId = null) {
    logger.debug('Getting journal by ID', { journalId: id });
    if (!id) {
      throw new ValidationError('Journal ID is required', 'id');
    }

    const journal = await Journal.findOne({ _id: id, isDeleted: false })
      .select(selectedFields)
      .lean();
    if (!journal) {
      throw new NotFoundError('Journal');
    }

    // Only check user access if currentUserId is provided
    if (currentUserId && journal.userId && journal.userId.toString() !== currentUserId) {
      throw new ValidationError('You do not have access to this journal');
    }

    // Try to get from cache first
    const cachedJournal = await redisClient.getEntity('journal', id);
    if (cachedJournal) {
      logger.debug('Journal retrieved from cache', { journalId: id });
      return cachedJournal;
    }
    // Cache the journal
    await redisClient.setEntity(
      'journal',
      id.toString(),
      journal,
      [journal.userId?.toString() || 'anonymous', ...(Array.isArray(journal.tags) ? journal.tags : [])],
      journal.userId?.toString() || 'anonymous'
    );
    logger.debug('Journal cached', { journalId: id });

    logger.info('Journal retrieved successfully', { journalId: id });
    return journal;
  }

  /**
   * Get journals with pagination, filtering, and search (with cache)
   * @param {Object} params - Pagination, filtering, and search parameters
   * @param {string} selectedFields - Fields to select
   * @returns {Object} Paginated journals response
   */
  async getJournals({ userId, search, page = 1, limit = 10, sortField = 'date', sortOrder = -1, filter = {} }, selectedFields = '') {
    logger.debug('Getting journals', { userId, page, limit, sortField, sortOrder, filter, search });
    // Always require userId from argument, never from input
    if (!userId) {
      throw new ValidationError('User ID is required', 'userId');
    }

    const cacheKey = redisClient.generateListKey(userId, 'journals', { page, limit, sortField, sortOrder, filter, search });
    const cachedResult = await redisClient.getList(cacheKey);
    if (cachedResult) {
      logger.debug('Journals retrieved from cache', { userId, page, limit });
      return cachedResult;
    }
    const skip = (page - 1) * limit;
    const query = { userId, isDeleted: false };
    // Apply filters if provided
    if (filter) {
      // Handle archived status
      if (filter.archived !== undefined) {
        query.archived = filter.archived;
      } else {
        query.archived = false; // Default behavior
      }

      // Handle tags
      if (filter.tags?.length > 0) {
        query.tags = { $in: filter.tags };
      }

      // Handle mood
      if (filter.mood) {
        query.mood = filter.mood;
      }

      // Handle date range
      if (filter.dateFrom || filter.dateTo) {
        query.date = {};
        if (filter.dateFrom) {
          query.date.$gte = new Date(filter.dateFrom);
        }
        if (filter.dateTo) {
          query.date.$lte = new Date(filter.dateTo);
        }
      }

      // Handle word count range
      if (filter.wordCountMin !== undefined || filter.wordCountMax !== undefined) {
        query.wordCount = {};
        if (filter.wordCountMin !== undefined) {
          query.wordCount.$gte = filter.wordCountMin;
        }
        if (filter.wordCountMax !== undefined) {
          query.wordCount.$lte = filter.wordCountMax;
        }
      }

      // Handle AI generated filter
      if (filter.aiGenerated !== undefined) {
        query.aiGenerated = filter.aiGenerated;
      }
    }
    const sortOptions = { [sortField]: sortOrder };
    let journals;
    let totalItems;
    if (search?.query) {
      logger.debug('Performing text search', { query: search.query });
      query.$text = {
        $search: search.query,
        $caseSensitive: false,
        $diacriticSensitive: false
      };
      journals = await Journal.find(query)
        .select(selectedFields || '_id userId title content date mood tags aiPromptUsed aiGenerated archived wordCount createdAt updatedAt')
        .sort({ score: { $meta: 'textScore' }, ...sortOptions })
        .skip(skip)
        .limit(limit)
        .lean();
    } else {
      journals = await Journal.find(query)
        .select(selectedFields || '_id userId title content date mood tags aiPromptUsed aiGenerated archived wordCount createdAt updatedAt')
        .sort(sortOptions)
        .skip(skip)
        .limit(limit)
        .lean();
    }
    totalItems = await Journal.countDocuments(query);
    const result = {
      success: true,
      message: 'Journals retrieved successfully',
      data: journals,
      pageInfo: {
        totalItems,
        currentPage: page,
        totalPages: Math.ceil(totalItems / limit)
      }
    };
    // Collect all tags from the result set for robust invalidation
    const allTags = Array.from(new Set(journals.flatMap(j => j.tags || [])));
    await redisClient.setList(cacheKey, result, [userId, ...allTags], userId);
    logger.debug('Journals cached', { userId, page, limit, totalItems });
    logger.info('Journals retrieved successfully', { userId, page, limit, totalItems, hasSearch: !!search?.query });
    return result;
  }
}

module.exports = new JournalService();