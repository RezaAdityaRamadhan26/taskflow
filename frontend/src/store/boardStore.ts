import { create } from 'zustand';
import type { Board, List, Card, Workspace } from '../types';

interface BoardState {
  board: Board | null;
  workspace: Workspace | null;
  lists: List[];
  cards: Record<string, Card[]>; // mapping listId -> cards array
  isLoading: boolean;
  setBoardData: (board: Board, workspace: Workspace, lists: List[], cards: Card[]) => void;
  setLoading: (loading: boolean) => void;
  
  // Optimistic UI updates
  addList: (list: List) => void;
  updateList: (listId: string, name: string) => void;
  addCard: (card: Card) => void;
  moveCard: (cardId: string, sourceListId: string, destListId: string, newIndex: number) => void;
}

export const useBoardStore = create<BoardState>((set) => ({
  board: null,
  workspace: null,
  lists: [],
  cards: {},
  isLoading: true,

  setBoardData: (board, workspace, lists, allCards) => {
    // Group cards by list_id
    const cardsByList: Record<string, Card[]> = {};
    lists.forEach(list => {
      cardsByList[list.id] = [];
    });
    allCards.forEach(card => {
      if (!cardsByList[card.list_id]) cardsByList[card.list_id] = [];
      cardsByList[card.list_id].push(card);
    });

    // Ensure cards are sorted by position
    Object.keys(cardsByList).forEach(listId => {
      cardsByList[listId].sort((a, b) => a.position - b.position);
    });

    set({ board, workspace, lists, cards: cardsByList, isLoading: false });
  },

  setLoading: (isLoading) => set({ isLoading }),

  addList: (list) => set((state) => ({
    lists: [...state.lists, list],
    cards: { ...state.cards, [list.id]: [] }
  })),

  updateList: (listId, name) => set((state) => ({
    lists: state.lists.map(l => l.id === listId ? { ...l, name } : l)
  })),

  addCard: (card) => set((state) => {
    const listCards = [...(state.cards[card.list_id] || []), card];
    return {
      cards: { ...state.cards, [card.list_id]: listCards }
    };
  }),

  moveCard: (cardId, sourceListId, destListId, newIndex) => set((state) => {
    const sourceCards = [...(state.cards[sourceListId] || [])];
    const destCards = sourceListId === destListId ? sourceCards : [...(state.cards[destListId] || [])];
    
    // Find card
    const cardIndex = sourceCards.findIndex(c => c.id === cardId);
    if (cardIndex === -1) return state; // Safety check
    
    const [movedCard] = sourceCards.splice(cardIndex, 1);
    
    // Update its list_id if changed
    movedCard.list_id = destListId;
    
    // Insert at new index
    destCards.splice(newIndex, 0, movedCard);
    
    return {
      cards: {
        ...state.cards,
        [sourceListId]: sourceCards,
        [destListId]: destCards
      }
    };
  }),
}));
