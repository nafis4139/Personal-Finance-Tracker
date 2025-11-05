// frontend/src/pages/Budgets.tsx

// Budgets management page:
// - Displays monthly budgets and quick stats.
// - Supports creating and deleting budgets.
// - Fetches categories (for expense-only selection) and budgets for a selected month.

import { useEffect, useMemo, useState } from "react";
import { api, apiList } from "../lib/api";
import EmptyState from "../components/EmptyState";

// Minimal API models used by this page.
type Category = { id:number; name:string; type:"income"|"expense" };
type Budget = { id:number; user_id:number; category_id:number|null; period_month:string; limit_amount:number; created_at:string; };

// Helper: format current date as YYYY-MM for <input type="month"> default value.
function yyyyMm(d=new Date()){
  const m=(d.getMonth()+1).toString().padStart(2,"0");
  return `${d.getFullYear()}-${m}`;
}
// Helper: render a YYYY-MM string as a localized month label (e.g., "March 2025").
function niceMonth(yyyyMm:string){
  const [y,m] = yyyyMm.split("-").map(Number);
  return new Date(y, m-1, 1).toLocaleString(undefined, { month:"long", year:"numeric" });
}

export default function Budgets(){
  // Current month selection.
  const [month, setMonth] = useState(yyyyMm());
  // Cached categories for the selector.
  const [cats, setCats] = useState<Category[]>([]);
  // Budgets for the selected month.
  const [items, setItems] = useState<Budget[]>([]);
  // Busy flag for mutating requests (create/delete).
  const [busy, setBusy] = useState(false);
  // Loading flag for initial/fetch operations.
  const [loading, setLoading] = useState(false);
  // Message for surface-level error reporting.
  const [msg, setMsg] = useState<string | null>(null);

  // Form state for creation.
  const [catId, setCatId] = useState<number | "none">("none");
  const [limit, setLimit] = useState<string>("");

  // Fetch budgets and categories for the selected month.
  async function load(){
    try{
      setLoading(true); setMsg(null);
      const [list, cs] = await Promise.all([
        apiList<Budget>(`/budgets?month=${month}`),
        apiList<Category>("/categories")
      ]);
      setItems(list); setCats(cs);
    }catch(e:any){ setMsg(e.message); setItems([]);
    } finally{ setLoading(false); }
  }
  // Re-load whenever the selected month changes.
  useEffect(()=>{ load(); }, [month]);

  // Create a budget and prepend it to the list on success.
  async function addBudget(){
    if(catId==="none" || !limit) return;
    try{
      setBusy(true); setMsg(null);
      const b = await api<Budget>("/budgets", { method:"POST", body: JSON.stringify({
        category_id: catId, period_month: month, limit_amount: Number(limit)
      })});
      setItems([b, ...items]);
      setLimit(""); setCatId("none");
    }catch(e:any){ setMsg(e.message);
    } finally{ setBusy(false); }
  }

  // Delete a budget and optimistically remove it from local state.
  async function delBudget(id:number){
    try{
      setBusy(true); setMsg(null);
      await api(`/budgets/${id}`, { method:"DELETE" });
      setItems(items.filter(i=>i.id!==id));
    }catch(e:any){ setMsg(e.message);
    } finally{ setBusy(false); }
  }

  // Aggregate total of all budget limits shown.
  const totalLimit = useMemo(()=> items.reduce((a,b)=>a+b.limit_amount,0), [items]);

  return (
    <div className="card section">
      <h1 className="h1">Budgets</h1>
      <p className="muted">Track monthly caps for expense categories.</p>

      <div className="spacer" />

      {/* Month selector drives the query parameter used for loading data. */}
      <div className="row">
        <input className="input" type="month" value={month} onChange={e=>setMonth(e.target.value)} />
      </div>

      <div className="spacer" />

      {/* Create budget form: select expense category and enter a numeric limit. */}
      <div className="card section">
        <div className="h2" style={{marginBottom:10}}>Create budget</div>
        <div className="row">
          <select className="select" value={catId} onChange={e=>setCatId(e.target.value==="none"?"none":Number(e.target.value))}>
            <option value="none">Select category</option>
            {cats.filter(c=>c.type==="expense").map(c=><option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
          <input className="input" placeholder="Limit amount" inputMode="decimal" value={limit} onChange={e=>setLimit(e.target.value)} />
          <button className="btn btn-primary" onClick={addBudget} disabled={busy}>Add</button>
        </div>
      </div>

      {/* Inline error message for fetch/mutation failures. */}
      {msg && (<><div className="spacer" /><div className="card section" style={{borderColor:"rgba(239,68,68,.35)"}}>{msg}</div></>)}

      <div className="spacer" />

      {/* Quick stats card. */}
      <div className="card section" style={{display:"flex", gap:16, flexWrap:"wrap"}}>
        <div>
          <div className="muted">Budgets this month</div>
          <div style={{fontSize:24, fontWeight:800}}>{items.length}</div>
        </div>
        <div>
          <div className="muted">Total limit</div>
          <div style={{fontSize:24, fontWeight:800}}>{totalLimit.toFixed(2)}</div>
        </div>
      </div>

      <div className="spacer" />

      {/* Conditional content: loader, empty state, or the list of budgets. */}
      {loading ? (
        <div className="muted">Loading…</div>
      ) : items.length === 0 ? (
        <EmptyState
          title={`No budgets for ${niceMonth(month)}`}
          subtitle="Create a budget to set a monthly spending cap."
          action={<button className="btn btn-primary" onClick={addBudget} disabled={busy || catId==="none" || !limit}>Add Budget</button>}
        />
      ) : (
        <ul className="list">
          {items.map(b=>(
            <li key={b.id} className="list-item">
              <div>
                <div className="h2" style={{fontSize:18, marginBottom:4}}>
                  {/* Resolve category name if available; fallback to an em dash. */}
                  {cats.find(c=>c.id===b.category_id)?.name || "—"}
                </div>
                <div className="muted">Limit: {b.limit_amount.toFixed(2)} • {b.period_month}</div>
              </div>
              <button className="btn btn-danger" onClick={()=>delBudget(b.id)} disabled={busy}>Delete</button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
