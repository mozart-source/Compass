import { gql } from '@apollo/client';

export const GET_JOURNAL = gql`
  query GetJournal($id: ID!) {
    journal(id: $id) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const GET_JOURNALS = gql`
  query GetJournals(
    $page: Int
    $limit: Int
    $sortField: JournalSortField
    $sortOrder: JournalSortOrder
    $filter: JournalFilter
  ) {
    journals(
      page: $page
      limit: $limit
      sortField: $sortField
      sortOrder: $sortOrder
      filter: $filter
    ) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      pageInfo {
        totalItems
        currentPage
        totalPages
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const GET_JOURNALS_BY_DATE_RANGE = gql`
  query GetJournalsByDateRange($startDate: String!, $endDate: String!) {
    journalsByDateRange(startDate: $startDate, endDate: $endDate) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      pageInfo {
        totalItems
        currentPage
        totalPages
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const CREATE_JOURNAL = gql`
  mutation CreateJournal($input: JournalInput!) {
    createJournal(input: $input) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const UPDATE_JOURNAL = gql`
  mutation UpdateJournal($id: ID!, $input: JournalInput!) {
    updateJournal(id: $id, input: $input) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const DELETE_JOURNAL = gql`
  mutation DeleteJournal($id: ID!) {
    deleteJournal(id: $id) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      errors {
        message
        field
        code
      }
    }
  }
`;

export const ARCHIVE_JOURNAL = gql`
  mutation ArchiveJournal($id: ID!) {
    archiveJournal(id: $id) {
      success
      message
      data {
        id
        title
        content
        date
        tags
        mood
        aiPromptUsed
        aiGenerated
        archived
        wordCount
        createdAt
        updatedAt
      }
      errors {
        message
        field
        code
      }
    }
  }
`;
