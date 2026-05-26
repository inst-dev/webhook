'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Key, Shield, CreditCard, User, Copy } from 'lucide-react';
import { apiKeysAPI, api } from '@/lib/api';
import { useAuthStore } from '@/store/auth';
import toast from 'react-hot-toast';

export default function SettingsPage() {
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('profile');

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-4 py-4">
          <h1 className="text-xl font-bold text-white">Settings</h1>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 py-8">
        {/* Tabs */}
        <div className="flex space-x-1 bg-gray-900/50 rounded-lg p-1 mb-8 w-fit">
          {[
            { id: 'profile', label: 'Profile', icon: User },
            { id: 'api-keys', label: 'API Keys', icon: Key },
            { id: 'security', label: 'Security', icon: Shield },
            { id: 'billing', label: 'Billing', icon: CreditCard },
          ].map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => setActiveTab(id)}
              className={`flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === id
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:text-white'
              }`}
            >
              <Icon className="w-4 h-4" />
              <span>{label}</span>
            </button>
          ))}
        </div>

        {activeTab === 'profile' && <ProfileSection user={user} />}
        {activeTab === 'api-keys' && <APIKeysSection />}
        {activeTab === 'security' && <SecuritySection />}
        {activeTab === 'billing' && <BillingSection />}
      </main>
    </div>
  );
}

function ProfileSection({ user }: { user: any }) {
  const [displayName, setDisplayName] = useState(user?.display_name || '');

  return (
    <div className="space-y-6">
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Profile Information</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Email</label>
            <input type="email" value={user?.email || ''} disabled className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-gray-400 cursor-not-allowed" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Display Name</label>
            <input type="text" value={displayName} onChange={(e) => setDisplayName(e.target.value)} className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-brand-500 outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Plan</label>
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-brand-500/10 text-brand-400 capitalize">
              {user?.plan || 'free'}
            </span>
          </div>
          <button className="bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg text-sm font-medium">
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}

function APIKeysSection() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [keyName, setKeyName] = useState('');
  const [newKey, setNewKey] = useState('');

  const { data } = useQuery({
    queryKey: ['api-keys'],
    queryFn: () => apiKeysAPI.list(),
  });

  const createMutation = useMutation({
    mutationFn: () => apiKeysAPI.create({ name: keyName, scopes: ['*'] }),
    onSuccess: (response) => {
      setNewKey(response.data.key);
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      setKeyName('');
    },
    onError: () => toast.error('Failed to create API key'),
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => apiKeysAPI.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      toast.success('API key revoked');
    },
  });

  const keys = data?.data?.api_keys || [];

  return (
    <div className="space-y-6">
      {newKey && (
        <div className="bg-green-500/10 border border-green-500/30 rounded-xl p-4">
          <p className="text-green-400 text-sm mb-2 font-medium">API key created! Copy it now - it won&apos;t be shown again.</p>
          <div className="flex items-center space-x-2">
            <code className="flex-1 text-sm text-green-300 font-mono bg-gray-900 rounded px-3 py-2 truncate">{newKey}</code>
            <button onClick={() => { navigator.clipboard.writeText(newKey); toast.success('Copied!'); }} className="text-green-400 hover:text-green-300">
              <Copy className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">API Keys</h2>
          <button onClick={() => setShowCreate(!showCreate)} className="bg-brand-600 hover:bg-brand-700 text-white px-3 py-1.5 rounded-lg text-sm font-medium">
            Create Key
          </button>
        </div>

        {showCreate && (
          <form onSubmit={(e) => { e.preventDefault(); createMutation.mutate(); }} className="flex items-center space-x-3 mb-4 p-4 bg-gray-800/50 rounded-lg">
            <input type="text" value={keyName} onChange={(e) => setKeyName(e.target.value)} placeholder="Key name" required className="flex-1 px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white text-sm focus:ring-2 focus:ring-brand-500 outline-none" />
            <button type="submit" className="bg-brand-600 text-white px-4 py-2 rounded-lg text-sm">Create</button>
          </form>
        )}

        <div className="space-y-3">
          {keys.length === 0 ? (
            <p className="text-gray-400 text-sm">No API keys created yet.</p>
          ) : (
            keys.map((key: any) => (
              <div key={key.id} className="flex items-center justify-between p-3 bg-gray-800/30 rounded-lg">
                <div>
                  <span className="text-white text-sm font-medium">{key.name}</span>
                  <span className="ml-3 text-gray-500 text-xs font-mono">{key.key_prefix}...</span>
                </div>
                <div className="flex items-center space-x-3">
                  <span className="text-xs text-gray-500">{key.last_used_at ? `Used ${new Date(key.last_used_at).toLocaleDateString()}` : 'Never used'}</span>
                  <button onClick={() => revokeMutation.mutate(key.id)} className="text-red-400 hover:text-red-300 text-xs">Revoke</button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function SecuritySection() {
  return (
    <div className="space-y-6">
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Change Password</h2>
        <div className="space-y-4 max-w-md">
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Current Password</label>
            <input type="password" className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-brand-500 outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">New Password</label>
            <input type="password" className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-brand-500 outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1.5">Confirm New Password</label>
            <input type="password" className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-brand-500 outline-none" />
          </div>
          <button className="bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg text-sm font-medium">
            Update Password
          </button>
        </div>
      </div>

      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Two-Factor Authentication</h2>
        <p className="text-gray-400 text-sm mb-4">Add an extra layer of security to your account.</p>
        <button className="border border-gray-700 hover:border-gray-600 text-gray-300 hover:text-white px-4 py-2 rounded-lg text-sm font-medium">
          Enable 2FA
        </button>
      </div>

      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Active Sessions</h2>
        <p className="text-gray-400 text-sm mb-4">Manage your active login sessions.</p>
        <button className="border border-red-700 hover:border-red-600 text-red-400 hover:text-red-300 px-4 py-2 rounded-lg text-sm font-medium">
          Revoke All Sessions
        </button>
      </div>
    </div>
  );
}

function BillingSection() {
  const { data: plansData } = useQuery({
    queryKey: ['plans'],
    queryFn: () => api.get('/billing/plans'),
  });

  const { data: subData } = useQuery({
    queryKey: ['subscription'],
    queryFn: () => api.get('/billing/subscription'),
  });

  const plans = plansData?.data?.plans || [];
  const subscription = subData?.data?.subscription;

  return (
    <div className="space-y-6">
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Current Plan</h2>
        <div className="flex items-center space-x-3">
          <span className="text-2xl font-bold text-brand-400 capitalize">{subscription?.plan || 'Free'}</span>
          {subscription && (
            <span className="text-sm text-gray-400">
              Renews {new Date(subscription.current_period_end).toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
        {plans.map((plan: any) => (
          <div key={plan.id} className={`border rounded-xl p-5 ${
            plan.id === (subscription?.plan || 'free')
              ? 'border-brand-500 bg-brand-500/5'
              : 'border-gray-800 bg-gray-900/50'
          }`}>
            <h3 className="text-white font-semibold">{plan.name}</h3>
            <div className="mt-2">
              <span className="text-2xl font-bold text-white">${plan.price}</span>
              <span className="text-gray-400 text-sm">/mo</span>
            </div>
            <p className="text-gray-400 text-xs mt-2">{plan.description}</p>
            <button
              disabled={plan.id === (subscription?.plan || 'free')}
              className={`w-full mt-4 py-2 rounded-lg text-sm font-medium ${
                plan.id === (subscription?.plan || 'free')
                  ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                  : 'bg-brand-600 hover:bg-brand-700 text-white'
              }`}
            >
              {plan.id === (subscription?.plan || 'free') ? 'Current Plan' : 'Upgrade'}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
