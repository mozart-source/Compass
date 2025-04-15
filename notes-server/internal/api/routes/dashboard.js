const express = require('express');
const { logger } = require('../../../pkg/utils/logger');
const { dashboardMetrics } = require('../graphql/resolvers/dashboardQueries');
const { userContextMiddleware } = require('../middleware');

const router = express.Router();

/**
 * @route   GET /api/dashboard/metrics
 * @desc    Get dashboard metrics for a user
 * @access  Private
 */
router.get('/metrics', userContextMiddleware, async (req, res) => {
  try {
    const userId = req.user.id;

    logger.info('REST API request for dashboard metrics', { userId });

    // Reuse the GraphQL resolver logic
    const metrics = await dashboardMetrics.resolve(null, { userId }, {});

    res.json(metrics);
  } catch (error) {
    logger.error('Error fetching dashboard metrics', { error: error.message });
    res.status(500).json({ error: 'Failed to fetch dashboard metrics' });
  }
});

module.exports = router;