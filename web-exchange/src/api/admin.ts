const BASE = '/api/marketplace/admin';

async function request<T>(url: string, options?: RequestInit): Promise<{ success: boolean; message: string; data: T }> {
  const token = localStorage.getItem('token') || '';
  const res = await fetch(BASE + url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'New-Api-User': localStorage.getItem('user_id') || '',
      ...options?.headers,
    },
  });
  return res.json();
}

export interface PlatformStats {
  total_trades: number;
  total_tokens: number;
  total_retail_amount: number;
  total_buyer_amount: number;
  total_seller_amount: number;
  total_platform_amount: number;
}

export interface SellerWithUser {
  id: number;
  user_id: number;
  status: number;
  balance: number;
  total_earned: number;
  username: string;
  email: string;
  listing_count: number;
  created_at: string;
}

export interface SellerDetail {
  seller: SellerWithUser;
  username: string;
  email: string;
  listings: Array<{
    id: number;
    provider_type: string;
    discount: number;
    total_cap: number;
    tokens_used: number;
    label: string;
    status: number;
  }>;
  withdrawals: Array<Withdrawal>;
}

export interface Withdrawal {
  id: number;
  seller_id: number;
  amount: number;
  status: number;
  period_start: string;
  period_end: string;
  created_at: string;
  paid_at: string | null;
  username?: string;
}

export const adminApi = {
  getStats: () => request<{ stats: PlatformStats; total_sellers: number; active_sellers: number; active_listings: number }>('/stats'),
  getSellers: () => request<SellerWithUser[]>('/sellers'),
  getSellerDetail: (id: number) => request<SellerDetail>(`/sellers/${id}`),
  suspendSeller: (id: number) => request<null>(`/sellers/${id}/suspend`, { method: 'POST' }),
  enableSeller: (id: number) => request<null>(`/sellers/${id}/enable`, { method: 'POST' }),
  getWithdrawals: (status = 0) => request<Withdrawal[]>(`/withdrawals?status=${status}`),
  approveWithdrawal: (id: number) => request<null>(`/withdrawals/${id}/approve`, { method: 'POST' }),
  markWithdrawalPaid: (id: number) => request<null>(`/withdrawals/${id}/paid`, { method: 'POST' }),
  rejectWithdrawal: (id: number) => request<null>(`/withdrawals/${id}/reject`, { method: 'POST' }),
  triggerSettlement: () => request<null>('/settlement/trigger', { method: 'POST' }),
};
