import { useEffect, useState } from 'react';
import { adminApi, type Withdrawal } from '../api/admin';

const STATUS_MAP: Record<number, { text: string; cls: string }> = {
  1: { text: 'Pending', cls: 'status-paused' },
  2: { text: 'Approved', cls: 'status-active' },
  3: { text: 'Paid', cls: 'status-active' },
  4: { text: 'Rejected', cls: 'status-exhausted' },
};

function formatQuota(n: number): string {
  return (n / 500000).toFixed(2);
}

function formatDate(s: string): string {
  if (!s) return '-';
  return new Date(s).toLocaleDateString();
}

export default function AdminWithdrawals() {
  const [withdrawals, setWithdrawals] = useState<Withdrawal[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState(0); // 0 = all
  const [actionLoading, setActionLoading] = useState<number | null>(null);
  const [settling, setSettling] = useState(false);

  async function load() {
    const res = await adminApi.getWithdrawals(filter);
    if (res.success) setWithdrawals(res.data || []);
    setLoading(false);
  }

  useEffect(() => { setLoading(true); load(); }, [filter]);

  async function handleApprove(id: number) {
    setActionLoading(id);
    await adminApi.approveWithdrawal(id);
    await load();
    setActionLoading(null);
  }

  async function handlePaid(id: number) {
    setActionLoading(id);
    await adminApi.markWithdrawalPaid(id);
    await load();
    setActionLoading(null);
  }

  async function handleReject(id: number) {
    if (!confirm('Reject this withdrawal? The amount will be refunded to the seller\'s balance.')) return;
    setActionLoading(id);
    await adminApi.rejectWithdrawal(id);
    await load();
    setActionLoading(null);
  }

  async function handleTriggerSettlement() {
    if (!confirm('Manually trigger monthly settlement? This will create withdrawal records for all sellers with balance.')) return;
    setSettling(true);
    await adminApi.triggerSettlement();
    await load();
    setSettling(false);
  }

  return (
    <div className="container" style={{ maxWidth: 960 }}>
      <div className="card">
        <div className="card-header">
          <h2>Withdrawals</h2>
          <button
            className="submit-btn"
            style={{ width: 'auto', padding: '10px 20px', marginTop: 0 }}
            onClick={handleTriggerSettlement}
            disabled={settling}
          >
            {settling && <span className="loading" />}
            {settling ? 'Settling...' : 'Trigger Settlement'}
          </button>
        </div>

        {/* Filter tabs */}
        <div style={{ display: 'flex', gap: 12, marginBottom: 24 }}>
          {[
            { label: 'All', value: 0 },
            { label: 'Pending', value: 1 },
            { label: 'Approved', value: 2 },
            { label: 'Paid', value: 3 },
            { label: 'Rejected', value: 4 },
          ].map((f) => (
            <button
              key={f.value}
              className={`limit-btn ${filter === f.value ? 'selected' : ''}`}
              onClick={() => setFilter(f.value)}
            >
              {f.label}
            </button>
          ))}
        </div>

        {loading ? (
          <div style={{ textAlign: 'center', padding: 32 }}>
            <span className="loading" /> Loading...
          </div>
        ) : withdrawals.length === 0 ? (
          <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: 32 }}>
            No withdrawals found.
          </p>
        ) : (
          <table className="listing-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Seller</th>
                <th>Amount</th>
                <th>Period</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {withdrawals.map((w) => {
                const status = STATUS_MAP[w.status] || { text: 'Unknown', cls: '' };
                return (
                  <tr key={w.id}>
                    <td>#{w.id}</td>
                    <td>{w.username || `Seller #${w.seller_id}`}</td>
                    <td><strong>${formatQuota(w.amount)}</strong></td>
                    <td style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
                      {formatDate(w.period_start)} ~ {formatDate(w.period_end)}
                    </td>
                    <td><span className={status.cls}>{status.text}</span></td>
                    <td>
                      {actionLoading === w.id ? (
                        <span className="loading" />
                      ) : (
                        <>
                          {w.status === 1 && (
                            <>
                              <button className="action-btn" onClick={() => handleApprove(w.id)}>Approve</button>
                              <button className="action-btn danger" onClick={() => handleReject(w.id)}>Reject</button>
                            </>
                          )}
                          {w.status === 2 && (
                            <button className="action-btn" onClick={() => handlePaid(w.id)}>Mark Paid</button>
                          )}
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
