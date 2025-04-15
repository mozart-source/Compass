const { 
  GraphQLObjectType, 
  GraphQLString, 
  GraphQLBoolean, 
  GraphQLList,
  GraphQLInt,
  GraphQLNonNull
} = require('graphql');

const ErrorType = new GraphQLObjectType({
  name: 'Error',
  fields: {
    message: { type: new GraphQLNonNull(GraphQLString) },
    field: { type: GraphQLString },
    code: { type: new GraphQLNonNull(GraphQLString) },
    errorId: { type: GraphQLString }
  }
});

const PageInfoType = new GraphQLObjectType({
  name: 'PageInfo',
  fields: {
    hasNextPage: { type: new GraphQLNonNull(GraphQLBoolean) },
    hasPreviousPage: { type: new GraphQLNonNull(GraphQLBoolean) },
    totalPages: { type: new GraphQLNonNull(GraphQLInt) },
    totalItems: { type: new GraphQLNonNull(GraphQLInt) },
    currentPage: { type: new GraphQLNonNull(GraphQLInt) },
    limit: { type: new GraphQLNonNull(GraphQLInt) }
  }
});

const createResponseType = (dataType, name) => {
  return new GraphQLObjectType({
    name: `${name}Response`,
    fields: {
      success: { 
        type: new GraphQLNonNull(GraphQLBoolean),
        description: 'Indicates if the operation was successful'
      },
      message: { 
        type: GraphQLString,
        description: 'A message describing the result of the operation'
      },
      data: { 
        type: dataType,
        description: 'The actual data returned by the operation'
      },
      errors: { 
        type: new GraphQLList(ErrorType),
        description: 'List of errors that occurred during the operation'
      },
      pageInfo: { 
        type: PageInfoType,
        description: 'Pagination information for list responses'
      }
    }
  });
};

// Helper function to create a paginated response
const createPaginatedResponse = (data, page, limit, totalItems) => {
  const totalPages = Math.ceil(totalItems / limit);
  return {
    success: true,
    data,
    pageInfo: {
      hasNextPage: page < totalPages,
      hasPreviousPage: page > 1,
      totalPages,
      totalItems,
      currentPage: page,
      limit
    }
  };
};

// Helper function to create an error response
const createErrorResponse = (message, errors) => {
  return {
    success: false,
    message,
    data: null,
    errors: Array.isArray(errors) ? errors : [{ message: errors, code: 'INTERNAL_ERROR' }]
  };
};

module.exports = { 
  createResponseType, 
  PageInfoType, 
  ErrorType,
  createPaginatedResponse,
  createErrorResponse
}; 