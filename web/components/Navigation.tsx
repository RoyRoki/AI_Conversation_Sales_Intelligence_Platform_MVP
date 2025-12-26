'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { clearAuth, getUser, isAdmin, isAgent, isCustomer } from '@/lib/auth';
import { apiClient } from '@/lib/api';
import { useRouter } from 'next/navigation';
import type { User } from '@/types';
import {
  HomeIcon,
  ChatBubbleLeftRightIcon,
  ChartBarIcon,
  Cog6ToothIcon,
  BookOpenIcon,
  UserCircleIcon,
  PresentationChartLineIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

export default function Navigation() {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [admin, setAdmin] = useState(false);
  const [agent, setAgent] = useState(false);
  const [customer, setCustomer] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  useEffect(() => {
    setMounted(true);
    const currentUser = getUser();
    setUser(currentUser);
    setAdmin(isAdmin());
    setAgent(isAgent());
    setCustomer(isCustomer());
  }, []);

  const handleLogout = () => {
    apiClient.logout();
    clearAuth();
    router.push('/login');
  };

  const isActive = (path: string) => pathname === path;

  const navigationItems = customer ? [
    { name: 'Dashboard', href: '/customer', icon: HomeIcon },
    { name: 'Products', href: '/customer/products', icon: BookOpenIcon },
    { name: 'Chat', href: '/customer/chat', icon: ChatBubbleLeftRightIcon },
    { name: 'Conversations', href: '/conversations', icon: ChatBubbleLeftRightIcon },
  ] : [
    { name: 'Dashboard', href: '/', icon: HomeIcon },
    { name: 'Conversations', href: '/conversations', icon: ChatBubbleLeftRightIcon },
    ...((admin || agent) ? [{ name: 'Analytics', href: '/analytics', icon: PresentationChartLineIcon }] : []),
    ...(admin ? [
      { name: 'Rules', href: '/rules', icon: Cog6ToothIcon },
      { name: 'Knowledge Base', href: '/knowledge-base', icon: BookOpenIcon },
      { name: 'Customer Memory', href: '/memory', icon: UserCircleIcon },
      { name: 'Agent Performance', href: '/performance', icon: ChartBarIcon },
    ] : []),
  ];

  if (!mounted) {
    return (
      <div className="fixed inset-y-0 left-0 z-50 w-64 bg-gray-900 text-white">
        <div className="flex h-full flex-col">
          <div className="flex h-16 items-center justify-between px-6">
            <span className="text-xl font-bold">AI Platform</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      {/* Mobile menu button */}
      <div className="lg:hidden fixed top-4 left-4 z-50">
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="p-2 rounded-md bg-gray-900 text-white hover:bg-gray-800"
        >
          {sidebarOpen ? (
            <XMarkIcon className="h-6 w-6" />
          ) : (
            <Bars3Icon className="h-6 w-6" />
          )}
        </button>
      </div>

      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-40 w-64 bg-gray-900 text-white border-r border-gray-700 transform transition-transform duration-300 ease-in-out ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        } lg:translate-x-0`}
      >
        <div className="flex h-full flex-col">
          {/* Logo/Brand */}
          <div className="flex h-16 items-center justify-between px-6 border-b border-gray-800">
            <Link href="/" className="text-xl font-bold text-white hover:text-blue-400 transition-colors">
              AI Platform
            </Link>
            <button
              onClick={() => setSidebarOpen(false)}
              className="lg:hidden text-gray-400 hover:text-white"
            >
              <XMarkIcon className="h-6 w-6" />
            </button>
          </div>

          {/* Navigation Links */}
          <nav className="flex-1 space-y-1 px-3 py-4 overflow-y-auto">
            {navigationItems.map((item) => {
              const Icon = item.icon;
              const active = isActive(item.href);
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={() => setSidebarOpen(false)}
                  className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                    active
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                  }`}
                >
                  <Icon className="mr-3 h-5 w-5 flex-shrink-0" />
                  <span>{item.name}</span>
                </Link>
              );
            })}
          </nav>

          {/* User Info & Logout */}
          <div className="border-t border-gray-800 p-4">
            <div className="mb-3">
              <p className="text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">
                Logged in as
              </p>
              <p className="text-sm font-medium text-white truncate">{user?.email}</p>
              <p className="text-xs text-gray-400 capitalize">{user?.role}</p>
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center w-full px-3 py-2 text-sm font-medium text-red-400 rounded-lg hover:bg-gray-800 hover:text-red-300 transition-colors"
            >
              <ArrowRightOnRectangleIcon className="mr-3 h-5 w-5" />
              <span>Logout</span>
            </button>
          </div>
        </div>
      </aside>

      {/* Overlay for mobile */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-30 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </>
  );
}
