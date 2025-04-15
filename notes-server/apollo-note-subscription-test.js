const { ApolloClient, InMemoryCache, split, HttpLink } = require('@apollo/client/core');
const { GraphQLWsLink } = require('@apollo/client/link/subscriptions');
const { createClient } = require('graphql-ws');
const { getMainDefinition } = require('@apollo/client/utilities');
const WebSocket = require('ws');
const fetch = require('cross-fetch');
const gql = require('graphql-tag');

// Set your test userId here or use process.env.TEST_USER_ID
const jwtToken = process.env.TEST_JWT_TOKEN || '<YOUR_JWT_TOKEN_HERE>';

const wsLink = new GraphQLWsLink(createClient({
  url: 'ws://localhost:5000/notes/graphql',
  webSocketImpl: WebSocket,
  connectionParams: {
    'Authorization': `Bearer ${jwtToken}`
  }
}));

const httpLink = new HttpLink({
  uri: 'http://localhost:5000/notes/graphql',
  fetch,
  headers: {
    'Authorization': `Bearer ${jwtToken}`
  }
});

const link = split(
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

const client = new ApolloClient({
  link,
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: { fetchPolicy: 'no-cache' },
    query: { fetchPolicy: 'no-cache' },
    mutate: { fetchPolicy: 'no-cache' },
  }
});

const NOTE_CREATED = gql`
  subscription($userId: ID!) {
    notePageCreated(userId: $userId) {
      success
      message
      data { id title userId createdAt }
    }
  }
`;
const NOTE_UPDATED = gql`
  subscription($userId: ID!) {
    notePageUpdated(userId: $userId) {
      success
      message
      data { id title userId updatedAt }
    }
  }
`;
const NOTE_DELETED = gql`
  subscription($userId: ID!) {
    notePageDeleted(userId: $userId) {
      success
      message
      data { id title userId updatedAt }
    }
  }
`;

function subscribeAndLog(name, query) {
  const observable = client.subscribe({ query, variables: { userId } });
  const sub = observable.subscribe({
    next: (data) => {
      console.log(`\n[${name}] Subscription event received:`);
      console.dir(data, { depth: null });
    },
    error: (err) => {
      console.error(`[${name}] Subscription error:`, err);
    },
    complete: () => {
      console.log(`[${name}] Subscription complete`);
    }
  });
  return sub;
}

console.log('Subscribing to note events for userId:', jwtToken);
subscribeAndLog('notePageCreated', NOTE_CREATED);
subscribeAndLog('notePageUpdated', NOTE_UPDATED);
subscribeAndLog('notePageDeleted', NOTE_DELETED);

// Keep the process alive
setInterval(() => {}, 1000); 