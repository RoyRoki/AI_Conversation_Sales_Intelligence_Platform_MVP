'use client';

import { useEffect, useState, useRef } from 'react';
import { apiClient } from '@/lib/api';
import type { Message, Conversation } from '@/types';
import { format } from 'date-fns';

interface CustomerChatProps {
  productId: string;
}

export default function CustomerChat({ productId }: CustomerChatProps) {
  const [conversation, setConversation] = useState<Conversation | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    initializeConversation();
  }, [productId]);

  useEffect(() => {
    if (conversation) {
      loadMessages();
      // Poll for new messages
      const interval = setInterval(() => {
        loadMessages();
      }, 3000);
      return () => clearInterval(interval);
    }
  }, [conversation]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const initializeConversation = async () => {
    try {
      setLoading(true);
      
      // Check localStorage for existing conversation ID
      const storageKey = `conversation_${productId}`;
      const existingConversationId = typeof window !== 'undefined' 
        ? localStorage.getItem(storageKey) 
        : null;

      // Try to load existing conversation first
      if (existingConversationId) {
        try {
          const existingConv = await apiClient.getConversation(existingConversationId);
          if (existingConv && existingConv.id) {
            setConversation(existingConv);
            // Messages are already included in the response from getConversation
            setMessages(existingConv.messages || []);
            setLoading(false);
            return;
          }
        } catch (err) {
          // Conversation doesn't exist or is invalid, create new one
          console.log('Existing conversation not found, creating new one:', err);
          if (typeof window !== 'undefined') {
            localStorage.removeItem(storageKey);
          }
        }
      }

      // Don't create conversation here - it will be created when user sends first message
    } catch (err: any) {
      console.error('Failed to initialize conversation:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadMessages = async () => {
    if (!conversation || !conversation.id) return;
    try {
      const data = await apiClient.getConversation(conversation.id);
      if (data && data.messages) {
        setMessages(data.messages);
      }
    } catch (err) {
      console.error('Failed to load messages:', err);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || sending) return;

    setSending(true);
    try {
      // Use "new" as conversation_id if no conversation exists (will create on backend)
      const conversationId = conversation?.id || 'new';
      const response = await apiClient.sendMessage(conversationId, 'customer', input, 'web');
      
      // If we just created a conversation, set it and store in localStorage
      if (!conversation && response.conversation_id) {
        const newConv = await apiClient.getConversation(response.conversation_id);
        setConversation(newConv);
        setMessages(newConv.messages || []);
        
        // Store conversation ID in localStorage
        const storageKey = productId 
          ? `conversation_${productId}` 
          : 'conversation_global';
        if (typeof window !== 'undefined' && newConv.id) {
          localStorage.setItem(storageKey, newConv.id);
        }
      }
      
      setInput('');
      // Reload messages after a short delay
      setTimeout(() => {
        if (conversation?.id || response.conversation_id) {
          loadMessages();
        }
      }, 500);
    } catch (err: any) {
      console.error('Failed to send message:', err);
    } finally {
      setSending(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center text-gray-400 py-4">Initializing chat...</div>
    );
  }

  return (
    <div className="flex flex-col h-[600px] bg-gray-800 rounded-lg border border-gray-700">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            <p className="mb-2">Start a conversation!</p>
            <p className="text-sm">Ask us anything about this product.</p>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.sender === 'agent' ? 'justify-start' : 'justify-end'}`}
            >
              <div
                className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                  message.sender === 'agent'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-700 text-gray-100'
                }`}
              >
                <p className="text-sm">{message.content}</p>
                <p
                  className={`text-xs mt-1 ${
                    message.sender === 'agent' ? 'text-blue-200' : 'text-gray-400'
                  }`}
                >
                  {format(new Date(message.timestamp), 'HH:mm')}
                </p>
              </div>
            </div>
          ))
        )}
        {sending && (
          <div className="flex justify-start">
            <div className="bg-gray-700 text-gray-100 max-w-xs lg:max-w-md px-4 py-2 rounded-lg">
              <p className="text-sm">Sending...</p>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      <form onSubmit={handleSendMessage} className="border-t border-gray-700 p-4">
        <div className="flex space-x-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type your message..."
            className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={sending}
          />
          <button
            type="submit"
            disabled={!input.trim() || sending}
            className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {sending ? 'Sending...' : 'Send'}
          </button>
        </div>
      </form>
    </div>
  );
}

