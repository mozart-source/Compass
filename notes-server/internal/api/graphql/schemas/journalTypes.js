const { 
  GraphQLObjectType, 
  GraphQLString, 
  GraphQLID, 
  GraphQLList, 
  GraphQLBoolean,
  GraphQLEnumType,
  GraphQLInputObjectType,
  GraphQLInt
} = require('graphql');
const Journal = require('../../../domain/journals/model');
const { createResponseType } = require('./responseTypes');

// Enum for mood values
const MoodEnum = new GraphQLEnumType({
  name: 'Mood',
  values: {
    HAPPY: { value: 'happy' },
    SAD: { value: 'sad' },
    ANGRY: { value: 'angry' },
    NEUTRAL: { value: 'neutral' },
    EXCITED: { value: 'excited' },
    ANXIOUS: { value: 'anxious' },
    TIRED: { value: 'tired' },
    GRATEFUL: { value: 'grateful' }
  }
});

// Input type for journal mutations
const JournalInput = new GraphQLInputObjectType({
  name: 'JournalInput',
  fields: {
    userId: { 
      type: GraphQLID,
      description: '[IGNORED] ID of the user who owns the journal. Always set from backend.'
    },
    title: { 
      type: GraphQLString,
      description: 'Title of the journal (max 200 characters)'
    },
    date: {
      type: GraphQLString,
      description: 'Date of the journal'
    },
    content: { 
      type: GraphQLString,
      description: 'Content of the journal (max 10000 characters)'
    },
    mood: {
      type: MoodEnum,
      description: 'Mood associated with the journal'
    },
    tags: { 
      type: new GraphQLList(GraphQLString),
      description: 'List of tags (max 10 tags, each max 50 characters)'
    },
    aiPromptUsed: {
      type: GraphQLString,
      description: 'AI prompt used to generate the journal'
    },
    aiGenerated: {
      type: GraphQLBoolean,
      description: 'Whether the journal is AI generated'
    },
    archived: {
      type: GraphQLBoolean,
      description: 'Whether the journal is archived'
    },
    wordCount: {
      type: GraphQLInt,
      description: 'Word count of the journal'
    },
    isDeleted: {
      type: GraphQLBoolean,
      description: 'Whether the journal is deleted (soft delete)'
    }
  }
});

const PaginationInput = new GraphQLInputObjectType({
  name: 'PaginationInput',
  fields: {
    page: { type: GraphQLInt, defaultValue: 1 },
    limit: { type: GraphQLInt, defaultValue: 10 }
  }
});

const JournalSortFieldEnum = new GraphQLEnumType({
  name: 'JournalSortField',
  description: 'Fields by which journals can be sorted',
  values: {
    DATE: { value: 'date' },
    CREATED_AT: { value: 'createdAt' },
    UPDATED_AT: { value: 'updatedAt' },
    TITLE: { value: 'title' },
    WORD_COUNT: { value: 'wordCount' },
    MOOD: { value: 'mood' }
  }
});

const SortOrderEnum = new GraphQLEnumType({
  name: 'JournalSortOrder',
  values: {
    ASC: { value: 1 },
    DESC: { value: -1 }
  }
});

const JournalFilterInput = new GraphQLInputObjectType({
  name: 'JournalFilter',
  description: 'Filters for journals',
  fields: {
    wordCountMin: { type: GraphQLInt },
    wordCountMax: { type: GraphQLInt },
    aiGenerated: { type: GraphQLBoolean },
    tags: { type: new GraphQLList(GraphQLString) },
    mood: { type: MoodEnum },
    archived: { type: GraphQLBoolean },
    dateFrom: { 
      type: GraphQLString,
      description: 'Filter journals from this date (inclusive)'
    },
    dateTo: { 
      type: GraphQLString,
      description: 'Filter journals to this date (inclusive)'
    }
  }
});

// Helper function to get selected fields from GraphQL query
const getSelectedFields = (info) => {
  try {
    const selections = info.fieldNodes[0].selectionSet.selections;
    const dataSelection = selections.find(selection => selection.name.value === 'data');
    if (dataSelection && dataSelection.selectionSet) {
      return dataSelection.selectionSet.selections.map(sel => sel.name.value).join(' ');
    }
    // If 'data' is not present, fallback to default fields
    return 'title content tags mood date wordCount createdAt updatedAt userId';
  } catch (e) {
    // Fallback in case of any error
    return 'title content tags mood date wordCount createdAt updatedAt userId';
  }
};

const JournalType = new GraphQLObjectType({
  name: 'Journal',
  fields: () => ({
    id: {
      type: GraphQLID,
      description: 'String ID for GraphQL compatibility',
      resolve: (parent) => parent._id ? parent._id.toString() : null
    },
    _id: { type: GraphQLID },
    userId: { type: GraphQLID },
    title: { type: GraphQLString },
    date: { type: GraphQLString },
    content: { type: GraphQLString },
    mood: { type: MoodEnum },
    tags: { type: new GraphQLList(GraphQLString) },
    aiPromptUsed: { type: GraphQLString },
    aiGenerated: { type: GraphQLBoolean },
    archived: { type: GraphQLBoolean },
    wordCount: { type: GraphQLInt },
    createdAt: { type: GraphQLString },
    updatedAt: { type: GraphQLString },
    isDeleted: { type: GraphQLBoolean }
  })
});

// Create response types for single and list responses
const JournalResponseType = createResponseType(JournalType, 'Journal');
const JournalListResponseType = createResponseType(new GraphQLList(JournalType), 'JournalList');

// --- Subscription Fields for Journals ---
const journalSubscriptionFields = {
  journalCreated: {
    type: JournalResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a journal is created.'
  },
  journalUpdated: {
    type: JournalResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a journal is updated.'
  },
  journalDeleted: {
    type: JournalResponseType,
    args: { userId: { type: GraphQLID } },
    description: 'Triggered when a journal is deleted.'
  }
};

module.exports = { 
  JournalType,
  JournalResponseType,
  JournalListResponseType,
  MoodEnum,
  JournalSortFieldEnum,
  SortOrderEnum,
  JournalFilterInput,
  JournalInput,
  PaginationInput,
  getSelectedFields,
  journalSubscriptionFields
}; 