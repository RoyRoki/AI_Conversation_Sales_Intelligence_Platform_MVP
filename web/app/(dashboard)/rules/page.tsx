'use client';

import ProtectedRoute from '@/components/ProtectedRoute';
import RuleConfig from '@/components/RuleConfig';

export default function RulesPage() {
  return (
    <ProtectedRoute requireAdmin>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Rule Configuration</h1>
          <p className="mt-2 text-gray-400">Manage business rules and policies</p>
        </div>
        <RuleConfig />
      </div>
    </ProtectedRoute>
  );
}

