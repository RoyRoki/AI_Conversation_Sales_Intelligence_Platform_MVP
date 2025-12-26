'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api';
import type { Rule } from '@/types';
import CustomDropdown from '@/components/ui/CustomDropdown';

export default function RuleConfig() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [editingRule, setEditingRule] = useState<Rule | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'block' as 'block' | 'correct' | 'flag',
    pattern: '',
    action: 'block' as 'block' | 'auto_correct' | 'flag',
    is_active: true,
  });

  useEffect(() => {
    loadRules();
  }, []);

  const loadRules = async () => {
    try {
      setLoading(true);
      const data = await apiClient.listRules();
      setRules(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load rules');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingRule(null);
    setFormData({
      name: '',
      description: '',
      type: 'block',
      pattern: '',
      action: 'block',
      is_active: true,
    });
    setShowForm(true);
  };

  const handleEdit = (rule: Rule) => {
    setEditingRule(rule);
    setFormData({
      name: rule.name,
      description: rule.description,
      type: rule.type as 'block' | 'correct' | 'flag',
      pattern: rule.pattern,
      action: rule.action as 'block' | 'auto_correct' | 'flag',
      is_active: rule.is_active,
    });
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingRule) {
        await apiClient.updateRule(editingRule.id, formData);
      } else {
        await apiClient.createRule({
          tenant_id: '', // Will be set by backend from token
          ...formData,
        });
      }
      setShowForm(false);
      setEditingRule(null);
      await loadRules();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save rule');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this rule?')) return;
    try {
      await apiClient.deleteRule(id);
      await loadRules();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete rule');
    }
  };

  const handleToggleActive = async (rule: Rule) => {
    try {
      await apiClient.updateRule(rule.id, { is_active: !rule.is_active });
      await loadRules();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update rule');
    }
  };

  if (loading) {
    return <div className="text-center py-8 text-gray-400">Loading rules...</div>;
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-white">Rules Configuration</h2>
        <button
          onClick={handleCreate}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Create Rule
        </button>
      </div>

      {showForm && (
        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">
            {editingRule ? 'Edit Rule' : 'Create Rule'}
          </h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300">Name</label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Description</label>
              <textarea
                required
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={3}
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Type</label>
              <div className="mt-1">
                <CustomDropdown
                  value={formData.type}
                  onChange={(value) =>
                    setFormData({ ...formData, type: value as 'block' | 'correct' | 'flag' })
                  }
                  options={[
                    { value: 'block', label: 'Block' },
                    { value: 'correct', label: 'Correct' },
                    { value: 'flag', label: 'Flag' },
                  ]}
                  placeholder="Select type"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Pattern</label>
              <input
                type="text"
                required
                value={formData.pattern}
                onChange={(e) => setFormData({ ...formData, pattern: e.target.value })}
                placeholder="Regex or keyword pattern"
                className="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300">Action</label>
              <div className="mt-1">
                <CustomDropdown
                  value={formData.action}
                  onChange={(value) =>
                    setFormData({
                      ...formData,
                      action: value as 'block' | 'auto_correct' | 'flag',
                    })
                  }
                  options={[
                    { value: 'block', label: 'Block' },
                    { value: 'auto_correct', label: 'Auto Correct' },
                    { value: 'flag', label: 'Flag' },
                  ]}
                  placeholder="Select action"
                />
              </div>
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="is_active"
                checked={formData.is_active}
                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-600 rounded bg-gray-700"
              />
              <label htmlFor="is_active" className="ml-2 block text-sm text-white">
                Active
              </label>
            </div>
            <div className="flex space-x-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                {editingRule ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingRule(null);
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
        <table className="min-w-full divide-y divide-gray-700">
          <thead className="bg-gray-700">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Pattern
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Action
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-gray-800 divide-y divide-gray-700">
            {rules.map((rule) => (
              <tr key={rule.id} className="hover:bg-gray-700/50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm font-medium text-white">{rule.name}</div>
                  <div className="text-sm text-gray-400">{rule.description}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{rule.type}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300 font-mono">
                  {rule.pattern}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">{rule.action}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <button
                    onClick={() => handleToggleActive(rule)}
                    className={`px-2 py-1 rounded text-xs font-medium ${
                      rule.is_active
                        ? 'bg-green-900/30 text-green-400'
                        : 'bg-gray-700 text-gray-300'
                    }`}
                  >
                    {rule.is_active ? 'Active' : 'Inactive'}
                  </button>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                  <button
                    onClick={() => handleEdit(rule)}
                    className="text-blue-400 hover:text-blue-300"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(rule.id)}
                    className="text-red-400 hover:text-red-300"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {rules.length === 0 && (
          <div className="text-center py-12 text-gray-400">No rules configured</div>
        )}
      </div>
    </div>
  );
}

