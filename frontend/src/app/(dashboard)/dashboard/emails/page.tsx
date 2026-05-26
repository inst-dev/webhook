'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Mail, Paperclip, Clock, ChevronRight } from 'lucide-react';
import { api } from '@/lib/api';

export default function EmailsPage() {
  const [selectedEmail, setSelectedEmail] = useState<any>(null);
  const [viewMode, setViewMode] = useState<'html' | 'text' | 'raw'>('html');

  const { data, isLoading } = useQuery({
    queryKey: ['email-logs'],
    queryFn: () => api.get('/endpoints/email-logs', { params: { limit: 50 } }),
    refetchInterval: 5000,
  });

  const emails = data?.data?.emails || [];

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="border-b border-gray-800/50 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-xl font-bold text-white">Email Hooks</h1>
              <p className="text-sm text-gray-400 mt-1">Captured inbound emails</p>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6">
        {/* Usage Info */}
        <div className="mb-6 bg-gray-900/50 border border-gray-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-gray-300 mb-2">How to use Email Hooks</h3>
          <p className="text-xs text-gray-400 mb-2">
            Send emails to your endpoint token address to capture them:
          </p>
          <code className="text-xs text-brand-400 font-mono bg-gray-800 px-3 py-1.5 rounded block">
            YOUR_TOKEN@emailhook.webhook.inst.lk
          </code>
        </div>

        <div className="flex h-[calc(100vh-220px)]">
          {/* Email List */}
          <div className="w-1/3 border-r border-gray-800 overflow-y-auto">
            {isLoading ? (
              <div className="p-8 text-center text-gray-400">Loading...</div>
            ) : emails.length === 0 ? (
              <div className="text-center py-20 px-4">
                <Mail className="w-10 h-10 text-gray-600 mx-auto mb-3" />
                <p className="text-gray-400 text-sm">No emails received yet</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-800/50">
                {emails.map((email: any) => (
                  <button
                    key={email.id}
                    onClick={() => setSelectedEmail(email)}
                    className={`w-full text-left p-4 hover:bg-gray-900/50 transition-colors ${
                      selectedEmail?.id === email.id ? 'bg-gray-900/70 border-l-2 border-brand-500' : ''
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-white font-medium truncate">{email.from}</span>
                      <ChevronRight className="w-3 h-3 text-gray-500" />
                    </div>
                    <p className="text-sm text-gray-300 mt-1 truncate">{email.subject || '(no subject)'}</p>
                    <div className="flex items-center space-x-3 mt-2 text-xs text-gray-500">
                      <span className="flex items-center space-x-1">
                        <Clock className="w-3 h-3" />
                        <span>{new Date(email.created_at).toLocaleTimeString()}</span>
                      </span>
                      {email.attachments && JSON.parse(email.attachments).length > 0 && (
                        <span className="flex items-center space-x-1">
                          <Paperclip className="w-3 h-3" />
                          <span>{JSON.parse(email.attachments).length}</span>
                        </span>
                      )}
                      <span>{Math.round(email.size / 1024)}KB</span>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Email Detail */}
          <div className="flex-1 overflow-y-auto">
            {selectedEmail ? (
              <div className="p-6">
                <div className="mb-4">
                  <h2 className="text-lg font-semibold text-white">{selectedEmail.subject || '(no subject)'}</h2>
                  <div className="mt-2 space-y-1 text-sm">
                    <p className="text-gray-400"><span className="text-gray-500">From:</span> {selectedEmail.from}</p>
                    <p className="text-gray-400"><span className="text-gray-500">To:</span> {selectedEmail.to}</p>
                    <p className="text-gray-400"><span className="text-gray-500">IP:</span> {selectedEmail.source_ip}</p>
                    <p className="text-gray-400"><span className="text-gray-500">Received:</span> {new Date(selectedEmail.created_at).toLocaleString()}</p>
                  </div>
                </div>

                {/* View Mode Toggle */}
                <div className="flex space-x-1 bg-gray-900/50 rounded-lg p-1 mb-4 w-fit">
                  {(['html', 'text', 'raw'] as const).map((mode) => (
                    <button
                      key={mode}
                      onClick={() => setViewMode(mode)}
                      className={`px-3 py-1.5 rounded-md text-xs font-medium capitalize ${
                        viewMode === mode ? 'bg-gray-800 text-white' : 'text-gray-400 hover:text-white'
                      }`}
                    >
                      {mode}
                    </button>
                  ))}
                </div>

                {/* Email Body */}
                <div className="bg-gray-900 rounded-lg p-4 overflow-auto max-h-[500px]">
                  {viewMode === 'html' && selectedEmail.html_body ? (
                    <div className="prose prose-invert prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: selectedEmail.html_body }} />
                  ) : viewMode === 'text' ? (
                    <pre className="text-sm text-gray-300 font-mono whitespace-pre-wrap">{selectedEmail.body || 'No text body'}</pre>
                  ) : (
                    <pre className="text-xs text-gray-400 font-mono whitespace-pre-wrap">{JSON.stringify(JSON.parse(selectedEmail.headers || '{}'), null, 2)}</pre>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center h-full text-gray-400">
                <p>Select an email to view</p>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
