import { useState, useEffect } from 'react';
import { useParams, Link, useLocation } from 'react-router-dom';
import { api } from '../../services/api';
import type { Workspace, Board, ApiResponse } from '../../types';
import { Layout, Plus, Users, Settings, Search, Loader2 } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

const createBoardSchema = z.object({
  name: z.string().min(2, 'Name is required').max(100),
  color: z.string().optional(),
});

type CreateBoardValues = z.infer<typeof createBoardSchema>;

const bgColors = [
  '#3b82f6', '#10b981', '#f59e0b', '#ef4444', 
  '#8b5cf6', '#ec4899', '#f43f5e', '#6366f1'
];

const WorkspacePage = () => {
  const { slug } = useParams<{ slug: string }>();
  const location = useLocation();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [boards, setBoards] = useState<Board[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const { register, handleSubmit, formState: { errors, isSubmitting }, reset, setValue, watch } = useForm<CreateBoardValues>({
    resolver: zodResolver(createBoardSchema),
    defaultValues: { color: bgColors[0] }
  });

  const selectedColor = watch('color');

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Try to get workspace from context/state or fetch if not available
        let wsId = location.state?.workspaceId;
        
        if (!wsId) {
          // If we only have slug (direct link), we need to fetch all and find it
          // In a real app with proper API, we'd have a getBySlug endpoint
          const wsRes = await api.get<ApiResponse<Workspace[]>>('/workspaces');
          if (wsRes.data.success && wsRes.data.data) {
            const ws = wsRes.data.data.find(w => w.slug === slug);
            if (ws) {
              wsId = ws.id;
              setWorkspace(ws);
            }
          }
        } else {
          const wsRes = await api.get<ApiResponse<Workspace>>(`/workspaces/${wsId}`);
          if (wsRes.data.success && wsRes.data.data) {
            setWorkspace(wsRes.data.data);
          }
        }

        if (wsId) {
          const boardsRes = await api.get<ApiResponse<Board[]>>(`/workspaces/${wsId}/boards`);
          if (boardsRes.data.success && boardsRes.data.data) {
            setBoards(boardsRes.data.data);
          }
        }
      } catch (error) {
        console.error('Failed to fetch workspace data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [slug, location.state]);

  const onCreateBoard = async (data: CreateBoardValues) => {
    if (!workspace) return;
    try {
      const res = await api.post<ApiResponse<Board>>('/boards', {
        ...data,
        workspace_id: workspace.id
      });
      if (res.data.success && res.data.data) {
        setBoards([res.data.data, ...boards]);
        setIsCreateModalOpen(false);
        reset();
      }
    } catch (error) {
      console.error('Failed to create board:', error);
    }
  };

  const filteredBoards = boards.filter(b => b.name.toLowerCase().includes(searchQuery.toLowerCase()));

  if (isLoading) {
    return <div className="flex-1 flex items-center justify-center"><Loader2 className="w-8 h-8 animate-spin text-primary-500" /></div>;
  }

  if (!workspace) {
    return <div className="flex-1 flex items-center justify-center text-surface-500">Workspace not found</div>;
  }

  return (
    <div className="flex-1 flex flex-col bg-surface-100 overflow-hidden">
      {/* Workspace Header */}
      <div className="bg-white border-b border-surface-200 shrink-0">
        <div className="max-w-6xl mx-auto px-6 py-8">
          <div className="flex items-center space-x-4 mb-6">
            <div className="w-16 h-16 bg-gradient-to-br from-primary-500 to-primary-700 rounded-xl flex items-center justify-center text-white font-bold text-3xl shadow-sm">
              {workspace.name.charAt(0).toUpperCase()}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-surface-900">{workspace.name}</h1>
              <p className="text-surface-500 text-sm mt-1">{workspace.role} • Workspace</p>
            </div>
          </div>

          <div className="flex space-x-1 border-b border-surface-200">
            <button className="px-4 py-2 border-b-2 border-primary-600 text-primary-600 font-medium text-sm flex items-center">
              <Layout className="w-4 h-4 mr-2" />
              Boards
            </button>
            <button className="px-4 py-2 border-b-2 border-transparent text-surface-500 hover:text-surface-700 font-medium text-sm flex items-center">
              <Users className="w-4 h-4 mr-2" />
              Members
            </button>
            {(workspace.role === 'OWNER' || workspace.role === 'ADMIN') && (
              <button className="px-4 py-2 border-b-2 border-transparent text-surface-500 hover:text-surface-700 font-medium text-sm flex items-center">
                <Settings className="w-4 h-4 mr-2" />
                Settings
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Workspace Content */}
      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-6xl mx-auto">
          <div className="flex justify-between items-center mb-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-surface-400 w-4 h-4" />
              <input
                type="text"
                placeholder="Search boards..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9 pr-4 py-2 border border-surface-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 text-sm w-64"
              />
            </div>
            
            {(workspace.role === 'OWNER' || workspace.role === 'ADMIN') && (
              <button
                onClick={() => setIsCreateModalOpen(true)}
                className="flex items-center px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors shadow-sm font-medium text-sm"
              >
                <Plus className="w-4 h-4 mr-2" />
                Create Board
              </button>
            )}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {filteredBoards.map(board => (
              <Link
                key={board.id}
                to={`/b/${board.id}`}
                state={{ board, workspace }}
                className="group relative h-28 rounded-lg overflow-hidden shadow-sm hover:shadow-card transition-shadow flex flex-col justify-between p-3"
                style={{ backgroundColor: board.color || '#3b82f6' }}
              >
                {/* Overlay for better text readability */}
                <div className="absolute inset-0 bg-black/10 group-hover:bg-black/20 transition-colors"></div>
                <h3 className="relative z-10 text-white font-bold truncate">{board.name}</h3>
              </Link>
            ))}
            
            {(workspace.role === 'OWNER' || workspace.role === 'ADMIN') && (
              <button
                onClick={() => setIsCreateModalOpen(true)}
                className="h-28 rounded-lg bg-surface-200 border border-surface-300 flex flex-col items-center justify-center text-surface-600 hover:bg-surface-300 hover:text-surface-800 transition-colors"
              >
                <Plus className="w-6 h-6 mb-1" />
                <span className="font-medium text-sm">Create new board</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Create Board Modal */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-surface-900/50 backdrop-blur-sm">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-sm overflow-hidden">
            <div className="px-6 py-4 border-b border-surface-200 flex justify-between items-center">
              <h3 className="text-lg font-bold text-surface-900">Create board</h3>
              <button onClick={() => { setIsCreateModalOpen(false); reset(); }} className="text-surface-400 hover:text-surface-600">
                &times;
              </button>
            </div>
            
            <form onSubmit={handleSubmit(onCreateBoard)} className="p-6">
              {/* Preview preview */}
              <div 
                className="w-full h-24 rounded-md mb-6 flex items-center justify-center shadow-inner"
                style={{ backgroundColor: selectedColor }}
              >
                <img src="https://trello.com/assets/14cda5dc635d1f13bc48.svg" alt="Preview" className="h-20" />
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-surface-700 mb-1">Board title</label>
                  <input 
                    type="text" 
                    className="w-full px-3 py-2 border border-surface-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                    {...register('name')}
                    autoFocus
                  />
                  {errors.name && <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>}
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-surface-700 mb-2">Background color</label>
                  <div className="flex flex-wrap gap-2">
                    {bgColors.map(color => (
                      <button
                        key={color}
                        type="button"
                        onClick={() => setValue('color', color)}
                        className={`w-8 h-8 rounded-md hover:opacity-80 transition-opacity ${selectedColor === color ? 'ring-2 ring-offset-2 ring-primary-500' : ''}`}
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </div>
                </div>

                <div className="pt-4">
                  <button 
                    type="submit" 
                    disabled={isSubmitting}
                    className="w-full py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 font-medium text-sm disabled:opacity-70 flex justify-center items-center"
                  >
                    {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Create'}
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default WorkspacePage;
