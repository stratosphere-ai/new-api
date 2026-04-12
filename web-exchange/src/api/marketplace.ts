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

export interface BuyerStats {
  total_trades: number;
  total_tokens: number;
  total_spent: number;
  total_saved: number;
}

export interface Trade {
  id: number;
  listing_id: number;
  model_name: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  discount: number;
  retail_amount: number;
  buyer_amount: number;
  created_at: string;
}

export interface ModelUsage {
  model_name: string;
  trade_count: number;
  total_tokens: number;
  total_spent: number;
  avg_discount: number;
}

export interface APIToken {
  id: number;
  name: string;
  key: string;
  status?: number;
}

export const api = {
  // Public
  getMarketInfo: () => request<MarketInfo>('/info'),

  // Seller
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

  // Buyer
  getBuyerStats: () => request<{ stats: BuyerStats; balance: number }>('/buyer/stats'),
  getBuyerTrades: (limit = 50, offset = 0) =>
    request<Trade[]>(`/buyer/trades?limit=${limit}&offset=${offset}`),
  getBuyerModelUsage: () => request<ModelUsage[]>('/buyer/models'),
  getBuyerAPIInfo: () => request<{ tokens: APIToken[]; base_url: string; group: string }>('/buyer/api-info'),
  createBuyerAPIToken: () => request<APIToken>('/buyer/api-token', { method: 'POST' }),
};
