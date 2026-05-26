'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Webhook, Check, Zap } from 'lucide-react';
import toast from 'react-hot-toast';

const plans = [
  {
    id: 'free',
    name: 'Free',
    price: 0,
    description: 'Perfect for trying out the platform',
    features: [
      '5 endpoints',
      'Unlimited requests',
      '72 hour retention',
      'Realtime WebSocket updates',
      'Request inspector',
      'Basic search',
    ],
    limitations: ['No custom domains', 'No DNS hooks', 'No email hooks', 'No API access'],
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 9.99,
    popular: true,
    description: 'For professional developers and security researchers',
    features: [
      '50 endpoints',
      'Unlimited requests',
      '30 day retention',
      '3 custom domains',
      'DNS interaction hooks',
      'Email hooks (SMTP)',
      'Full API access',
      'Custom responses',
      'Request replay',
      'Advanced search & filtering',
    ],
    limitations: [],
  },
  {
    id: 'team',
    name: 'Team',
    price: 29.99,
    description: 'For development teams and organizations',
    features: [
      '200 endpoints',
      'Unlimited requests',
      '90 day retention',
      '10 custom domains',
      'DNS interaction hooks',
      'Email hooks (SMTP)',
      'Full API access',
      'Up to 10 team members',
      'Analytics dashboard',
      'Priority support',
    ],
    limitations: [],
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 99.99,
    description: 'For enterprises requiring full access and support',
    features: [
      'Unlimited endpoints',
      'Unlimited requests',
      '1 year retention',
      'Unlimited custom domains',
      'Unlimited team members',
      'Admin console access',
      'SLA guarantee',
      'Priority support',
      'Custom integrations',
      'Dedicated account manager',
    ],
    limitations: [],
  },
];

export default function PricingPage() {
  const [billingInterval, setBillingInterval] = useState<'monthly' | 'yearly'>('monthly');

  const handleSubscribe = (planId: string) => {
    if (planId === 'free') {
      window.location.href = '/register';
      return;
    }
    // Redirect to register/login with plan selection
    window.location.href = `/register?plan=${planId}`;
  };

  return (
    <div className="min-h-screen bg-gray-950">
      {/* Navigation */}
      <nav className="border-b border-gray-800/50 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link href="/" className="flex items-center space-x-2">
              <Webhook className="w-8 h-8 text-brand-500" />
              <span className="text-xl font-bold text-white">Webhook.inst.lk</span>
            </Link>
            <div className="flex items-center space-x-4">
              <Link href="/login" className="text-gray-300 hover:text-white px-4 py-2 text-sm transition-colors">
                Sign In
              </Link>
              <Link href="/register" className="bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors">
                Get Started
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Header */}
      <section className="py-16 px-4 text-center">
        <h1 className="text-4xl md:text-5xl font-bold text-white mb-4">Simple, transparent pricing</h1>
        <p className="text-xl text-gray-400 max-w-2xl mx-auto mb-8">
          All plans include unlimited requests. No hidden fees.
        </p>

        {/* Billing toggle */}
        <div className="inline-flex items-center bg-gray-900 rounded-lg p-1">
          <button
            onClick={() => setBillingInterval('monthly')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              billingInterval === 'monthly' ? 'bg-brand-600 text-white' : 'text-gray-400 hover:text-white'
            }`}
          >
            Monthly
          </button>
          <button
            onClick={() => setBillingInterval('yearly')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              billingInterval === 'yearly' ? 'bg-brand-600 text-white' : 'text-gray-400 hover:text-white'
            }`}
          >
            Yearly <span className="text-green-400 text-xs ml-1">-20%</span>
          </button>
        </div>
      </section>

      {/* Plans Grid */}
      <section className="max-w-7xl mx-auto px-4 pb-20">
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          {plans.map((plan) => (
            <div
              key={plan.id}
              className={`relative rounded-2xl p-6 ${
                plan.popular
                  ? 'bg-brand-500/5 border-2 border-brand-500'
                  : 'bg-gray-900/50 border border-gray-800'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="bg-brand-600 text-white text-xs font-bold px-3 py-1 rounded-full flex items-center space-x-1">
                    <Zap className="w-3 h-3" />
                    <span>Most Popular</span>
                  </span>
                </div>
              )}

              <h3 className="text-lg font-semibold text-white">{plan.name}</h3>
              <p className="text-gray-400 text-sm mt-1">{plan.description}</p>

              <div className="mt-4 mb-6">
                <span className="text-4xl font-bold text-white">
                  ${billingInterval === 'yearly' ? (plan.price * 0.8).toFixed(2) : plan.price.toFixed(2)}
                </span>
                {plan.price > 0 && <span className="text-gray-400 text-sm">/month</span>}
              </div>

              <button
                onClick={() => handleSubscribe(plan.id)}
                className={`w-full py-2.5 rounded-lg font-medium text-sm transition-colors ${
                  plan.popular
                    ? 'bg-brand-600 hover:bg-brand-700 text-white'
                    : 'bg-gray-800 hover:bg-gray-700 text-white'
                }`}
              >
                {plan.price === 0 ? 'Get Started Free' : 'Subscribe'}
              </button>

              <ul className="mt-6 space-y-3">
                {plan.features.map((feature, i) => (
                  <li key={i} className="flex items-start space-x-2 text-sm">
                    <Check className="w-4 h-4 text-green-400 mt-0.5 flex-shrink-0" />
                    <span className="text-gray-300">{feature}</span>
                  </li>
                ))}
                {plan.limitations.map((limitation, i) => (
                  <li key={i} className="flex items-start space-x-2 text-sm">
                    <span className="w-4 h-4 text-gray-600 mt-0.5 flex-shrink-0 text-center">-</span>
                    <span className="text-gray-500">{limitation}</span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>

      {/* FAQ */}
      <section className="max-w-3xl mx-auto px-4 pb-20">
        <h2 className="text-2xl font-bold text-white text-center mb-8">Frequently Asked Questions</h2>
        <div className="space-y-4">
          <FaqItem question="Are requests really unlimited?" answer="Yes! All plans include unlimited webhook requests. We never throttle or limit the number of incoming webhooks you can receive." />
          <FaqItem question="What payment methods do you accept?" answer="We accept PayPal for international payments and PayHere for local Sri Lankan payments (credit/debit cards, bank transfers)." />
          <FaqItem question="Can I cancel anytime?" answer="Yes, you can cancel your subscription at any time. You'll retain access until the end of your billing period." />
          <FaqItem question="What happens when retention expires?" answer="Captured requests older than your plan's retention period are automatically cleaned up. Upgrade your plan for longer retention." />
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

function FaqItem({ question, answer }: { question: string; answer: string }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl">
      <button onClick={() => setOpen(!open)} className="w-full text-left p-5 flex items-center justify-between">
        <span className="text-white font-medium">{question}</span>
        <span className="text-gray-400 text-xl">{open ? '-' : '+'}</span>
      </button>
      {open && <p className="px-5 pb-5 text-gray-400 text-sm">{answer}</p>}
    </div>
  );
}
