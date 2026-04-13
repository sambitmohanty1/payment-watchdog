import { renderHook, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './useAuth';
import { onAuthStateChanged, signOut, getIdToken, getIdTokenResult } from 'firebase/auth';
import api from '@/lib/api';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock Firebase
vi.mock('firebase/auth', () => ({
  getAuth: vi.fn(),
  onAuthStateChanged: vi.fn(),
  signOut: vi.fn(),
  getIdToken: vi.fn().mockResolvedValue('mock-token'),
  getIdTokenResult: vi.fn().mockResolvedValue({
    claims: { tenant_id: 'test-tenant-123' }
  }),
}));

vi.mock('@/lib/firebase', () => ({
  auth: {},
}));

// Mock API
vi.mock('@/lib/api', () => ({
  default: {
    setToken: vi.fn(),
    setCompanyId: vi.fn(),
  }
}));

describe('useAuth Hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with loading state', () => {
    (onAuthStateChanged as any).mockReturnValue(() => {});
    
    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider
    });

    expect(result.current.loading).toBe(true);
    expect(result.current.user).toBe(null);
  });

  it('should handle authenticated state and map tenant_id', async () => {
    const mockUser = { email: 'test@example.com', uid: '123' };
    
    // Simulate auth state change
    (onAuthStateChanged as any).mockImplementation((auth: any, callback: any) => {
      callback(mockUser);
      return vi.fn();
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider
    });

    // Wait for the useEffect to finish
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.user).toEqual(mockUser);
    expect(result.current.tenantId).toBe('test-tenant-123');
    
    // Verify API injection
    expect(api.setToken).toHaveBeenCalledWith('mock-token');
    expect(api.setCompanyId).toHaveBeenCalledWith('test-tenant-123');
  });

  it('should handle unauthenticated state', async () => {
    (onAuthStateChanged as any).mockImplementation((auth: any, callback: any) => {
      callback(null);
      return vi.fn();
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider
    });

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.user).toBe(null);
    expect(result.current.tenantId).toBe(null);
    expect(api.setToken).toHaveBeenCalledWith(null);
  });

  it('should handle logout', async () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider
    });

    await act(async () => {
      await result.current.logout();
    });

    expect(signOut).toHaveBeenCalled();
  });
});
