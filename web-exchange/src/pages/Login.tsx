import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi, saveUser } from '../api/auth';

export default function Login() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!username || !password) return;
    setLoading(true);
    setError('');

    if (mode === 'register') {
      const res = await authApi.register(username, password, email);
      if (!res.success) {
        setError(res.message);
        setLoading(false);
        return;
      }
      // Auto-login after register
    }

    const res = await authApi.login(username, password);
    setLoading(false);

    if (res.success) {
      saveUser(res.data);
      navigate('/sell');
    } else {
      setError(res.message);
    }
  }

  return (
    <div className="container" style={{ maxWidth: 440, paddingTop: 60 }}>
      <div className="card">
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h2 style={{ marginBottom: 8 }}>
            {mode === 'login' ? 'Welcome back' : 'Create account'}
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: 14 }}>
            {mode === 'login'
              ? 'Sign in to your Token Exchange account'
              : 'Join the marketplace as a buyer or seller'}
          </p>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <div className="limit-label">Username</div>
            <input
              className="input-field"
              type="text"
              placeholder="Enter username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
            />
          </div>

          {mode === 'register' && (
            <div className="input-group">
              <div className="limit-label">Email</div>
              <input
                className="input-field"
                type="email"
                placeholder="Enter email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
              />
            </div>
          )}

          <div className="input-group">
            <div className="limit-label">Password</div>
            <input
              className="input-field"
              type="password"
              placeholder="Enter password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            />
          </div>

          {error && <div className="validation-msg error">{error}</div>}

          <button className="submit-btn" type="submit" disabled={loading || !username || !password}>
            {loading && <span className="loading" />}
            {mode === 'login' ? 'Sign In' : 'Create Account'}
          </button>
        </form>

        <div style={{ textAlign: 'center', marginTop: 20 }}>
          {mode === 'login' ? (
            <p style={{ color: 'var(--text-secondary)', fontSize: 14 }}>
              Don't have an account?{' '}
              <button
                onClick={() => { setMode('register'); setError(''); }}
                style={{ background: 'none', border: 'none', color: 'var(--accent)', cursor: 'pointer', fontSize: 14 }}
              >
                Sign up
              </button>
            </p>
          ) : (
            <p style={{ color: 'var(--text-secondary)', fontSize: 14 }}>
              Already have an account?{' '}
              <button
                onClick={() => { setMode('login'); setError(''); }}
                style={{ background: 'none', border: 'none', color: 'var(--accent)', cursor: 'pointer', fontSize: 14 }}
              >
                Sign in
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
