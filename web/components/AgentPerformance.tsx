'use client';

import { useEffect, useState } from 'react';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { apiClient } from '@/lib/api';

interface AgentMetrics {
  agent_id: string;
  agent_name: string;
  response_time_avg: number;
  conversation_quality_score: number;
  sentiment_improvement_rate: number;
  total_conversations: number;
  win_rate: number;
}

interface AgentPerformanceProps {
  dateRange?: { start: string; end: string };
}

export default function AgentPerformance({ dateRange }: AgentPerformanceProps) {
  const [metrics, setMetrics] = useState<AgentMetrics[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [trendData, setTrendData] = useState<any[]>([]);

  useEffect(() => {
    // Note: This would need backend API endpoints for agent performance
    // For now, this is a placeholder with mock data
    setTimeout(() => {
      setMetrics([
        {
          agent_id: 'agent1',
          agent_name: 'John Doe',
          response_time_avg: 2.5,
          conversation_quality_score: 0.85,
          sentiment_improvement_rate: 0.72,
          total_conversations: 150,
          win_rate: 0.68,
        },
        {
          agent_id: 'agent2',
          agent_name: 'Jane Smith',
          response_time_avg: 3.2,
          conversation_quality_score: 0.78,
          sentiment_improvement_rate: 0.65,
          total_conversations: 120,
          win_rate: 0.62,
        },
      ]);
      setLoading(false);
    }, 1000);
  }, [dateRange]);

  if (loading) {
    return <div className="text-center py-8 text-gray-400">Loading agent performance data...</div>;
  }

  const chartData = metrics.map((m) => ({
    name: m.agent_name,
    'Response Time (min)': m.response_time_avg,
    'Quality Score': m.conversation_quality_score * 100,
    'Win Rate': m.win_rate * 100,
  }));

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-white">Agent Performance Review</h2>
        <div className="flex space-x-2">
          <input
            type="date"
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="date"
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map((metric) => (
          <div key={metric.agent_id} className="bg-gray-800 rounded-lg shadow p-4 border border-gray-700">
            <h3 className="text-sm font-medium text-gray-400">{metric.agent_name}</h3>
            <div className="mt-2 space-y-1 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-400">Avg Response Time:</span>
                <span className="font-medium text-white">{metric.response_time_avg.toFixed(1)} min</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Quality Score:</span>
                <span className="font-medium text-white">
                  {(metric.conversation_quality_score * 100).toFixed(0)}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Sentiment Improvement:</span>
                <span className="font-medium text-white">
                  {(metric.sentiment_improvement_rate * 100).toFixed(0)}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Total Conversations:</span>
                <span className="font-medium text-white">{metric.total_conversations}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Win Rate:</span>
                <span className="font-medium text-white">{(metric.win_rate * 100).toFixed(0)}%</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
        <h3 className="text-lg font-semibold text-white mb-4">Performance Comparison</h3>
        <ResponsiveContainer width="100%" height={400}>
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="name" stroke="#9CA3AF" />
            <YAxis stroke="#9CA3AF" />
            <Tooltip />
            <Legend />
            <Bar dataKey="Response Time (min)" fill="#4F46E5" />
            <Bar dataKey="Quality Score" fill="#10B981" />
            <Bar dataKey="Win Rate" fill="#F59E0B" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
        <h3 className="text-lg font-semibold text-white mb-4">Performance Trends</h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={trendData.length > 0 ? trendData : chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="name" stroke="#9CA3AF" />
            <YAxis stroke="#9CA3AF" />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="Quality Score" stroke="#4F46E5" />
            <Line type="monotone" dataKey="Win Rate" stroke="#10B981" />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
