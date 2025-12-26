'use client';

import { useEffect, useState, Fragment } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api';
import type { PrioritizedLead } from '@/types';
import CustomDropdown from '@/components/ui/CustomDropdown';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDistanceToNow } from 'date-fns';

type SortField = 'score' | 'win_probability' | 'urgency' | 'last_message_time';
type SortDirection = 'asc' | 'desc';

export default function LeadPipeline() {
  const [leads, setLeads] = useState<PrioritizedLead[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterUrgency, setFilterUrgency] = useState<string>('all');
  const [filterStage, setFilterStage] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState<SortField>('score');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadLeads();
  }, []);

  const loadLeads = async () => {
    try {
      setLoading(true);
      const { leads: leadsData } = await apiClient.getLeads();
      setLeads(leadsData);
    } catch (err) {
      console.error('Failed to load leads:', err);
    } finally {
      setLoading(false);
    }
  };

  // Transform backend data to match frontend expectations
  const transformedLeads = leads.map((lead: any) => {
    const score = lead.score ?? lead.priority_score ?? 0;
    const winProbability = lead.win_probability ?? 0;
    
    // Convert urgency_score (number) to urgency (string) if needed
    let urgency = lead.urgency;
    if (!urgency && lead.urgency_score !== undefined) {
      const urgencyScore = lead.urgency_score;
      if (urgencyScore >= 0.7) urgency = 'high';
      else if (urgencyScore >= 0.4) urgency = 'medium';
      else urgency = 'low';
    }
    urgency = urgency || 'low';
    
    return {
      ...lead,
      score,
      win_probability: winProbability,
      urgency,
      urgency_score: lead.urgency_score ?? 0,
    };
  });

  // Filter leads
  const filteredLeads = transformedLeads.filter((lead) => {
    if (filterUrgency !== 'all' && lead.urgency !== filterUrgency) return false;
    if (filterStage !== 'all' && lead.lead_stage !== filterStage) return false;
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      const email = lead.customer_email?.toLowerCase() || '';
      const convId = lead.conversation_id.toLowerCase();
      if (!email.includes(term) && !convId.includes(term)) return false;
    }
    return true;
  });

  // Sort leads
  const sortedLeads = [...filteredLeads].sort((a, b) => {
    let aVal: any;
    let bVal: any;

    switch (sortField) {
      case 'score':
        aVal = a.score || 0;
        bVal = b.score || 0;
        break;
      case 'win_probability':
        aVal = a.win_probability || 0;
        bVal = b.win_probability || 0;
        break;
      case 'urgency':
        const urgencyOrder: Record<string, number> = { high: 3, medium: 2, low: 1 };
        aVal = urgencyOrder[a.urgency || 'low'] || 0;
        bVal = urgencyOrder[b.urgency || 'low'] || 0;
        break;
      case 'last_message_time':
        aVal = a.engagement?.last_message_time ? new Date(a.engagement.last_message_time).getTime() : 0;
        bVal = b.engagement?.last_message_time ? new Date(b.engagement.last_message_time).getTime() : 0;
        break;
      default:
        return 0;
    }

    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
    }
  });

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const toggleRow = (conversationId: string) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(conversationId)) {
      newExpanded.delete(conversationId);
    } else {
      newExpanded.add(conversationId);
    }
    setExpandedRows(newExpanded);
  };

  const getUrgencyBadge = (lead: PrioritizedLead) => {
    const urgency = lead.urgency || 'low';
    const urgencyScore = lead.urgency_score || 0;
    
    if (urgency === 'high' || urgencyScore >= 0.7) {
      return (
        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-semibold bg-red-600 text-white border border-red-500 shadow-sm whitespace-nowrap">
          <span>🔥</span>
          <span>Hot</span>
        </span>
      );
    } else if (urgency === 'medium' || urgencyScore >= 0.4) {
      return (
        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-semibold bg-yellow-600 text-white border border-yellow-500 shadow-sm whitespace-nowrap">
          <span>⚠</span>
          <span>Warm</span>
        </span>
      );
    } else {
      return (
        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-semibold bg-blue-600 text-white border border-blue-500 shadow-sm whitespace-nowrap">
          <span>❄</span>
          <span>Cold</span>
        </span>
      );
    }
  };

  const getSentimentTrendIcon = (trend?: string) => {
    if (!trend) return null;
    const trendLower = trend.toLowerCase();
    if (trendLower === 'improving' || trendLower === 'up') {
      return (
        <svg className="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
      );
    } else if (trendLower === 'deteriorating' || trendLower === 'down') {
      return (
        <svg className="w-4 h-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
        </svg>
      );
    } else {
      return (
        <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
        </svg>
      );
    }
  };

  const getUrgencyColor = (urgency: string) => {
    if (urgency === 'high') return 'bg-red-900/30 text-red-400 border-red-700';
    if (urgency === 'medium') return 'bg-yellow-900/30 text-yellow-400 border-yellow-700';
    return 'bg-green-900/30 text-green-400 border-green-700';
  };

  const getWinProbabilityColor = (probability: number) => {
    if (probability >= 0.7) return 'bg-green-900/30 text-green-400 border-green-700';
    if (probability >= 0.4) return 'bg-yellow-900/30 text-yellow-400 border-yellow-700';
    return 'bg-red-900/30 text-red-400 border-red-700';
  };

  const getStageColor = (stage?: string) => {
    if (!stage) return 'bg-gray-700 text-gray-300';
    switch (stage) {
      case 'decision':
        return 'bg-green-900/30 text-green-400 border-green-700';
      case 'evaluation':
        return 'bg-yellow-900/30 text-yellow-400 border-yellow-700';
      default:
        return 'bg-blue-900/30 text-blue-400 border-blue-700';
    }
  };

  const formatLastMessageTime = (timeStr?: string) => {
    if (!timeStr) return 'N/A';
    try {
      return formatDistanceToNow(new Date(timeStr), { addSuffix: true });
    } catch {
      return 'N/A';
    }
  };

  if (loading) {
    return <div className="text-center py-8 text-gray-400">Loading leads...</div>;
  }

  return (
    <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-white">Lead Pipeline</h2>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder="Search by email or ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-4 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <CustomDropdown
            value={filterUrgency}
            onChange={(value) => setFilterUrgency(value)}
            options={[
              { value: 'all', label: 'All Urgency' },
              { value: 'high', label: 'High' },
              { value: 'medium', label: 'Medium' },
              { value: 'low', label: 'Low' },
            ]}
            placeholder="Filter by Urgency"
            className="min-w-[150px]"
          />
          <CustomDropdown
            value={filterStage}
            onChange={(value) => setFilterStage(value)}
            options={[
              { value: 'all', label: 'All Stages' },
              { value: 'discovery', label: 'Discovery' },
              { value: 'evaluation', label: 'Evaluation' },
              { value: 'decision', label: 'Decision' },
            ]}
            placeholder="Filter by Stage"
            className="min-w-[150px]"
          />
        </div>
      </div>

      <div className="overflow-x-auto -mx-6 px-6 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
        <Table className="min-w-full">
          <TableHeader>
            <TableRow>
              <TableHead className="w-12 sticky left-0 bg-gray-800 z-10"></TableHead>
              <TableHead className="min-w-[180px]">
                <button
                  onClick={() => handleSort('score')}
                  className="flex items-center gap-1 hover:text-white"
                >
                  Customer Email
                  {sortField === 'score' && (sortDirection === 'asc' ? '↑' : '↓')}
                </button>
              </TableHead>
              <TableHead className="min-w-[120px]">
                <button
                  onClick={() => handleSort('score')}
                  className="flex items-center gap-1 hover:text-white"
                >
                  Score
                  {sortField === 'score' && (sortDirection === 'asc' ? '↑' : '↓')}
                </button>
              </TableHead>
              <TableHead className="min-w-[100px]">
                <button
                  onClick={() => handleSort('win_probability')}
                  className="flex items-center gap-1 hover:text-white"
                >
                  Win Prob
                  {sortField === 'win_probability' && (sortDirection === 'asc' ? '↑' : '↓')}
                </button>
              </TableHead>
              <TableHead className="min-w-[100px]">
                <button
                  onClick={() => handleSort('urgency')}
                  className="flex items-center gap-1 hover:text-white"
                >
                  Urgency
                  {sortField === 'urgency' && (sortDirection === 'asc' ? '↑' : '↓')}
                </button>
              </TableHead>
              <TableHead className="min-w-[180px]">Summary</TableHead>
              <TableHead className="min-w-[100px] sticky right-0 bg-gray-800 z-10">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedLeads.map((lead) => {
              const isExpanded = expandedRows.has(lead.conversation_id);
              return (
                <Fragment key={lead.conversation_id}>
                  <TableRow className="hover:bg-gray-700/50 transition-colors cursor-pointer group">
                    <TableCell 
                      className="sticky left-0 bg-gray-800 z-10 cursor-pointer"
                      onClick={() => toggleRow(lead.conversation_id)}
                    >
                      <div className="flex items-center justify-center w-full h-full py-2">
                        <span className="text-gray-400 hover:text-white transition-colors">
                          {isExpanded ? '▼' : '▶'}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium text-white">
                          {lead.customer_email || 'No email'}
                        </div>
                        <div className="text-xs text-gray-400">
                          {lead.conversation_id.slice(0, 8)}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="font-semibold text-white">
                        {typeof lead.score === 'number' ? lead.score.toFixed(1) : 'N/A'}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="flex-1 min-w-[60px]">
                          <div className="relative h-2 bg-gray-700 rounded-full overflow-hidden">
                            <div
                              className={`h-full transition-all duration-300 ${
                                (lead.win_probability || 0) >= 0.7 ? 'bg-green-500' :
                                (lead.win_probability || 0) >= 0.4 ? 'bg-yellow-500' : 'bg-red-500'
                              }`}
                              style={{ width: `${((lead.win_probability || 0) * 100)}%` }}
                            />
                          </div>
                        </div>
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold border whitespace-nowrap ${getWinProbabilityColor(lead.win_probability || 0)}`}>
                          {typeof lead.win_probability === 'number' 
                            ? (lead.win_probability * 100).toFixed(0) 
                            : '0'}%
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      {getUrgencyBadge(lead)}
                    </TableCell>
                    <TableCell>
                      <div className="text-sm space-y-1.5">
                        {lead.ai_insights?.intent && (
                          <div className="flex items-center gap-2">
                            <span className="text-gray-500 text-xs">Intent:</span>
                            <span className="text-white capitalize text-xs font-medium">{lead.ai_insights.intent}</span>
                          </div>
                        )}
                        {lead.engagement && (
                          <div className="flex items-center gap-2">
                            <span className="text-gray-500 text-xs">Last:</span>
                            <span className="text-white text-xs font-medium">{formatLastMessageTime(lead.engagement.last_message_time)}</span>
                          </div>
                        )}
                        {lead.recommended_action && (
                          <div className="truncate max-w-[150px]" title={lead.recommended_action}>
                            <span className="text-blue-400 text-xs font-medium">{lead.recommended_action}</span>
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="sticky right-0 bg-gray-800 z-10">
                      <Link
                        href={`/conversations/${lead.conversation_id}`}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-md transition-colors"
                        onClick={(e) => e.stopPropagation()}
                      >
                        View
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </Link>
                    </TableCell>
                  </TableRow>
                  {isExpanded && (
                    <TableRow>
                      <TableCell colSpan={6} className="bg-gray-700/30 p-0">
                        <div className="p-6">
                          {/* Recommended Action - Prominent at Top */}
                          {lead.recommended_action && (
                            <div className="mb-5 flex items-center justify-between">
                              <h3 className="text-sm font-medium text-gray-400">Recommended Action</h3>
                              <Link
                                href={`/conversations/${lead.conversation_id}`}
                                className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition-colors shadow-lg flex items-center gap-2"
                                onClick={(e) => e.stopPropagation()}
                              >
                                Continue Conversation
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                                </svg>
                              </Link>
                            </div>
                          )}
                          
                          {/* Two-Column Grid Layout */}
                          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                            {/* Left Column */}
                            <div className="space-y-4">
                              {/* Lead Context Card */}
                              {lead.lead_context && (
                                <div className="bg-gray-900/50 rounded-lg border border-gray-700 p-4">
                                  <h4 className="font-semibold text-white mb-3 text-sm">Lead Context</h4>
                                  <div className="grid grid-cols-2 gap-3 text-sm">
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Source:</span>
                                      <span className="text-white font-medium">{lead.lead_context.source}</span>
                                    </div>
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Channel:</span>
                                      <span className="text-white font-medium">{lead.lead_context.channel}</span>
                                    </div>
                                    {lead.lead_context.product_interest && (
                                      <div className="break-words col-span-2">
                                        <span className="text-gray-500 block mb-1 text-xs">Product Interest:</span>
                                        <span className="text-white font-medium break-all">{lead.lead_context.product_interest}</span>
                                      </div>
                                    )}
                                    <div className="break-words col-span-2">
                                      <span className="text-gray-500 block mb-1 text-xs">Customer Type:</span>
                                      <span className="text-white font-medium">{lead.lead_context.customer_type}</span>
                                    </div>
                                  </div>
                                </div>
                              )}

                              {/* AI Insights Card */}
                              {lead.ai_insights && (
                                <div className="bg-gray-900/50 rounded-lg border border-gray-700 p-4">
                                  <h4 className="font-semibold text-white mb-3 text-sm">AI Insights</h4>
                                  <div className="space-y-3 text-sm">
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Intent:</span>
                                      <span className="text-white font-medium capitalize">{lead.ai_insights.intent || 'Unknown'}</span>
                                    </div>
                                    {lead.ai_insights.primary_objection && (
                                      <div className="break-words">
                                        <span className="text-gray-500 block mb-1 text-xs">Primary Objection:</span>
                                        <span className="text-white font-medium capitalize">{lead.ai_insights.primary_objection}</span>
                                      </div>
                                    )}
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Sentiment Trend:</span>
                                      <div className="flex items-center gap-2">
                                        {getSentimentTrendIcon(lead.ai_insights.sentiment_trend)}
                                        <span className="text-white font-medium capitalize">{lead.ai_insights.sentiment_trend || 'Stable'}</span>
                                      </div>
                                    </div>
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Confidence:</span>
                                      <div className="flex items-center gap-2">
                                        <div className="flex-1 h-2 bg-gray-700 rounded-full overflow-hidden">
                                          <div
                                            className={`h-full transition-all duration-300 ${
                                              (lead.ai_insights.confidence || 0) >= 0.7 ? 'bg-green-500' :
                                              (lead.ai_insights.confidence || 0) >= 0.4 ? 'bg-yellow-500' : 'bg-red-500'
                                            }`}
                                            style={{ width: `${((lead.ai_insights.confidence || 0) * 100)}%` }}
                                          />
                                        </div>
                                        <span className="text-white font-medium">{(lead.ai_insights.confidence * 100).toFixed(0)}%</span>
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              )}
                            </div>

                            {/* Right Column */}
                            <div className="space-y-4">
                              {/* Engagement Card */}
                              {lead.engagement && (
                                <div className="bg-gray-900/50 rounded-lg border border-gray-700 p-4">
                                  <h4 className="font-semibold text-white mb-3 text-sm flex items-center gap-2">
                                    Engagement
                                    {lead.engagement.silence_detected && (
                                      <span className="relative flex h-2 w-2">
                                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                                        <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
                                      </span>
                                    )}
                                  </h4>
                                  <div className="space-y-3 text-sm">
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Last Message:</span>
                                      <span className="text-white font-medium">{formatLastMessageTime(lead.engagement.last_message_time)}</span>
                                    </div>
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Response Delay Risk:</span>
                                      <span className={`font-medium capitalize inline-flex items-center gap-1 px-2 py-1 rounded ${
                                        lead.engagement.response_delay_risk === 'high' ? 'bg-red-900/30 text-red-400 border border-red-700' :
                                        lead.engagement.response_delay_risk === 'medium' ? 'bg-yellow-900/30 text-yellow-400 border border-yellow-700' :
                                        'bg-green-900/30 text-green-400 border border-green-700'
                                      }`}>
                                        {lead.engagement.response_delay_risk || 'low'}
                                      </span>
                                    </div>
                                    <div className="break-words">
                                      <span className="text-gray-500 block mb-1 text-xs">Silence Detected:</span>
                                      <span className={`font-medium inline-flex items-center gap-1 px-2 py-1 rounded ${
                                        lead.engagement.silence_detected 
                                          ? 'bg-red-900/30 text-red-400 border border-red-700' 
                                          : 'bg-green-900/30 text-green-400 border border-green-700'
                                      }`}>
                                        {lead.engagement.silence_detected ? 'Yes' : 'No'}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              )}

                              {/* Stage and Risk Flags Card */}
                              <div className="bg-gray-900/50 rounded-lg border border-gray-700 p-4">
                                <div className="space-y-4">
                                  {lead.lead_stage && (
                                    <div>
                                      <h4 className="font-semibold text-white mb-3 text-sm">Stage</h4>
                                      <span className={`px-3 py-1.5 rounded-md text-sm font-medium border ${getStageColor(lead.lead_stage)} capitalize inline-block`}>
                                        {lead.lead_stage}
                                      </span>
                                    </div>
                                  )}
                                  
                                  {/* Risk Flags */}
                                  {lead.risk_flags && lead.risk_flags.length > 0 && (
                                    <div>
                                      <h4 className="font-semibold text-white mb-3 text-sm">Risk Flags</h4>
                                      <div className="flex flex-wrap gap-2">
                                        {lead.risk_flags.map((flag: string, idx: number) => (
                                          <span
                                            key={idx}
                                            className="px-2.5 py-1 rounded text-xs bg-red-900/30 text-red-400 border border-red-700 break-words"
                                          >
                                            ⚠ {flag}
                                          </span>
                                        ))}
                                      </div>
                                    </div>
                                  )}
                                </div>
                              </div>
                            </div>
                          </div>

                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </Fragment>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {sortedLeads.length === 0 && (
        <div className="text-center text-gray-400 py-8">No leads found</div>
      )}
    </div>
  );
}
