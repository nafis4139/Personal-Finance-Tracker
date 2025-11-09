import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = { vus: 10, duration: '30s' }; // small smoke perf

const BASE = __ENV.BASE || 'http://localhost:8081/api';

export default function () {
  // health
  let res = http.get(`${BASE}/healthz`);
  check(res, { '200 health': (r) => r.status === 200 });

  // login (use a seeded user)
  res = http.post(`${BASE}/login`, JSON.stringify({ email: 'nafis@example.com', password: 'secret123' }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { '200/401 login': (r) => r.status === 200 || r.status === 401 });

  sleep(1);
}
