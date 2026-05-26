import type { Metadata } from 'next';
import { Inter, JetBrains_Mono } from 'next/font/google';
import './globals.css';
import { Providers } from '@/components/providers';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const jetbrains = JetBrains_Mono({ subsets: ['latin'], variable: '--font-mono' });

export const metadata: Metadata = {
  title: 'Webhook.inst.lk - Developer Webhook & Interaction Platform',
  description: 'Capture, inspect, and debug webhooks, DNS interactions, and emails in real-time.',
  keywords: ['webhook', 'api', 'developer tools', 'request inspector', 'dns logging'],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable} ${jetbrains.variable}`} suppressHydrationWarning>
      <body className="min-h-screen bg-white dark:bg-gray-950 antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
