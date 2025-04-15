const {
  GraphQLObjectType,
  GraphQLInt,
  GraphQLString,
  GraphQLList,
  GraphQLObjectTypeConfig
} = require('graphql');
const { GraphQLJSON } = require('graphql-type-json');
const { NotePageType } = require('./noteTypes');
const { JournalType } = require('./journalTypes');

const DashboardMetricsType = new GraphQLObjectType({
  name: 'DashboardMetrics',
  fields: {
    moodSummary: { type: GraphQLString },
    notesCount: { type: GraphQLInt },
    journalsCount: { type: GraphQLInt },
    recentNotes: { type: new GraphQLList(NotePageType) },
    recentJournals: { type: new GraphQLList(JournalType) },
    tagCounts: { type: GraphQLJSON },
    moodDistribution: { type: GraphQLJSON },
    timestamp: { type: GraphQLString }
  }
});

module.exports = { DashboardMetricsType };