'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { getUser } from '@/lib/auth';
import type { Product } from '@/types';

export default function ProductKB() {
  const [entries, setEntries] = useState<Product[]>([]);
  const [filteredEntries, setFilteredEntries] = useState<Product[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editingEntry, setEditingEntry] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: '',
    price: '',
    price_currency: 'INR',
    features: '',
    limitations: '',
    target_audience: '',
    common_questions: '',
  });

  useEffect(() => {
    loadProducts();
  }, []);

  useEffect(() => {
    // Filter entries based on search query
    if (!searchQuery.trim()) {
      setFilteredEntries(entries);
    } else {
      const query = searchQuery.toLowerCase();
      setFilteredEntries(
        entries.filter(
          (entry) =>
            entry.name.toLowerCase().includes(query) ||
            entry.description.toLowerCase().includes(query) ||
            entry.category.toLowerCase().includes(query) ||
            entry.features.some((f) => f.toLowerCase().includes(query))
        )
      );
    }
  }, [searchQuery, entries]);

  const loadProducts = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Check if user is logged in
      const user = getUser();
      if (!user) {
        setError('Please log in to view products');
        setLoading(false);
        return;
      }
      
      const products = await apiClient.listProducts();
      setEntries(products);
      setFilteredEntries(products);
    } catch (err: any) {
      console.error('Failed to load products - Full error:', err);
      console.error('Error details:', {
        message: err.message,
        code: err.code,
        response: err.response,
        responseData: err.response?.data,
        responseStatus: err.response?.status,
        responseHeaders: err.response?.headers,
        request: err.request,
        config: err.config,
      });
      
      let errorMessage = 'Failed to load products';
      
      if (err.code === 'ECONNREFUSED' || err.message?.includes('Network Error')) {
        errorMessage = 'Cannot connect to server. Please ensure the server is running on port 8080.';
      } else if (err.response?.status === 401) {
        errorMessage = 'Authentication required. Please log in again.';
        if (typeof window !== 'undefined') {
          setTimeout(() => {
            window.location.href = '/login';
          }, 2000);
        }
      } else if (err.response?.status === 400) {
        errorMessage = err.response?.data?.error || 'Invalid request. Please log out and log in again.';
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error;
      } else if (err.message) {
        errorMessage = err.message;
      }
      
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingEntry(null);
    setFormData({
      name: '',
      description: '',
      category: '',
      price: '',
      price_currency: 'INR',
      features: '',
      limitations: '',
      target_audience: '',
      common_questions: '',
    });
    setShowForm(true);
  };

  const handleEdit = (entry: Product) => {
    setEditingEntry(entry);
    setFormData({
      name: entry.name,
      description: entry.description,
      category: entry.category || '',
      price: entry.price?.toString() || '',
      price_currency: entry.price_currency || 'INR',
      features: entry.features.join(', '),
      limitations: entry.limitations.join(', '),
      target_audience: entry.target_audience || '',
      common_questions: entry.common_questions.join(', '),
    });
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setError('');
      const productData = {
        name: formData.name,
        description: formData.description,
        category: formData.category,
        price: parseFloat(formData.price) || 0,
        price_currency: formData.price_currency,
        features: formData.features.split(',').map((f) => f.trim()).filter(Boolean),
        limitations: formData.limitations.split(',').map((l) => l.trim()).filter(Boolean),
        target_audience: formData.target_audience,
        common_questions: formData.common_questions.split(',').map((q) => q.trim()).filter(Boolean),
      };

      if (editingEntry) {
        await apiClient.updateProduct(editingEntry.id, productData);
      } else {
        await apiClient.createProduct(productData);
      }

      setShowForm(false);
      setEditingEntry(null);
      await loadProducts();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save product');
      console.error('Failed to save product:', err);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this entry?')) {
      return;
    }

    try {
      setError('');
      await apiClient.deleteProduct(id);
      await loadProducts();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete product');
      console.error('Failed to delete product:', err);
    }
  };

  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setError('');
      const text = await file.text();
      const data = JSON.parse(text);
      const products = Array.isArray(data) ? data : [data];

      // Create each product via API
      for (const product of products) {
        try {
          await apiClient.createProduct({
            name: product.name,
            description: product.description,
            category: product.category || '',
            price: product.price || 0,
            price_currency: product.price_currency || 'INR',
            features: product.features || [],
            limitations: product.limitations || [],
            target_audience: product.target_audience || '',
            common_questions: product.common_questions || [],
          });
        } catch (err) {
          console.error('Failed to import product:', product.name, err);
        }
      }

      await loadProducts();
      alert(`Successfully imported ${products.length} product(s)`);
    } catch (err) {
      setError('Failed to import file. Please check the format.');
      console.error('Import error:', err);
    }
  };

  const handleExport = () => {
    const dataStr = JSON.stringify(entries, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'product-kb.json';
    link.click();
  };

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-md">
          {error}
        </div>
      )}

      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-white">Product Knowledge Base</h2>
        <div className="flex space-x-2">
          <label className="px-4 py-2 bg-gray-700 text-gray-200 rounded-md hover:bg-gray-600 cursor-pointer">
            Import JSON/CSV
            <input
              type="file"
              accept=".json,.csv"
              onChange={handleImport}
              className="hidden"
            />
          </label>
          <button
            onClick={handleExport}
            disabled={entries.length === 0}
            className="px-4 py-2 bg-gray-700 text-gray-200 rounded-md hover:bg-gray-600 disabled:opacity-50"
          >
            Export
          </button>
          <button
            onClick={handleCreate}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Add Entry
          </button>
        </div>
      </div>

      {showForm && (
        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">
            {editingEntry ? 'Edit Entry' : 'Create Entry'}
          </h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-300">Name *</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-300">Category *</label>
                <input
                  type="text"
                  required
                  value={formData.category}
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                  className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Description *</label>
              <textarea
                required
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={3}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-300">Price *</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  value={formData.price}
                  onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                  className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-300">Currency</label>
                <select
                  value={formData.price_currency}
                  onChange={(e) => setFormData({ ...formData, price_currency: e.target.value })}
                  className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="INR">INR</option>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Features (comma-separated)</label>
              <input
                type="text"
                value={formData.features}
                onChange={(e) => setFormData({ ...formData, features: e.target.value })}
                placeholder="Feature 1, Feature 2, Feature 3"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Limitations (comma-separated)</label>
              <input
                type="text"
                value={formData.limitations}
                onChange={(e) => setFormData({ ...formData, limitations: e.target.value })}
                placeholder="Limitation 1, Limitation 2"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Target Audience</label>
              <input
                type="text"
                value={formData.target_audience}
                onChange={(e) => setFormData({ ...formData, target_audience: e.target.value })}
                placeholder="e.g., Small businesses, Enterprise customers"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Common Questions (comma-separated)</label>
              <input
                type="text"
                value={formData.common_questions}
                onChange={(e) => setFormData({ ...formData, common_questions: e.target.value })}
                placeholder="Question 1, Question 2"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex space-x-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingEntry ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingEntry(null);
                  setError('');
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
        <div className="px-4 py-3 bg-gray-700 border-b border-gray-600">
          <input
            type="text"
            placeholder="Search entries..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-3 py-2 bg-gray-600 border border-gray-500 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        {loading ? (
          <div className="text-center py-12 text-gray-400">Loading products...</div>
        ) : (
          <div className="divide-y divide-gray-700">
            {filteredEntries.map((entry) => (
              <div key={entry.id} className="p-4 hover:bg-gray-700/50">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-white">{entry.name}</h3>
                    <p className="text-sm text-gray-400 mt-1">{entry.description}</p>
                    <div className="mt-2 flex items-center space-x-4 text-sm text-gray-300">
                      <span>Category: {entry.category}</span>
                      {entry.price > 0 && (
                        <span>
                          Price: {entry.price_currency} {entry.price.toFixed(2)}
                        </span>
                      )}
                      {entry.target_audience && (
                        <span>Target: {entry.target_audience}</span>
                      )}
                    </div>
                    {entry.features.length > 0 && (
                      <div className="mt-2">
                        <p className="text-sm font-medium text-gray-300">Features:</p>
                        <ul className="list-disc list-inside text-sm text-gray-400">
                          {entry.features.map((feature, idx) => (
                            <li key={idx}>{feature}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                    {entry.limitations.length > 0 && (
                      <div className="mt-2">
                        <p className="text-sm font-medium text-gray-300">Limitations:</p>
                        <ul className="list-disc list-inside text-sm text-gray-400">
                          {entry.limitations.map((limitation, idx) => (
                            <li key={idx}>{limitation}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                    {entry.common_questions.length > 0 && (
                      <div className="mt-2">
                        <p className="text-sm font-medium text-gray-300">Common Questions:</p>
                        <ul className="list-disc list-inside text-sm text-gray-400">
                          {entry.common_questions.map((question, idx) => (
                            <li key={idx}>{question}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => handleEdit(entry)}
                      className="px-3 py-1 text-sm text-blue-400 hover:text-blue-300"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(entry.id)}
                      className="px-3 py-1 text-sm text-red-400 hover:text-red-300"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
        {!loading && filteredEntries.length === 0 && (
          <div className="text-center py-12 text-gray-400">
            {searchQuery ? 'No products match your search' : 'No entries in knowledge base'}
          </div>
        )}
      </div>
    </div>
  );
}
