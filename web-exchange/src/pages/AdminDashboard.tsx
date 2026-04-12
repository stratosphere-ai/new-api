import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { adminApi, type PlatformStats, type SellerWithUser } from '../api/admin';

function formatQuota(n: number): string {
  return (n / 500000).toFixed(2);
}

function formatTokens(n: number): string {
  if (n >= 1_000_000_000) return (n / 1_000_000_000).toFixed(1) + 'B';
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K';
  return n.toString();
}

export default function AdminDashboard() {
  const [stats, setStats] = useState<{ stats: PlatformStats; total_sellers: number; active_sellers: number; active_listings: number } | null>(null);
  const [sellers, setSellers] = useState<SellerWithUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  async function load() {
    const [statsRes, sellersRes] = await Promise.all([
      adminApi.getStats(),
      adminApi.getSellers(),
    ]);
    if (statsRes.success) setStats(statsRes.data);
    if (sellersRes.success) setSellers(sellersRes.data || []);
    setLoading(false);
  }

  useEffect(() => { load(); }, []);

  async function handleSuspend(id: number) {
    setActionLoading(id);
    await adminApi.suspendSeller(id);
    await load();
    setActionLoading(null);
  }

  async function handleEnable(id: number) {
    setActionLoading(id);
    await adminApi.enableSeller(id);
    await load();
    setActionLoading(null);
  }

  if (loading) {
    return (
      <div className="container">
        <div className="card" style={{ textAlign: 'center', padding: 48 }}>
          <span className="loading" /> Loading admin dashboard...
        </div>
      </div>
    );
  }

  const s = stats?.stats;

  return (
    <div className="container" style={{ maxWidth: 960 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2>Admin Dashboard</h2>
        <Link to="/admin/withdrawals">
          <button className="submit-btn" style={{ width: 'auto', padding: '10px 20px', marginTop: 0 }}>
            Manage Withdrawals
          </button>
        </Link>
      </div>

      {/* Platform stats */}
      <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)', marginBottom: 24 }}>
        <div className="stat-card">
          <div className="label">Platform Revenue</div>
          <div className="value" style={{ color: 'var(--accent)' }}>${s ? formatQuota(s.total_platform_amount) : '0'}</div>
        </div>
        <div className="stat-card">
          <div className="label">Total Trades</div>
          <div className="value">{s?.total_trades || 0}</div>
        </div>
        <div className="stat-card">
          <div className="label">Active Sellers</div>
          <div className="value">{stats?.active_sellers || 0} / {stats?.total_sellers || 0}</div>
        </div>
        <div className="stat-card">
          <div className="label">Active Listings</div>
          <div className="value">{stats?.active_listings || 0}</div>
        </div>
      </div>

      <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)', marginBottom: 24 }}>
        <div className="stat-card">
          <div className="label">Buyer Payments</div>
          <div className="value">${s ? formatQuota(s.total_buyer_amount) : '0'}</div>
        </div>
        <div className="stat-card">
          <div className="label">Seller Payouts</div>
          <div className="value">${s ? formatQuota(s.total_seller_amount) : '0'}</div>
        </div>
        <div className="stat-card">
          <div className="label">Tokens Processed</div>
          <div className="value">{s ? formatTokens(s.total_tokens) : '0'}</div>
        </div>
      </div>

      {/* Sellers table */}
      <div className="card">
        <div className="card-header">
          <h3>Sellers</h3>
          <span className="badge">{sellers.length} total</span>
        </div>

        {sellers.length === 0 ? (
          <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>No sellers registered yet.</p>
        ) : (
          <table className="listing-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>User</th>
                <th>Listings</th>
                <th>Balance</th>
                <th>Total Earned</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {sellers.map((seller) => (
                <tr key={seller.id}>
                  <td>#{seller.id}</td>
                  <td>
                    <strong>{seller.username || `User #${seller.user_id}`}</strong>
                    {seller.email && <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{seller.email}</div>}
                  </td>
                  <td>{seller.listing_count}</td>
                  <td>${formatQuota(seller.balance)}</td>
                  <td>${formatQuota(seller.total_earned)}</td>
                  <td>
                    <span className={seller.status === 1 ? 'status-active' : 'status-paused'}>
                      {seller.status === 1 ? 'Active' : 'Suspended'}
                    </span>
                  </td>
                  <td>
                    {actionLoading === seller.id ? (
                      <span className="loading" />
                    ) : seller.status === 1 ? (
                      <button className="action-btn danger" onClick={() => handleSuspend(seller.id)}>Suspend</button>
                    ) : (
                      <button className="action-btn" onClick={() => handleEnable(seller.id)}>Enable</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
