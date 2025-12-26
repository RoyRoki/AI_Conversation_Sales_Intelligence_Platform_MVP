'use client';

import { useState } from 'react';
import { ChatBubbleLeftRightIcon, XMarkIcon } from '@heroicons/react/24/outline';
import GlobalChat from './GlobalChat';

export default function FloatingChatButton() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      {/* Floating Chat Button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="fixed bottom-6 right-6 z-[9999] bg-blue-600 hover:bg-blue-700 text-white rounded-full p-4 shadow-lg hover:shadow-xl transition-all duration-300 group relative"
          style={{
            animation: 'float 3s ease-in-out infinite',
            position: 'fixed',
            bottom: '24px',
            right: '24px',
          }}
          aria-label="Open chat"
        >
          <ChatBubbleLeftRightIcon className="w-6 h-6 relative z-10" />
          {/* Pulse ring animation */}
          <span 
            className="absolute inset-0 rounded-full bg-blue-600 opacity-75"
            style={{
              animation: 'pulse-ring 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
            }}
          ></span>
        </button>
      )}

      {/* Chat Widget */}
      {isOpen && (
        <div 
          className="fixed bottom-6 right-6 z-[9999] w-full sm:w-96 h-[600px] max-h-[calc(100vh-3rem)] bg-gray-800 rounded-lg shadow-2xl border border-gray-700 flex flex-col animate-slide-up"
          style={{
            position: 'fixed',
            bottom: '24px',
            right: '24px',
          }}
        >
          {/* Header */}
          <div className="bg-blue-600 text-white p-4 rounded-t-lg flex items-center justify-between flex-shrink-0">
            <div className="flex items-center space-x-2">
              <ChatBubbleLeftRightIcon className="w-5 h-5" />
              <h3 className="font-semibold">Chat with Sales</h3>
            </div>
            <button
              onClick={() => setIsOpen(false)}
              className="hover:bg-blue-700 rounded-full p-1 transition-colors"
              aria-label="Close chat"
            >
              <XMarkIcon className="w-5 h-5" />
            </button>
          </div>

          {/* Chat Content */}
          <div className="flex-1 overflow-hidden min-h-0">
            <GlobalChat />
          </div>
        </div>
      )}

      <style jsx>{`
        .animate-slide-up {
          animation: slide-up 0.3s ease-out;
        }
      `}</style>
    </>
  );
}

