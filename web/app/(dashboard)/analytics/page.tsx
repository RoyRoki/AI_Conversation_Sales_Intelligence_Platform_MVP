'use client';

import AnalyticsDashboard from '@/components/AnalyticsDashboard';
import LeadPipeline from '@/components/LeadPipeline';

export default function AnalyticsPage() {
  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white">Analytics Dashboard</h1>
        <p className="mt-2 text-gray-400">View detailed analytics and insights</p>
      </div>

      <AnalyticsDashboard />

      <div className="mt-8">
        <LeadPipeline />
      </div>
    </div>
  );
}

