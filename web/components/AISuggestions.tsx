'use client';

import { useState } from 'react';
import type { AISuggestion } from '@/types';

interface AISuggestionsProps {
  suggestions: AISuggestion[];
  onUseSuggestion: (suggestion: string) => void;
  onEditSuggestion?: (suggestion: string) => void;
  onRegenerate?: () => void;
  regenerating?: boolean;
  autoReplyEnabled?: boolean;
}

export default function AISuggestions({
  suggestions,
  onUseSuggestion,
  onEditSuggestion,
  onRegenerate,
  regenerating = false,
  autoReplyEnabled = false,
}: AISuggestionsProps) {
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editedText, setEditedText] = useState('');
  const [expandedReasoning, setExpandedReasoning] = useState<Set<number>>(new Set());

  const toggleReasoning = (index: number) => {
    const newExpanded = new Set(expandedReasoning);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedReasoning(newExpanded);
  };

  if (regenerating && suggestions.length === 0) {
    return (
      <div className="bg-gray-800/50 rounded-lg p-6 border border-gray-700">
        <div className="flex items-center gap-3 mb-3">
          <div className="animate-spin rounded-full h-5 w-5 border-2 border-blue-400 border-t-transparent"></div>
          <div className="text-gray-300 font-medium">Generating AI suggestions...</div>
        </div>
        <div className="text-sm text-gray-400">Analyzing conversation and generating personalized responses</div>
      </div>
    );
  }

  if (suggestions.length === 0) {
    return (
      <div className="bg-gray-800/50 rounded-lg p-6 border border-gray-700">
        <div className="text-center">
          <div className="text-gray-500 text-2xl mb-3">💡</div>
          <div className="text-gray-300 font-medium mb-2">No AI suggestions available</div>
          <div className="text-sm text-gray-400 mb-4">
            AI suggestions may be temporarily unavailable. You can still respond manually.
          </div>
          {onRegenerate && (
            <button
              onClick={onRegenerate}
              disabled={regenerating}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed transition-colors font-medium"
            >
              {regenerating ? 'Regenerating...' : 'Regenerate Suggestions'}
            </button>
          )}
        </div>
      </div>
    );
  }

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return { bg: 'bg-green-500', text: 'text-green-400', border: 'border-green-500/30' };
    if (confidence >= 0.6) return { bg: 'bg-yellow-500', text: 'text-yellow-400', border: 'border-yellow-500/30' };
    return { bg: 'bg-red-500', text: 'text-red-400', border: 'border-red-500/30' };
  };

  const getConfidenceLabel = (confidence: number) => {
    if (confidence >= 0.8) return 'High';
    if (confidence >= 0.6) return 'Medium';
    return 'Low';
  };

  const handleEdit = (index: number, suggestion: AISuggestion) => {
    setEditingIndex(index);
    setEditedText(suggestion.reply);
  };

  const handleSaveEdit = (index: number) => {
    if (onEditSuggestion && editedText.trim()) {
      onEditSuggestion(editedText);
    }
    setEditingIndex(null);
    setEditedText('');
  };

  const handleCancelEdit = () => {
    setEditingIndex(null);
    setEditedText('');
  };

  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="text-base font-semibold text-white">AI Suggestions</h3>
          {autoReplyEnabled && (
            <span className="text-xs bg-green-900/30 text-green-400 px-2 py-1 rounded-full border border-green-700/50">
              Auto-Reply ON
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {regenerating && (
            <div className="flex items-center gap-2 text-xs text-gray-400">
              <div className="animate-spin rounded-full h-3 w-3 border-2 border-blue-400 border-t-transparent"></div>
              <span>Generating...</span>
            </div>
          )}
          {onRegenerate && (
            <button
              onClick={onRegenerate}
              disabled={regenerating}
              className="px-3 py-1.5 bg-gray-700 text-gray-200 rounded-lg text-xs hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed transition-colors font-medium flex items-center gap-1.5"
            >
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {regenerating ? 'Regenerating...' : 'Regenerate'}
            </button>
          )}
        </div>
      </div>

      {/* Suggestions List */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {suggestions.map((suggestion, index) => {
          const colors = getConfidenceColor(suggestion.confidence);
          const isEditing = editingIndex === index;
          const isReasoningExpanded = expandedReasoning.has(index);

          return (
            <div
              key={index}
              className="bg-gray-800/50 rounded-lg border border-gray-700 p-4 hover:border-gray-600 transition-colors flex flex-col"
            >
              {/* Confidence Badge */}
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className={`text-xs font-medium px-2 py-1 rounded-full ${colors.text} bg-gray-900/50 border ${colors.border}`}>
                    {getConfidenceLabel(suggestion.confidence)} Confidence
                  </span>
                  <span className="text-xs text-gray-500">
                    {(suggestion.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                <div className="w-16 h-1.5 bg-gray-700 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${colors.bg} transition-all duration-300`}
                    style={{ width: `${suggestion.confidence * 100}%` }}
                  />
                </div>
              </div>

              {/* Suggestion Content */}
              {isEditing ? (
                <div className="space-y-3">
                  <textarea
                    value={editedText}
                    onChange={(e) => setEditedText(e.target.value)}
                    className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white placeholder-gray-400 resize-none"
                    rows={4}
                    placeholder="Edit the suggestion..."
                  />
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleSaveEdit(index)}
                      className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium transition-colors"
                    >
                      Save & Send
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="px-4 py-2 bg-gray-700 text-gray-200 rounded-lg hover:bg-gray-600 text-sm font-medium transition-colors"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <>
                  <p className="text-sm text-gray-100 leading-relaxed mb-3 whitespace-pre-wrap break-words">
                    {suggestion.reply}
                  </p>

                  {/* Reasoning (Collapsible) */}
                  {suggestion.reasoning && (
                    <div className="mb-3">
                      <button
                        onClick={() => toggleReasoning(index)}
                        className="flex items-center gap-2 text-xs text-gray-400 hover:text-gray-300 transition-colors w-full text-left"
                      >
                        <svg
                          className={`w-4 h-4 transition-transform ${isReasoningExpanded ? 'rotate-180' : ''}`}
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                        <span>{isReasoningExpanded ? 'Hide' : 'Show'} reasoning</span>
                      </button>
                      {isReasoningExpanded && (
                        <div className="mt-2 p-3 bg-gray-900/50 rounded-lg border border-gray-700">
                          <p className="text-xs text-gray-400 leading-relaxed whitespace-pre-wrap break-words">
                            {suggestion.reasoning}
                          </p>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Action Buttons */}
                  <div className="flex gap-2 pt-3 mt-auto border-t border-gray-700">
                    <button
                      onClick={() => onUseSuggestion(suggestion.reply)}
                      className="flex-1 px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-xs font-medium transition-colors flex items-center justify-center gap-1.5"
                    >
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Use
                    </button>
                    {onEditSuggestion && (
                      <button
                        onClick={() => handleEdit(index, suggestion)}
                        className="px-3 py-1.5 bg-gray-700 text-gray-200 rounded-lg hover:bg-gray-600 text-xs font-medium transition-colors flex items-center gap-1.5"
                      >
                        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                        Edit
                      </button>
                    )}
                  </div>
                </>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

