'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';

export function GuestGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const [checked, setChecked] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      // Check if there's a redirect target stored (e.g., /console)
      const redirect = typeof window !== 'undefined' ? sessionStorage.getItem('redirect_after_login') : null;
      if (redirect) {
        sessionStorage.removeItem('redirect_after_login');
        router.replace(redirect);
      } else {
        router.replace('/dashboard');
      }
      return;
    }
    setChecked(true);
  }, [isAuthenticated, router]);

  if (!checked) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-950">
        <div className="w-8 h-8 border-2 border-brand-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return <>{children}</>;
}
