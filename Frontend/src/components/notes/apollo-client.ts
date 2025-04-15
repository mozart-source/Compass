import { ApolloClient, InMemoryCache, split, HttpLink, ApolloLink } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { getMainDefinition } from '@apollo/client/utilities';
import { Observable } from 'rxjs';
import { getApiUrls } from '@/config';

// Get the URLs using the configuration system
const { NOTES_API_URL, WS_NOTES_URL } = getApiUrls();

// Construct WebSocket URL properly
const wsUrl = `${WS_NOTES_URL.replace(/^http/, 'ws')}/graphql`;

console.log('Apollo Client URLs:', {
  NOTES_API_URL,
  WS_NOTES_URL,
  constructedWsUrl: wsUrl
});

const wsLink = new GraphQLWsLink(createClient({
  url: wsUrl,
  connectionParams: () => {
    const token = localStorage.getItem('token');
    return token ? {
      'Authorization': `Bearer ${token}`
    } : {};
  }
}));

// Create auth middleware
const authMiddleware = new ApolloLink((operation, forward) => {
  const token = localStorage.getItem('token');
  if (!token) {
    // If no token, reject the operation immediately
    return new Observable(observer => {
      observer.error(new Error('No authentication token available'));
    });
  }
  
  operation.setContext({
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return forward(operation);
});

const httpLink = new HttpLink({
  uri: `${NOTES_API_URL}/graphql`
});

const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === 'OperationDefinition' &&
      definition.operation === 'subscription'
    );
  },
  wsLink,
  httpLink
);

export const client = new ApolloClient({
  link: ApolloLink.from([authMiddleware, splitLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: { fetchPolicy: 'network-only' },
    query: { fetchPolicy: 'network-only' },
    mutate: { fetchPolicy: 'network-only' },
  }
});

// GraphQL Operations
export const NOTE_CREATED = `
  subscription NoteCreated($userId: ID!) {
    notePageCreated(userId: $userId) {
      success
      message
      data { id title userId createdAt }
    }
  }
`;

export const NOTE_UPDATED = `
  subscription NoteUpdated($userId: ID!) {
    notePageUpdated(userId: $userId) {
      success
      message
      data { id title userId updatedAt }
    }
  }
`;

export const NOTE_DELETED = `
  subscription NoteDeleted($userId: ID!) {
    notePageDeleted(userId: $userId) {
      success
      message
      data { id title userId updatedAt }
    }
  }
`;
