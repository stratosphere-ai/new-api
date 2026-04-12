import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, type MarketInfo } from '../api/marketplace';
import { isLoggedIn } from '../api/auth';

export default function Landing() {
  const navigate = useNavigate();
  const [market, setMarket] = useState<MarketInfo | null>(null);

  useEffect(() => {
    api.getMarketInfo().then((res) => {
      if (res.success) setMarket(res.data);
    });
  }, []);

  return (
    <div className="container" style={{ maxWidth: 800, paddingTop: 60 }}>
      {/* Hero */}
      <div style={{ textAlign: 'center', marginBottom: 64 }}>
        <h1 style={{ fontSize: 48, fontWeight: 800, lineHeight: 1.2, marginBottom: 20 }}>
          AI API tokens,
          <br />
          <span style={{ color: 'var(--accent)' }}>cheaper than retail.</span>
        </h1>
        <p style={{ fontSize: 18, color: 'var(--text-secondary)', maxWidth: 520, margin: '0 auto 32px' }}>
          Buy AI API access at up to {market?.max_discount || 40}% off official prices.
          Sell your unused API quota and earn.
        </p>
        <div style={{ display: 'flex', gap: 16, justifyContent: 'center' }}>
          {isLoggedIn() ? (
            <>
              <button className="submit-btn" style={{ width: 'auto', padding: '14px 32px' }} onClick={() => navigate('/sell')}>
                Start Selling
              </button>
              <button
                className="submit-btn"
                style={{ width: 'auto', padding: '14px 32px', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                onClick={() => navigate('/api')}
              >
                Get API Access
              </button>
            </>
          ) : (
            <>
              <button className="submit-btn" style={{ width: 'auto', padding: '14px 32px' }} onClick={() => navigate('/login')}>
                Get Started
              </button>
              <button
                className="submit-btn"
                style={{ width: 'auto', padding: '14px 32px', background: 'var(--bg-input)', color: 'var(--text-primary)' }}
                onClick={() => navigate('/login')}
              >
                Sign In
              </button>
            </>
          )}
        </div>
      </div>

      {/* Feature cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 20, marginBottom: 64 }}>
        <div className="card" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 32, marginBottom: 12 }}>&#x1f512;</div>
          <h3 style={{ marginBottom: 8 }}>AES-256 Encrypted</h3>
          <p style={{ fontSize: 14, color: 'var(--text-secondary)' }}>
            Seller keys are encrypted at rest. Never visible to buyers or platform operators.
          </p>
        </div>
        <div className="card" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 32, marginBottom: 12 }}>&#x26a1;</div>
          <h3 style={{ marginBottom: 8 }}>Smart Routing</h3>
          <p style={{ fontSize: 14, color: 'var(--text-secondary)' }}>
            Auto-failover with circuit breakers. If one key fails, requests seamlessly switch to another.
          </p>
        </div>
        <div className="card" style={{ textAlign: 'center' }}>
          <div style={{ fontSize: 32, marginBottom: 12 }}>&#x1f4b0;</div>
          <h3 style={{ marginBottom: 8 }}>Save up to {market?.max_discount || 40}%</h3>
          <p style={{ fontSize: 14, color: 'var(--text-secondary)' }}>
            Current market discount: {market?.min_discount || 10}% – {market?.max_discount || 40}% off retail.
          </p>
        </div>
      </div>

      {/* How it works */}
      <div className="card" style={{ marginBottom: 64 }}>
        <h2 style={{ textAlign: 'center', marginBottom: 32 }}>How it works</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 40 }}>
          <div>
            <div className="step-label" style={{ color: 'var(--accent)' }}>For Buyers</div>
            <div style={{ fontSize: 15, lineHeight: 2, color: 'var(--text-secondary)' }}>
              1. Create an account and top up your balance<br />
              2. Generate a marketplace API token<br />
              3. Use standard OpenAI-compatible endpoints<br />
              4. Pay discounted prices per request
            </div>
          </div>
          <div>
            <div className="step-label" style={{ color: 'var(--accent)' }}>For Sellers</div>
            <div style={{ fontSize: 15, lineHeight: 2, color: 'var(--text-secondary)' }}>
              1. Paste your API key (Anthropic / OpenAI / Google)<br />
              2. Set your discount and token limits<br />
              3. Earn revenue on every request served<br />
              4. Monthly payouts with full transparency
            </div>
          </div>
        </div>
      </div>

      {/* Supported providers */}
      <div style={{ textAlign: 'center', marginBottom: 40 }}>
        <p style={{ color: 'var(--text-muted)', fontSize: 14, marginBottom: 16 }}>SUPPORTED PROVIDERS</p>
        <div style={{ display: 'flex', justifyContent: 'center', gap: 48, fontSize: 18, fontWeight: 600 }}>
          <span>Anthropic</span>
          <span>OpenAI</span>
          <span>Google</span>
        </div>
      </div>
    </div>
  );
}
