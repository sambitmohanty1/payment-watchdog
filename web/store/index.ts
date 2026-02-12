import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import {
  PaymentFailureEvent,
  DashboardStats,
  Alert,
  Company,
  FilterOptions,
  SortOptions,
  CommunicationTemplate,
} from '@/types';

interface AppState {
  // Company and authentication
  company: Company | null;
  companyId: string | null;
  
  // Dashboard data
  dashboardStats: DashboardStats | null;
  paymentFailures: PaymentFailureEvent[];
  alerts: Alert[];
  
  // UI state
  isLoading: boolean;
  error: string | null;
  currentView: 'dashboard' | 'failures' | 'alerts' | 'settings';
  
  // Filters and pagination
  filters: FilterOptions;
  sortOptions: SortOptions;
  currentPage: number;
  totalPages: number;
  totalItems: number;
  
  // Communication templates
  communicationTemplates: CommunicationTemplate[];
  
  // Actions
  setCompany: (company: Company) => void;
  setCompanyId: (id: string) => void;
  setDashboardStats: (stats: DashboardStats) => void;
  setPaymentFailures: (failures: PaymentFailureEvent[]) => void;
  addPaymentFailure: (failure: PaymentFailureEvent) => void;
  updatePaymentFailure: (id: string, updates: Partial<PaymentFailureEvent>) => void;
  setAlerts: (alerts: Alert[]) => void;
  markAlertAsRead: (id: string) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setCurrentView: (view: AppState['currentView']) => void;
  setFilters: (filters: FilterOptions) => void;
  setSortOptions: (sort: SortOptions) => void;
  setPagination: (page: number, totalPages: number, totalItems: number) => void;
  setCommunicationTemplates: (templates: CommunicationTemplate[]) => void;
  clearError: () => void;
  resetState: () => void;
  clearPersistedData: () => void;
}

const initialState = {
  company: null,
  companyId: null,
  dashboardStats: null,
  paymentFailures: [],
  alerts: [],
  isLoading: false,
  error: null,
  currentView: 'dashboard' as const,
  filters: {},
  sortOptions: { field: 'created_at', direction: 'desc' as const },
  currentPage: 1,
  totalPages: 1,
  totalItems: 0,
  communicationTemplates: [],
};

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set, get) => ({
        ...initialState,

        setCompany: (company) => set({ company, companyId: company.id }),
        
        setCompanyId: (id) => set({ companyId: id }),
        
        setDashboardStats: (stats) => set({ dashboardStats: stats }),
        
        setPaymentFailures: (failures) => set({ paymentFailures: failures }),
        
        addPaymentFailure: (failure) => {
          const { paymentFailures } = get();
          // Add to beginning for real-time updates
          set({ paymentFailures: [failure, ...paymentFailures] });
        },
        
        updatePaymentFailure: (id, updates) => {
          const { paymentFailures } = get();
          const updatedFailures = paymentFailures.map(failure =>
            failure.id === id ? { ...failure, ...updates } : failure
          );
          set({ paymentFailures: updatedFailures });
        },
        
        setAlerts: (alerts) => set({ alerts }),
        
        markAlertAsRead: (id) => {
          const { alerts } = get();
          const updatedAlerts = alerts.map(alert =>
            alert.id === id ? { ...alert, status: 'read' as const, read_at: new Date().toISOString() } : alert
          );
          set({ alerts: updatedAlerts });
        },
        
        setLoading: (loading) => set({ isLoading: loading }),
        
        setError: (error) => set({ error }),
        
        setCurrentView: (view) => set({ currentView: view }),
        
        setFilters: (filters) => set({ filters, currentPage: 1 }),
        
        setSortOptions: (sort) => set({ sortOptions: sort, currentPage: 1 }),
        
        setPagination: (page, totalPages, totalItems) => 
          set({ currentPage: page, totalPages, totalItems }),
        
        setCommunicationTemplates: (templates) => set({ communicationTemplates: templates }),
        
        clearError: () => set({ error: null }),
        
        resetState: () => set(initialState),
        
        clearPersistedData: () => {
          // Clear localStorage
          if (typeof window !== 'undefined') {
            localStorage.removeItem('lexure-intelligence-store');
          }
          // Reset to initial state
          set(initialState);
        },
      }),
      {
        name: 'lexure-intelligence-store',
        partialize: (state) => ({
          company: state.company,
          companyId: state.companyId,
          currentView: state.currentView,
          filters: state.filters,
          sortOptions: state.sortOptions,
        }),
      }
    ),
    {
      name: 'lexure-intelligence-store',
    }
  )
);

// Selector hooks for better performance
export const useCompany = () => useAppStore((state) => state.company);
export const useCompanyId = () => useAppStore((state) => state.companyId);
export const useDashboardStats = () => useAppStore((state) => state.dashboardStats);
export const usePaymentFailures = () => useAppStore((state) => state.paymentFailures);
export const useAlerts = () => useAppStore((state) => state.alerts);
export const useIsLoading = () => useAppStore((state) => state.isLoading);
export const useError = () => useAppStore((state) => state.error);
export const useCurrentView = () => useAppStore((state) => state.currentView);
export const useFilters = () => useAppStore((state) => state.filters);
export const useSortOptions = () => useAppStore((state) => state.sortOptions);
export const usePagination = () => useAppStore((state) => ({
  currentPage: state.currentPage,
  totalPages: state.totalPages,
  totalItems: state.totalItems,
}));
export const useCommunicationTemplates = () => useAppStore((state) => state.communicationTemplates);
