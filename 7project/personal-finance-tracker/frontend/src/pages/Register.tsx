// frontend/src/pages/Register.tsx

// Registration form:
// - Collects name, email, and password.
// - Submits to API, then navigates to login on success.
// - Shows inline error and loading state.

import { useState } from "react";
import { api } from "../lib/api";
import { useNavigate, Link } from "react-router-dom";

export default function Register() {
  const n = useNavigate();

  // Controlled inputs with dev-friendly defaults.
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  // UI state for error feedback and submit progress.
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  // POST /register with the provided credentials; redirect to /login on success.
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    try {
      await api(`/register`, {
        method: "POST",
        body: JSON.stringify({ name, email, password }),
      });
      n("/login");
    } catch (e: any) {
      setErr(e.message || "Registration failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card section" style={{ maxWidth: 400, margin: "60px auto" }}>
      <h1 className="h1" style={{ textAlign: "center" }}>Create account</h1>
      <p className="muted" style={{ textAlign: "center", marginBottom: 20 }}>
        Start tracking spending and income today
      </p>

      {/* Registration form */}
      <form onSubmit={submit} className="space-y-3">
        <div className="row" style={{ flexDirection: "column", gap: 12 }}>
          <input
            className="input w-100"
            placeholder="Full name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
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
          {/* Inline error banner if API call fails */}
          {err && <div className="card section" style={{ borderColor: "rgba(239,68,68,.35)" }}>{err}</div>}
          <button
            disabled={busy}
            className="btn btn-primary w-100"
            style={{ marginTop: 8 }}
          >
            {busy ? "Registering..." : "Register"}
          </button>
        </div>
      </form>

      <div className="spacer" />
      {/* Link to login for existing accounts */}
      <p className="muted" style={{ textAlign: "center" }}>
        Already have an account?{" "}
        <Link to="/login" className="underline" style={{ color: "var(--brand)" }}>
          Login here
        </Link>
      </p>
    </div>
  );
}
