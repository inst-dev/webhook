'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Copy, ArrowLeft, Clock, Globe, Code, Send } from 'lucide-react';
import { endpointsAPI, requestsAPI } from '@/lib/api';
import { wsClient } from '@/lib/websocket';
import { useAuthStore } from '@/store/auth';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function EndpointDetailPage() {
  const params = useParams();
  const endpointId = params.id as string;
  const { accessToken } = useAuthStore();
  const [selectedRequest, setSelectedRequest] = useState<any>(null);
  const [requests, setRequests] = useState<any[]>([]);

  const { data: endpointData } = useQuery({
    queryKey: ['endpoint', endpointId],
    queryFn: () => endpointsAPI.get(endpointId),
  });

  const { data: requestsData, refetch } = useQuery({
    queryKey: ['requests', endpointId],
    queryFn: () => requestsAPI.list(endpointId, { limit: 50 }),
    refetchInterval: 5000,
  });

  useEffect(() => {
    if (requestsData?.data?.requests) {
      setRequests(requestsData.data.requests);
    }
  }, [requestsData]);

  useEffect(() => {
    if (accessToken) {
      try {
        wsClient.connect(accessToken, endpointId);
        const unsub = wsClient.on('new_request', (data: any) => {
          setRequests((prev) => [data, ...prev]);
          refetch();
        });
        return () => {
          unsub();
          wsClient.disconnect();
        };
      } catch (e) {
        // WebSocket connection failed - non-critical, page still works
        console.warn('WebSocket connection failed:', e);
      }
    }
  }, [accessToken, endpointId, refetch]);

  const endpoint = endpointData?.data;
  const domain = process.env.NEXT_PUBLIC_DOMAIN || 'webhook.inst.lk';

  const safeJsonStringify = (data: any): string => {
    if (!data) return '{}';
    try {
      if (typeof data === 'string') {
        return JSON.stringify(JSON.parse(data), null, 2);
      }
      return JSON.stringify(data, null, 2);
    } catch {
      return String(data);
    }
  };

  const formatBody = (body: any, contentType: string) => {
    if (!body) return 'No body';
    try {
      if (typeof body === 'object') {
        return JSON.stringify(body, null, 2);
      }
      if (contentType?.includes('json')) {
        return JSON.stringify(JSON.parse(body), null, 2);
      }
      return String(body);
    } catch {
      return String(body);
    }
  };

  const methodColor = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'text-green-400 bg-green-400/10',
      POST: 'text-blue-400 bg-blue-400/10',
      PUT: 'text-yellow-400 bg-yellow-400/10',
      PATCH: 'text-orange-400 bg-orange-400/10',
      DELETE: 'text-red-400 bg-red-400/10',
    };
    return colors[method] || 'text-gray-400 bg-gray-400/10';
  };

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center space-x-4">
            <Link href="/dashboard" className="text-gray-400 hover:text-white">
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div className="flex-1">
              <h1 className="text-lg font-bold text-white">{endpoint?.name || endpoint?.token}</h1>
              <div className="flex items-center space-x-2 mt-1">
                <code className="text-sm text-gray-400 font-mono">https://{domain}/{endpoint?.token}</code>
                <button
                  onClick={() => { navigator.clipboard.writeText(`https://${domain}/${endpoint?.token}`); toast.success('Copied!'); }}
                  className="text-gray-500 hover:text-brand-500"
                >
                  <Copy className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <div className="flex items-center space-x-2 text-sm text-gray-400">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse-dot" />
              <span>Live</span>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto flex h-[calc(100vh-80px)]">
        {/* Request List */}
        <div className="w-1/3 border-r border-gray-800 overflow-y-auto">
          {requests.length === 0 ? (
            <div className="text-center py-20 px-4">
              <Globe className="w-10 h-10 text-gray-600 mx-auto mb-3" />
              <p className="text-gray-400 text-sm">Waiting for requests...</p>
              <p className="text-gray-500 text-xs mt-2">Send a request to your endpoint URL</p>
            </div>
          ) : (
            <div className="divide-y divide-gray-800/50">
              {requests.map((req: any) => (
                <button
                  key={req.id}
                  onClick={() => setSelectedRequest(req)}
                  className={`w-full text-left p-4 hover:bg-gray-900/50 transition-colors ${
                    selectedRequest?.id === req.id ? 'bg-gray-900/70 border-l-2 border-brand-500' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className={`text-xs font-mono font-bold px-1.5 py-0.5 rounded ${methodColor(req.method)}`}>
                      {req.method}
                    </span>
                    <span className="text-xs text-gray-500">
                      {new Date(req.created_at).toLocaleTimeString()}
                    </span>
                  </div>
                  <div className="mt-1.5 text-sm text-gray-400 truncate">{req.source_ip}</div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Request Detail */}
        <div className="flex-1 overflow-y-auto p-6">
          {selectedRequest ? (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span className={`text-sm font-mono font-bold px-2 py-1 rounded ${methodColor(selectedRequest.method)}`}>
                    {selectedRequest.method}
                  </span>
                  <span className="text-white font-mono text-sm">{selectedRequest.path}</span>
                </div>
                <div className="flex items-center space-x-2 text-xs text-gray-400">
                  <Clock className="w-3.5 h-3.5" />
                  <span>{new Date(selectedRequest.created_at).toLocaleString()}</span>
                </div>
              </div>

              {/* Headers */}
              <section>
                <h3 className="text-sm font-semibold text-gray-300 mb-2 flex items-center space-x-2">
                  <Code className="w-4 h-4" />
                  <span>Headers</span>
                </h3>
                <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
                  <pre className="text-sm text-gray-300 font-mono whitespace-pre-wrap">
                    {safeJsonStringify(selectedRequest.headers)}
                  </pre>
                </div>
              </section>

              {/* Body */}
              <section>
                <h3 className="text-sm font-semibold text-gray-300 mb-2 flex items-center space-x-2">
                  <Send className="w-4 h-4" />
                  <span>Body</span>
                </h3>
                <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
                  <pre className="text-sm text-gray-300 font-mono whitespace-pre-wrap">
                    {formatBody(selectedRequest.body, selectedRequest.content_type)}
                  </pre>
                </div>
              </section>

              {/* Metadata */}
              <section className="grid grid-cols-2 gap-4">
                <div className="bg-gray-900 rounded-lg p-4">
                  <span className="text-xs text-gray-500 block mb-1">Source IP</span>
                  <span className="text-sm text-white font-mono">{selectedRequest.source_ip}</span>
                </div>
                <div className="bg-gray-900 rounded-lg p-4">
                  <span className="text-xs text-gray-500 block mb-1">Content Type</span>
                  <span className="text-sm text-white font-mono">{selectedRequest.content_type || 'N/A'}</span>
                </div>
                <div className="bg-gray-900 rounded-lg p-4">
                  <span className="text-xs text-gray-500 block mb-1">Content Length</span>
                  <span className="text-sm text-white font-mono">{selectedRequest.content_length} bytes</span>
                </div>
                <div className="bg-gray-900 rounded-lg p-4">
                  <span className="text-xs text-gray-500 block mb-1">Response Code</span>
                  <span className="text-sm text-white font-mono">{selectedRequest.response_code}</span>
                </div>
              </section>
            </div>
          ) : (
            <div className="flex items-center justify-center h-full text-gray-400">
              <p>Select a request to inspect</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
