import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Navbar from './components/Navbar';
import StartSelling from './pages/StartSelling';
import Dashboard from './pages/Dashboard';
import BuyerUsage from './pages/BuyerUsage';
import APIAccess from './pages/APIAccess';
import './index.css';

function App() {
  return (
    <BrowserRouter>
      <Navbar />
      <Routes>
        <Route path="/" element={<StartSelling />} />
        <Route path="/sell" element={<StartSelling />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/usage" element={<BuyerUsage />} />
        <Route path="/api" element={<APIAccess />} />
      </Routes>
    </BrowserRouter>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
