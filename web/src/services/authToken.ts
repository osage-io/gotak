// Single source of truth for the Authorization header. Login stores the JWT in
// localStorage as 'authToken' (see Login.tsx); every API client must send it or
// protected endpoints (mapping, positions, entities) return 401.
export function authHeaders(): Record<string, string> {
  const t = typeof localStorage !== 'undefined' ? localStorage.getItem('authToken') : null;
  return t && t !== 'demo-token' ? { Authorization: `Bearer ${t}` } : {};
}
