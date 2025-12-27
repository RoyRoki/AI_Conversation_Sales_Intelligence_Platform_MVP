'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { canAccessAgentFeatures } from '@/lib/auth';
import ChatInterface from '@/components/ChatInterface';
import SuggestionsModal from '@/components/SuggestionsModal';
import type { ConversationWithMetadata, Message, SuggestionsResponse, WinProbability, ChurnRisk, EffectiveAutoReplyConfig } from '@/types';

export default function ConversationDetailPage() {
  const params = useParams();
  const router = useRouter();
  const conversationId = params.id as string;
  const [conversation, setConversation] = useState<ConversationWithMetadata | null>(null);
  const [suggestions, setSuggestions] = useState<SuggestionsResponse | null>({ suggestions: [], contextUsed: false });
  const [winProbability, setWinProbability] = useState<WinProbability | null>(null);
  const [churnRisk, setChurnRisk] = useState<ChurnRisk | null>(null);
  const [autoReplyConfig, setAutoReplyConfig] = useState<EffectiveAutoReplyConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [sendingMessage, setSendingMessage] = useState(false);
  const [updatingAutoReply, setUpdatingAutoReply] = useState(false);
  const [regeneratingSuggestions, setRegeneratingSuggestions] = useState(false);
  const [showSuggestionsModal, setShowSuggestionsModal] = useState(false);
  const lastCustomerMessageRef = useRef<string | null>(null);
  const loadingSuggestionsRef = useRef<boolean>(false);
  const canAccessAgent = canAccessAgentFeatures();

  useEffect(() => {
    if (conversationId) {
      loadConversation();
      loadAnalytics();
      if (canAccessAgent) {
        loadSuggestions();
        loadAutoReplyConfig();
      }
    }
  }, [conversationId, canAccessAgent]);

  // Detect new customer messages and auto-regenerate suggestions
  useEffect(() => {
    if (!canAccessAgent || !conversation?.messages) return;
    
    const messages = conversation.messages || [];
    const lastCustomerMessage = messages
      .filter((msg: Message) => msg.sender === 'customer')
      .sort((a: Message, b: Message) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())[0];
    
    if (lastCustomerMessage && lastCustomerMessage.id !== lastCustomerMessageRef.current) {
      lastCustomerMessageRef.current = lastCustomerMessage.id;
      // New customer message detected - regenerate suggestions with loading state
      setRegeneratingSuggestions(true);
      loadSuggestions(); // loadSuggestions now handles clearing regenerating state in finally block
    }
  }, [conversation?.messages, canAccessAgent]);

  // Poll for updates every 5 seconds (conversation and auto-reply config only)
  // Suggestions are loaded on-demand when modal opens or when new customer message is detected
  useEffect(() => {
    if (!conversationId) return;
    const interval = setInterval(() => {
      loadConversation();
      if (canAccessAgent) {
        loadAutoReplyConfig();
        // Only poll suggestions if modal is open
        if (showSuggestionsModal) {
          loadSuggestions();
        }
      }
    }, 5000);
    return () => clearInterval(interval);
  }, [conversationId, canAccessAgent, showSuggestionsModal]);

  const loadConversation = async () => {
    try {
      const data = await apiClient.getConversation(conversationId);
      setConversation(data);
      
      // Initialize last customer message ref on first load
      if (data?.messages && !lastCustomerMessageRef.current) {
        const messages = data.messages || [];
        const lastCustomerMessage = messages
          .filter((msg: Message) => msg.sender === 'customer')
          .sort((a: Message, b: Message) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())[0];
        if (lastCustomerMessage) {
          lastCustomerMessageRef.current = lastCustomerMessage.id;
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load conversation');
    } finally {
      setLoading(false);
    }
  };

  const loadSuggestions = async (forceRegenerate: boolean = false) => {
    // Prevent concurrent requests
    if (loadingSuggestionsRef.current) {
      return;
    }
    
    loadingSuggestionsRef.current = true;
    try {
      const data = await apiClient.getSuggestions(conversationId, forceRegenerate);
      setSuggestions(data);
    } catch (err: any) {
      console.error('Failed to load suggestions:', err);
      // Always set empty suggestions on error to show graceful degradation message
      setSuggestions({ suggestions: [], contextUsed: false });
    } finally {
      // Ensure regenerating state is cleared
      setRegeneratingSuggestions(false);
      loadingSuggestionsRef.current = false;
    }
  };

  const loadAnalytics = async () => {
    try {
      const [winProb, churn] = await Promise.all([
        apiClient.getWinProbability(conversationId).catch(() => null),
        apiClient.getChurnRisk(conversationId).catch(() => null),
      ]);
      
      // Transform backend WinProbability format to frontend format
      if (winProb) {
        const transformedWinProb: WinProbability = {
          probability: (winProb as any).probability ?? 0,
          factors: (winProb as any).factors || [],
          trend: (winProb as any).trend || 'stable',
        };
        setWinProbability(transformedWinProb);
      } else {
        setWinProbability(null);
      }
      
      // Transform backend ChurnRisk format to frontend format
      if (churn) {
        const riskScore = (churn as any).risk_score ?? (churn as any).score ?? 0;
        const transformedChurn: ChurnRisk = {
          risk_level: (churn as any).risk_level || 
            (riskScore >= 0.7 ? 'high' : 
             riskScore >= 0.4 ? 'medium' : 'low'),
          score: riskScore,
          factors: (churn as any).factors || [],
        };
        setChurnRisk(transformedChurn);
      } else {
        setChurnRisk(null);
      }
    } catch (err) {
      console.error('Failed to load analytics:', err);
    }
  };

  const loadAutoReplyConfig = async () => {
    try {
      const data = await apiClient.getConversationAutoReply(conversationId);
      console.log('Loaded auto-reply config:', data);
      if (data && data.effective) {
        setAutoReplyConfig(data.effective);
      } else {
        console.warn('No effective config in response:', data);
      }
    } catch (err) {
      console.error('Failed to load auto-reply config:', err);
    }
  };

  const handleToggleAutoReply = async (enabled: boolean) => {
    try {
      setUpdatingAutoReply(true);
      setError(''); // Clear any previous errors
      await apiClient.updateConversationAutoReply(conversationId, { enabled });
      
      // Always reload config from backend to ensure UI shows actual state
      await loadAutoReplyConfig();
    } catch (err: any) {
      console.error('Failed to update auto-reply:', err);
      const errorMessage = err.response?.data?.error || err.message || 'Failed to update auto-reply';
      setError(errorMessage);
      // Reload config on error to ensure UI reflects actual state
      await loadAutoReplyConfig();
    } finally {
      setUpdatingAutoReply(false);
    }
  };

  const handleSendMessage = async (message: string) => {
    if (!conversation) return;
    setSendingMessage(true);
    try {
      await apiClient.sendMessage(conversationId, 'agent', message);
      await loadConversation();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to send message');
      throw err;
    } finally {
      setSendingMessage(false);
    }
  };

  const handleUseSuggestion = async (suggestion: string) => {
    await handleSendMessage(suggestion);
    setShowSuggestionsModal(false);
  };

  const handleEditSuggestion = async (suggestion: string) => {
    await handleSendMessage(suggestion);
    setShowSuggestionsModal(false);
  };

  const handleRegenerateSuggestions = async () => {
    try {
      setRegeneratingSuggestions(true);
      await loadSuggestions(true); // Pass true to force regeneration
    } catch (err) {
      console.error('Failed to regenerate suggestions:', err);
    } finally {
      setRegeneratingSuggestions(false);
    }
  };

  // Load suggestions when modal opens (if not already loaded)
  useEffect(() => {
    if (showSuggestionsModal && canAccessAgent) {
      // Only load if we don't have suggestions yet
      const hasSuggestions = suggestions?.suggestions && suggestions.suggestions.length > 0;
      if (!hasSuggestions && !loadingSuggestionsRef.current) {
        loadSuggestions();
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showSuggestionsModal, canAccessAgent]);

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="text-center">Loading conversation...</div>
      </div>
    );
  }

  if (error && !conversation) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      </div>
    );
  }

  if (!conversation) {
    return null;
  }

  const messages = conversation.messages || [];

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-6">
        <button
          onClick={() => router.back()}
          className="text-blue-400 hover:text-blue-300 mb-4"
        >
          ← Back to conversations
        </button>
        <h1 className="text-3xl font-bold text-white">Conversation {conversation.id}</h1>
        <p className="mt-2 text-gray-400">
          Status: <span className="font-medium">{conversation.status}</span>
        </p>
      </div>

      {error && (
        <div className="mb-4 bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Main Chat Area */}
        <div className="lg:col-span-3 flex flex-col">
          <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 flex-1 min-h-[600px] flex flex-col">
            <ChatInterface
              conversationId={conversationId}
              messages={messages}
              onSendMessage={handleSendMessage}
              isLoading={sendingMessage}
              autoReplyConfig={canAccessAgent ? autoReplyConfig : undefined}
              onToggleAutoReply={canAccessAgent ? handleToggleAutoReply : undefined}
              updatingAutoReply={updatingAutoReply}
              onShowSuggestions={canAccessAgent ? () => setShowSuggestionsModal(true) : undefined}
              suggestionsCount={canAccessAgent ? suggestions?.suggestions?.length || 0 : 0}
            />
          </div>

          {/* AI Suggestions Modal */}
          {canAccessAgent && (
            <SuggestionsModal
              isOpen={showSuggestionsModal}
              onClose={() => setShowSuggestionsModal(false)}
              suggestions={suggestions?.suggestions || []}
              onUseSuggestion={handleUseSuggestion}
              onEditSuggestion={handleEditSuggestion}
              onRegenerate={handleRegenerateSuggestions}
              regenerating={regeneratingSuggestions}
              autoReplyEnabled={autoReplyConfig?.enabled || false}
            />
          )}

          {/* Conversation Insights */}
          {conversation.metadata && (
            <div className="mt-6 bg-gray-800/50 rounded-lg shadow-lg p-5 border border-gray-700">
              <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
                <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                Conversation Insights
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 bg-gray-900/50 rounded-lg border border-gray-700/50">
                  <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">Intent</p>
                  <p className="text-base font-semibold text-white mb-1">{conversation.metadata.intent || 'N/A'}</p>
                  <p className="text-xs text-gray-500">
                    Score: {conversation.metadata.intent_score ? (conversation.metadata.intent_score * 100).toFixed(0) : '0'}%
                  </p>
                </div>
                <div className="p-3 bg-gray-900/50 rounded-lg border border-gray-700/50">
                  <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">Sentiment</p>
                  <p className="text-base font-semibold capitalize text-white mb-1">{conversation.metadata.sentiment || 'N/A'}</p>
                  <p className="text-xs text-gray-500">
                    Score: {conversation.metadata.sentiment_score ? (conversation.metadata.sentiment_score * 100).toFixed(0) : '0'}%
                  </p>
                </div>
                {conversation.metadata.emotions && conversation.metadata.emotions.length > 0 && (
                  <div className="p-3 bg-gray-900/50 rounded-lg border border-gray-700/50 col-span-2">
                    <p className="text-xs text-gray-400 uppercase tracking-wide mb-2">Emotions</p>
                    <div className="flex flex-wrap gap-2">
                      {conversation.metadata.emotions.map((emotion, idx) => (
                        <span
                          key={idx}
                          className="px-2.5 py-1 bg-blue-900/30 text-blue-300 rounded-md text-xs border border-blue-700/30"
                        >
                          {emotion}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                {conversation.metadata.objections && conversation.metadata.objections.length > 0 && (
                  <div className="p-3 bg-gray-900/50 rounded-lg border border-gray-700/50 col-span-2">
                    <p className="text-xs text-gray-400 uppercase tracking-wide mb-2">Objections</p>
                    <div className="flex flex-wrap gap-2">
                      {conversation.metadata.objections.map((objection, idx) => (
                        <span
                          key={idx}
                          className="px-2.5 py-1 bg-red-900/30 text-red-300 rounded-md text-xs border border-red-700/30"
                        >
                          {objection}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar for Analytics */}
        <div className="space-y-6 flex flex-col">

          {winProbability && (
            <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
              <h3 className="text-lg font-semibold text-white mb-4">Win Probability</h3>
              <div className="text-3xl font-bold text-blue-400 mb-2">
                {typeof winProbability.probability === 'number' 
                  ? (winProbability.probability * 100).toFixed(0) 
                  : '0'}%
              </div>
              {winProbability.trend && (
                <p className="text-sm text-gray-400 capitalize mb-4">
                  Trend: {winProbability.trend}
                </p>
              )}
              <div className="space-y-1">
                <p className="text-xs font-medium text-gray-300">Factors:</p>
                {winProbability.factors && winProbability.factors.length > 0 ? (
                  winProbability.factors.map((factor, idx) => (
                    <p key={idx} className="text-xs text-gray-400">
                      • {factor}
                    </p>
                  ))
                ) : (
                  <p className="text-xs text-gray-500">No factors available</p>
                )}
              </div>
            </div>
          )}

          {churnRisk && (
            <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
              <h3 className="text-lg font-semibold text-white mb-4">Churn Risk</h3>
              <div
                className={`text-3xl font-bold mb-2 ${
                  churnRisk.risk_level === 'high'
                    ? 'text-red-400'
                    : churnRisk.risk_level === 'medium'
                    ? 'text-yellow-400'
                    : 'text-green-400'
                }`}
              >
                {churnRisk.risk_level ? churnRisk.risk_level.toUpperCase() : 'UNKNOWN'}
              </div>
              <p className="text-sm text-gray-400 mb-4">
                Score: {typeof churnRisk.score === 'number' ? (churnRisk.score * 100).toFixed(0) : '0'}%
              </p>
              <div className="space-y-1">
                <p className="text-xs font-medium text-gray-300">Factors:</p>
                {churnRisk.factors && churnRisk.factors.length > 0 ? (
                  churnRisk.factors.map((factor, idx) => (
                    <p key={idx} className="text-xs text-gray-400">
                      • {factor}
                    </p>
                  ))
                ) : (
                  <p className="text-xs text-gray-500">No factors available</p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

