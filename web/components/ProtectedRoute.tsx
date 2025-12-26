'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { isAuthenticated, getUser, canAccessAdminFeatures, canAccessAgentFeatures } from '@/lib/auth';
import type { UserRole } from '@/types';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: UserRole;
  requireAdmin?: boolean;
  requireAgent?: boolean;
}

export default function ProtectedRoute({
  children,
  requiredRole,
  requireAdmin,
  requireAgent,
}: ProtectedRouteProps) {
  const router = useRouter();
  const [mounted, setMounted] = useState(false);
  const [authorized, setAuthorized] = useState(false);

  useEffect(() => {
    setMounted(true);
    
    if (!isAuthenticated()) {
      router.push('/login');
      return;
    }

    const user = getUser();
    if (!user) {
      router.push('/login');
      return;
    }

    if (requireAdmin && !canAccessAdminFeatures()) {
      router.push('/');
      return;
    }

    if (requireAgent && !canAccessAgentFeatures()) {
      router.push('/');
      return;
    }

    if (requiredRole && user.role !== requiredRole) {
      router.push('/');
      return;
    }

    setAuthorized(true);
  }, [router, requiredRole, requireAdmin, requireAgent]);

  // Don't render until mounted to prevent hydration mismatch
  if (!mounted) {
    return null;
  }

  if (!authorized) {
    return null;
  }

  return <>{children}</>;
}

