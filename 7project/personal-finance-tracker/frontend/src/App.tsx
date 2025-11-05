import { BrowserRouter, Routes, Route, NavLink, Navigate, useNavigate } from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Categories from "./pages/Categories";
import Transactions from "./pages/Transactions";
import Budgets from "./pages/Budgets";
import { getToken, clearToken } from "./lib/auth";

function Shell({ children }: { children: React.ReactNode }) {
  const n = useNavigate();
  const authed = !!getToken();
  return (
    <>
      <header className="navbar">
        <div className="navbar-inner">
          <div className="brand"><a href="/"><span className="badge">PFT</span></a></div>
          <nav className="nav">
            <NavLink to="/dashboard" className={({isActive})=> isActive ? "active" : ""}>Dashboard</NavLink>
            <NavLink to="/categories" className={({isActive})=> isActive ? "active" : ""}>Categories</NavLink>
            <NavLink to="/transactions" className={({isActive})=> isActive ? "active" : ""}>Transactions</NavLink>
            <NavLink to="/budgets" className={({isActive})=> isActive ? "active" : ""}>Budgets</NavLink>
          </nav>
          <div className="right">
            {authed ? (
              <button className="btn btn-ghost" onClick={()=>{ clearToken(); n("/login"); }}>Logout</button>
            ) : (
              <NavLink to="/login" className="btn btn-primary">Login</NavLink>
            )}
          </div>
        </div>
      </header>
      <div className="container">
        <div className="spacer" />
        {children}
      </div>
    </>
  );
}

function Private({ children }: { children: React.ReactNode }) {
  return getToken() ? <>{children}</> : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <BrowserRouter>
      <Shell>
        <Routes>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/login" element={<Login/>} />
          <Route path="/register" element={<Register/>} />
          <Route path="/dashboard" element={<Private><Dashboard/></Private>} />
          <Route path="/categories" element={<Private><Categories/></Private>} />
          <Route path="/transactions" element={<Private><Transactions/></Private>} />
          <Route path="/budgets" element={<Private><Budgets/></Private>} />
          <Route path="*" element={<div className="card section">Not Found</div>} />
        </Routes>
      </Shell>
    </BrowserRouter>
  );
}
