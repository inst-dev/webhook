'use client';

import { useQuery } from '@tanstack/react-query';
import { Activity, Users, Globe, Mail, Server, Database, Wifi, Clock } from 'lucide-react';
import { api } from '@/lib/api';

export default function AdminConsolePage() {
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
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-white">Admin Console</h1>
            <p className="text-sm text-gray-400">console.webhook.inst.lk</p>
          </div>
          <div className="flex items-center space-x-2">
            <div className={`w-2.5 h-2.5 rounded-full ${health.status === 'healthy' ? 'bg-green-500' : 'bg-yellow-500'} animate-pulse`} />
            <span className="text-sm text-gray-300 capitalize">{health.status || 'checking...'}</span>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 space-y-8">
        {/* Realtime Metrics */}
        <section>
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Realtime Traffic</h2>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <MetricCard label="Req/sec" value={metrics.requests_per_second || 0} icon={Activity} color="text-green-400" />
            <MetricCard label="Req/min" value={metrics.requests_per_minute || 0} icon={Activity} color="text-blue-400" />
            <MetricCard label="Req/hour" value={metrics.requests_per_hour || 0} icon={Clock} color="text-purple-400" />
            <MetricCard label="Req/day" value={metrics.requests_per_day || 0} icon={Clock} color="text-orange-400" />
            <MetricCard label="WebSockets" value={metrics.active_websockets || 0} icon={Wifi} color="text-cyan-400" />
          </div>
        </section>

        {/* Platform Stats */}
        <section>
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Platform Overview</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard label="Total Users" value={stats.users?.total || 0} sublabel={`${stats.users?.active_today || 0} active today`} icon={Users} />
            <StatCard label="Endpoints" value={stats.endpoints?.total || 0} sublabel={`${stats.endpoints?.active || 0} active`} icon={Globe} />
            <StatCard label="Requests Today" value={stats.requests?.today || 0} sublabel={`${stats.requests?.total || 0} total`} icon={Activity} />
            <StatCard label="Emails Today" value={stats.emails?.today || 0} sublabel={`DNS: ${stats.dns?.today || 0}`} icon={Mail} />
          </div>
        </section>

        {/* System Health */}
        <section>
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">System Health</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <HealthCard
              title="PostgreSQL"
              icon={Database}
              status={health.services?.database?.status}
              latency={health.services?.database?.latency_ms}
            />
            <HealthCard
              title="Redis"
              icon={Server}
              status={health.services?.redis?.status}
              latency={health.services?.redis?.latency_ms}
            />
            <HealthCard
              title="WebSocket Hub"
              icon={Wifi}
              status={true}
              extra={`${health.services?.websocket?.active_connections || 0} connections`}
            />
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
      </main>
    </div>
  );
}

function MetricCard({ label, value, icon: Icon, color }: { label: string; value: number; icon: any; color: string }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-4">
      <div className="flex items-center space-x-2 mb-2">
        <Icon className={`w-4 h-4 ${color}`} />
        <span className="text-xs text-gray-400">{label}</span>
      </div>
      <span className={`text-2xl font-bold ${color}`}>{value.toLocaleString()}</span>
    </div>
  );
}

function StatCard({ label, value, sublabel, icon: Icon }: { label: string; value: number; sublabel: string; icon: any }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5">
      <div className="flex items-center justify-between mb-3">
        <Icon className="w-5 h-5 text-gray-400" />
      </div>
      <span className="text-2xl font-bold text-white">{value.toLocaleString()}</span>
      <p className="text-xs text-gray-500 mt-1">{label}</p>
      <p className="text-xs text-gray-600 mt-0.5">{sublabel}</p>
    </div>
  );
}

function HealthCard({ title, icon: Icon, status, latency, extra }: { title: string; icon: any; status: boolean; latency?: number; extra?: string }) {
  return (
    <div className="bg-gray-900/50 border border-gray-800 rounded-xl p-5">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-3">
          <Icon className="w-5 h-5 text-gray-400" />
          <span className="text-white font-medium">{title}</span>
        </div>
        <div className={`w-3 h-3 rounded-full ${status ? 'bg-green-500' : 'bg-red-500'}`} />
      </div>
      <div className="flex items-center space-x-4 text-sm">
        <span className={status ? 'text-green-400' : 'text-red-400'}>
          {status ? 'Healthy' : 'Down'}
        </span>
        {latency !== undefined && (
          <span className="text-gray-400">{latency}ms</span>
        )}
        {extra && <span className="text-gray-400">{extra}</span>}
      </div>
    </div>
  );
}
