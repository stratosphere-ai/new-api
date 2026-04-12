import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, type CreateListingRequest } from '../api/marketplace';

const PROVIDERS = [
  { id: 'anthropic', name: 'Anthropic', sub: 'Claude' },
  { id: 'openai', name: 'OpenAI', sub: 'GPT' },
  { id: 'google', name: 'Google', sub: 'Gemini' },
];

const TOTAL_CAP_OPTIONS = [
  { label: '1M', value: 1_000_000 },
  { label: '10M', value: 10_000_000 },
  { label: '100M', value: 100_000_000 },
  { label: 'No limit', value: 0 },
];

const HOURLY_CAP_OPTIONS = [
  { label: '100K', value: 100_000 },
  { label: '500K', value: 500_000 },
  { label: '1M', value: 1_000_000 },
  { label: 'None', value: 0 },
];

const DAILY_CAP_OPTIONS = [
  { label: '1M', value: 1_000_000 },
  { label: '5M', value: 5_000_000 },
  { label: '10M', value: 10_000_000 },
  { label: 'None', value: 0 },
];

export default function StartSelling() {
  const navigate = useNavigate();
  const [provider, setProvider] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [discount, setDiscount] = useState(30);
  const [totalCap, setTotalCap] = useState(0);
  const [hourlyCap, setHourlyCap] = useState(0);
  const [dailyCap, setDailyCap] = useState(0);
  const [label, setLabel] = useState('');

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [validating, setValidating] = useState(false);

  const canSubmit = provider && apiKey && discount > 0 && !loading;

  async function handleSubmit() {
    if (!canSubmit) return;
    setLoading(true);
    setError('');

    // First register as seller (idempotent)
    const regResult = await api.registerSeller();
    if (!regResult.success) {
      setError(regResult.message);
      setLoading(false);
      return;
    }

    const req: CreateListingRequest = {
      provider_type: provider,
      api_key: apiKey,
      discount,
      total_cap: totalCap,
      hourly_cap: hourlyCap,
      daily_cap: dailyCap,
      label,
    };

    const result = await api.createListing(req);
    setLoading(false);

    if (result.success) {
      navigate('/dashboard');
    } else {
      setError(result.message);
    }
  }

  return (
    <div className="container">
      <div className="card">
        <div className="card-header">
          <h2>Start Selling</h2>
          <span className="badge">3 steps</span>
        </div>

        {/* Step 1: Provider */}
        <div className="step-label">Step 1 &mdash; Provider</div>
        <div className="provider-grid">
          {PROVIDERS.map((p) => (
            <button
              key={p.id}
              className={`provider-btn ${provider === p.id ? 'selected' : ''}`}
              onClick={() => setProvider(p.id)}
            >
              <div className="name">{p.name}</div>
              <div className="sub">{p.sub}</div>
            </button>
          ))}
        </div>

        {/* Step 2: API Key */}
        <div className="step-label">Step 2 &mdash; Paste your API Key</div>
        <div className="input-group">
          <input
            className="input-field"
            type="password"
            placeholder="sk-ant-... or sk-..."
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
          />
          <div className="input-hint">
            <span className="check">{validating ? '' : '\u2705'}</span>
            Encrypted with AES-256. Never visible to buyers.
          </div>
        </div>

        {/* Step 3: Discount */}
        <div className="step-label">Step 3 &mdash; Set your discount</div>
        <div className="discount-section">
          <div className="discount-header">
            <div />
            <div className="discount-value">
              {discount}%
              <span className="unit">off retail</span>
            </div>
          </div>
          <input
            className="slider"
            type="range"
            min={5}
            max={80}
            value={discount}
            onChange={(e) => setDiscount(Number(e.target.value))}
            style={{
              background: `linear-gradient(to right, var(--accent) ${((discount - 5) / 75) * 100}%, var(--bg-input) ${((discount - 5) / 75) * 100}%)`,
            }}
          />
        </div>

        {/* Token limits */}
        <div className="step-label" style={{ cursor: 'pointer' }}>
          Token limits
        </div>

        <div className="limits-section">
          <div className="limit-label">Total Token Cap</div>
          <div className="limit-options">
            {TOTAL_CAP_OPTIONS.map((opt) => (
              <button
                key={opt.label}
                className={`limit-btn ${totalCap === opt.value ? 'selected' : ''}`}
                onClick={() => setTotalCap(opt.value)}
              >
                {opt.label}
              </button>
            ))}
          </div>
          <div className="limit-hint">Total tokens (input+output) your key will serve. Auto-pauses when hit.</div>
        </div>

        <div className="limits-section">
          <div className="limit-label">Max Tokens/Hour</div>
          <div className="limit-options">
            {HOURLY_CAP_OPTIONS.map((opt) => (
              <button
                key={opt.label}
                className={`limit-btn ${hourlyCap === opt.value ? 'selected' : ''}`}
                onClick={() => setHourlyCap(opt.value)}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        <div className="limits-section">
          <div className="limit-label">Max Tokens/Day</div>
          <div className="limit-options">
            {DAILY_CAP_OPTIONS.map((opt) => (
              <button
                key={opt.label}
                className={`limit-btn ${dailyCap === opt.value ? 'selected' : ''}`}
                onClick={() => setDailyCap(opt.value)}
              >
                {opt.label}
              </button>
            ))}
          </div>
          <div className="limit-hint">Resets automatically every hour/day. Prevents runaway usage on your key.</div>
        </div>

        {/* Label */}
        <div className="input-group">
          <div className="limit-label">Label (optional)</div>
          <input
            className="input-field"
            type="text"
            placeholder="e.g. My Claude Pro key"
            value={label}
            onChange={(e) => setLabel(e.target.value)}
          />
        </div>

        {/* Error */}
        {error && <div className="validation-msg error">{error}</div>}

        {/* Submit */}
        <button
          className="submit-btn"
          disabled={!canSubmit}
          onClick={handleSubmit}
        >
          {loading && <span className="loading" />}
          {!provider ? 'Select a provider to continue' : loading ? 'Creating listing...' : 'Start Selling'}
        </button>
      </div>
    </div>
  );
}
