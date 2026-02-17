import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Payment Watchdog - Payment Recovery Management',
  description: 'Monitor and manage payment recovery workflows',
  keywords: 'payment recovery, payment failures, cashflow, fintech, SaaS',
  authors: [{ name: 'Payment Watchdog Team' }],
  viewport: 'width=device-width, initial-scale=1',
  robots: 'index, follow',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="font-sans">
        <div className="min-h-screen bg-gray-50">
          {children}
        </div>
      </body>
    </html>
  );
}
