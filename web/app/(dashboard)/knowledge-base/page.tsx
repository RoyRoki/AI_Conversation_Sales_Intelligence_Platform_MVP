'use client';

import ProtectedRoute from '@/components/ProtectedRoute';
import ProductKB from '@/components/ProductKB';

export default function KnowledgeBasePage() {
  return (
    <ProtectedRoute requireAdmin>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Product Knowledge Base</h1>
          <p className="mt-2 text-gray-400">Manage product information and knowledge base</p>
        </div>
        <ProductKB />
      </div>
    </ProtectedRoute>
  );
}

