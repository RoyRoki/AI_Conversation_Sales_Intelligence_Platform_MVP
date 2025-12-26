'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api';
import type { Conversation, PrioritizedLead } from '@/types';
import { format } from 'date-fns';
import CustomDropdown from '@/components/ui/CustomDropdown';

export default function ConversationsPage() {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [leads, setLeads] = useState<PrioritizedLead[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'date' | 'status'>('date');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  useEffect(() => {
    loadConversations();
  }, []);

  useEffect(() => {
    if (conversations.length > 0) {
      loadLeads();
    }
  }, [conversations]);

  const loadConversations = async () => {
    try {
      setLoading(true);
      const data = await apiClient.listConversations();
      setConversations(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load conversations');
    } finally {
      setLoading(false);
    }
  };

  const loadLeads = async () => {
    try {
      const conversationIds = conversations.map((c) => c.id);
      const { leads } = await apiClient.getLeads(conversationIds);
      setLeads(leads);
    } catch (err) {
      // Silently fail - leads are optional
      console.error('Failed to load leads:', err);
    }
  };

  const getLeadData = (conversationId: string): PrioritizedLead | undefined => {
    return leads.find((l) => l.conversation_id === conversationId);
  };

  const getWinProbabilityColor = (probability: number) => {
    if (probability >= 0.7) return 'text-green-400 bg-green-900/30';
    if (probability >= 0.4) return 'text-yellow-400 bg-yellow-900/30';
    return 'text-red-400 bg-red-900/30';
  };

  const getChurnRiskColor = (risk: string) => {
    if (risk === 'high') return 'text-red-400 bg-red-900/30';
    if (risk === 'medium') return 'text-yellow-400 bg-yellow-900/30';
    return 'text-green-400 bg-green-900/30';
  };

  const filteredConversations = conversations
    .filter((conv) => {
      if (filterStatus !== 'all' && conv.status !== filterStatus) return false;
      if (searchTerm) {
        const term = searchTerm.toLowerCase();
        return conv.id.toLowerCase().includes(term);
      }
      return true;
    })
    .sort((a, b) => {
      if (sortBy === 'date') {
        return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
      }
      return a.status.localeCompare(b.status);
    });

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="text-center text-gray-400">Loading conversations...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white">Conversations</h1>
        <p className="mt-2 text-gray-400">View and manage all conversations</p>
      </div>

      {error && (
        <div className="mb-4 bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="mb-6 flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-4 py-2 bg-gray-800 border border-gray-700 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <CustomDropdown
          value={filterStatus}
          onChange={(value) => setFilterStatus(value)}
          options={[
            { value: 'all', label: 'All Status' },
            { value: 'active', label: 'Active' },
            { value: 'closed', label: 'Closed' },
            { value: 'archived', label: 'Archived' },
          ]}
          placeholder="Filter by Status"
          className="min-w-[150px]"
        />
        <CustomDropdown
          value={sortBy}
          onChange={(value) => setSortBy(value as 'date' | 'status')}
          options={[
            { value: 'date', label: 'Sort by Date' },
            { value: 'status', label: 'Sort by Status' },
          ]}
          placeholder="Sort by"
          className="min-w-[150px]"
        />
      </div>

      <div className="bg-gray-800 shadow overflow-hidden sm:rounded-md border border-gray-700">
        <ul className="divide-y divide-gray-700">
          {filteredConversations.map((conversation) => {
            const lead = getLeadData(conversation.id);
            return (
              <li key={conversation.id}>
                <Link
                  href={`/conversations/${conversation.id}`}
                  className="block hover:bg-gray-700/50 transition-colors"
                >
                  <div className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div>
                          <p className="text-sm font-medium text-blue-400 truncate">
                            {conversation.id}
                          </p>
                          {conversation.customer_email && (
                            <p className="text-sm text-gray-400 truncate">
                              {conversation.customer_email}
                            </p>
                          )}
                        </div>
                        <span
                          className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            conversation.status === 'active'
                              ? 'bg-green-900/30 text-green-400'
                              : conversation.status === 'closed'
                              ? 'bg-gray-700 text-gray-300'
                              : 'bg-yellow-900/30 text-yellow-400'
                          }`}
                        >
                          {conversation.status}
                        </span>
                      </div>
                      <div className="ml-2 flex-shrink-0 flex">
                        {lead && (
                          <div className="flex items-center space-x-2">
                            <span
                              className={`px-2 py-1 rounded text-xs font-medium ${getWinProbabilityColor(
                                lead.win_probability
                              )}`}
                            >
                              Win: {(lead.win_probability * 100).toFixed(0)}%
                            </span>
                            <span
                              className={`px-2 py-1 rounded text-xs font-medium ${
                                lead.urgency === 'high'
                                  ? 'bg-red-900/30 text-red-400'
                                  : lead.urgency === 'medium'
                                  ? 'bg-yellow-900/30 text-yellow-400'
                                  : 'bg-green-900/30 text-green-400'
                              }`}
                            >
                              {lead.urgency} priority
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="mt-2 sm:flex sm:justify-between">
                      <div className="sm:flex">
                        <p className="flex items-center text-sm text-gray-400">
                          Updated: {format(new Date(conversation.updated_at), 'PPp')}
                        </p>
                      </div>
                    </div>
                  </div>
                </Link>
              </li>
            );
          })}
        </ul>
        {filteredConversations.length === 0 && (
          <div className="text-center py-12 text-gray-400">No conversations found</div>
        )}
      </div>
    </div>
  );
}

