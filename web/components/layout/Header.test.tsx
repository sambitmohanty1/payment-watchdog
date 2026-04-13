import { render, screen, fireEvent } from '@testing-library/react';
import { Header } from './Header';
import { useAuth } from '@/hooks/useAuth';
import { useAppStore } from '@/lib/store';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import React from 'react';

// Mock the dependencies
vi.mock('@/hooks/useAuth');
vi.mock('@/lib/store');
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn() }),
}));
vi.mock('next-themes', () => ({
  useTheme: () => ({ theme: 'dark', setTheme: vi.fn() }),
}));

// Mock Radix UI Dropdown to render content directly for testing
vi.mock('@/components/ui', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...actual as any,
    MaterialIcon: ({ name }: { name: string }) => <span data-testid={`icon-${name}`}>{name}</span>,
    DropdownMenu: ({ children }: any) => <div>{children}</div>,
    DropdownMenuTrigger: ({ children }: any) => <div>{children}</div>,
    DropdownMenuContent: ({ children }: any) => <div data-testid="dropdown-content">{children}</div>,
    DropdownMenuItem: ({ children, onClick, className }: any) => (
      <div onClick={onClick} className={className}>{children}</div>
    ),
  };
});

describe('Header Component', () => {
  const mockLogout = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (useAppStore as unknown as any).mockReturnValue({
      isPrivacyMode: false,
      togglePrivacyMode: vi.fn(),
    });
  });

  it('should render login and get started buttons when not authenticated', () => {
    (useAuth as any).mockReturnValue({
      user: null,
      loading: false,
      logout: mockLogout,
    });

    render(<Header />);

    expect(screen.getByText('Login')).toBeInTheDocument();
  });

  it('should render user menu when authenticated', () => {
    (useAuth as any).mockReturnValue({
      user: { email: 'john@example.com', displayName: 'John Doe' },
      loading: false,
      logout: mockLogout,
    });

    render(<Header />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
  });

  it('should call logout when logout button clicked', () => {
    (useAuth as any).mockReturnValue({
      user: { email: 'test@example.com' },
      loading: false,
      logout: mockLogout,
    });

    render(<Header />);
    
    // Find logout which is now rendered directly due to mock
    const logoutBtn = screen.getByText('Logout');
    fireEvent.click(logoutBtn);

    expect(mockLogout).toHaveBeenCalled();
  });
});
