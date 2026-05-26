'use client';

import { useQuery } from '@tanstack/react-query';
import { Globe, Clock, ArrowUpRight } from 'lucide-react';
import { api } from '@/lib/api';

export default function DNSPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['dns-logs'],
    queryFn: () => api.get('/endpoints/dns-logs', { params: { limit: 50 } }),
    refetchInterval: 5000,
  });

  const logs = data?.data?.logs || [];

  const queryTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      A: 'text-green-400 bg-green-400/10',
      AAAA: 'text-blue-400 bg-blue-400/10',
      TXT: 'text-yellow-400 bg-yellow-400/10',
      MX: 'text-purple-400 bg-purple-400/10',
      CNAME: 'text-orange-400 bg-orange-400/10',
      NS: 'text-cyan-400 bg-cyan-400/10',
    };
    return colors[type] || 'text-gray-400 bg-gray-400/10';
  };

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-xl font-bold text-white">DNS Interaction Logs</h1>
              <p className="text-sm text-gray-400 mt-1">Monitor DNS queries to your tokens</p>
            </div>
            <div className="flex items-center space-x-2 text-sm text-gray-400">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              <span>Live</span>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6">
        {/* Usage Info */}
        <div className="mb-6 bg-gray-900/50 border border-gray-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-gray-300 mb-2">How to use DNS Hooks</h3>
          <p className="text-xs text-gray-400 mb-2">
            Send DNS queries to any subdomain of your endpoint token to log interactions:
          </p>
          <code className="text-xs text-brand-400 font-mono bg-gray-800 px-3 py-1.5 rounded block">
            nslookup test.YOUR_TOKEN.dns.webhook.inst.lk
          </code>
        </div>

        {/* Logs Table */}
        {isLoading ? (
          <div className="text-center py-20 text-gray-400">Loading DNS logs...</div>
        ) : logs.length === 0 ? (
          <div className="text-center py-20">
            <Globe className="w-12 h-12 text-gray-600 mx-auto mb-4" />
            <h2 className="text-lg font-semibold text-white mb-2">No DNS queries yet</h2>
            <p className="text-gray-400 text-sm">DNS interactions will appear here in real-time.</p>
          </div>
        ) : (
          <div className="bg-gray-900/50 border border-gray-800 rounded-xl overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-800">
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Type</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Query Name</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Source IP</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Port</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-gray-400 uppercase">Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800/50">
                {logs.map((log: any) => (
                  <tr key={log.id} className="hover:bg-gray-800/30 transition-colors">
                    <td className="px-4 py-3">
                      <span className={`text-xs font-mono font-bold px-2 py-0.5 rounded ${queryTypeColor(log.query_type)}`}>
                        {log.query_type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-white font-mono">{log.query_name}</td>
                    <td className="px-4 py-3 text-sm text-gray-300 font-mono">{log.source_ip}</td>
                    <td className="px-4 py-3 text-sm text-gray-400">{log.source_port}</td>
                    <td className="px-4 py-3 text-sm text-gray-400 flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>{new Date(log.created_at).toLocaleTimeString()}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </div>
  );
}
