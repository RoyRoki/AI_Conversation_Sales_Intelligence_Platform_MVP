'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { getUser, isAdmin, isAgent, isCustomer } from '@/lib/auth';
import type { DashboardMetrics } from '@/types';
import MetricCard from '@/components/dashboard/MetricCard';
import QuickActionCard from '@/components/dashboard/QuickActionCard';
import LoadingSkeleton from '@/components/dashboard/LoadingSkeleton';
import {
  ChatBubbleLeftRightIcon,
  BoltIcon,
  FaceSmileIcon,
  TrophyIcon,
  DocumentTextIcon,
  ChartBarIcon,
  Cog6ToothIcon,
  BookOpenIcon,
  UserCircleIcon,
  PresentationChartLineIcon,
} from '@heroicons/react/24/outline';

export default function Dashboard() {
  const router = useRouter();
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const user = getUser();
  const admin = isAdmin();
  const agent = isAgent();
  const customer = isCustomer();

  useEffect(() => {
    // Redirect customers to their own dashboard
    if (customer) {
      router.push('/customer');
      return;
    }
    loadDashboard();
  }, [customer, router]);

  const loadDashboard = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getDashboard();
      setMetrics(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load dashboard');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <LoadingSkeleton />;
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="bg-red-900/20 border-l-4 border-red-500 text-red-400 p-4 rounded-r-lg shadow-professional">
          <div className="flex">
            <div className="ml-3">
              <p className="text-sm font-medium">{error}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Calculate total for percentage calculations
  const totalIntents = metrics?.top_intents?.reduce((sum, intent) => sum + intent.count, 0) || 0;
  const totalObjections = metrics?.top_objections?.reduce((sum, obj) => sum + obj.count, 0) || 0;

  return (
    <div className="max-w-7xl mx-auto">
      {/* Header Section */}
      <div className="mb-10">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-4xl font-bold text-white mb-2">
              Welcome back, <span className="text-blue-400">{user?.email?.split('@')[0]}</span>
            </h1>
            <p className="text-lg text-gray-400">
              {admin ? 'Admin Dashboard' : agent ? 'Agent Dashboard' : 'Customer Dashboard'}
            </p>
          </div>
          <div className="hidden md:flex items-center space-x-2 text-sm text-gray-500">
            <span>Last updated:</span>
            <span className="font-medium text-gray-400">{new Date().toLocaleDateString()}</span>
          </div>
        </div>
      </div>

      {metrics && (
        <>
          {/* Metrics Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-10">
            <MetricCard
              title="Total Conversations"
              value={metrics.total_conversations}
              icon={<ChatBubbleLeftRightIcon className="w-6 h-6" />}
              gradient="blue"
            />
            <MetricCard
              title="Active Conversations"
              value={metrics.active_conversations}
              icon={<BoltIcon className="w-6 h-6" />}
              gradient="blue-light"
            />
            <MetricCard
              title="Average Sentiment"
              value={
                metrics.average_sentiment != null
                  ? `${(metrics.average_sentiment * 100).toFixed(1)}%`
                  : '0%'
              }
              icon={<FaceSmileIcon className="w-6 h-6" />}
              gradient="blue"
            />
            <MetricCard
              title="Win Rate"
              value={
                metrics.win_rate != null ? `${(metrics.win_rate * 100).toFixed(1)}%` : '0%'
              }
              icon={<TrophyIcon className="w-6 h-6" />}
              gradient="blue-dark"
            />
          </div>

          {/* Data Visualization Section */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-10">
            {/* Top Intents */}
            <div className="bg-gray-800 rounded-xl shadow-professional p-6 border border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-white">Top Intents</h3>
                <DocumentTextIcon className="w-6 h-6 text-blue-400" />
              </div>
              {metrics.top_intents && metrics.top_intents.length > 0 ? (
                <div className="space-y-4">
                  {metrics.top_intents.map((intent, idx) => {
                    const percentage = totalIntents > 0 ? (intent.count / totalIntents) * 100 : 0;
                    return (
                      <div key={idx} className="space-y-2">
                        <div className="flex items-center justify-between text-sm">
                          <span className="font-medium text-gray-300 capitalize">
                            {intent.intent}
                          </span>
                          <span className="text-gray-400">
                            {intent.count} ({percentage.toFixed(0)}%)
                          </span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2.5 overflow-hidden">
                          <div
                            className="bg-gradient-to-r from-blue-500 to-blue-600 h-2.5 rounded-full transition-all duration-500"
                            style={{ width: `${percentage}%` }}
                          ></div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <DocumentTextIcon className="w-12 h-12 mx-auto mb-3 text-gray-600" />
                  <p className="text-sm">No intents data available</p>
                </div>
              )}
            </div>

            {/* Top Objections */}
            <div className="bg-gray-800 rounded-xl shadow-professional p-6 border border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-white">Top Objections</h3>
                <ChartBarIcon className="w-6 h-6 text-red-400" />
              </div>
              {metrics.top_objections && metrics.top_objections.length > 0 ? (
                <div className="space-y-4">
                  {metrics.top_objections.map((objection, idx) => {
                    const percentage =
                      totalObjections > 0 ? (objection.count / totalObjections) * 100 : 0;
                    return (
                      <div key={idx} className="space-y-2">
                        <div className="flex items-center justify-between text-sm">
                          <span className="font-medium text-gray-300 capitalize">
                            {objection.objection}
                          </span>
                          <span className="text-gray-400">
                            {objection.count} ({percentage.toFixed(0)}%)
                          </span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2.5 overflow-hidden">
                          <div
                            className="bg-gradient-to-r from-red-500 to-red-600 h-2.5 rounded-full transition-all duration-500"
                            style={{ width: `${percentage}%` }}
                          ></div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <ChartBarIcon className="w-12 h-12 mx-auto mb-3 text-gray-600" />
                  <p className="text-sm">No objections data available</p>
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {/* Quick Actions Section */}
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-white mb-4">Quick Actions</h2>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <QuickActionCard
          title="View Conversations"
          description="Browse and manage all conversations"
          href="/conversations"
          icon={<ChatBubbleLeftRightIcon className="w-6 h-6" />}
        />
        {(admin || agent) && (
          <QuickActionCard
            title="Analytics"
            description="View detailed analytics and insights"
            href="/analytics"
            icon={<PresentationChartLineIcon className="w-6 h-6" />}
          />
        )}
        {admin && (
          <>
            <QuickActionCard
              title="Configure Rules"
              description="Manage business rules and policies"
              href="/rules"
              icon={<Cog6ToothIcon className="w-6 h-6" />}
            />
            <QuickActionCard
              title="Knowledge Base"
              description="Manage product knowledge base"
              href="/knowledge-base"
              icon={<BookOpenIcon className="w-6 h-6" />}
            />
            <QuickActionCard
              title="Customer Memory"
              description="View and edit customer preferences"
              href="/memory"
              icon={<UserCircleIcon className="w-6 h-6" />}
            />
            <QuickActionCard
              title="Agent Performance"
              description="Review agent performance metrics"
              href="/performance"
              icon={<ChartBarIcon className="w-6 h-6" />}
            />
          </>
        )}
      </div>
    </div>
  );
}
