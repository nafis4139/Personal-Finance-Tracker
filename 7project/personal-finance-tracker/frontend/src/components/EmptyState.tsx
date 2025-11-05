// frontend/src/components/EmptyState.tsx
// Presentational component for displaying an "empty state" UI.
// Renders an icon, a title, an optional subtitle, and an optional action area (e.g., button/link).
// Styling is expected to be provided by CSS classes: empty-state, empty-card, empty-icon, empty-action.

import React from "react";

export default function EmptyState({
  title,
  subtitle,
  action,
}: {
  // Required headline text for the empty state.
  title: string;
  // Optional supporting text beneath the title.
  subtitle?: string;
  // Optional React node for actions (e.g., <button/>, <Link/>).
  action?: React.ReactNode;
}) {
  return (
    <div className="empty-state">
      <div className="empty-card">
        {/* Decorative icon only; aria-hidden prevents it from being read by assistive tech. */}
        <div className="empty-icon" aria-hidden>
          📭
        </div>
        <h3>{title}</h3>
        {/* Conditionally render supporting text when provided. */}
        {subtitle && <p>{subtitle}</p>}
        {/* Conditionally render action area for interactive controls. */}
        {action && <div className="empty-action">{action}</div>}
      </div>
    </div>
  );
}
