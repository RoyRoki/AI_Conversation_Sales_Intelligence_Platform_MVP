'use client';

import { useState, useEffect } from 'react';
import type { AISuggestion } from '@/types';

interface SuggestionsModalProps {
  isOpen: boolean;
  onClose: () => void;
  suggestions: AISuggestion[];
  onUseSuggestion: (suggestion: string) => void;
  onEditSuggestion?: (suggestion: string) => void;
  onRegenerate?: () => void;
  regenerating?: boolean;
  autoReplyEnabled?: boolean;
}

export default function SuggestionsModal({
  isOpen,
  onClose,
  suggestions,
  onUseSuggestion,
  onEditSuggestion,
  onRegenerate,
  regenerating = false,
  autoReplyEnabled = false,
}: SuggestionsModalProps) {
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editedText, setEditedText] = useState('');

  // Close modal on Escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;


  const handleEdit = (index: number, suggestion: AISuggestion) => {
    setEditingIndex(index);
    setEditedText(suggestion.reply);
  };

  const handleSaveEdit = (index: number) => {
    if (onEditSuggestion && editedText.trim()) {
      onEditSuggestion(editedText);
      onClose();
    }
    setEditingIndex(null);
    setEditedText('');
  };

  const handleCancelEdit = () => {
    setEditingIndex(null);
    setEditedText('');
  };

  const handleUse = (suggestion: string) => {
    onUseSuggestion(suggestion);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-gray-800 rounded-lg shadow-xl border border-gray-700 w-full max-w-4xl max-h-[85vh] flex flex-col">
          {/* Header */}
          <div className="flex items-center justify-between p-5 border-b border-gray-700">
            <div className="flex items-center gap-3">
              <h2 className="text-xl font-semibold text-white">AI Suggestions</h2>
              {autoReplyEnabled && (
                <span className="text-xs bg-green-900/30 text-green-400 px-2 py-1 rounded-full border border-green-700/50">
                  Auto-Reply ON
                </span>
              )}
              {suggestions.length > 0 && (
                <span className="text-sm text-gray-400">
                  ({suggestions.length} {suggestions.length === 1 ? 'suggestion' : 'suggestions'})
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              {regenerating && (
                <div className="flex items-center gap-2 text-sm text-gray-400">
                  <div className="animate-spin rounded-full h-4 w-4 border-2 border-blue-400 border-t-transparent"></div>
                  <span>Generating...</span>
                </div>
              )}
              {onRegenerate && (
                <button
                  onClick={onRegenerate}
                  disabled={regenerating}
                  className="px-3 py-1.5 bg-gray-700 text-gray-200 rounded-lg text-sm hover:bg-gray-600 disabled:bg-gray-800 disabled:cursor-not-allowed transition-colors font-medium flex items-center gap-1.5"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Regenerate
                </button>
              )}
              <button
                onClick={onClose}
                className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-5">
            {regenerating && suggestions.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12">
                <div className="animate-spin rounded-full h-12 w-12 border-4 border-blue-400 border-t-transparent mb-4"></div>
                <div className="text-gray-300 font-medium mb-2">Generating AI suggestions...</div>
                <div className="text-sm text-gray-400">Analyzing conversation and generating personalized responses</div>
              </div>
            ) : suggestions.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12">
                <div className="text-gray-500 text-4xl mb-4">💡</div>
                <div className="text-gray-300 font-medium mb-2">No AI suggestions available</div>
                <div className="text-sm text-gray-400 mb-4 text-center max-w-md">
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
            ) : (
              <div className="space-y-3">
                {suggestions.map((suggestion, index) => {
                  const isEditing = editingIndex === index;

                  return (
                    <div
                      key={index}
                      className="bg-gray-900/50 rounded-lg border border-gray-700 p-4 hover:border-gray-600 transition-colors"
                    >
                      {/* Suggestion Content */}
                      {isEditing ? (
                        <div className="space-y-3">
                          <textarea
                            value={editedText}
                            onChange={(e) => setEditedText(e.target.value)}
                            className="w-full px-4 py-3 bg-gray-800 border border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-white placeholder-gray-400 resize-none"
                            rows={5}
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
                        <div className="flex items-start gap-4">
                          {/* Suggestion Text */}
                          <div className="flex-1">
                            <p className="text-sm text-gray-100 leading-relaxed whitespace-pre-wrap break-words">
                              {suggestion.reply}
                            </p>
                          </div>

                          {/* Action Buttons */}
                          <div className="flex gap-2 flex-shrink-0">
                            <button
                              onClick={() => handleUse(suggestion.reply)}
                              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium transition-colors flex items-center gap-2"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                              </svg>
                              Use
                            </button>
                            {onEditSuggestion && (
                              <button
                                onClick={() => handleEdit(index, suggestion)}
                                className="px-4 py-2 bg-gray-700 text-gray-200 rounded-lg hover:bg-gray-600 text-sm font-medium transition-colors flex items-center gap-2"
                              >
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                </svg>
                                Edit
                              </button>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

