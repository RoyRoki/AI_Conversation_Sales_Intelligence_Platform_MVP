import type { User, UserRole } from '@/types';

export function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('token');
}

export function getUser(): User | null {
  if (typeof window === 'undefined') return null;
  const userStr = localStorage.getItem('user');
  if (!userStr) return null;
  try {
    return JSON.parse(userStr);
  } catch {
    return null;
  }
}

export function setAuth(token: string, user: User): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
}

export function clearAuth(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  localStorage.removeItem('customer_email');
}

export function getCustomerEmail(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('customer_email');
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}

export function hasRole(role: UserRole): boolean {
  const user = getUser();
  return user?.role === role;
}

export function isAdmin(): boolean {
  return hasRole('admin');
}

export function isAgent(): boolean {
  return hasRole('agent');
}

export function canAccessAdminFeatures(): boolean {
  return isAdmin();
}

export function canAccessAgentFeatures(): boolean {
  return isAdmin() || isAgent();
}

export function isCustomer(): boolean {
  return hasRole('customer');
}

export function customerLogin(email: string, tenantId: string): Promise<{ token: string; user: User }> {
  // This will be implemented via API client
  // For now, return a promise that will be handled by the component
  return Promise.reject('Use apiClient.customerLogin instead');
}

