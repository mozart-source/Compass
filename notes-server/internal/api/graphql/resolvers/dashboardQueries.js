const { DashboardMetricsType } = require('../schemas/dashboardTypes');
const Journal = require('../../../domain/journals/model');
const Note = require('../../../domain/notes/model');
const RedisService = require('../../../infrastructure/cache/redisService');
const redisConfig = require('../../../infrastructure/cache/config');
const { dashboardEvents } = require('../../../infrastructure/cache/dashboardEvents');
const { logger } = require('../../../../pkg/utils/logger');

const redisClient = new RedisService(redisConfig);

const dashboardMetrics = {
  type: DashboardMetricsType,
  args: { userId: { type: require('graphql').GraphQLID } },
  async resolve(parent, args, context) {
    try {
      const userId = args.userId;

      // Check cache first
      const cacheKey = `compass:notes:dashboard:metrics:${userId}`;
      const cachedMetrics = await redisClient.get(cacheKey);

      if (cachedMetrics) {
        logger.debug('Returning cached dashboard metrics', { userId });
        return {
          success: true,
          data: JSON.parse(cachedMetrics)
        };
      }

      logger.info('Generating dashboard metrics', { userId });

      // Get mood summary from journals with error handling
      let moodSummary = null;
      let moodDistribution = {};
      try {
        moodSummary = await Journal.getMoodSummary(userId);
        if (moodSummary) {
          moodDistribution = JSON.parse(moodSummary);
        }
      } catch (error) {
        logger.warn('Failed to get mood summary', { error: error.message, userId });
      }

      // Count notes and journals with error handling
      let notesCount = 0;
      let journalsCount = 0;
      try {
        notesCount = await Note.countDocuments({ userId, isDeleted: false });
        journalsCount = await Journal.countDocuments({ userId, isDeleted: false });
      } catch (error) {
        logger.warn('Failed to count documents', { error: error.message, userId });
      }

      // Get recent notes (last 5) with error handling
      let recentNotes = [];
      try {
        recentNotes = await Note.find({ userId, isDeleted: false })
          .sort({ updatedAt: -1 })
          .limit(5)
          .select('title content updatedAt')
          .lean();
      } catch (error) {
        logger.warn('Failed to get recent notes', { error: error.message, userId });
      }

      // Get recent journals (last 5) with error handling
      let recentJournals = [];
      try {
        recentJournals = await Journal.find({ userId, isDeleted: false })
          .sort({ date: -1 })
          .limit(5)
          .select('title content date mood')
          .lean();
      } catch (error) {
        logger.warn('Failed to get recent journals', { error: error.message, userId });
      }

      // Get tag distribution with error handling
      let tagCounts = [];
      try {
        tagCounts = await Note.aggregate([
          { $match: { userId, isDeleted: false } },
          { $unwind: "$tags" },
          { $group: { _id: "$tags", count: { $sum: 1 } } },
          { $sort: { count: -1 } },
          { $limit: 10 }
        ]);
      } catch (error) {
        logger.warn('Failed to get tag distribution', { error: error.message, userId });
      }

      const metrics = {
        mood: moodSummary,
        moodDistribution,
        notes: {
          count: notesCount,
          recent: recentNotes
        },
        journals: {
          count: journalsCount,
          recent: recentJournals
        },
        tags: tagCounts,
        timestamp: new Date().toISOString()
      };

      // Cache the metrics
      await redisClient.set(cacheKey, JSON.stringify(metrics), 300); // Cache for 5 minutes

      // Publish metrics update event
      try {
        await dashboardEvents.publishMetricsUpdate(userId, null, metrics);
      } catch (error) {
        logger.warn('Failed to publish metrics update event', { error: error.message, userId });
      }

      return {
        success: true,
        data: metrics
      };
    } catch (error) {
      logger.error('Error fetching dashboard metrics', { error: error.message, userId: args.userId });
      
      // Return empty structure on error instead of throwing
      return {
        success: false,
        data: {
          mood: null,
          moodDistribution: {},
          notes: {
            count: 0,
            recent: []
          },
          journals: {
            count: 0,
            recent: []
          },
          tags: [],
          timestamp: new Date().toISOString()
        },
        error: 'Failed to fetch dashboard metrics'
      };
    }
  }
};

module.exports = { dashboardMetrics };