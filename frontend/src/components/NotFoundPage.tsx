import { Link } from 'react-router-dom';
import { Home } from 'lucide-react';

const NotFoundPage = () => {
  return (
    <div className="min-h-screen bg-surface-100 flex flex-col justify-center items-center px-4">
      <h1 className="text-6xl font-bold text-primary-600">404</h1>
      <h2 className="mt-4 text-2xl font-bold text-surface-900">Page not found</h2>
      <p className="mt-2 text-surface-600 text-center max-w-md">
        Sorry, we couldn't find the page you're looking for. It might have been removed or the link might be broken.
      </p>
      <Link
        to="/"
        className="mt-8 flex items-center space-x-2 bg-primary-600 text-white px-6 py-3 rounded-md hover:bg-primary-700 transition-colors font-medium"
      >
        <Home className="w-5 h-5" />
        <span>Back to Home</span>
      </Link>
    </div>
  );
};

export default NotFoundPage;
