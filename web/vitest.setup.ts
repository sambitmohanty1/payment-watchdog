import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Global mock for Firebase to avoid initialization errors in tests
vi.mock('firebase/app', () => ({
  initializeApp: vi.fn(),
  getApps: vi.fn(() => []),
  getApp: vi.fn(),
}));

vi.mock('firebase/auth', () => ({
  getAuth: vi.fn(() => ({})),
  onAuthStateChanged: vi.fn(() => vi.fn()),
  signOut: vi.fn(),
  signInWithEmailAndPassword: vi.fn(),
  signInWithPopup: vi.fn(),
  GoogleAuthProvider: vi.fn(),
  getIdToken: vi.fn(),
  getIdTokenResult: vi.fn(),
}));

// Mock MaterialIcon to avoid rendering issues globally
vi.stubGlobal('MaterialIcon', () => null);
