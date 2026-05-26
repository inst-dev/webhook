'use client';

import Link from 'next/link';
import { Globe, Zap, Shield, Terminal, Mail, Webhook } from 'lucide-react';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-950 via-gray-900 to-gray-950">
      {/* Navigation */}
      <nav className="border-b border-gray-800/50 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-2">
              <Webhook className="w-8 h-8 text-brand-500" />
              <span className="text-xl font-bold text-white">Webhook.inst.lk</span>
            </div>
            <div className="flex items-center space-x-4">
              <Link href="/pricing" className="text-gray-300 hover:text-white px-4 py-2 text-sm transition-colors">
                Pricing
              </Link>
              <Link href="/login" className="text-gray-300 hover:text-white px-4 py-2 text-sm transition-colors">
                Sign In
              </Link>
              <Link
                href="/register"
                className="bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
              >
                Get Started
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="py-24 px-4">
        <div className="max-w-4xl mx-auto text-center">
          <div className="inline-flex items-center px-3 py-1 rounded-full bg-brand-500/10 border border-brand-500/20 text-brand-400 text-sm mb-8">
            <Zap className="w-3 h-3 mr-2" />
            Realtime webhook inspection & interaction platform
          </div>
          <h1 className="text-5xl md:text-6xl font-bold text-white mb-6 leading-tight">
            Capture. Inspect.
            <br />
            <span className="text-brand-500">Debug.</span>
          </h1>
          <p className="text-xl text-gray-400 mb-10 max-w-2xl mx-auto">
            A production-grade platform for developers, security researchers, and automation engineers 
            to capture webhooks, DNS queries, and emails in real-time.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/register"
              className="w-full sm:w-auto bg-brand-600 hover:bg-brand-700 text-white px-8 py-3 rounded-lg font-medium transition-colors text-lg"
            >
              Create Free Endpoint
            </Link>
            <Link
              href="/pricing"
              className="w-full sm:w-auto border border-gray-700 hover:border-gray-600 text-gray-300 hover:text-white px-8 py-3 rounded-lg font-medium transition-colors text-lg"
            >
              View Pricing
            </Link>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="py-20 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<Globe className="w-6 h-6" />}
              title="Webhook Capture"
              description="Generate unique URLs to capture HTTP requests with full header, body, and metadata inspection."
            />
            <FeatureCard
              icon={<Terminal className="w-6 h-6" />}
              title="DNS Interaction Logging"
              description="Monitor DNS queries with wildcard subdomain support. Perfect for SSRF detection and bug bounty."
            />
            <FeatureCard
              icon={<Mail className="w-6 h-6" />}
              title="Email Hooks"
              description="Receive and inspect emails with full MIME parsing, attachment handling, and HTML rendering."
            />
            <FeatureCard
              icon={<Zap className="w-6 h-6" />}
              title="Realtime Updates"
              description="WebSocket-powered live request streaming with instant UI updates and live counters."
            />
            <FeatureCard
              icon={<Shield className="w-6 h-6" />}
              title="Security First"
              description="Built with strict CSP, XSS prevention, CSRF protection, and comprehensive input validation."
            />
            <FeatureCard
              icon={<Webhook className="w-6 h-6" />}
              title="Request Replay"
              description="Resend captured requests with modified headers, body, and target URL for debugging."
            />
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-gray-800/50 py-8 px-4">
        <div className="max-w-6xl mx-auto text-center text-gray-500 text-sm">
          <p>&copy; 2024 Webhook.inst.lk. Built for developers, by developers.</p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <div className="p-6 rounded-xl bg-gray-900/50 border border-gray-800/50 hover:border-brand-500/30 transition-colors">
      <div className="w-12 h-12 rounded-lg bg-brand-500/10 flex items-center justify-center text-brand-500 mb-4">
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
      <p className="text-gray-400 text-sm leading-relaxed">{description}</p>
    </div>
  );
}
