'use client';

import { Suspense } from 'react';
import GlobalChat from '@/components/GlobalChat';

function ChatContent() {
  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white mb-2">Chat with Sales</h1>
        <p className="text-gray-400">Get instant answers about our products and services</p>
      </div>

      <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6" style={{ height: '700px' }}>
        <GlobalChat />
      </div>
    </div>
  );
}

export default function CustomerChatPage() {
  return (
    <Suspense fallback={<div className="text-center text-gray-400 py-8">Loading chat...</div>}>
      <ChatContent />
    </Suspense>
  );
}


