import { useEffect, useState } from 'react';
import { api } from '../api/marketplace';

interface Withdrawal {
  id: number;
  seller_id: number;
  amount: number;
  status: number;
  period_start: string;
  period_end: string;
  created_at: string;
  paid_at: string | null;
}

interface Trade {
  id: number;
  model_name: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  discount: number;
  seller_amount: number;
  created_at: string;
}

const STATUS_MAP: Record<number, { text: string; cls: string }> = {
  1: { text: 'Pending', cls: 'status-paused' },
  2: { text: 'Approved', cls: 'status-active' },
  3: { text: 'Paid', cls: 'status-active' },
  4: { text: 'Rejected', cls: 'status-exhausted' },
};

function formatQuota(n: number): string {
  return '$' + (n / 500000).toFixed(4);
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K';
  return n.toString();
}

function formatDate(s: string): string {
  if (!s) return '-';
  return new Date(s).toLocaleDateString();
}

function formatDateTime(s: string): string {
  if (!s) return '-';
  return new Date(s).toLocaleString();
}

// Extended API client for seller-specific endpoints
async function request<T>(url: string, options?: RequestInit): Promise<{ success: boolean; message: string; data: T }> {
  const token = localStorage.getItem('token') || '';
  const res = await fetch('/api/marketplace' + url, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });
  return res.json();
}

export default function SellerWithdrawals() {
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([]);
  const [trades, setTrades] = useState<Trade[]>([]);
  const [balance, setBalance] = useState(0);
  const [totalEarned, setTotalEarned] = useState(0);
  const [loading, setLoading] = useState(true);
  const [withdrawing, setWithdrawing] = useState(false);
  const [tab, setTab] = useState<'withdrawals' | 'trades'>('withdrawals');

  async function load() {
    const [wRes, tRes] = await Promise.all([
      request<{ withdrawals: Withdrawal[]; balance: number; total_earned: number }>('/seller/withdrawals'),
      request<Trade[]>('/seller/trades'),
    ]);
    if (wRes.success) {
      setWithdrawals(wRes.data.withdrawals || []);
      setBalance(wRes.data.balance);
      setTotalEarned(wRes.data.total_earned);
    }
    if (tRes.success) setTrades(tRes.data || []);
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  async function handleRequestWithdrawal() {
    if (!confirm('Request withdrawal of your entire available balance?')) return;
    setWithdrawing(true);
    await request('/seller/withdrawal/request', { method: 'POST' });
    await load();
    setWithdrawing(false);
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

  // Calculate stats
  const totalPaid = withdrawals
    .filter((w) => w.status === 3)
    .reduce((sum, w) => sum + w.amount, 0);
  const totalPending = withdrawals
    .filter((w) => w.status === 1 || w.status === 2)
    .reduce((sum, w) => sum + w.amount, 0);

  return (
    <div className="container">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2>Earnings & Withdrawals</h2>
        {balance > 0 && (
          <button
            className="submit-btn"
            style={{ width: 'auto', padding: '10px 24px', marginTop: 0 }}
            onClick={handleRequestWithdrawal}
            disabled={withdrawing}
          >
            {withdrawing && <span className="loading" />}
            {withdrawing ? 'Requesting...' : `Withdraw ${formatQuota(balance)}`}
          </button>
        )}
      </div>

      {/* Stats */}
      <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="stat-card">
          <div className="label">Available Balance</div>
          <div className="value" style={{ color: 'var(--accent)' }}>{formatQuota(balance)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Total Earned</div>
          <div className="value">{formatQuota(totalEarned)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Total Paid Out</div>
          <div className="value">{formatQuota(totalPaid)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Pending Payout</div>
          <div className="value">{formatQuota(totalPending)}</div>
        </div>
      </div>

      {/* Info banner */}
      <div className="card" style={{ padding: 16, marginBottom: 24, background: 'var(--accent-dim)' }}>
        <p style={{ color: 'var(--accent)', fontSize: 14, margin: 0 }}>
          Withdrawals are processed monthly on the 1st. Your accumulated balance is automatically
          converted to a withdrawal request pending admin approval.
        </p>
      </div>

      {/* Tabs */}
      <div className="card">
        <div style={{ display: 'flex', gap: 24, marginBottom: 24, borderBottom: '1px solid var(--border)', paddingBottom: 16 }}>
          <button
            onClick={() => setTab('withdrawals')}
            style={{
              background: 'none', border: 'none', color: tab === 'withdrawals' ? 'var(--accent)' : 'var(--text-secondary)',
              fontWeight: tab === 'withdrawals' ? 700 : 400, fontSize: 15, cursor: 'pointer',
              borderBottom: tab === 'withdrawals' ? '2px solid var(--accent)' : 'none', paddingBottom: 4,
            }}
          >
            Withdrawal History
          </button>
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
        </div>

        {tab === 'withdrawals' && (
          withdrawals.length === 0 ? (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
              No withdrawals yet. Your first withdrawal will be created automatically at the end of the month.
            </p>
          ) : (
            <table className="listing-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Period</th>
                  <th>Amount</th>
                  <th>Status</th>
                  <th>Paid Date</th>
                </tr>
              </thead>
              <tbody>
                {withdrawals.map((w) => {
                  const status = STATUS_MAP[w.status] || { text: 'Unknown', cls: '' };
                  return (
                    <tr key={w.id}>
                      <td>#{w.id}</td>
                      <td style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
                        {formatDate(w.period_start)} ~ {formatDate(w.period_end)}
                      </td>
                      <td><strong>{formatQuota(w.amount)}</strong></td>
                      <td><span className={status.cls}>{status.text}</span></td>
                      <td style={{ fontSize: 13, color: 'var(--text-muted)' }}>
                        {w.paid_at ? formatDate(w.paid_at) : '-'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )
        )}

        {tab === 'trades' && (
          trades.length === 0 ? (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
              No trades yet. Once buyers start using your keys, trades will appear here.
            </p>
          ) : (
            <table className="listing-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Model</th>
                  <th>Tokens</th>
                  <th>Discount</th>
                  <th>Your Earnings</th>
                </tr>
              </thead>
              <tbody>
                {trades.map((t) => (
                  <tr key={t.id}>
                    <td style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{formatDateTime(t.created_at)}</td>
                    <td><strong>{t.model_name}</strong></td>
                    <td>{formatTokens(t.total_tokens)}</td>
                    <td>{t.discount}% off</td>
                    <td style={{ color: 'var(--accent)' }}>{formatQuota(t.seller_amount)}</td>
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
