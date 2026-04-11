import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AppState {
  isPrivacyMode: boolean
  togglePrivacyMode: () => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      isPrivacyMode: false,
      togglePrivacyMode: () => set((state) => ({ isPrivacyMode: !state.isPrivacyMode })),
    }),
    {
      name: 'payment-watchdog-storage',
    }
  )
)
