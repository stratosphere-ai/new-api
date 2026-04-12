const BASE = '/api/user';

async function request<T>(url: string, options?: RequestInit): Promise<{ success: boolean; message: string; data: T }> {
  const res = await fetch(BASE + url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  return res.json();
}

export interface UserInfo {
  id: number;
  username: string;
  display_name: string;
  role: number;
  status: number;
  group: string;
}

export const authApi = {
  login: (username: string, password: string) =>
    request<UserInfo>('/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  register: (username: string, password: string, email: string) =>
    request<null>('/register', {
      method: 'POST',
      body: JSON.stringify({ username, password, email }),
    }),

  logout: () => request<null>('/logout', { method: 'GET' }),

  getSelf: () => request<UserInfo>('/self', { method: 'GET' }),
};

// Simple auth state helpers
export function saveUser(user: UserInfo) {
  localStorage.setItem('user', JSON.stringify(user));
  localStorage.setItem('user_id', String(user.id));
}

export function getUser(): UserInfo | null {
  const raw = localStorage.getItem('user');
  if (!raw) return null;
  try { return JSON.parse(raw); } catch { return null; }
}

export function clearUser() {
  localStorage.removeItem('user');
  localStorage.removeItem('user_id');
  localStorage.removeItem('token');
}

export function isLoggedIn(): boolean {
  return getUser() !== null;
}

export function isAdmin(): boolean {
  const user = getUser();
  return user !== null && user.role >= 10;
}
