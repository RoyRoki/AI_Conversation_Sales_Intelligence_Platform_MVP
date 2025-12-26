'use client';

import ProtectedRoute from '@/components/ProtectedRoute';
import AgentPerformance from '@/components/AgentPerformance';

export default function PerformancePage() {
  return (
    <ProtectedRoute requireAdmin>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Agent Performance</h1>
          <p className="mt-2 text-gray-400">Review agent performance metrics and trends</p>
        </div>
        <AgentPerformance />
      </div>
    </ProtectedRoute>
  );
}

