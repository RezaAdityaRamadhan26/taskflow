import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { api } from '../../services/api';
import { CheckSquare, LogOut, LayoutDashboard, User as UserIcon } from 'lucide-react';
import { useState } from 'react';

const MainLayout = () => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout failed:', error);
    } finally {
      logout();
      navigate('/login');
    }
  };

  return (
    <div className="min-h-screen bg-surface-100 flex flex-col">
      {/* Top Navbar */}
      <header className="bg-white border-b border-surface-200 h-14 flex items-center justify-between px-4 z-10">
        <div className="flex items-center space-x-6">
          <Link to="/" className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-primary-600 rounded-md flex items-center justify-center">
              <CheckSquare className="text-white w-5 h-5" />
            </div>
            <span className="font-bold text-xl text-surface-900 tracking-tight hidden sm:block">TaskFlow</span>
          </Link>
          
          <nav className="hidden md:flex space-x-4">
            <Link to="/" className="text-surface-600 hover:text-primary-600 hover:bg-primary-50 px-3 py-1.5 rounded-md font-medium text-sm transition-colors flex items-center space-x-1">
              <LayoutDashboard className="w-4 h-4" />
              <span>Dashboard</span>
            </Link>
          </nav>
        </div>

        <div className="flex items-center">
          <div className="relative">
            <button 
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="flex items-center space-x-2 hover:bg-surface-100 p-1.5 rounded-full transition-colors focus:outline-none"
            >
              {user?.avatar_url ? (
                <img src={user.avatar_url} alt={user.name} className="w-8 h-8 rounded-full object-cover border border-surface-300" />
              ) : (
                <div className="w-8 h-8 bg-primary-100 text-primary-700 rounded-full flex items-center justify-center font-bold text-sm border border-primary-200">
                  {user?.name.charAt(0).toUpperCase()}
                </div>
              )}
            </button>

            {isMenuOpen && (
              <>
                <div 
                  className="fixed inset-0 z-40" 
                  onClick={() => setIsMenuOpen(false)}
                ></div>
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-surface-200">
                  <div className="px-4 py-3 border-b border-surface-100">
                    <p className="text-sm font-medium text-surface-900 truncate">{user?.name}</p>
                    <p className="text-xs text-surface-500 truncate">{user?.email}</p>
                  </div>
                  <button
                    onClick={() => {
                      setIsMenuOpen(false);
                      // TODO: Profile modal
                    }}
                    className="flex w-full items-center px-4 py-2 text-sm text-surface-700 hover:bg-surface-100"
                  >
                    <UserIcon className="w-4 h-4 mr-2" />
                    Profile
                  </button>
                  <button
                    onClick={handleLogout}
                    className="flex w-full items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                  >
                    <LogOut className="w-4 h-4 mr-2" />
                    Sign out
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col overflow-hidden relative">
        <Outlet />
      </main>
    </div>
  );
};

export default MainLayout;
