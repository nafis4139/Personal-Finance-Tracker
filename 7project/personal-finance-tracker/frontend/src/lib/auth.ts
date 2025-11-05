// frontend/src/lib/auth.ts

// Minimal token storage utilities backed by localStorage.
// Intended for persisting a short-lived JWT or similar bearer token on the client.

export function getToken(): string | null {
  // Retrieve the current token value, or null if not set.
  return localStorage.getItem("token");
}

export function setToken(t: string) {
  // Persist a token value for subsequent requests.
  localStorage.setItem("token", t);
}

export function clearToken() {
  // Remove any stored token to effectively sign out the session on the client.
  localStorage.removeItem("token");
}
