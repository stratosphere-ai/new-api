import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, type SellerDashboard, type Listing } from '../api/marketplace';

const STATUS_MAP: Record<number, { text: string; cls: string }> = {
  1: { text: 'Active', cls: 'status-active' },
  2: { text: 'Paused', cls: 'status-paused' },
  3: { text: 'Exhausted', cls: 'status-exhausted' },
};

const PROVIDER_NAMES: Record<string, string> = {
  anthropic: 'Anthropic',
  openai: 'OpenAI',
  google: 'Google',
};

function formatTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K';
  return n.toString();
}

export default function Dashboard() {
  const [dashboard, setDashboard] = useState<SellerDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  async function loadDashboard() {
    const res = await api.getDashboard();
    setLoading(false);
    if (res.success) {
      setDashboard(res.data);
    } else {
      setError(res.message);
    }
  }

  useEffect(() => {
    loadDashboard();
  }, []);

  async function handlePause(id: number) {
    setActionLoading(id);
    await api.pauseListing(id);
    await loadDashboard();
    setActionLoading(null);
  }

  async function handleResume(id: number) {
    setActionLoading(id);
    await api.resumeListing(id);
    await loadDashboard();
    setActionLoading(null);
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this listing?')) return;
    setActionLoading(id);
    await api.deleteListing(id);
    await loadDashboard();
    setActionLoading(null);
  }

  if (loading) {
    return (
      <div className="container">
        <div className="card" style={{ textAlign: 'center', padding: '48px' }}>
          <span className="loading" /> Loading dashboard...
        </div>
      </div>
    );
  }

  if (error || !dashboard) {
    return (
      <div className="container">
        <div className="card">
          <p style={{ color: 'var(--text-secondary)', marginBottom: 16 }}>
            {error || 'Not registered as a seller yet.'}
          </p>
          <Link to="/sell" className="submit-btn" style={{ display: 'inline-block', textAlign: 'center', textDecoration: 'none' }}>
            Start Selling
          </Link>
        </div>
      </div>
    );
  }

  const { seller, listings } = dashboard;

  return (
    <div className="container">
      {/* Stats */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="label">Balance</div>
          <div className="value">{(seller.balance / 500000).toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Total Earned</div>
          <div className="value">{(seller.total_earned / 500000).toFixed(2)}</div>
        </div>
        <div className="stat-card">
          <div className="label">Listings</div>
          <div className="value">{listings?.length || 0}</div>
        </div>
      </div>

      {/* Listings */}
      <div className="card">
        <div className="card-header">
          <h2>Your Listings</h2>
          <Link to="/sell">
            <button className="submit-btn" style={{ width: 'auto', padding: '10px 24px', marginTop: 0 }}>
              + New Listing
            </button>
          </Link>
        </div>

        {!listings || listings.length === 0 ? (
          <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
            No listings yet. Create your first listing to start earning.
          </p>
        ) : (
          <table className="listing-table">
            <thead>
              <tr>
                <th>Provider</th>
                <th>Discount</th>
                <th>Usage</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {listings.map((listing: Listing) => {
                const status = STATUS_MAP[listing.status] || { text: 'Unknown', cls: '' };
                const capText = listing.total_cap > 0
                  ? `${formatTokens(listing.tokens_used)} / ${formatTokens(listing.total_cap)}`
                  : formatTokens(listing.tokens_used);

                return (
                  <tr key={listing.id}>
                    <td>
                      <strong>{PROVIDER_NAMES[listing.provider_type] || listing.provider_type}</strong>
                      {listing.label && <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{listing.label}</div>}
                    </td>
                    <td>{listing.discount}% off</td>
                    <td>{capText}</td>
                    <td><span className={status.cls}>{status.text}</span></td>
                    <td>
                      {actionLoading === listing.id ? (
                        <span className="loading" />
                      ) : (
                        <>
                          {listing.status === 1 ? (
                            <button className="action-btn" onClick={() => handlePause(listing.id)}>Pause</button>
                          ) : listing.status === 2 ? (
                            <button className="action-btn" onClick={() => handleResume(listing.id)}>Resume</button>
                          ) : null}
                          <button className="action-btn danger" onClick={() => handleDelete(listing.id)}>Delete</button>
                        </>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
