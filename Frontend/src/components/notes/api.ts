import { gql } from '@apollo/client';

export const NOTE_CREATED_SUBSCRIPTION = gql`
  subscription {
    notePageCreated {
      success
      message
      data { id title createdAt }
    }
  }
`;

export const NOTE_UPDATED_SUBSCRIPTION = gql`
  subscription {
    notePageUpdated {
      success
      message
      data { id title updatedAt }
    }
  }
`;

export const NOTE_DELETED_SUBSCRIPTION = gql`
  subscription {
    notePageDeleted {
      success
      message
      data { id title updatedAt }
    }
  }
`;

export const GET_NOTES = gql`
  query GetNotes($page: Int!) {
    notePages(page: $page) {
      success
      message
      data {
        id
        title
        content
        tags
        favorited
        createdAt
        updatedAt
      }
      pageInfo {
        totalPages
        totalItems
        currentPage
      }
    }
  }
`;

export const CREATE_NOTE = gql`
  mutation CreateNote($input: NotePageInput!) {
    createNotePage(input: $input) {
      success
      message
      data {
        id
        title
        content
        tags
        favorited
        createdAt
        updatedAt
      }
    }
  }
`;

export const UPDATE_NOTE = gql`
  mutation UpdateNote($id: ID!, $input: NotePageInput!) {
    updateNotePage(id: $id, input: $input) {
      success
      message
      data {
        id
        title
        content
        tags
        favorited
        updatedAt
      }
    }
  }
`;

export const DELETE_NOTE = gql`
  mutation DeleteNote($id: ID!) {
    deleteNotePage(id: $id) {
      success
      message
      data {
        id
      }
    }
  }
`; 