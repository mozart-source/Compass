import { create } from 'zustand';

type DragStore = {
  lastDroppedId: string | null;
  setLastDroppedId: (id: string | null) => void;
  chatbotAttachedTo: string | null;
  setChatbotAttachedTo: (id: string | null) => void;
  attachmentPosition: { x: number; y: number; side: 'left' | 'right' } | null;
  setAttachmentPosition: (pos: { x: number; y: number; side: 'left' | 'right' } | null) => void;
};

export const useDragStore = create<DragStore>((set) => ({
  lastDroppedId: null,
  setLastDroppedId: (id) => set({ lastDroppedId: id }),
  chatbotAttachedTo: null,
  setChatbotAttachedTo: (id) => set({ chatbotAttachedTo: id }),
  attachmentPosition: null,
  setAttachmentPosition: (pos) => set({ attachmentPosition: pos }),
}));