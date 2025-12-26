'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api';
import type { CustomerMemory } from '@/types';
import CustomDropdown from '@/components/ui/CustomDropdown';

export default function CustomerMemoryComponent() {
  const [memories, setMemories] = useState<CustomerMemory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingMemory, setEditingMemory] = useState<CustomerMemory | null>(null);
  const [formData, setFormData] = useState({
    customer_id: '',
    preferred_language: '',
    pricing_sensitivity: 'medium' as 'high' | 'medium' | 'low',
    product_interests: '',
    past_objections: '',
  });

  useEffect(() => {
    loadMemories();
  }, []);

  const loadMemories = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await apiClient.listMemories(100, 0);
      setMemories(response.memories || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load customer memories');
      console.error('Failed to load memories:', err);
      setMemories([]); // Ensure memories is always an array
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingMemory(null);
    setFormData({
      customer_id: '',
      preferred_language: '',
      pricing_sensitivity: 'medium',
      product_interests: '',
      past_objections: '',
    });
    setShowForm(true);
  };

  const handleEdit = (memory: CustomerMemory) => {
    setEditingMemory(memory);
    setFormData({
      customer_id: memory.customer_id,
      preferred_language: memory.preferred_language,
      pricing_sensitivity: memory.pricing_sensitivity,
      product_interests: memory.product_interests.join(', '),
      past_objections: memory.past_objections.join(', '),
    });
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setError('');
      if (editingMemory) {
        // Update existing memory
        await apiClient.updateMemory(editingMemory.id, {
          preferred_language: formData.preferred_language,
          pricing_sensitivity: formData.pricing_sensitivity,
          product_interests: formData.product_interests.split(',').map((s) => s.trim()).filter(Boolean),
          past_objections: formData.past_objections.split(',').map((s) => s.trim()).filter(Boolean),
        });
      } else {
        // Create new memory
        await apiClient.createMemory({
          customer_id: formData.customer_id,
          preferred_language: formData.preferred_language,
          pricing_sensitivity: formData.pricing_sensitivity,
          product_interests: formData.product_interests.split(',').map((s) => s.trim()).filter(Boolean),
          past_objections: formData.past_objections.split(',').map((s) => s.trim()).filter(Boolean),
        });
      }
      setShowForm(false);
      setEditingMemory(null);
      await loadMemories();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save customer memory');
      console.error('Failed to save memory:', err);
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm('Are you sure you want to delete this memory?')) {
      try {
        setError('');
        await apiClient.deleteMemory(id);
        await loadMemories();
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to delete customer memory');
        console.error('Failed to delete memory:', err);
      }
    }
  };

  const filteredMemories = (memories || []).filter((memory) => {
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      return (
        memory.customer_id.toLowerCase().includes(term) ||
        memory.preferred_language.toLowerCase().includes(term) ||
        (memory.product_interests || []).some((interest) => interest.toLowerCase().includes(term))
      );
    }
    return true;
  });

  if (loading) {
    return <div className="text-center py-8 text-gray-400">Loading customer memories...</div>;
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-white">Customer Memory</h2>
        <button
          onClick={handleCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Add Memory
        </button>
      </div>

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search by customer ID, language, or interests..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {showForm && (
        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">
            {editingMemory ? 'Edit Customer Memory' : 'Create Customer Memory'}
          </h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300">Customer ID</label>
              <input
                type="text"
                required
                value={formData.customer_id}
                onChange={(e) => setFormData({ ...formData, customer_id: e.target.value })}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Preferred Language</label>
              <input
                type="text"
                required
                value={formData.preferred_language}
                onChange={(e) => setFormData({ ...formData, preferred_language: e.target.value })}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Pricing Sensitivity</label>
              <div className="mt-1">
                <CustomDropdown
                  value={formData.pricing_sensitivity}
                  onChange={(value) =>
                    setFormData({
                      ...formData,
                      pricing_sensitivity: value as 'high' | 'medium' | 'low',
                    })
                  }
                  options={[
                    { value: 'high', label: 'High' },
                    { value: 'medium', label: 'Medium' },
                    { value: 'low', label: 'Low' },
                  ]}
                  placeholder="Select sensitivity"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">
                Product Interests (comma-separated)
              </label>
              <input
                type="text"
                value={formData.product_interests}
                onChange={(e) => setFormData({ ...formData, product_interests: e.target.value })}
                placeholder="Product 1, Product 2, Product 3"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">
                Past Objections (comma-separated)
              </label>
              <input
                type="text"
                value={formData.past_objections}
                onChange={(e) => setFormData({ ...formData, past_objections: e.target.value })}
                placeholder="Objection 1, Objection 2, Objection 3"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex space-x-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingMemory ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingMemory(null);
                }}
                className="px-4 py-2 bg-gray-600 text-gray-200 rounded-md hover:bg-gray-500"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="bg-gray-800 rounded-lg shadow overflow-hidden border border-gray-700">
        <div className="divide-y divide-gray-700">
          {filteredMemories.map((memory) => (
            <div key={memory.id} className="p-6 hover:bg-gray-700/50">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-white">
                    Customer: {memory.customer_id}
                  </h3>
                  <div className="mt-2 grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="font-medium text-gray-300">Preferred Language:</span>
                      <span className="ml-2 text-gray-400">{memory.preferred_language}</span>
                    </div>
                    <div>
                      <span className="font-medium text-gray-300">Pricing Sensitivity:</span>
                      <span className="ml-2 text-gray-400 capitalize">
                        {memory.pricing_sensitivity}
                      </span>
                    </div>
                  </div>
                  {(memory.product_interests || []).length > 0 && (
                    <div className="mt-2">
                      <span className="text-sm font-medium text-gray-300">Product Interests:</span>
                      <div className="mt-1 flex flex-wrap gap-2">
                        {(memory.product_interests || []).map((interest, idx) => (
                          <span
                            key={idx}
                            className="px-2 py-1 bg-blue-900/30 text-blue-400 rounded text-xs"
                          >
                            {interest}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                  {(memory.past_objections || []).length > 0 && (
                    <div className="mt-2">
                      <span className="text-sm font-medium text-gray-300">Past Objections:</span>
                      <div className="mt-1 flex flex-wrap gap-2">
                        {(memory.past_objections || []).map((objection, idx) => (
                          <span
                            key={idx}
                            className="px-2 py-1 bg-red-900/30 text-red-400 rounded text-xs"
                          >
                            {objection}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => handleEdit(memory)}
                    className="px-3 py-1 text-sm text-blue-400 hover:text-blue-300"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(memory.id)}
                    className="px-3 py-1 text-sm text-red-400 hover:text-red-300"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
        {filteredMemories.length === 0 && (
          <div className="text-center py-12 text-gray-400">No customer memories found</div>
        )}
      </div>
    </div>
  );
}
