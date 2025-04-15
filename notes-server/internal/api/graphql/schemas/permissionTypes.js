const { GraphQLEnumType, GraphQLInputObjectType, GraphQLObjectType, GraphQLID } = require('graphql');

const PermissionLevelEnum = new GraphQLEnumType({
  name: 'PermissionLevel',
  values: {
    VIEW: { value: 'view' },
    EDIT: { value: 'edit' },
    COMMENT: { value: 'comment' }
  }
});

const PermissionInput = new GraphQLInputObjectType({
  name: 'PermissionInput',
  fields: {
    userId: { type: GraphQLID },
    level: { type: PermissionLevelEnum }
  }
});

const PermissionType = new GraphQLObjectType({
  name: 'Permission',
  fields: {
    userId: { type: GraphQLID },
    level: { type: PermissionLevelEnum }
  }
});

module.exports = { PermissionLevelEnum, PermissionInput, PermissionType }; 