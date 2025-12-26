'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { setAuth } from '@/lib/auth';

export default function CustomerLogin() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) {
      setError('Please enter your email');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // For demo: Store email in localStorage and create mock user
      const trimmedEmail = email.trim();
      localStorage.setItem('customer_email', trimmedEmail);
      
      // Create mock user object for demo
      const mockUser = {
        id: `customer-${Date.now()}`,
        tenant_id: 'OMX26',
        email: trimmedEmail,
        role: 'customer' as const,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      // Generate a simple demo token (just a string, not a real JWT)
      const demoToken = `demo-token-${Date.now()}-${Math.random().toString(36).substring(7)}`;
      
      // Store auth data
      setAuth(demoToken, mockUser);
      
      // Try API call, but don't fail if it errors (for demo)
      try {
        await apiClient.customerLogin(trimmedEmail, 'OMX26');
      } catch (apiErr) {
        // API call failed, but we continue with demo mode
        console.log('API call failed, using demo mode:', apiErr);
      }
      
      router.push('/customer');
    } catch (err: any) {
      setError(err.message || 'Failed to login. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center px-4" style={{ minHeight: '100vh' }}>
      <div className="max-w-md w-full bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Welcome</h1>
          <p className="text-gray-400">Enter your email to get started</p>
        </div>

        {error && (
          <div className="mb-4 bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-300 mb-2">
              Email Address
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="your.email@example.com"
              className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
              disabled={loading}
            />
          </div>

          <button
            type="submit"
            disabled={loading || !email.trim()}
            className="w-full px-4 py-3 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? 'Logging in...' : 'Continue'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-400">
          No password required. We'll use your email to track your conversations.
        </p>
      </div>
    </div>
  );
}

