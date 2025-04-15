const mongoose = require('mongoose');
const { DatabaseError } = require('../../../pkg/utils/errorHandler');
const { logger } = require('../../../pkg/utils/logger');

// Helper function to update bi-directional links
const updateBidirectionalLinks = async (noteId, oldLinksOut = [], newLinksOut = []) => {
  try {
    const bulkOps = [];
    
    // Remove old links
    const removedLinks = oldLinksOut.filter(link => !newLinksOut.includes(link.toString()));
    if (removedLinks.length > 0) {
      bulkOps.push({
        updateMany: {
          filter: { _id: { $in: removedLinks } },
          update: { $pull: { linksIn: noteId } }
        }
      });
    }
    
    // Add new links
    const newLinks = newLinksOut.filter(link => !oldLinksOut.map(l => l.toString()).includes(link.toString()));
    if (newLinks.length > 0) {
      bulkOps.push({
        updateMany: {
          filter: { _id: { $in: newLinks } },
          update: { $addToSet: { linksIn: noteId } }
        }
      });
    }
    
    if (bulkOps.length > 0) {
      const result = await mongoose.model('NotePage').bulkWrite(bulkOps, { ordered: false });
      
      if (result.modifiedCount !== (removedLinks.length + newLinks.length)) {
        throw new DatabaseError('Some link updates failed');
      }
    }

    return true;
  } catch (error) {
    if (error instanceof DatabaseError) {
      throw error;
    }
    throw new DatabaseError(`Failed to update links: ${error.message}`);
  }
};

// Helper function to handle cascading deletes
const handleCascadingDelete = async (noteId) => {
  try {
    const note = await mongoose.model('NotePage').findById(noteId);
    if (!note) return;

    const bulkOps = [];

    // Remove this note from linksIn of all linked notes
    if (note.linksOut.length > 0) {
      bulkOps.push({
        updateMany: {
          filter: { _id: { $in: note.linksOut } },
          update: { $pull: { linksIn: noteId } }
        }
      });
    }

    // Remove this note from linksOut of all notes that link to it
    if (note.linksIn.length > 0) {
      bulkOps.push({
        updateMany: {
          filter: { _id: { $in: note.linksIn } },
          update: { $pull: { linksOut: noteId } }
        }
      });
    }

    if (bulkOps.length > 0) {
      const result = await mongoose.model('NotePage').bulkWrite(bulkOps, { ordered: false });
      
      if (result.modifiedCount !== (note.linksOut.length + note.linksIn.length)) {
        throw new DatabaseError('Some cascading delete operations failed');
      }
    }

    return true;
  } catch (error) {
    if (error instanceof DatabaseError) {
      throw error;
    }
    throw new DatabaseError(`Failed to handle cascading delete: ${error.message}`);
  }
};

// Helper function to validate links
async function validateLinks(linksOut) {
  try {
    if (!Array.isArray(linksOut)) {
      logger.error('Invalid linksOut type', { type: typeof linksOut });
      throw new DatabaseError('linksOut must be an array');
    }

    if (linksOut.length === 0) {
      logger.debug('Empty linksOut array, validation passed');
      return true;
    }

    // Convert all IDs to strings for comparison
    const linkIds = linksOut.map(id => id.toString());
    logger.debug('Validating links', { linkIds });

    // Check for duplicate links
    const uniqueLinks = new Set(linkIds);
    if (uniqueLinks.size !== linkIds.length) {
      logger.warn('Duplicate links detected', { linkIds });
      throw new DatabaseError('Duplicate links are not allowed');
    }

    // Find all notes that exist and are not deleted
    const existingNotes = await mongoose.model('NotePage').find({
      _id: { $in: linksOut },
      isDeleted: false
    }).select('_id');

    const existingIds = existingNotes.map(note => note._id.toString());
    const invalidLinks = linkIds.filter(id => !existingIds.includes(id));

    if (invalidLinks.length > 0) {
      logger.warn('Invalid or deleted notes found', { invalidLinks });
      throw new DatabaseError(`Invalid or deleted notes found: ${invalidLinks.join(', ')}`);
    }

    logger.debug('Link validation successful', { linkIds });
    return true;
  } catch (error) {
    if (error instanceof DatabaseError) {
      throw error;
    }
    logger.error('Link validation failed', { error: error.message });
    throw new DatabaseError(`Failed to validate links: ${error.message}`);
  }
}

// Helper function to get linked notes
async function getLinkedNotes(noteId) {
  try {
    logger.debug('Getting linked notes', { noteId });
    const note = await mongoose.model('NotePage').findById(noteId)
      .populate('linksOut', 'title content tags favorited icon')
      .populate('linksIn', 'title content tags favorited icon');

    if (!note) {
      logger.warn('Note not found', { noteId });
      throw new DatabaseError('Note not found');
    }

    logger.debug('Linked notes retrieved', { 
      noteId,
      outgoingLinks: note.linksOut.length,
      incomingLinks: note.linksIn.length
    });

    return {
      linksOut: note.linksOut,
      linksIn: note.linksIn
    };
  } catch (error) {
    if (error instanceof DatabaseError) {
      throw error;
    }
    logger.error('Failed to get linked notes', { error: error.message, noteId });
    throw new DatabaseError(`Failed to get linked notes: ${error.message}`);
  }
}

// Helper function to get note link statistics
async function getNoteLinkStats(noteId) {
  try {
    logger.debug('Getting note link stats', { noteId });
    const note = await mongoose.model('NotePage').findById(noteId);
    if (!note) {
      logger.warn('Note not found', { noteId });
      throw new DatabaseError('Note not found');
    }

    const stats = {
      totalLinks: note.linksOut.length + note.linksIn.length,
      outgoingLinks: note.linksOut.length,
      incomingLinks: note.linksIn.length
    };

    logger.debug('Note link stats retrieved', { noteId, stats });
    return stats;
  } catch (error) {
    if (error instanceof DatabaseError) {
      throw error;
    }
    logger.error('Failed to get note link stats', { error: error.message, noteId });
    throw new DatabaseError(`Failed to get note link stats: ${error.message}`);
  }
}

module.exports = { 
  updateBidirectionalLinks,
  handleCascadingDelete,
  validateLinks,
  getLinkedNotes,
  getNoteLinkStats
};