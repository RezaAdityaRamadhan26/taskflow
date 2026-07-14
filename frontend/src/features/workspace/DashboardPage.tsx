import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../../services/api';
import type { Workspace, ApiResponse } from '../../types';
import { Plus, Users, Settings, Briefcase, Loader2, AlertCircle } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';

const createWsSchema = z.object({
  name: z.string().min(2, 'Name is required (min 2 chars)').max(100),
  slug: z.string().min(2).max(100).regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, 'Slug must be lowercase alphanumeric with hyphens'),
});

type CreateWsValues = z.infer<typeof createWsSchema>;

const DashboardPage = () => {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const { register, handleSubmit, formState: { errors, isSubmitting }, reset, watch, setValue } = useForm<CreateWsValues>({
    resolver: zodResolver(createWsSchema)
  });

  // Auto-generate slug from name
  const wsName = watch('name');
  useEffect(() => {
    if (wsName) {
      const generatedSlug = wsName.toLowerCase().trim().replace(/[^a-z0-9-]+/g, '-').replace(/-{2,}/g, '-').replace(/^-|-$/g, '');
      setValue('slug', generatedSlug, { shouldValidate: true });
    }
  }, [wsName, setValue]);

  const fetchWorkspaces = async () => {
    try {
      const res = await api.get<ApiResponse<Workspace[]>>('/workspaces');
      if (res.data.success && res.data.data) {
        setWorkspaces(res.data.data);
      }
    } catch (error) {
      console.error('Failed to fetch workspaces:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkspaces();
  }, []);

  const onCreateWorkspace = async (data: CreateWsValues) => {
    setCreateError(null);
    try {
      const res = await api.post<ApiResponse<Workspace>>('/workspaces', data);
      if (res.data.success && res.data.data) {
        setWorkspaces([res.data.data, ...workspaces]);
        setIsCreateModalOpen(false);
        reset();
      }
    } catch (err: any) {
      setCreateError(err.response?.data?.error || 'Failed to create workspace');
    }
  };

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-primary-500" />
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-auto bg-surface-100 p-6 md:p-10">
      <div className="max-w-5xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-2xl font-bold text-surface-900">Your Workspaces</h1>
            <p className="text-surface-500 text-sm mt-1">Manage your teams and organizations</p>
          </div>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="flex items-center px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors shadow-sm font-medium text-sm"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Workspace
          </button>
        </div>

        {workspaces.length === 0 ? (
          <div className="bg-white border border-surface-200 rounded-lg p-12 text-center shadow-sm">
            <div className="mx-auto w-16 h-16 bg-primary-50 rounded-full flex items-center justify-center mb-4">
              <Briefcase className="w-8 h-8 text-primary-500" />
            </div>
            <h3 className="text-lg font-medium text-surface-900 mb-2">No workspaces yet</h3>
            <p className="text-surface-500 mb-6 max-w-sm mx-auto">
              Create a workspace to start organizing your boards, lists, and cards with your team.
            </p>
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors shadow-sm font-medium text-sm"
            >
              <Plus className="w-4 h-4 mr-2" />
              Create your first Workspace
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {workspaces.map((ws) => (
              <Link
                key={ws.id}
                to={`/w/${ws.slug}`}
                state={{ workspaceId: ws.id }}
                className="bg-white border border-surface-200 rounded-lg p-6 hover:shadow-card-hover hover:border-primary-300 transition-all group"
              >
                <div className="flex justify-between items-start mb-4">
                  <div className="w-12 h-12 bg-gradient-to-br from-primary-500 to-primary-700 rounded-lg flex items-center justify-center text-white font-bold text-xl shadow-sm">
                    {ws.name.charAt(0).toUpperCase()}
                  </div>
                  <span className={`text-xs px-2 py-1 rounded-full font-medium ${
                    ws.role === 'OWNER' ? 'bg-purple-100 text-purple-700' :
                    ws.role === 'ADMIN' ? 'bg-blue-100 text-blue-700' :
                    'bg-surface-100 text-surface-600'
                  }`}>
                    {ws.role}
                  </span>
                </div>
                <h3 className="text-lg font-semibold text-surface-900 group-hover:text-primary-600 transition-colors">
                  {ws.name}
                </h3>
                <p className="text-surface-500 text-sm mt-1 mb-4 truncate">
                  {ws.description || 'No description provided.'}
                </p>
                <div className="flex items-center space-x-4 border-t border-surface-100 pt-4 mt-auto">
                  <div className="flex items-center text-surface-500 text-sm">
                    <Users className="w-4 h-4 mr-1.5" />
                    <span>Team</span>
                  </div>
                  <div className="flex items-center text-surface-500 text-sm hover:text-primary-600">
                    <Settings className="w-4 h-4 mr-1.5" />
                    <span>Settings</span>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}

        {/* Create Workspace Modal */}
        {isCreateModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-surface-900/50 backdrop-blur-sm">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md overflow-hidden">
              <div className="px-6 py-4 border-b border-surface-200 flex justify-between items-center">
                <h3 className="text-lg font-bold text-surface-900">Create Workspace</h3>
                <button onClick={() => { setIsCreateModalOpen(false); reset(); setCreateError(null); }} className="text-surface-400 hover:text-surface-600">
                  &times;
                </button>
              </div>
              
              <div className="p-6">
                {createError && (
                  <div className="mb-4 bg-red-50 p-3 rounded-md flex items-start">
                    <AlertCircle className="w-5 h-5 text-red-500 mr-2 shrink-0" />
                    <p className="text-sm text-red-700">{createError}</p>
                  </div>
                )}
                
                <form onSubmit={handleSubmit(onCreateWorkspace)} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-surface-700 mb-1">Workspace Name</label>
                    <input 
                      type="text" 
                      className="w-full px-3 py-2 border border-surface-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                      placeholder="e.g. Engineering Team"
                      {...register('name')}
                    />
                    {errors.name && <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>}
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-surface-700 mb-1">URL Slug</label>
                    <div className="flex rounded-md shadow-sm">
                      <span className="inline-flex items-center px-3 rounded-l-md border border-r-0 border-surface-300 bg-surface-50 text-surface-500 sm:text-sm">
                        taskflow.com/w/
                      </span>
                      <input 
                        type="text" 
                        className="flex-1 min-w-0 block w-full px-3 py-2 rounded-none rounded-r-md border border-surface-300 focus:outline-none focus:ring-2 focus:ring-primary-500 sm:text-sm"
                        placeholder="engineering-team"
                        {...register('slug')}
                      />
                    </div>
                    {errors.slug && <p className="mt-1 text-xs text-red-600">{errors.slug.message}</p>}
                  </div>

                  <div className="pt-4 flex justify-end space-x-3">
                    <button 
                      type="button" 
                      onClick={() => { setIsCreateModalOpen(false); reset(); }}
                      className="px-4 py-2 border border-surface-300 rounded-md text-sm font-medium text-surface-700 hover:bg-surface-50"
                    >
                      Cancel
                    </button>
                    <button 
                      type="submit" 
                      disabled={isSubmitting}
                      className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 font-medium text-sm disabled:opacity-70 flex items-center"
                    >
                      {isSubmitting && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                      Create Workspace
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default DashboardPage;
