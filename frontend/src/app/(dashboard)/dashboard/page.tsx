'use client';

import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Copy, Trash2, Globe, Activity, ExternalLink } from 'lucide-react';
import { endpointsAPI } from '@/lib/api';
import { wsClient } from '@/lib/websocket';
import { useAuthStore } from '@/store/auth';
import toast from 'react-hot-toast';
import Link from 'next/link';

export default function DashboardPage() {
  const { user, accessToken } = useAuthStore();
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['endpoints'],
    queryFn: () => endpointsAPI.list({ limit: 50 }),
  });

  const createMutation = useMutation({
    mutationFn: (name: string) => endpointsAPI.create({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['endpoints'] });
      setShowCreate(false);
      setNewName('');
      toast.success('Endpoint created!');
    },
    onError: () => toast.error('Failed to create endpoint'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => endpointsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['endpoints'] });
      toast.success('Endpoint deleted');
    },
  });

  useEffect(() => {
    if (accessToken) {
      wsClient.connect(accessToken);
      const unsub = wsClient.on('new_request', () => {
        queryClient.invalidateQueries({ queryKey: ['endpoints'] });
      });
      return () => {
        unsub();
        wsClient.disconnect();
      };
    }
  }, [accessToken, queryClient]);

  const endpoints = data?.data?.endpoints || [];
  const domain = process.env.NEXT_PUBLIC_DOMAIN || 'webhook.inst.lk';

  const copyUrl = (token: string) => {
    navigator.clipboard.writeText(`https://${domain}/${token}`);
    toast.success('URL copied to clipboard');
  };

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-white">Dashboard</h1>
            <p className="text-sm text-gray-400">Welcome back, {user?.display_name}</p>
          </div>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center space-x-2 bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            <Plus className="w-4 h-4" />
            <span>New Endpoint</span>
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Create Modal */}
        {showCreate && (
          <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-full max-w-md">
              <h2 className="text-lg font-semibold text-white mb-4">Create New Endpoint</h2>
              <form onSubmit={(e) => { e.preventDefault(); createMutation.mutate(newName); }}>
                <input
                  type="text"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="Endpoint name (optional)"
                  className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:ring-2 focus:ring-brand-500 outline-none mb-4"
                />
                <div className="flex justify-end space-x-3">
                  <button type="button" onClick={() => setShowCreate(false)} className="text-gray-400 hover:text-white px-4 py-2">
                    Cancel
                  </button>
                  <button type="submit" className="bg-brand-600 hover:bg-brand-700 text-white px-4 py-2 rounded-lg font-medium">
                    Create
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Endpoints List */}
        {isLoading ? (
          <div className="text-center py-20 text-gray-400">Loading endpoints...</div>
        ) : endpoints.length === 0 ? (
          <div className="text-center py-20">
            <Globe className="w-12 h-12 text-gray-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-white mb-2">No endpoints yet</h2>
            <p className="text-gray-400 mb-6">Create your first webhook endpoint to start capturing requests.</p>
            <button
              onClick={() => setShowCreate(true)}
              className="bg-brand-600 hover:bg-brand-700 text-white px-6 py-2.5 rounded-lg font-medium"
            >
              Create Endpoint
            </button>
          </div>
        ) : (
          <div className="grid gap-4">
            {endpoints.map((endpoint: any) => (
              <div key={endpoint.id} className="bg-gray-900/50 border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-colors">
                <div className="flex items-center justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-3">
                      <h3 className="text-white font-medium truncate">
                        {endpoint.name || endpoint.token}
                      </h3>
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                        endpoint.is_active ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'
                      }`}>
                        {endpoint.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                    <div className="flex items-center mt-2 space-x-2">
                      <code className="text-sm text-brand-400 font-mono truncate">
                        https://{domain}/{endpoint.token}
                      </code>
                      <button onClick={() => copyUrl(endpoint.token)} className="text-gray-500 hover:text-brand-500 flex-shrink-0" title="Copy URL">
                        <Copy className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                  <div className="flex items-center space-x-4 ml-4">
                    <div className="text-right">
                      <div className="flex items-center space-x-1 text-gray-400">
                        <Activity className="w-3.5 h-3.5" />
                        <span className="text-sm">{endpoint.request_count} requests</span>
                      </div>
                    </div>
                    <Link href={`/dashboard/endpoints/${endpoint.id}`} className="text-gray-400 hover:text-brand-500">
                      <ExternalLink className="w-4 h-4" />
                    </Link>
                    <button onClick={() => deleteMutation.mutate(endpoint.id)} className="text-gray-400 hover:text-red-500">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
