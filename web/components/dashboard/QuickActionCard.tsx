'use client';

import Link from 'next/link';
import { ReactNode } from 'react';

interface QuickActionCardProps {
  title: string;
  description: string;
  href: string;
  icon: ReactNode;
  className?: string;
}

export default function QuickActionCard({
  title,
  description,
  href,
  icon,
  className = '',
}: QuickActionCardProps) {
  return (
    <Link
      href={href}
      className={`group bg-gray-800 rounded-xl shadow-professional hover:shadow-professional-lg transition-all duration-300 p-6 block border border-gray-700 hover:border-blue-500 ${className}`}
    >
      <div className="flex items-start space-x-4">
        <div className="flex-shrink-0 p-3 bg-blue-900/50 rounded-lg text-blue-400 group-hover:bg-blue-800 group-hover:scale-110 transition-transform duration-300">
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-lg font-semibold text-white mb-2 group-hover:text-blue-400 transition-colors">
            {title}
          </h3>
          <p className="text-sm text-gray-400 line-clamp-2">{description}</p>
        </div>
        <div className="flex-shrink-0 text-gray-500 group-hover:text-blue-400 transition-colors">
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </div>
      </div>
    </Link>
  );
}

