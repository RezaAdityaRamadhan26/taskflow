import React, { useState, useEffect } from 'react';
import { useParams, useLocation } from 'react-router-dom';
import { 
  DndContext, 
  DragOverlay, 
  closestCorners, 
  KeyboardSensor, 
  PointerSensor, 
  useSensor, 
  useSensors,
} from '@dnd-kit/core';
import type {
  DragStartEvent,
  DragOverEvent,
  DragEndEvent
} from '@dnd-kit/core';
import { 
  SortableContext, 
  horizontalListSortingStrategy, 
  verticalListSortingStrategy 
} from '@dnd-kit/sortable';
import { api } from '../../services/api';
import { useBoardStore } from '../../store/boardStore';
import type { Board, List, Card as CardType, ApiResponse } from '../../types';
import { Plus, MoreHorizontal, Loader2, Calendar } from 'lucide-react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { cn } from '../../utils/cn';

// ==========================================
// Sortable Card Component
// ==========================================
const SortableCard = ({ card }: { card: CardType }) => {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: card.id,
    data: { type: 'Card', card }
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const priorityColors = {
    'LOW': 'bg-blue-100 text-blue-800',
    'MEDIUM': 'bg-green-100 text-green-800',
    'HIGH': 'bg-orange-100 text-orange-800',
    'URGENT': 'bg-red-100 text-red-800'
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        "bg-white p-3 rounded shadow-sm border border-surface-200 cursor-grab hover:border-primary-300 relative group",
        isDragging && "opacity-30 border-dashed border-2 border-primary-500"
      )}
    >
      <div className="flex justify-between items-start mb-2">
        <span className={cn("text-[10px] font-bold px-2 py-0.5 rounded", priorityColors[card.priority])}>
          {card.priority}
        </span>
      </div>
      <p className="text-sm font-medium text-surface-900 mb-2">{card.title}</p>
      
      {/* Badges row */}
      <div className="flex items-center space-x-3 text-surface-400 text-xs mt-3">
        {card.description && (
          <div className="flex items-center">
            <MoreHorizontal className="w-3.5 h-3.5" />
          </div>
        )}
        {card.due_date && (
          <div className="flex items-center">
            <Calendar className="w-3.5 h-3.5 mr-1" />
            <span>{new Date(card.due_date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}</span>
          </div>
        )}
      </div>
    </div>
  );
};

// ==========================================
// Sortable List/Column Component
// ==========================================
const SortableList = ({ list, cards, onAddCard }: { list: List, cards: CardType[], onAddCard: (listId: string, title: string) => void }) => {
  const [isAdding, setIsAdding] = useState(false);
  const [newCardTitle, setNewCardTitle] = useState('');

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: list.id,
    data: { type: 'List', list }
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const handleAddCard = (e: React.FormEvent) => {
    e.preventDefault();
    if (newCardTitle.trim()) {
      onAddCard(list.id, newCardTitle.trim());
      setNewCardTitle('');
      setIsAdding(false);
    }
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        "bg-surface-200 w-72 rounded-lg flex flex-col shrink-0 max-h-full",
        isDragging && "opacity-50"
      )}
    >
      <div 
        {...attributes} 
        {...listeners}
        className="p-3 flex justify-between items-center cursor-grab active:cursor-grabbing hover:bg-surface-300 rounded-t-lg transition-colors"
      >
        <h3 className="font-semibold text-surface-900 text-sm">{list.name}</h3>
        <button className="text-surface-500 hover:bg-surface-400 hover:text-surface-900 p-1 rounded transition-colors">
          <MoreHorizontal className="w-4 h-4" />
        </button>
      </div>

      <div className="p-2 flex-1 overflow-y-auto scrollbar-thin flex flex-col gap-2">
        <SortableContext items={cards.map(c => c.id)} strategy={verticalListSortingStrategy}>
          {cards.map(card => (
            <SortableCard key={card.id} card={card} />
          ))}
        </SortableContext>

        {isAdding ? (
          <form onSubmit={handleAddCard} className="mt-1">
            <textarea
              className="w-full text-sm p-2 border border-primary-500 rounded shadow-sm focus:outline-none resize-none"
              placeholder="Enter a title for this card..."
              autoFocus
              value={newCardTitle}
              onChange={e => setNewCardTitle(e.target.value)}
              onKeyDown={e => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleAddCard(e);
                }
                if (e.key === 'Escape') setIsAdding(false);
              }}
              rows={2}
            />
            <div className="flex items-center space-x-2 mt-2">
              <button type="submit" className="bg-primary-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-primary-700 transition-colors">
                Add card
              </button>
              <button type="button" onClick={() => setIsAdding(false)} className="text-surface-500 hover:text-surface-800 text-sm p-1">
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <button 
            onClick={() => setIsAdding(true)}
            className="flex items-center text-surface-500 hover:bg-surface-300 hover:text-surface-800 p-2 rounded w-full transition-colors text-sm font-medium mt-1"
          >
            <Plus className="w-4 h-4 mr-2" />
            Add a card
          </button>
        )}
      </div>
    </div>
  );
};

// ==========================================
// Main Board Page
// ==========================================
const BoardPage = () => {
  const { boardId } = useParams<{ boardId: string }>();
  const location = useLocation();
  const { board, lists, cards, isLoading, setBoardData, addList, addCard, moveCard } = useBoardStore();
  
  const [activeCard, setActiveCard] = useState<CardType | null>(null);
  const [_activeList, setActiveList] = useState<List | null>(null);
  const [isAddingList, setIsAddingList] = useState(false);
  const [newListTitle, setNewListTitle] = useState('');

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor)
  );

  // Initial Data Fetch
  useEffect(() => {
    const fetchData = async () => {
      try {
        const boardRes = await api.get<ApiResponse<Board>>(`/boards/${boardId}`);
        const boardData = boardRes.data.data;
        if (!boardData) return;

        const listsRes = await api.get<ApiResponse<List[]>>(`/boards/${boardId}/lists`);
        const listsData = listsRes.data.data || [];

        // In a real app we'd want a single endpoint for this or Promise.all
        // For MVP, fetch cards for all lists one by one
        let allCards: CardType[] = [];
        for (const list of listsData) {
          const cardsRes = await api.get<ApiResponse<CardType[]>>(`/lists/${list.id}/cards`);
          if (cardsRes.data.success && cardsRes.data.data) {
            allCards = [...allCards, ...cardsRes.data.data];
          }
        }

        setBoardData(boardData, location.state?.workspace, listsData, allCards);
      } catch (error) {
        console.error('Failed to fetch board data:', error);
      }
    };
    fetchData();
  }, [boardId]);

  // Handle Add List
  const handleAddList = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newListTitle.trim() || !boardId) return;

    try {
      const res = await api.post<ApiResponse<List>>('/lists', {
        board_id: boardId,
        name: newListTitle.trim()
      });
      if (res.data.success && res.data.data) {
        addList(res.data.data);
        setNewListTitle('');
        setIsAddingList(false);
      }
    } catch (error) {
      console.error('Failed to create list', error);
    }
  };

  // Handle Add Card
  const handleAddCard = async (listId: string, title: string) => {
    try {
      const res = await api.post<ApiResponse<CardType>>('/cards', {
        list_id: listId,
        title: title
      });
      if (res.data.success && res.data.data) {
        addCard(res.data.data);
      }
    } catch (error) {
      console.error('Failed to create card', error);
    }
  };

  // ==========================================
  // Drag and Drop Logic
  // ==========================================
  const handleDragStart = (event: DragStartEvent) => {
    const { current } = event.active.data;
    if (current?.type === 'Card') setActiveCard(current.card);
    if (current?.type === 'List') setActiveList(current.list);
  };

  const handleDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over) return;

    const activeData = active.data.current;
    const overData = over.data.current;

    // We only care about Card over Card or Card over List
    if (activeData?.type !== 'Card') return;

    const activeCard = activeData.card as CardType;
    const activeListId = activeCard.list_id;

    if (overData?.type === 'List') {
      const overListId = over.id as string;
      if (activeListId !== overListId) {
        // Optimistic UI update: move card to the bottom of the empty list
        moveCard(active.id as string, activeListId, overListId, cards[overListId]?.length || 0);
      }
      return;
    }

    if (overData?.type === 'Card') {
      const overCard = overData.card as CardType;
      const overListId = overCard.list_id;

      if (activeListId !== overListId) {
        // Moving to a different list
        const overIndex = cards[overListId].findIndex(c => c.id === over.id);
        moveCard(active.id as string, activeListId, overListId, overIndex);
      }
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    setActiveCard(null);
    setActiveList(null);

    const { active, over } = event;
    if (!over) return;

    if (active.data.current?.type === 'Card') {
      // Find the card's final list and index
      let targetListId = "";
      let targetIndex = 0;

      if (over.data.current?.type === 'List') {
        targetListId = over.id as string;
        targetIndex = cards[targetListId].length;
      } else if (over.data.current?.type === 'Card') {
        const overCard = over.data.current.card as CardType;
        targetListId = overCard.list_id;
        targetIndex = cards[targetListId].findIndex(c => c.id === over.id);
      }

      if (targetListId) {
        const targetListCards = cards[targetListId] || [];
        const cardToMove = targetListCards.find(c => c.id === active.id);
        
        if (cardToMove) {
          // Reorder locally first if within same list
          const oldListId = (active.data.current.card as CardType).list_id;
          if (oldListId === targetListId) {
             const oldIndex = targetListCards.findIndex(c => c.id === active.id);
             if (oldIndex !== targetIndex) {
                 moveCard(active.id as string, oldListId, targetListId, targetIndex);
             }
          }

          // Calculate new FLOAT position based on neighbors
          // Refreshed target array after potential moveCard
          const currentListCards = useBoardStore.getState().cards[targetListId];
          const newIdx = currentListCards.findIndex(c => c.id === active.id);
          
          let newPosition = 65536.0;
          if (currentListCards.length > 1) {
            if (newIdx === 0) {
              // At the top: half of the first item's pos
              newPosition = currentListCards[1].position / 2;
            } else if (newIdx === currentListCards.length - 1) {
              // At the bottom: last item's pos + 65536
              newPosition = currentListCards[newIdx - 1].position + 65536.0;
            } else {
              // In the middle: average of above and below
              const prev = currentListCards[newIdx - 1].position;
              const next = currentListCards[newIdx + 1].position;
              newPosition = (prev + next) / 2;
            }
          }

          // API Call
          try {
            await api.put(`/cards/${active.id}`, {
              list_id: targetListId,
              title: cardToMove.title,
              position: newPosition,
              priority: cardToMove.priority,
              description: cardToMove.description,
              due_date: cardToMove.due_date
            });
            // Re-fetch or rely on optimistic? Optimistic is mostly fine.
            // A perfect solution would update the state with the exact backend position.
          } catch (error) {
            console.error('Failed to move card via API', error);
            // Need rollback logic here in real app
          }
        }
      }
    }
  };

  if (isLoading) {
    return <div className="flex-1 flex items-center justify-center bg-surface-100"><Loader2 className="w-8 h-8 animate-spin text-primary-500" /></div>;
  }

  if (!board) return <div>Board not found</div>;

  return (
    <div className="flex-1 flex flex-col h-full" style={{ backgroundColor: board.color || '#3b82f6' }}>
      {/* Board Header */}
      <div className="bg-black/20 backdrop-blur-sm p-3 px-6 flex items-center shrink-0">
        <h1 className="text-xl font-bold text-white tracking-wide">{board.name}</h1>
      </div>

      {/* Board Canvas */}
      <div className="flex-1 overflow-x-auto overflow-y-hidden p-6 relative">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCorners}
          onDragStart={handleDragStart}
          onDragOver={handleDragOver}
          onDragEnd={handleDragEnd}
        >
          <div className="flex items-start gap-4 h-full">
            <SortableContext items={lists.map(l => l.id)} strategy={horizontalListSortingStrategy}>
              {lists.map(list => (
                <SortableList 
                  key={list.id} 
                  list={list} 
                  cards={cards[list.id] || []} 
                  onAddCard={handleAddCard} 
                />
              ))}
            </SortableContext>

            {/* Add new list button */}
            <div className="w-72 shrink-0">
              {isAddingList ? (
                <form onSubmit={handleAddList} className="bg-surface-200 p-2 rounded-lg shadow-sm border border-surface-300">
                  <input
                    type="text"
                    className="w-full px-3 py-2 text-sm border border-primary-500 rounded focus:outline-none mb-2"
                    placeholder="Enter list title..."
                    autoFocus
                    value={newListTitle}
                    onChange={(e) => setNewListTitle(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Escape') setIsAddingList(false);
                    }}
                  />
                  <div className="flex items-center space-x-2">
                    <button type="submit" className="bg-primary-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-primary-700">
                      Add list
                    </button>
                    <button type="button" onClick={() => setIsAddingList(false)} className="text-surface-600 hover:text-surface-900 p-1 text-sm">
                      Cancel
                    </button>
                  </div>
                </form>
              ) : (
                <button 
                  onClick={() => setIsAddingList(true)}
                  className="bg-white/20 hover:bg-white/30 backdrop-blur-sm text-white flex items-center w-full px-4 py-3 rounded-lg font-medium transition-colors border border-white/10"
                >
                  <Plus className="w-5 h-5 mr-2" />
                  Add another list
                </button>
              )}
            </div>
          </div>

          <DragOverlay>
            {activeCard ? <SortableCard card={activeCard} /> : null}
            {/* Can add List overlay styling here later */}
          </DragOverlay>
        </DndContext>
      </div>
    </div>
  );
};

export default BoardPage;
