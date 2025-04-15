import { useQuery, useMutation } from '@apollo/client';
import { JournalSortField, JournalSortOrder } from './types';
import type { Journal, JournalInput, JournalFilter, ApiResponse, JournalListResponse } from './types';
import {
  GET_JOURNAL,
  GET_JOURNALS,
  GET_JOURNALS_BY_DATE_RANGE,
  CREATE_JOURNAL,
  UPDATE_JOURNAL,
  DELETE_JOURNAL,
  ARCHIVE_JOURNAL
} from './api';

export const useJournal = (id: string) => {
  return useQuery<{ journal: ApiResponse<Journal> }>(GET_JOURNAL, {
    variables: { id },
    skip: !id
  });
};

export const useJournals = (
  page: number = 1,
  limit: number = 10,
  sortField: JournalSortField = JournalSortField.DATE,
  sortOrder: JournalSortOrder = JournalSortOrder.DESC,
  filter?: JournalFilter
) => {
  return useQuery<{ journals: JournalListResponse }>(GET_JOURNALS, {
    variables: { 
      page,
      limit,
      sortField,
      sortOrder,
      filter
    }
  });
};

export const useJournalsByDateRange = (startDate: string, endDate: string) => {
  return useQuery<{ journalsByDateRange: JournalListResponse }>(GET_JOURNALS_BY_DATE_RANGE, {
    variables: { startDate, endDate },
    skip: !startDate || !endDate
  });
};

interface CreateJournalResponse {
  createJournal: ApiResponse<Journal>;
}

interface CreateJournalVariables {
  input: JournalInput;
}

interface UpdateJournalResponse {
  updateJournal: ApiResponse<Journal>;
}

interface UpdateJournalVariables {
  id: string;
  input: Partial<JournalInput>;
}

interface DeleteJournalResponse {
  deleteJournal: ApiResponse<Journal>;
}

interface DeleteJournalVariables {
  id: string;
}

interface ArchiveJournalResponse {
  archiveJournal: ApiResponse<Journal>;
}

interface ArchiveJournalVariables {
  id: string;
}

export const useCreateJournal = () => {
  return useMutation<CreateJournalResponse, CreateJournalVariables>(CREATE_JOURNAL);
};

export const useUpdateJournal = () => {
  return useMutation<UpdateJournalResponse, UpdateJournalVariables>(UPDATE_JOURNAL);
};

export const useDeleteJournal = () => {
  return useMutation<DeleteJournalResponse, DeleteJournalVariables>(DELETE_JOURNAL);
};

export const useArchiveJournal = () => {
  return useMutation<ArchiveJournalResponse, ArchiveJournalVariables>(ARCHIVE_JOURNAL);
};
