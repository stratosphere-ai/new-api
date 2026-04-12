import { useEffect, useState } from 'react';
import { api, type BuyerStats, type Trade, type ModelUsage } from '../api/marketplace';

function formatTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K';
  return n.toString();
}

function formatQuota(n: number): string {
  return (n / 500000).toFixed(4);
}

function formatDate(s: string): string {
  return new Date(s).toLocaleString();
}

export default function BuyerUsage() {
  const [stats, setStats] = useState<BuyerStats | null>(null);
  const [balance, setBalance] = useState(0);
  const [trades, setTrades] = useState<Trade[]>([]);
  const [modelUsage, setModelUsage] = useState<ModelUsage[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<'trades' | 'models'>('trades');

  useEffect(() => {
    async function load() {
      const [statsRes, tradesRes, modelsRes] = await Promise.all([
        api.getBuyerStats(),
        api.getBuyerTrades(),
        api.getBuyerModelUsage(),
      ]);
      if (statsRes.success) {
        setStats(statsRes.data.stats);
        setBalance(statsRes.data.balance);
      }
      if (tradesRes.success) setTrades(tradesRes.data || []);
      if (modelsRes.success) setModelUsage(modelsRes.data || []);
      setLoading(false);
    }
    load();
  }, []);

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
      {/* Stats cards */}
      <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card">
          <div className="label">Balance</div>
          <div className="value">${formatQuota(balance)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Total Spent</div>
          <div className="value">${stats ? formatQuota(stats.total_spent) : '0'}</div>
        </div>
        <div className="stat-card">
          <div className="label">Saved</div>
          <div className="value" style={{ color: 'var(--accent)' }}>
            ${stats ? formatQuota(stats.total_saved) : '0'}
          </div>
        </div>
        <div className="stat-card">
          <div className="label">Requests</div>
          <div className="value">{stats?.total_trades || 0}</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="card">
        <div style={{ display: 'flex', gap: 24, marginBottom: 24, borderBottom: '1px solid var(--border)', paddingBottom: 16 }}>
          <button
            onClick={() => setTab('trades')}
            style={{
              background: 'none', border: 'none', color: tab === 'trades' ? 'var(--accent)' : 'var(--text-secondary)',
              fontWeight: tab === 'trades' ? 700 : 400, fontSize: 15, cursor: 'pointer',
              borderBottom: tab === 'trades' ? '2px solid var(--accent)' : 'none', paddingBottom: 4,
            }}
          >
            Trade History
          </button>
          <button
            onClick={() => setTab('models')}
            style={{
              background: 'none', border: 'none', color: tab === 'models' ? 'var(--accent)' : 'var(--text-secondary)',
              fontWeight: tab === 'models' ? 700 : 400, fontSize: 15, cursor: 'pointer',
              borderBottom: tab === 'models' ? '2px solid var(--accent)' : 'none', paddingBottom: 4,
            }}
          >
            By Model
          </button>
        </div>

        {tab === 'trades' && (
          trades.length === 0 ? (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
              No trades yet. Use your API token to start making requests.
            </p>
          ) : (
            <table className="listing-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Model</th>
                  <th>Tokens</th>
                  <th>Discount</th>
                  <th>Cost</th>
                </tr>
              </thead>
              <tbody>
                {trades.map((t) => (
                  <tr key={t.id}>
                    <td style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{formatDate(t.created_at)}</td>
                    <td><strong>{t.model_name}</strong></td>
                    <td>{formatTokens(t.total_tokens)}</td>
                    <td style={{ color: 'var(--accent)' }}>{t.discount}% off</td>
                    <td>${formatQuota(t.buyer_amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}

        {tab === 'models' && (
          modelUsage.length === 0 ? (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
              No usage data yet.
            </p>
          ) : (
            <table className="listing-table">
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Requests</th>
                  <th>Tokens</th>
                  <th>Avg Discount</th>
                  <th>Total Cost</th>
                </tr>
              </thead>
              <tbody>
                {modelUsage.map((m) => (
                  <tr key={m.model_name}>
                    <td><strong>{m.model_name}</strong></td>
                    <td>{m.trade_count}</td>
                    <td>{formatTokens(m.total_tokens)}</td>
                    <td style={{ color: 'var(--accent)' }}>{m.avg_discount}% off</td>
                    <td>${formatQuota(m.total_spent)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}
      </div>
    </div>
  );
}
