import { useQuery, useMutation, useSubscription } from '@apollo/client';
import { GET_NOTES, CREATE_NOTE, UPDATE_NOTE, DELETE_NOTE, NOTE_CREATED_SUBSCRIPTION, NOTE_UPDATED_SUBSCRIPTION, NOTE_DELETED_SUBSCRIPTION } from '@/components/notes/api';
import { gql } from '@apollo/client';
import { Note, UseNotesResult } from './types';

export function useNotes(): UseNotesResult {
  // Query for fetching notes
  const { data, loading, error, client } = useQuery(GET_NOTES, {
    variables: { 
      page: 1,
      limit: 50
    },
    fetchPolicy: 'cache-first'
  });

  // Mutations with optimistic updates
  const [createNoteMutation] = useMutation(CREATE_NOTE);
  const [updateNoteMutation] = useMutation(UPDATE_NOTE, {
    onError: (error) => {
      console.error('Error updating note:', error);
      // Refetch on error to ensure consistency
      client.refetchQueries({
        include: [GET_NOTES],
      });
    }
  });
  const [deleteNoteMutation] = useMutation(DELETE_NOTE);

  // Subscriptions with cache updates
  const createdSubscription = useSubscription(NOTE_CREATED_SUBSCRIPTION, {
    onData: ({ data }) => {
      if (data?.data?.notePageCreated?.success) {
        const newNote = data.data.notePageCreated.data;
        client.cache.modify({
          fields: {
            notePages(existingPages = { data: [] }) {
              return {
                ...existingPages,
                data: [...existingPages.data, newNote]
              };
            }
          }
        });
      }
    },
    onError: (error) => {
      console.error('[notePageCreated] Subscription error:', error);
    }
  });

  const updatedSubscription = useSubscription(NOTE_UPDATED_SUBSCRIPTION, {
    onData: ({ data }) => {
      if (data?.data?.notePageUpdated?.success) {
        const updatedNote = data.data.notePageUpdated.data;
        client.cache.modify({
          fields: {
            notePages(existingPages = { data: [] }) {
              return {
                ...existingPages,
                data: existingPages.data.map((note: Note) =>
                  note.id === updatedNote.id ? { ...note, ...updatedNote } : note
                )
              };
            }
          }
        });
      }
    },
    onError: (error) => {
      console.error('[notePageUpdated] Subscription error:', error);
    }
  });

  const deletedSubscription = useSubscription(NOTE_DELETED_SUBSCRIPTION, {
    onData: ({ data }) => {
      if (data?.data?.notePageDeleted?.success) {
        const deletedNoteId = data.data.notePageDeleted.data.id;
        client.cache.modify({
          fields: {
            notePages(existingPages = { data: [] }) {
              return {
                ...existingPages,
                data: existingPages.data.filter((note: Note) => note.id !== deletedNoteId)
              };
            }
          }
        });
      }
    },
    onError: (error) => {
      console.error('[notePageDeleted] Subscription error:', error);
    }
  });

  // Helper functions for mutations with optimistic updates
  const createNote = async (input: Partial<Note>): Promise<Note> => {
    const optimisticId = `temp-${Date.now()}`;
    const optimisticNote = {
      __typename: 'NotePage',
      id: optimisticId,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      ...input
    };

    const { data } = await createNoteMutation({
      variables: { input },
      optimisticResponse: {
        createNotePage: {
          success: true,
          message: 'Note created',
          data: optimisticNote
        }
      },
      update: (cache, { data: responseData }) => {
        const newNote = responseData?.createNotePage?.data;
        cache.modify({
          fields: {
            notePages(existingPages = { data: [] }) {
              return {
                ...existingPages,
                data: [...existingPages.data, newNote]
              };
            }
          }
        });
      }
    });
    return data.createNotePage.data;
  };

  const updateNote = async (id: string, input: Partial<Note>): Promise<Note> => {
    // Get the current note from cache
    const currentNote = client.cache.readFragment<Note>({
      id: `NotePage:${id}`,
      fragment: gql`
        fragment CurrentNote on NotePage {
          id
          title
          content
          tags
          favorited
          updatedAt
          __typename
        }
      `
    });

    // Create optimistic response with special handling for tags and favorited
    const optimisticNote = {
      __typename: 'NotePage',
      ...(currentNote || { tags: [], favorited: false }),
      ...input,
      id,
      // If updating tags, merge with existing tags
      tags: input.tags !== undefined 
        ? input.tags 
        : currentNote?.tags || [],
      // If updating favorited status, use the new value
      favorited: input.favorited !== undefined 
        ? input.favorited 
        : currentNote?.favorited || false,
      updatedAt: new Date().toISOString()
    };

    const { data } = await updateNoteMutation({
      variables: { id, input },
      optimisticResponse: {
        updateNotePage: {
          success: true,
          message: 'Note updated',
          data: optimisticNote
        }
      },
      update: (cache, { data: responseData }) => {
        if (!responseData?.updateNotePage?.success) return;
        
        const updatedNote = responseData.updateNotePage.data;
        
        // Update the note in the cache
        cache.modify({
          fields: {
            notePages(existingPages = { data: [] }, { readField }) {
              return {
                ...existingPages,
                data: existingPages.data.map((note: any) => 
                  readField('id', note) === id ? updatedNote : note
                )
              };
            }
          }
        });
      }
    });
    return data.updateNotePage.data;
  };

  const deleteNote = async (id: string): Promise<void> => {
    await deleteNoteMutation({
      variables: { id },
      optimisticResponse: {
        deleteNotePage: {
          success: true,
          message: 'Note deleted',
          data: { id }
        }
      },
      update: (cache) => {
        // Remove the deleted note from the cache
        cache.modify({
          fields: {
            notePages(existingPages = { data: [] }, { readField }) {
              return {
                ...existingPages,
                data: existingPages.data.filter((note: any) => id !== readField('id', note))
              };
            }
          }
        });
      },
      refetchQueries: [
        {
          query: GET_NOTES,
          variables: { page: 1 }
        }
      ]
    });
  };

  return {
    notes: data?.notePages?.data || [],
    loading,
    error,
    createNote,
    updateNote,
    deleteNote,
    pageInfo: data?.notePages?.pageInfo
  };
} 