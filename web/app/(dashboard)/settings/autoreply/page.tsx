'use client';

import { useEffect, useState } from 'react';
import ProtectedRoute from '@/components/ProtectedRoute';
import { apiClient } from '@/lib/api';
import type { AutoReplyGlobalConfig } from '@/types';

export default function AutoReplySettingsPage() {
  const [config, setConfig] = useState<AutoReplyGlobalConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getGlobalAutoReply();
      setConfig(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load auto-reply settings');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!config) return;

    try {
      setSaving(true);
      setError('');
      setSuccess('');
      await apiClient.updateGlobalAutoReply({
        enabled: config.enabled,
        confidence_threshold: config.confidence_threshold,
      });
      setSuccess('Settings saved successfully!');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <ProtectedRoute requireAdmin>
        <div className="max-w-7xl mx-auto">
          <div className="text-center text-gray-400">Loading settings...</div>
        </div>
      </ProtectedRoute>
    );
  }

  if (!config) {
    return (
      <ProtectedRoute requireAdmin>
        <div className="max-w-7xl mx-auto">
          <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
            Failed to load settings
          </div>
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute requireAdmin>
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white">Auto-Reply Settings</h1>
          <p className="mt-2 text-gray-400">Configure global auto-reply behavior for all conversations</p>
        </div>

        {error && (
          <div className="mb-4 bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
            {error}
          </div>
        )}

        {success && (
          <div className="mb-4 bg-green-900/20 border border-green-500 text-green-400 px-4 py-3 rounded">
            {success}
          </div>
        )}

        <div className="bg-gray-800 rounded-lg shadow p-6 border border-gray-700">
          <div className="space-y-6">
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={config.enabled}
                  onChange={(e) => setConfig({ ...config, enabled: e.target.checked })}
                  className="w-5 h-5 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
                />
                <span className="text-white font-medium">Enable Auto-Reply</span>
              </label>
              <p className="mt-2 text-sm text-gray-400">
                When enabled, AI will automatically send replies to customer messages that meet the confidence threshold.
              </p>
            </div>

            <div>
              <label className="block text-white font-medium mb-2">
                Confidence Threshold: {(config.confidence_threshold * 100).toFixed(0)}%
              </label>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={config.confidence_threshold}
                onChange={(e) =>
                  setConfig({ ...config, confidence_threshold: parseFloat(e.target.value) })
                }
                className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
              />
              <div className="flex justify-between text-xs text-gray-400 mt-1">
                <span>0% (Low)</span>
                <span>50%</span>
                <span>100% (High)</span>
              </div>
              <p className="mt-2 text-sm text-gray-400">
                Only suggestions with confidence above this threshold will be sent automatically.
              </p>
            </div>

            <div className="bg-gray-700/50 rounded p-4">
              <h3 className="text-white font-medium mb-2">Current Settings Preview</h3>
              <div className="space-y-1 text-sm text-gray-300">
                <p>
                  <span className="font-medium">Status:</span>{' '}
                  <span className={config.enabled ? 'text-green-400' : 'text-red-400'}>
                    {config.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </p>
                <p>
                  <span className="font-medium">Confidence Threshold:</span>{' '}
                  {(config.confidence_threshold * 100).toFixed(0)}%
                </p>
                <p className="text-gray-400 text-xs mt-2">
                  Note: Individual conversations can override these global settings.
                </p>
              </div>
            </div>

            <div className="flex space-x-4">
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {saving ? 'Saving...' : 'Save Settings'}
              </button>
              <button
                onClick={loadConfig}
                className="px-6 py-2 bg-gray-600 text-gray-200 rounded-md hover:bg-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-500"
              >
                Reset
              </button>
            </div>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}

