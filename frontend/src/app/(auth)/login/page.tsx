'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Webhook, Eye, EyeOff } from 'lucide-react';
import { authAPI } from '@/lib/api';
import { useAuthStore } from '@/store/auth';
import toast from 'react-hot-toast';

export default function LoginPage() {
  const router = useRouter();
  const { setUser, setToken } = useAuthStore();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const { data } = await authAPI.login({ email, password });
      setUser(data.user);
      setToken(data.access_token);
      toast.success('Welcome back!');
      router.push('/dashboard');
    } catch (error: any) {
      const message = error.response?.data?.message || 'Login failed';
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <Link href="/" className="inline-flex items-center space-x-2">
            <Webhook className="w-8 h-8 text-brand-500" />
            <span className="text-xl font-bold text-white">Webhook.inst.lk</span>
          </Link>
          <h1 className="mt-6 text-2xl font-bold text-white">Sign in to your account</h1>
          <p className="mt-2 text-gray-400">
            Don&apos;t have an account?{' '}
            <Link href="/register" className="text-brand-500 hover:text-brand-400">
              Create one
            </Link>
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-300 mb-1.5">
              Email address
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:ring-2 focus:ring-brand-500 focus:border-transparent outline-none transition"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-1.5">
              Password
            </label>
            <div className="relative">
              <input
                id="password"
                type={showPassword ? 'text' : 'password'}
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-2.5 bg-gray-900 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:ring-2 focus:ring-brand-500 focus:border-transparent outline-none transition pr-10"
                placeholder="Enter your password"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          <div className="flex items-center justify-between">
            <label className="flex items-center">
              <input type="checkbox" className="rounded border-gray-700 bg-gray-900 text-brand-500 focus:ring-brand-500" />
              <span className="ml-2 text-sm text-gray-400">Remember me</span>
            </label>
            <Link href="/forgot-password" className="text-sm text-brand-500 hover:text-brand-400">
              Forgot password?
            </Link>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-brand-600 hover:bg-brand-700 disabled:bg-brand-800 disabled:cursor-not-allowed text-white py-2.5 rounded-lg font-medium transition-colors"
          >
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
