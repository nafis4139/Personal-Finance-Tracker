// frontend/src/pages/Login.tsx

// Simple email/password login form.
// - Submits credentials to the API, stores the returned JWT, and navigates to the dashboard.
// - Displays inline error feedback and a busy state while the request is in flight.

import { useState } from "react";
import { api } from "../lib/api";
import { setToken } from "../lib/auth";
import { useNavigate, Link } from "react-router-dom";

export default function Login() {
  const n = useNavigate();

  // Local form state (pre-filled defaults for convenience in development).
  const [email, setEmail] = useState("nafis@example.com");
  const [password, setPassword] = useState("secret123");

  // UI state for error message and submit progress.
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  // Handle form submit: POST /login → { token }, persist, then redirect.
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    try {
      const resp = await api<{ token: string }>(`/login`, {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      setToken(resp.token);
      n("/dashboard");
    } catch (e: any) {
      setErr(e.message || "Login failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card section" style={{ maxWidth: 400, margin: "60px auto" }}>
      <h1 className="h1" style={{ textAlign: "center" }}>Welcome back</h1>
      <p className="muted" style={{ textAlign: "center", marginBottom: 20 }}>
        Log in to manage personal finances
      </p>

      {/* Login form with controlled inputs */}
      <form onSubmit={submit} className="space-y-3">
        <div className="row" style={{ flexDirection: "column", gap: 12 }}>
          <input
            className="input w-100"
            placeholder="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className="input w-100"
            placeholder="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          {/* Inline error banner */}
          {err && <div className="card section" style={{ borderColor: "rgba(239,68,68,.35)" }}>{err}</div>}
          <button
            disabled={busy}
            className="btn btn-primary w-100"
            style={{ marginTop: 8 }}
          >
            {busy ? "Logging in..." : "Login"}
          </button>
        </div>
      </form>

      <div className="spacer" />
      {/* Registration link for accounts not yet created */}
      <p className="muted" style={{ textAlign: "center" }}>
        Don’t have an account?{" "}
        <Link to="/register" className="underline" style={{ color: "var(--brand)" }}>
          Register here
        </Link>
      </p>
    </div>
  );
}
