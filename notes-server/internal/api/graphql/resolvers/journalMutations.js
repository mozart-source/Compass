const { GraphQLID } = require('graphql');
const { 
  JournalResponseType, 
  JournalInput,
  getSelectedFields
} = require('../schemas/journalTypes');
const journalService = require('../../../domain/journals/journalService');
const { ValidationError, NotFoundError, DatabaseError } = require('../../../../pkg/utils/errorHandler');
const { logger } = require('../../../../pkg/utils/logger');
const { pubsub } = require('../../../infrastructure/cache/pubsub');

const journalMutations = {
  createJournal: {
    type: JournalResponseType,
    args: { 
      input: { type: JournalInput }
    },
    async resolve(parent, args, context, info) {
      try {
        const { input } = args;
        const selectedFields = getSelectedFields(info) || 'id _id userId title content date mood tags aiPromptUsed aiGenerated archived wordCount createdAt updatedAt isDeleted';
        const currentUserId = context.user && context.user.id;
        const inputWithUser = { ...input, userId: currentUserId };
        const savedJournal = await journalService.createJournal(inputWithUser, selectedFields);
        
        // Publish subscription event
        pubsub.publish('JOURNAL_CREATED', { journalCreated: { success: true, message: 'Journal created', data: savedJournal, errors: null } });
        
        return {
          success: true,
          message: 'Journal entry created successfully',
          data: savedJournal,
          errors: null
        };
      } catch (error) {
        logger.error('Error in createJournal', {
          error: error.message,
          stack: error.stack,
          input: args.input
        });
        return {
          success: false,
          message: error.message,
          data: null,
          errors: [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 
                  error instanceof DatabaseError ? 'DATABASE_ERROR' : 'INTERNAL_ERROR'
          }]
        };
      }
    }
  },
  updateJournal: {
    type: JournalResponseType,
    args: { 
      id: { type: GraphQLID },
      input: { type: JournalInput }
    },
    async resolve(parent, args, context, info) {
      try {
        const { id, input } = args;
        const selectedFields = getSelectedFields(info);
        const currentUserId = context.user && context.user.id;
        const inputWithUser = { ...input, userId: currentUserId };
        const updatedJournal = await journalService.updateJournal(id, inputWithUser, selectedFields);
        
        // Publish subscription event
        pubsub.publish('JOURNAL_UPDATED', { journalUpdated: { success: true, message: 'Journal updated', data: updatedJournal, errors: null } });
        
        return {
          success: true,
          message: 'Journal entry updated successfully',
          data: updatedJournal,
          errors: null
        };
      } catch (error) {
        logger.error('Error in updateJournal', {
          error: error.message,
          stack: error.stack,
          journalId: args.id,
          input: args.input
        });
        return {
          success: false,
          message: error.message,
          data: null,
          errors: [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 
                  error instanceof NotFoundError ? 'NOT_FOUND' :
                  error instanceof DatabaseError ? 'DATABASE_ERROR' : 'INTERNAL_ERROR'
          }]
        };
      }
    }
  },
  deleteJournal: {
    type: JournalResponseType,
    args: { id: { type: GraphQLID } },
    async resolve(parent, args, context, info) {
      try {
        const { id } = args;
        const selectedFields = getSelectedFields(info);
        
        const deletedJournal = await journalService.deleteJournal(id, selectedFields);
        
        // Publish subscription event
        pubsub.publish('JOURNAL_DELETED', { journalDeleted: { success: true, message: 'Journal deleted', data: deletedJournal, errors: null } });
        
        return {
          success: true,
          message: 'Journal entry permanently deleted successfully',
          data: deletedJournal,
          errors: null
        };
      } catch (error) {
        logger.error('Error in deleteJournal', {
          error: error.message,
          stack: error.stack,
          journalId: args.id
        });
        return {
          success: false,
          message: error.message,
          data: null,
          errors: [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 
                  error instanceof NotFoundError ? 'NOT_FOUND' : 'INTERNAL_ERROR'
          }]
        };
      }
    }
  },
  archiveJournal: {
    type: JournalResponseType,
    args: { id: { type: GraphQLID } },
    async resolve(parent, args, context, info) {
      try {
        const { id } = args;
        const selectedFields = getSelectedFields(info);
        
        const archivedJournal = await journalService.archiveJournal(id, selectedFields);
        
        return {
          success: true,
          message: 'Journal entry archived successfully',
          data: archivedJournal,
          errors: null
        };
      } catch (error) {
        logger.error('Error in archiveJournal', {
          error: error.message,
          stack: error.stack,
          journalId: args.id
        });
        return {
          success: false,
          message: error.message,
          data: null,
          errors: [{
            message: error.message,
            field: error.field,
            code: error instanceof ValidationError ? 'VALIDATION_ERROR' : 
                  error instanceof NotFoundError ? 'NOT_FOUND' : 'INTERNAL_ERROR'
          }]
        };
      }
    }
  }
};

module.exports = journalMutations; 