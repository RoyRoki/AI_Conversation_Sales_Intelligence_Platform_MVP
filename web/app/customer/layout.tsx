'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { isAuthenticated, isCustomer } from '@/lib/auth';
import CustomerNavigation from '@/components/customer/Navigation';
import FloatingChatButton from '@/components/FloatingChatButton';

export default function CustomerLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [mounted, setMounted] = useState(false);
  const [authorized, setAuthorized] = useState(false);

  useEffect(() => {
    setMounted(true);
    
    if (!isAuthenticated() || !isCustomer()) {
      router.push('/customer-login');
      return;
    }

    setAuthorized(true);
  }, [router]);

  // Don't render until mounted to prevent hydration mismatch
  if (!mounted) {
    return null;
  }

  if (!authorized) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <CustomerNavigation />
      <main>{children}</main>
      <FloatingChatButton />
    </div>
  );
}

