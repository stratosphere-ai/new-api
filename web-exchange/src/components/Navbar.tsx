import { Link, useLocation } from 'react-router-dom';

export default function Navbar() {
  const location = useLocation();
  const isActive = (path: string) => location.pathname === path ? 'active' : '';

  return (
    <nav className="nav">
      <Link to="/" className="nav-brand">Token Exchange</Link>
      <div className="nav-links">
        <Link to="/sell" className={isActive('/sell')}>Sell</Link>
        <Link to="/dashboard" className={isActive('/dashboard')}>Dashboard</Link>
      </div>
    </nav>
  );
}
