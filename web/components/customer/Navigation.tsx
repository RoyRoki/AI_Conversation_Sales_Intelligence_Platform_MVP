'use client';

import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { getUser, clearAuth } from '@/lib/auth';
import { apiClient } from '@/lib/api';

export default function CustomerNavigation() {
  const router = useRouter();
  const user = getUser();

  const handleLogout = () => {
    apiClient.logout();
    clearAuth();
    router.push('/customer-login');
  };

  return (
    <nav className="bg-gray-800 border-b border-gray-700">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center space-x-8">
            <Link href="/customer" className="text-xl font-bold text-blue-400">
              OMX Digital
            </Link>
            <div className="flex space-x-4">
              <Link
                href="/customer"
                className="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
              >
                Home
              </Link>
              <Link
                href="/customer/products"
                className="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
              >
                Products
              </Link>
              <Link
                href="/customer/chat"
                className="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
              >
                Chat
              </Link>
            </div>
          </div>
          <div className="flex items-center space-x-4">
            {user && (
              <span className="text-gray-400 text-sm">{user.email}</span>
            )}
            <button
              onClick={handleLogout}
              className="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
}

