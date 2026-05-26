'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth';

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, user, accessToken } = useAuthStore();
  const [checked, setChecked] = useState(false);

  useEffect(() => {
    // Check auth state after hydration
    if (!isAuthenticated || !user || !accessToken) {
      // Store the intended destination so we can redirect back after login
      if (typeof window !== 'undefined') {
        sessionStorage.setItem('redirect_after_login', pathname);
      }
      router.replace('/login');
      return;
    }
    setChecked(true);
  }, [isAuthenticated, user, accessToken, router, pathname]);

  if (!checked) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-950">
        <div className="flex flex-col items-center space-y-4">
          <div className="w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full animate-spin" />
          <p className="text-gray-400 text-sm">Loading...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
