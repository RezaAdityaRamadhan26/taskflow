import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './store/authStore';
import { api } from './services/api';
import type { ApiResponse, User } from './types';

// Layouts
import MainLayout from './components/layouts/MainLayout';
import AuthLayout from './components/layouts/AuthLayout';

// Pages
import LoginPage from './features/auth/LoginPage';
import RegisterPage from './features/auth/RegisterPage';
import DashboardPage from './features/workspace/DashboardPage';
import WorkspacePage from './features/workspace/WorkspacePage';
import BoardPage from './features/board/BoardPage';
import NotFoundPage from './components/NotFoundPage';

// Protected Route Component
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated, isLoading } = useAuthStore();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-surface-100">
        <div className="w-10 h-10 border-4 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

function App() {
  const { setUser, setAccessToken, setLoading } = useAuthStore();

  useEffect(() => {
    // Check session on app load
    const initAuth = async () => {
      try {
        // Try to refresh token silently using httpOnly cookie
        const res = await api.post<{ success: boolean; data: { access_token: string } }>('/auth/refresh');
        if (res.data?.success && res.data.data?.access_token) {
          setAccessToken(res.data.data.access_token);
          
          // Fetch user profile
          const userRes = await api.get<ApiResponse<User>>('/auth/me');
          if (userRes.data?.success && userRes.data.data) {
            setUser(userRes.data.data);
          }
        }
      } catch (error) {
        // No valid session, that's fine, they just need to log in
        console.log('No active session found.');
      } finally {
        setLoading(false);
      }
    };

    initAuth();
  }, [setAccessToken, setUser, setLoading]);

  return (
    <BrowserRouter>
      <Routes>
        {/* Auth Routes */}
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
        </Route>

        {/* Protected Routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MainLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="w/:slug" element={<WorkspacePage />} />
          <Route path="b/:boardId" element={<BoardPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
