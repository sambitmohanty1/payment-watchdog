import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Payment Watchdog | Sovereign AU Command Center',
  description: 'Mission-critical payment recovery orchestration and health monitoring for the Australian Sovereign region.',
  keywords: 'payment recovery, sovereign cloud, fintech, payment watchdog, data residency',
  authors: [{ name: 'Payment Watchdog Principal Engineering' }],
  viewport: 'width=device-width, initial-scale=1, maximum-scale=1',
  icons: {
    icon: '/favicon.svg',
    shortcut: '/favicon-192x192.svg',
    apple: '/favicon-192x192.svg',
  },
  manifest: '/site.webmanifest',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <head>
        <link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons" />
      </head>
      <body className="font-sans antialiased overflow-x-hidden">
        {children}
      </body>
    </html>
  );
}
