'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { 
  onAuthStateChanged, 
  User, 
  signOut,
  getIdToken,
  getIdTokenResult
} from 'firebase/auth';
import { auth } from '@/lib/firebase';
import api from '@/lib/api';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  tenantId: string | null;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  tenantId: null,
  logout: async () => {},
});

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [tenantId, setTenantId] = useState<string | null>(null);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      if (user) {
        setUser(user);
        
        // Extract token and tenant_id (company_id) from custom claims
        const token = await getIdToken(user);
        const idTokenResult = await getIdTokenResult(user);
        const cid = idTokenResult.claims.tenant_id as string;
        
        setTenantId(cid || null);
        
        // Inject into API singleton
        api.setToken(token);
        if (cid) {
          api.setCompanyId(cid);
        }
      } else {
        setUser(null);
        setTenantId(null);
        api.setToken(null);
      }
      setLoading(false);
    });

    return () => unsubscribe();
  }, []);

  const logout = async () => {
    await signOut(auth);
    window.location.href = '/login';
  };

  // Automated Onboarding Guard
  useEffect(() => {
    // Only redirect if we are on the dashboard/app routes, not login/onboarding themselves
    const publicPaths = ['/login', '/onboarding'];
    const currentPath = window.location.pathname;

    if (!loading && user && !tenantId && !publicPaths.includes(currentPath)) {
      window.location.href = '/onboarding';
    }
  }, [user, tenantId, loading]);

  return (
    <AuthContext.Provider value={{ user, loading, tenantId, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
