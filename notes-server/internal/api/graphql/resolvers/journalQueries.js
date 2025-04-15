const { GraphQLID, GraphQLString, GraphQLInt, GraphQLBoolean } = require('graphql');
const { 
  JournalResponseType, 
  JournalListResponseType,
  JournalSortFieldEnum,
  SortOrderEnum,
  JournalFilterInput,
  getSelectedFields
} = require('../schemas/journalTypes');
const { createErrorResponse } = require('../schemas/responseTypes');
const journalService = require('../../../domain/journals/journalService');
const { ValidationError, NotFoundError } = require('../../../../pkg/utils/errorHandler');
const { logger } = require('../../../../pkg/utils/logger');

const journalQueries = {
  journal: {
    type: JournalResponseType,
    args: { 
      id: { type: GraphQLID },
      includeArchived: { type: GraphQLBoolean, defaultValue: false }
    },
    async resolve(parent, args, context, info) {
      try {
        const selectedFields = getSelectedFields(info);
        const journal = await journalService.getJournal(args.id, selectedFields, args.includeArchived);
        if (journal.isDeleted) {
          throw new NotFoundError('Journal');
        }
        
        return {
          success: true,
          message: 'Journal retrieved successfully',
          data: journal,
          errors: null
        };
      } catch (error) {
        logger.error('Error in journal query', {
          error: error.message,
          stack: error.stack,
          journalId: args.id
        });
        return createErrorResponse(
          error.message,
          [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 
                  error instanceof NotFoundError ? 'NOT_FOUND' : 'INTERNAL_ERROR'
          }]
        );
      }
    }
  },

  journals: {
    type: JournalListResponseType,
    args: { 
      userId: { type: GraphQLID },
      page: { type: GraphQLInt, defaultValue: 1 },
      limit: { type: GraphQLInt, defaultValue: 10 },
      sortField: { type: JournalSortFieldEnum, defaultValue: 'date' },
      sortOrder: { type: SortOrderEnum, defaultValue: -1 },
      filter: { type: JournalFilterInput }
    },
    async resolve(parent, args, context, info) {
      try {
        const selectedFields = getSelectedFields(info);
        const currentUserId = context.user && context.user.id;
        // Only allow userId if it matches context.user.id
        let userId = args.userId || currentUserId;
        if (args.userId && args.userId !== currentUserId) {
          throw new ValidationError('You are not authorized to access other users\' journals', 'userId');
        }
        return await journalService.getJournals({ ...args, userId }, selectedFields);
      } catch (error) {
        logger.error('Error in journals query', {
          error: error.message,
          stack: error.stack,
          userId: args.userId,
          page: args.page,
          limit: args.limit
        });
        return createErrorResponse(
          error.message,
          [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 'INTERNAL_ERROR'
          }]
        );
      }
    }
  },

  journalsByDateRange: {
    type: JournalListResponseType,
    args: {
      userId: { type: GraphQLID },
      startDate: { type: GraphQLString },
      endDate: { type: GraphQLString }
    },
    async resolve(parent, args, context, info) {
      try {
        const selectedFields = getSelectedFields(info);
        const currentUserId = context.user && context.user.id;
        // Only allow userId if it matches context.user.id
        let userId = args.userId || currentUserId;
        if (args.userId && args.userId !== currentUserId) {
          throw new ValidationError('You are not authorized to access other users\' journals', 'userId');
        }
        const journals = await journalService.getJournalsByDateRange(
          new Date(args.startDate),
          new Date(args.endDate),
          userId,
          selectedFields
        );
        // Filter out deleted journals
        const filteredJournals = journals.filter(j => !j.isDeleted);
        return {
          success: true,
          message: 'Journals retrieved successfully',
          data: filteredJournals,
          pageInfo: {
            totalItems: filteredJournals.length,
            currentPage: 1,
            totalPages: 1
          }
        };
      } catch (error) {
        logger.error('Error in journalsByDateRange query', {
          error: error.message,
          stack: error.stack,
          userId: args.userId,
          startDate: args.startDate,
          endDate: args.endDate
        });
        return createErrorResponse(
          error.message,
          [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 'INTERNAL_ERROR'
          }]
        );
      }
    }
  }
};

module.exports = journalQueries; 