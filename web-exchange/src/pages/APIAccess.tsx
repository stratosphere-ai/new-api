import { useEffect, useState } from 'react';
import { api, type APIToken } from '../api/marketplace';

export default function APIAccess() {
  const [tokens, setTokens] = useState<APIToken[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [copied, setCopied] = useState(false);

  async function load() {
    const res = await api.getBuyerAPIInfo();
    if (res.success) {
      setTokens(res.data.tokens || []);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  async function handleCreate() {
    setCreating(true);
    const res = await api.createBuyerAPIToken();
    setCreating(false);
    if (res.success) {
      setNewKey(res.data.key);
      load();
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  if (loading) {
    return (
      <div className="container">
        <div className="card" style={{ textAlign: 'center', padding: 48 }}>
          <span className="loading" /> Loading...
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <div className="card-header">
          <h2>API Access</h2>
        </div>

        {/* Quick start guide */}
        <div style={{ marginBottom: 32 }}>
          <div className="step-label">Quick Start</div>
          <div style={{
            background: 'var(--bg-input)', borderRadius: 12, padding: 20,
            fontFamily: 'monospace', fontSize: 13, lineHeight: 1.8, overflowX: 'auto',
          }}>
            <div style={{ color: 'var(--text-muted)' }}># Use any OpenAI-compatible client</div>
            <div>
              <span style={{ color: 'var(--accent)' }}>curl</span> {window.location.origin}/v1/chat/completions \
            </div>
            <div>&nbsp; -H <span style={{ color: '#f59e0b' }}>"Authorization: Bearer YOUR_TOKEN"</span> \</div>
            <div>&nbsp; -H <span style={{ color: '#f59e0b' }}>"Content-Type: application/json"</span> \</div>
            <div>&nbsp; -d <span style={{ color: '#f59e0b' }}>'{`{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"Hello"}]}`}'</span></div>
          </div>
        </div>

        {/* New key alert */}
        {newKey && (
          <div className="validation-msg success" style={{ marginBottom: 24 }}>
            <div style={{ fontWeight: 700, marginBottom: 8 }}>API Token created! Copy it now — it won't be shown again.</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <code style={{ flex: 1, wordBreak: 'break-all' }}>{newKey}</code>
              <button className="action-btn" onClick={() => copyToClipboard(newKey)}>
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
        )}

        {/* Existing tokens */}
        <div className="step-label">Your API Tokens</div>

        {tokens.length === 0 ? (
          <p style={{ color: 'var(--text-secondary)', marginBottom: 24 }}>
            No marketplace API tokens yet. Create one to start using the API.
          </p>
        ) : (
          <table className="listing-table" style={{ marginBottom: 24 }}>
            <thead>
              <tr>
                <th>Name</th>
                <th>Key</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((t) => (
                <tr key={t.id}>
                  <td>{t.name}</td>
                  <td>
                    <code style={{ fontSize: 13, color: 'var(--text-muted)' }}>
                      {t.key.slice(0, 8)}...{t.key.slice(-4)}
                    </code>
                  </td>
                  <td>
                    <span className={t.status === 1 ? 'status-active' : 'status-paused'}>
                      {t.status === 1 ? 'Active' : 'Disabled'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        <button className="submit-btn" onClick={handleCreate} disabled={creating}>
          {creating && <span className="loading" />}
          {creating ? 'Creating...' : '+ Create New API Token'}
        </button>
      </div>

      {/* Supported models */}
      <div className="card">
        <h3 style={{ marginBottom: 16 }}>Supported Models</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
          <div>
            <div style={{ fontWeight: 700, marginBottom: 8 }}>Anthropic</div>
            <div style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.8 }}>
              claude-sonnet-4-20250514<br />
              claude-opus-4-20250514<br />
              claude-haiku-4-5-20251001<br />
              claude-3-5-sonnet-20241022
            </div>
          </div>
          <div>
            <div style={{ fontWeight: 700, marginBottom: 8 }}>OpenAI</div>
            <div style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.8 }}>
              gpt-4o<br />
              gpt-4o-mini<br />
              gpt-4-turbo<br />
              o1, o3-mini
            </div>
          </div>
          <div>
            <div style={{ fontWeight: 700, marginBottom: 8 }}>Google</div>
            <div style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.8 }}>
              gemini-2.5-pro<br />
              gemini-2.5-flash<br />
              gemini-2.0-flash<br />
              gemini-1.5-pro
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
