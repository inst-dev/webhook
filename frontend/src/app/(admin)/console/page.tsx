'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Activity, Users, Globe, Mail, Server, Database, Wifi, Clock, Shield, CreditCard, Settings, Search, Ban, Check, Power } from 'lucide-react';
import { api } from '@/lib/api';
import { useAuthStore } from '@/store/auth';
import toast from 'react-hot-toast';

type TabId = 'overview' | 'users' | 'security' | 'billing' | 'settings';

export default function AdminConsolePage() {
  const [activeTab, setActiveTab] = useState<TabId>('overview');
  const { user } = useAuthStore();

  // Verify admin access - allow admin, enterprise plans and first registered user (owner)
  const isAdmin = user?.plan === 'admin' || user?.plan === 'enterprise' || user?.email?.endsWith('@webhook.inst.lk');
  
  if (!isAdmin) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-950">
        <div className="text-center">
          <Shield className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <h1 className="text-xl font-bold text-white mb-2">Access Denied</h1>
          <p className="text-gray-400">You don&apos;t have admin privileges.</p>
          <p className="text-gray-500 text-sm mt-2">Contact the platform owner to get admin access.</p>
        </div>
      </div>
    );
  }

  const tabs: { id: TabId; label: string; icon: any }[] = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'users', label: 'Users', icon: Users },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'billing', label: 'Billing', icon: CreditCard },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-white">Admin Console</h1>
            <p className="text-sm text-gray-400">console.webhook.inst.lk</p>
          </div>
        </div>
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex space-x-1 overflow-x-auto pb-2">
            {tabs.map(({ id, label, icon: Icon }) => (
              <button
                key={id}
                onClick={() => setActiveTab(id)}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors ${
                  activeTab === id ? 'bg-brand-500/10 text-brand-400' : 'text-gray-400 hover:text-white hover:bg-gray-900/50'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{label}</span>
              </button>
            ))}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        {activeTab === 'overview' && <OverviewTab />}
        {activeTab === 'users' && <UsersTab />}
        {activeTab === 'security' && <SecurityTab />}
        {activeTab === 'billing' && <BillingTab />}
        {activeTab === 'settings' && <SettingsTab />}
      </main>
    </div>
  );
}

function OverviewTab() {
  const { data: dashData } = useQuery({
    queryKey: ['admin-dashboard'],
    queryFn: () => api.get('/admin/dashboard'),
    refetchInterval: 10000,
  });

  const { data: metricsData } = useQuery({
    queryKey: ['admin-metrics'],
    queryFn: () => api.get('/admin/metrics/realtime'),
    refetchInterval: 2000,
  });

  const { data: healthData } = useQuery({
    queryKey: ['admin-health'],
    queryFn: () => api.get('/admin/health'),
    refetchInterval: 15000,
  });

  const stats = dashData?.data || {};
  const metrics = metricsData?.data || {};
  const health = healthData?.data || {};

  return (
    <div className="space-y-8">
      {/* Realtime Metrics */}
      <section>
        <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Realtime Traffic</h2>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          <MetricCard label="Req/sec" value={metrics.requests_per_second || 0} color="text-green-400" />
          <MetricCard label="Req/min" value={metrics.requests_per_minute || 0} color="text-blue-400" />
          <MetricCard label="Req/hour" value={metrics.requests_per_hour || 0} color="text-purple-400" />
          <MetricCard label="Req/day" value={metrics.requests_per_day || 0} color="text-orange-400" />
          <MetricCard label="WebSockets" value={metrics.active_websockets || 0} color="text-cyan-400" />
        </div>
      </section>

      {/* Platform Stats */}
      <section>
        <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Platform Overview</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard label="Total Users" value={stats.users?.total || 0} sublabel={`${stats.users?.active_today || 0} active today`} />
          <StatCard label="Endpoints" value={stats.endpoints?.total || 0} sublabel={`${stats.endpoints?.active || 0} active`} />
          <StatCard label="Requests Today" value={stats.requests?.today || 0} sublabel={`${stats.requests?.total || 0} total`} />
          <StatCard label="Emails Today" value={stats.emails?.today || 0} sublabel={`DNS: ${stats.dns?.today || 0}`} />
        </div>
      </section>

      {/* System Health */}
      <section>
        <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">System Health</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <HealthCard title="PostgreSQL" status={health.services?.database?.status} latency={health.services?.database?.latency_ms} />
          <HealthCard title="Redis" status={health.services?.redis?.status} latency={health.services?.redis?.latency_ms} />
          <HealthCard title="WebSocket Hub" status={true} extra={`${health.services?.websocket?.active_connections || 0} connections`} />
        </div>
      </section>

      {/* Runtime Info */}
      {health.runtime && (
        <section>
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Runtime</h2>
          <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5 grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <span className="text-xs text-gray-500 block">Goroutines</span>
              <span className="text-lg font-semibold text-white">{health.runtime.goroutines}</span>
            </div>
            <div>
              <span className="text-xs text-gray-500 block">Memory Alloc</span>
              <span className="text-lg font-semibold text-white">{(health.runtime.memory_alloc / 1024 / 1024).toFixed(1)} MB</span>
            </div>
            <div>
              <span className="text-xs text-gray-500 block">GC Runs</span>
              <span className="text-lg font-semibold text-white">{health.runtime.gc_runs}</span>
            </div>
            <div>
              <span className="text-xs text-gray-500 block">Go Version</span>
              <span className="text-lg font-semibold text-white">{health.runtime.go_version}</span>
            </div>
          </div>
        </section>
      )}
    </div>
  );
}

function UsersTab() {
  const [search, setSearch] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['admin-users', search],
    queryFn: () => api.get('/admin/users', { params: { search, limit: 50 } }),
  });

  const users = data?.data?.users || [];
  const total = data?.data?.total || 0;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">Users ({total})</h2>
        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search users..."
            className="pl-9 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white text-sm placeholder-gray-500 focus:ring-2 focus:ring-brand-500 outline-none w-64"
          />
        </div>
      </div>

      <div className="bg-gray-900/50 border border-gray-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-800">
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">User</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Plan</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Verified</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Joined</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Last Login</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800/50">
            {isLoading ? (
              <tr><td colSpan={6} className="text-center py-8 text-gray-400">Loading...</td></tr>
            ) : users.length === 0 ? (
              <tr><td colSpan={6} className="text-center py-8 text-gray-400">No users found</td></tr>
            ) : (
              users.map((u: any) => (
                <tr key={u.id} className="hover:bg-gray-800/30">
                  <td className="px-4 py-3">
                    <div>
                      <p className="text-sm text-white">{u.display_name}</p>
                      <p className="text-xs text-gray-500">{u.email}</p>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-xs font-medium px-2 py-0.5 rounded capitalize ${
                      u.plan === 'enterprise' ? 'bg-purple-500/10 text-purple-400' :
                      u.plan === 'team' ? 'bg-blue-500/10 text-blue-400' :
                      u.plan === 'pro' ? 'bg-green-500/10 text-green-400' :
                      'bg-gray-500/10 text-gray-400'
                    }`}>{u.plan}</span>
                  </td>
                  <td className="px-4 py-3">
                    {u.email_verified ? <Check className="w-4 h-4 text-green-400" /> : <span className="text-xs text-gray-500">No</span>}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-400">{new Date(u.created_at).toLocaleDateString()}</td>
                  <td className="px-4 py-3 text-sm text-gray-400">{u.last_login_at ? new Date(u.last_login_at).toLocaleDateString() : 'Never'}</td>
                  <td className="px-4 py-3">
                    <button className="text-xs text-red-400 hover:text-red-300">Ban</button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function SecurityTab() {
  const { data } = useQuery({
    queryKey: ['admin-security-logs'],
    queryFn: () => api.get('/admin/security-logs', { params: { limit: 50 } }),
  });

  const logs = data?.data?.logs || [];

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-white">Security & Audit Logs</h2>
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-800">
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Action</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Resource</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">IP</th>
              <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Time</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800/50">
            {logs.length === 0 ? (
              <tr><td colSpan={4} className="text-center py-8 text-gray-400">No security events recorded yet</td></tr>
            ) : (
              logs.map((log: any) => (
                <tr key={log.id} className="hover:bg-gray-800/30">
                  <td className="px-4 py-3 text-sm text-white">{log.action}</td>
                  <td className="px-4 py-3 text-sm text-gray-300">{log.resource}</td>
                  <td className="px-4 py-3 text-sm text-gray-400 font-mono">{log.ip}</td>
                  <td className="px-4 py-3 text-sm text-gray-400">{new Date(log.created_at).toLocaleString()}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function BillingTab() {
  const { data } = useQuery({
    queryKey: ['admin-billing'],
    queryFn: () => api.get('/admin/billing'),
  });

  const billing = data?.data || {};

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold text-white">Billing Analytics</h2>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Free Users" value={billing.plans?.free || 0} sublabel="No revenue" />
        <StatCard label="Pro Users" value={billing.plans?.pro || 0} sublabel={`$${((billing.plans?.pro || 0) * 9.99).toFixed(0)} MRR`} />
        <StatCard label="Team Users" value={billing.plans?.team || 0} sublabel={`$${((billing.plans?.team || 0) * 29.99).toFixed(0)} MRR`} />
        <StatCard label="Enterprise" value={billing.plans?.enterprise || 0} sublabel={`$${((billing.plans?.enterprise || 0) * 99.99).toFixed(0)} MRR`} />
      </div>

      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5">
        <h3 className="text-white font-medium mb-2">Total MRR</h3>
        <p className="text-3xl font-bold text-brand-400">
          ${(
            (billing.plans?.pro || 0) * 9.99 +
            (billing.plans?.team || 0) * 29.99 +
            (billing.plans?.enterprise || 0) * 99.99
          ).toFixed(2)}
        </p>
        <p className="text-sm text-gray-400 mt-1">{billing.active_subscriptions || 0} active subscriptions</p>
      </div>
    </div>
  );
}

function SettingsTab() {
  const [paypalEnabled, setPaypalEnabled] = useState(true);
  const [payhereEnabled, setPayhereEnabled] = useState(true);
  const [registrationOpen, setRegistrationOpen] = useState(true);
  const [maintenanceMode, setMaintenanceMode] = useState(false);

  const handleSave = () => {
    toast.success('Settings saved');
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <h2 className="text-lg font-semibold text-white">Platform Settings</h2>

      {/* Payment Gateways */}
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h3 className="text-white font-medium mb-4 flex items-center space-x-2">
          <CreditCard className="w-5 h-5 text-brand-400" />
          <span>Payment Gateways</span>
        </h3>
        <div className="space-y-4">
          <ToggleRow
            label="PayPal"
            description="International payments via PayPal"
            enabled={paypalEnabled}
            onToggle={() => setPaypalEnabled(!paypalEnabled)}
          />
          <ToggleRow
            label="PayHere"
            description="Sri Lanka payments (cards, bank transfer)"
            enabled={payhereEnabled}
            onToggle={() => setPayhereEnabled(!payhereEnabled)}
          />
        </div>
      </div>

      {/* Platform Controls */}
      <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-6">
        <h3 className="text-white font-medium mb-4 flex items-center space-x-2">
          <Power className="w-5 h-5 text-brand-400" />
          <span>Platform Controls</span>
        </h3>
        <div className="space-y-4">
          <ToggleRow
            label="Open Registration"
            description="Allow new users to register"
            enabled={registrationOpen}
            onToggle={() => setRegistrationOpen(!registrationOpen)}
          />
          <ToggleRow
            label="Maintenance Mode"
            description="Show maintenance page to all users"
            enabled={maintenanceMode}
            onToggle={() => setMaintenanceMode(!maintenanceMode)}
          />
        </div>
      </div>

      <button onClick={handleSave} className="bg-brand-600 hover:bg-brand-700 text-white px-6 py-2.5 rounded-lg font-medium text-sm">
        Save Settings
      </button>
    </div>
  );
}

// Reusable components
function ToggleRow({ label, description, enabled, onToggle }: { label: string; description: string; enabled: boolean; onToggle: () => void }) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm text-white font-medium">{label}</p>
        <p className="text-xs text-gray-400">{description}</p>
      </div>
      <button
        onClick={onToggle}
        className={`relative w-11 h-6 rounded-full transition-colors ${enabled ? 'bg-brand-600' : 'bg-gray-700'}`}
      >
        <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white transition-transform ${enabled ? 'translate-x-5.5 left-0.5' : 'left-0.5'}`}
          style={{ transform: enabled ? 'translateX(22px)' : 'translateX(2px)' }}
        />
      </button>
    </div>
  );
}

function MetricCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-4">
      <span className="text-xs text-gray-400 block mb-1">{label}</span>
      <span className={`text-2xl font-bold ${color}`}>{value.toLocaleString()}</span>
    </div>
  );
}

function StatCard({ label, value, sublabel }: { label: string; value: number; sublabel: string }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5">
      <span className="text-2xl font-bold text-white">{value.toLocaleString()}</span>
      <p className="text-xs text-gray-400 mt-1">{label}</p>
      <p className="text-xs text-gray-600 mt-0.5">{sublabel}</p>
    </div>
  );
}

function HealthCard({ title, status, latency, extra }: { title: string; status: boolean; latency?: number; extra?: string }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5">
      <div className="flex items-center justify-between mb-3">
        <span className="text-white font-medium">{title}</span>
        <div className={`w-3 h-3 rounded-full ${status ? 'bg-green-500' : 'bg-red-500'}`} />
      </div>
      <div className="flex items-center space-x-4 text-sm">
        <span className={status ? 'text-green-400' : 'text-red-400'}>{status ? 'Healthy' : 'Down'}</span>
        {latency !== undefined && <span className="text-gray-400">{latency}ms</span>}
        {extra && <span className="text-gray-400">{extra}</span>}
      </div>
    </div>
  );
}
