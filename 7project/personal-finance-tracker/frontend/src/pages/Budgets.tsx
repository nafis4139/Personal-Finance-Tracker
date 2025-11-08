// frontend/src/pages/Budgets.tsx
//
// Budgets screen
// - Shows this month's budgets (expense categories only)
// - Create, edit (inline), and delete a budget
// - Quick stats: count + total limit
//
// Notes:
// * We keep budgets scoped by `period_month` (YYYY-MM).
// * Inline edit is optimistic: we update local list after a successful PUT.
// * We don't allow changing `period_month` from the edit form (keeps the UX simple).

import { useEffect, useMemo, useState } from "react";
import { api, apiList } from "../lib/api";
import EmptyState from "../components/EmptyState";

// --- API models used here ---
type Category = { id: number; name: string; type: "income" | "expense" };
type Budget = {
  id: number;
  user_id: number;
  category_id: number | null;
  period_month: string;      // "YYYY-MM" (e.g., "2025-11")
  limit_amount: number;
  created_at: string;        // ISO timestamp
};

// --- Small date helpers ---

// Default value for <input type="month">
function yyyyMm(d = new Date()) {
  const m = (d.getMonth() + 1).toString().padStart(2, "0");
  return `${d.getFullYear()}-${m}`;
}

// Human-friendly label for a month key ("2025-11" -> "November 2025")
function niceMonth(yyyyMm: string) {
  const [y, m] = yyyyMm.split("-").map(Number);
  return new Date(y, m - 1, 1).toLocaleString(undefined, {
    month: "long",
    year: "numeric",
  });
}

export default function Budgets() {
  // --- Page state ---

  // The currently selected month; drives the query to the API.
  const [month, setMonth] = useState(yyyyMm());

  // Cached categories (we only allow expense categories to be selected).
  const [cats, setCats] = useState<Category[]>([]);

  // Budgets fetched for `month`.
  const [items, setItems] = useState<Budget[]>([]);

  // Flags and surface error message.
  const [busy, setBusy] = useState(false);     // true while mutating (create/edit/delete)
  const [loading, setLoading] = useState(false); // true while fetching
  const [msg, setMsg] = useState<string | null>(null);

  // Create form state.
  const [catId, setCatId] = useState<number | "none">("none");
  const [limit, setLimit] = useState<string>("");

  // Inline edit state (per-row).
  const [editId, setEditId] = useState<number | null>(null);
  const [editCatId, setEditCatId] = useState<number | "none">("none");
  const [editLimit, setEditLimit] = useState<string>("");

  // --- Data loading ---

  // Fetch budgets for the selected month and the user's categories.
  async function load() {
    try {
      setLoading(true);
      setMsg(null);

      const [list, cs] = await Promise.all([
        apiList<Budget>(`/budgets?month=${month}`),
        apiList<Category>("/categories"),
      ]);

      setItems(list);
      setCats(cs);
    } catch (e: any) {
      setMsg(e.message);
      setItems([]); // keep the UI stable
    } finally {
      setLoading(false);
    }
  }

  // Re-fetch when the month changes.
  useEffect(() => {
    load();
  }, [month]);

  // --- Create ---

  // Create a budget for the current month; prepend on success.
  async function addBudget() {
    if (catId === "none" || !limit) return;

    try {
      setBusy(true);
      setMsg(null);

      const b = await api<Budget>("/budgets", {
        method: "POST",
        body: JSON.stringify({
          category_id: catId,
          period_month: month,              // lock budget to current month
          limit_amount: Number(limit),
        }),
      });

      setItems([b, ...items]);
      setLimit("");
      setCatId("none");
    } catch (e: any) {
      setMsg(e.message);
    } finally {
      setBusy(false);
    }
  }

  // --- Delete ---

  // Remove a budget and drop it from local state.
  async function delBudget(id: number) {
    try {
      setBusy(true);
      setMsg(null);

      await api(`/budgets/${id}`, { method: "DELETE" });
      setItems(items.filter((i) => i.id !== id));
    } catch (e: any) {
      setMsg(e.message);
    } finally {
      setBusy(false);
    }
  }

  // --- Edit (inline) ---

  // Enter edit mode for a row; seed inputs with the existing values.
  function startEdit(b: Budget) {
    setEditId(b.id);
    setEditCatId(b.category_id ?? "none");
    setEditLimit(String(b.limit_amount));
  }

  // Leave edit mode without saving.
  function cancelEdit() {
    setEditId(null);
    setEditCatId("none");
    setEditLimit("");
  }

  // Save edits (category and/or limit). We intentionally do not change period_month here.
  async function saveEdit() {
    if (editId == null) return;

    try {
      setBusy(true);
      setMsg(null);

      const payload: any = {
        limit_amount: Number(editLimit),
      };
      if (editCatId !== "none") payload.category_id = editCatId;

      const updated = await api<Budget>(`/budgets/${editId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });

      // Replace the updated item in-place to keep ordering intact.
      setItems(items.map((i) => (i.id === updated.id ? updated : i)));
      cancelEdit();
    } catch (e: any) {
      setMsg(e.message);
    } finally {
      setBusy(false);
    }
  }

  // --- Derived values ---

  // Sum of all limits visible in the list (useful quick KPI).
  const totalLimit = useMemo(
    () => items.reduce((a, b) => a + b.limit_amount, 0),
    [items]
  );

  // --- Render ---

  return (
    <div className="card section">
      <h1 className="h1">Budgets</h1>
      <p className="muted">Track monthly caps for expense categories.</p>

      <div className="spacer" />

      {/* Month selector drives the API query window */}
      <div className="row">
        <input
          className="input"
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
        />
      </div>

      <div className="spacer" />

      {/* Create budget */}
      <div className="card section">
        <div className="h2" style={{ marginBottom: 10 }}>Create budget</div>
        <div className="row">
          <select
            className="select"
            value={catId}
            onChange={(e) =>
              setCatId(e.target.value === "none" ? "none" : Number(e.target.value))
            }
          >
            <option value="none">Select category</option>
            {/* Only expense categories can have limits */}
            {cats
              .filter((c) => c.type === "expense")
              .map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
          </select>

          <input
            className="input"
            placeholder="Limit amount"
            inputMode="decimal"
            value={limit}
            onChange={(e) => setLimit(e.target.value)}
          />

          <button className="btn btn-primary" onClick={addBudget} disabled={busy}>
            Add
          </button>
        </div>
      </div>

      {/* Inline error banner */}
      {msg && (
        <>
          <div className="spacer" />
          <div className="card section" style={{ borderColor: "rgba(239,68,68,.35)" }}>
            {msg}
          </div>
        </>
      )}

      <div className="spacer" />

      {/* Quick stats */}
      <div className="card section" style={{ display: "flex", gap: 16, flexWrap: "wrap" }}>
        <div>
          <div className="muted">Budgets this month</div>
          <div style={{ fontSize: 24, fontWeight: 800 }}>{items.length}</div>
        </div>
        <div>
          <div className="muted">Total limit</div>
          <div style={{ fontSize: 24, fontWeight: 800 }}>{totalLimit.toFixed(2)}</div>
        </div>
      </div>

      <div className="spacer" />

      {/* List (or empty state) */}
      {loading ? (
        <div className="muted">Loading…</div>
      ) : items.length === 0 ? (
        <EmptyState
          title={`No budgets for ${niceMonth(month)}`}
          subtitle="Create a budget to set a monthly spending cap."
          action={
            <button
              className="btn btn-primary"
              onClick={addBudget}
              disabled={busy || catId === "none" || !limit}
            >
              Add Budget
            </button>
          }
        />
      ) : (
        <ul className="list">
          {items.map((b) => (
            <li key={b.id} className="list-item">
              {/* Edit vs. view mode per row */}
              {editId === b.id ? (
                // --- EDIT MODE ---
                <div
                  style={{
                    display: "flex",
                    gap: 10,
                    alignItems: "center",
                    flexWrap: "wrap",
                    width: "100%",
                  }}
                >
                  <select
                    className="select"
                    value={editCatId}
                    onChange={(e) =>
                      setEditCatId(
                        e.target.value === "none" ? "none" : Number(e.target.value)
                      )
                    }
                  >
                    <option value="none">No category</option>
                    {cats
                      .filter((c) => c.type === "expense")
                      .map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.name}
                        </option>
                      ))}
                  </select>

                  <input
                    className="input"
                    placeholder="Limit amount"
                    inputMode="decimal"
                    value={editLimit}
                    onChange={(e) => setEditLimit(e.target.value)}
                    style={{ minWidth: 140 }}
                  />

                  <div style={{ marginLeft: "auto", display: "flex", gap: 8 }}>
                    <button className="btn btn-primary" onClick={saveEdit} disabled={busy}>
                      Save
                    </button>
                    <button className="btn btn-ghost" onClick={cancelEdit} disabled={busy}>
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                // --- VIEW MODE ---
                <>
                  <div>
                    <div className="h2" style={{ fontSize: 18, marginBottom: 4 }}>
                      {/* Fallback em dash if category was deleted or null */}
                      {cats.find((c) => c.id === b.category_id)?.name || "—"}
                    </div>
                    {/* Intentionally hide period/month in the subtitle per your request */}
                    <div className="muted">Limit: {b.limit_amount.toFixed(2)}</div>
                  </div>

                  <div style={{ display: "flex", gap: 8 }}>
                    <button className="btn" onClick={() => startEdit(b)} disabled={busy}>
                      Edit
                    </button>
                    <button
                      className="btn btn-danger"
                      onClick={() => delBudget(b.id)}
                      disabled={busy}
                    >
                      Delete
                    </button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
