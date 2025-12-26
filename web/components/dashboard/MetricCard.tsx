'use client';

import { ReactNode } from 'react';

interface MetricCardProps {
  title: string;
  value: string | number;
  icon: ReactNode;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  gradient?: 'blue' | 'blue-light' | 'blue-dark';
  className?: string;
}

export default function MetricCard({
  title,
  value,
  icon,
  trend,
  trendValue,
  gradient = 'blue',
  className = '',
}: MetricCardProps) {
  const gradientClasses = {
    blue: 'gradient-blue',
    'blue-light': 'gradient-blue-light',
    'blue-dark': 'gradient-blue-dark',
  };

  const trendColors = {
    up: 'text-green-500',
    down: 'text-red-500',
    neutral: 'text-gray-500',
  };

  const trendIcons = {
    up: '↑',
    down: '↓',
    neutral: '→',
  };

  return (
    <div
      className={`bg-gray-800 rounded-xl shadow-professional hover:shadow-professional-lg transition-all duration-300 overflow-hidden border border-gray-700 ${className}`}
    >
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <div className={`p-3 rounded-lg ${gradientClasses[gradient]} text-white`}>
            {icon}
          </div>
          {trend && trendValue && (
            <div className={`flex items-center space-x-1 text-sm font-medium ${trendColors[trend]}`}>
              <span>{trendIcons[trend]}</span>
              <span>{trendValue}</span>
            </div>
          )}
        </div>
        <h3 className="text-sm font-medium text-gray-400 mb-2">{title}</h3>
        <p className="text-3xl font-bold text-white">{value}</p>
      </div>
    </div>
  );
}

