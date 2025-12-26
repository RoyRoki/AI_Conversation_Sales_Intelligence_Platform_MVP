'use client';

import { useEffect, useState } from 'react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { apiClient } from '@/lib/api';
import type { DashboardMetrics, TrendAnalysis } from '@/types';

interface AnalyticsDashboardProps {
  conversationId?: string;
}

const COLORS = ['#4F46E5', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'];

export default function AnalyticsDashboard({ conversationId }: AnalyticsDashboardProps) {
  const [dashboardMetrics, setDashboardMetrics] = useState<DashboardMetrics | null>(null);
  const [trends, setTrends] = useState<TrendAnalysis | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboard();
    if (conversationId) {
      loadTrends();
    }
  }, [conversationId]);

  const loadDashboard = async () => {
    try {
      const data = await apiClient.getDashboard();
      setDashboardMetrics(data);
    } catch (err) {
      console.error('Failed to load dashboard:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadTrends = async () => {
    if (!conversationId) return;
    try {
      const data = await apiClient.getTrends(conversationId);
      setTrends(data);
    } catch (err) {
      console.error('Failed to load trends:', err);
    }
  };

  if (loading) {
    return <div className="text-center py-8 text-gray-400">Loading analytics...</div>;
  }

  const intentData =
    dashboardMetrics?.top_intents && Array.isArray(dashboardMetrics.top_intents)
      ? dashboardMetrics.top_intents.map((intent) => ({
        name: intent.intent,
        value: intent.count,
      }))
      : [];

  const objectionData =
    dashboardMetrics?.top_objections && Array.isArray(dashboardMetrics.top_objections)
      ? dashboardMetrics.top_objections.map((obj) => ({
        name: obj.objection,
        value: obj.count,
      }))
      : [];

  const trendData =
    trends?.data_points && Array.isArray(trends.data_points)
      ? trends.data_points.map((point) => ({
        date: new Date(point.timestamp).toLocaleDateString(),
        sentiment: point.sentiment_score,
        engagement: point.engagement_score,
      }))
      : [];

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg shadow p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400">Total Conversations</h3>
          <p className="text-2xl font-bold text-white mt-2">
            {dashboardMetrics?.total_conversations || 0}
          </p>
        </div>
        <div className="bg-gray-800 rounded-lg shadow p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400">Active Conversations</h3>
          <p className="text-2xl font-bold text-white mt-2">
            {dashboardMetrics?.active_conversations || 0}
          </p>
        </div>
        <div className="bg-gray-800 rounded-lg shadow p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400">Average Sentiment</h3>
          <p className="text-2xl font-bold text-white mt-2">
            {dashboardMetrics && dashboardMetrics.average_sentiment != null
              ? (dashboardMetrics.average_sentiment * 100).toFixed(1)
              : 0}%
          </p>
        </div>
        <div className="bg-gray-800 rounded-lg shadow p-4 border border-gray-700">
          <h3 className="text-sm font-medium text-gray-400">Win Rate</h3>
          <p className="text-2xl font-bold text-white mt-2">
            {dashboardMetrics && dashboardMetrics.win_rate != null
              ? (dashboardMetrics.win_rate * 100).toFixed(1)
              : 0}%
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Top Intents</h3>
          {intentData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={intentData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${percent ? (percent * 100).toFixed(0) : 0}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {intentData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="text-center text-gray-400 py-8">No data available</div>
          )}
        </div>

        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Top Objections</h3>
          {objectionData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={objectionData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis dataKey="name" stroke="#9CA3AF" />
                <YAxis stroke="#9CA3AF" />
                <Tooltip />
                <Bar dataKey="value" fill="#4F46E5" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="text-center text-gray-400 py-8">No data available</div>
          )}
        </div>
      </div>

      {trends && trendData.length > 0 && (
        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Trends Over Time</h3>
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={trendData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis dataKey="date" stroke="#9CA3AF" />
              <YAxis stroke="#9CA3AF" />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="sentiment" stroke="#4F46E5" name="Sentiment" />
              <Line type="monotone" dataKey="engagement" stroke="#10B981" name="Engagement" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}

