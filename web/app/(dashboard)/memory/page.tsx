'use client';

import ProtectedRoute from '@/components/ProtectedRoute';
import CustomerMemory from '@/components/CustomerMemory';

export default function MemoryPage() {
  return (
    <ProtectedRoute requireAdmin>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Customer Memory</h1>
          <p className="mt-2 text-gray-400">View and edit customer preferences and history</p>
        </div>
        <CustomerMemory />
      </div>
    </ProtectedRoute>
  );
}

