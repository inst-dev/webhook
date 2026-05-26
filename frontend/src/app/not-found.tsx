'use client';

import Link from 'next/link';
import { Webhook, Home, ArrowLeft } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950 px-4">
      <div className="text-center">
        <Webhook className="w-16 h-16 text-brand-500 mx-auto mb-6" />
        <h1 className="text-6xl font-bold text-white mb-2">404</h1>
        <h2 className="text-xl font-semibold text-gray-300 mb-4">Page not found</h2>
        <p className="text-gray-400 max-w-md mx-auto mb-8">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <div className="flex items-center justify-center space-x-4">
          <Link href="/" className="flex items-center space-x-2 bg-brand-600 hover:bg-brand-700 text-white px-5 py-2.5 rounded-lg font-medium transition-colors">
            <Home className="w-4 h-4" />
            <span>Home</span>
          </Link>
          <button onClick={() => window.history.back()} className="flex items-center space-x-2 border border-gray-700 hover:border-gray-600 text-gray-300 hover:text-white px-5 py-2.5 rounded-lg font-medium transition-colors">
            <ArrowLeft className="w-4 h-4" />
            <span>Go back</span>
          </button>
        </div>
      </div>
    </div>
  );
}
