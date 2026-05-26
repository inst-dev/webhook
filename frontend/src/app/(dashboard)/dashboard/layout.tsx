'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Webhook, Globe, Activity, Mail, Settings, Shield, LogOut } from 'lucide-react';
import { useAuthStore } from '@/store/auth';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { logout, user } = useAuthStore();

  const navItems = [
    { href: '/dashboard', label: 'Endpoints', icon: Globe },
    { href: '/dashboard/dns', label: 'DNS Hooks', icon: Activity },
    { href: '/dashboard/emails', label: 'Email Hooks', icon: Mail },
    { href: '/dashboard/settings', label: 'Settings', icon: Settings },
  ];

  const isActive = (href: string) => {
    if (href === '/dashboard') return pathname === '/dashboard';
    return pathname.startsWith(href);
  };

  return (
    <div className="flex min-h-screen bg-gray-950">
      {/* Sidebar */}
      <aside className="w-64 border-r border-gray-800/50 bg-gray-950 flex flex-col">
        <div className="p-4 border-b border-gray-800/50">
          <Link href="/" className="flex items-center space-x-2">
            <Webhook className="w-7 h-7 text-brand-500" />
            <span className="text-lg font-bold text-white">Webhook</span>
          </Link>
        </div>

        <nav className="flex-1 p-3 space-y-1">
          {navItems.map(({ href, label, icon: Icon }) => (
            <Link
              key={href}
              href={href}
              className={`flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive(href)
                  ? 'bg-brand-500/10 text-brand-400'
                  : 'text-gray-400 hover:text-white hover:bg-gray-900/50'
              }`}
            >
              <Icon className="w-4 h-4" />
              <span>{label}</span>
            </Link>
          ))}

          {user?.plan === 'admin' || user?.plan === 'enterprise' ? (
            <Link
              href="/console"
              className="flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-900/50"
            >
              <Shield className="w-4 h-4" />
              <span>Admin Console</span>
            </Link>
          ) : null}
        </nav>

        <div className="p-3 border-t border-gray-800/50">
          <div className="px-3 py-2 mb-2">
            <p className="text-sm text-white truncate">{user?.display_name}</p>
            <p className="text-xs text-gray-500 truncate">{user?.email}</p>
          </div>
          <button
            onClick={() => { logout(); window.location.href = '/login'; }}
            className="flex items-center space-x-3 px-3 py-2.5 rounded-lg text-sm font-medium text-gray-400 hover:text-red-400 hover:bg-red-500/5 w-full transition-colors"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  );
}
