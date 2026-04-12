import { Link, useLocation, useNavigate } from 'react-router-dom';
import { isLoggedIn, isAdmin, getUser, clearUser } from '../api/auth';

export default function Navbar() {
  const location = useLocation();
  const navigate = useNavigate();
  const isActive = (path: string) => location.pathname.startsWith(path) ? 'active' : '';
  const loggedIn = isLoggedIn();
  const user = getUser();

  function handleLogout() {
    clearUser();
    navigate('/');
  }

  return (
    <nav className="nav">
      <Link to="/" className="nav-brand">Token Exchange</Link>
      <div className="nav-links">
        {loggedIn ? (
          <>
            <Link to="/sell" className={isActive('/sell')}>Sell</Link>
            <Link to="/dashboard" className={isActive('/dashboard')}>Dashboard</Link>
            <Link to="/earnings" className={isActive('/earnings')}>Earnings</Link>
            <Link to="/usage" className={isActive('/usage')}>Usage</Link>
            <Link to="/api" className={isActive('/api')}>API</Link>
            {isAdmin() && (
              <>
                <span style={{ color: 'var(--border)' }}>|</span>
                <Link to="/admin" className={isActive('/admin')}>Admin</Link>
              </>
            )}
            <span style={{ color: 'var(--border)' }}>|</span>
            <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>{user?.username}</span>
            <button
              onClick={handleLogout}
              style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: 14 }}
            >
              Logout
            </button>
          </>
        ) : (
          <Link to="/login" className={isActive('/login')}>Sign In</Link>
        )}
      </div>
    </nav>
  );
}
