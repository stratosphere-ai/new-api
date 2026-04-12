const BASE = '/api/marketplace';

async function request<T>(url: string, options?: RequestInit): Promise<{ success: boolean; message: string; data: T }> {
  const token = localStorage.getItem('token') || '';
  const res = await fetch(BASE + url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options?.headers,
    },
  });
  return res.json();
}

export interface Seller {
  id: number;
  user_id: number;
  status: number;
  balance: number;
  total_earned: number;
}

export interface Listing {
  id: number;
  seller_id: number;
  channel_id: number;
  provider_type: string;
  discount: number;
  total_cap: number;
  hourly_cap: number;
  daily_cap: number;
  tokens_used: number;
  label: string;
  status: number;
  created_at: string;
}

export interface SellerDashboard {
  seller: Seller;
  listings: Listing[];
}

export interface MarketInfo {
  min_discount: number;
  max_discount: number;
}

export interface CreateListingRequest {
  provider_type: string;
  api_key: string;
  discount: number;
  total_cap: number;
  hourly_cap: number;
  daily_cap: number;
  label: string;
}

export const api = {
  getMarketInfo: () => request<MarketInfo>('/info'),
  registerSeller: () => request<Seller>('/seller/register', { method: 'POST' }),
  getDashboard: () => request<SellerDashboard>('/seller/dashboard'),
  createListing: (data: CreateListingRequest) =>
    request<Listing>('/listings', { method: 'POST', body: JSON.stringify(data) }),
  pauseListing: (id: number) =>
    request<null>(`/listings/${id}/pause`, { method: 'POST' }),
  resumeListing: (id: number) =>
    request<null>(`/listings/${id}/resume`, { method: 'POST' }),
  deleteListing: (id: number) =>
    request<null>(`/listings/${id}`, { method: 'DELETE' }),
};
