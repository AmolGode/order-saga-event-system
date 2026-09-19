import http from 'k6/http';
import { check } from 'k6';

// options = global test config, not per-request.
//
// constant-arrival-rate targets a fixed request rate directly, instead of a
// fixed VU count — with plain `vus`, actual req/sec depends on how fast each
// request completes, and slows down exactly when the server is under load
// (the opposite of what you want from a load test). This executor instead
// auto-scales VUs (up to maxVUs) to hold `rate` steady regardless of latency.
export const options = {
  scenarios: {
    saga_load: {
      executor: 'constant-arrival-rate',
      rate: 1000, // target: 1000 requests/sec
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 200, // VUs k6 starts with
      maxVUs: 2000, // ceiling k6 can grow to if responses get slow
    },
  },
};

// Real product IDs from order-service's data/product_data.json — the API
// validates product_id against this file, so a random UUID would just 400.
const PRODUCT_IDS = [
  '234bf452-a77c-48aa-86f4-bb0b358ba778',
  'd2350b96-e9fd-401d-8b47-3fa62f5249d9',
  'ee4c2554-84a3-4e78-94bc-8665326e5934',
];

function randomProductId() {
  return PRODUCT_IDS[Math.floor(Math.random() * PRODUCT_IDS.length)];
}

// This function is what k6 actually runs — every VU calls it in a tight
// loop, over and over, until `duration` elapses. One call = one iteration.
export default function () {
  const payload = JSON.stringify({
    customer_id: crypto.randomUUID(),
    product_id: randomProductId(),
    qty: Math.floor(Math.random() * 5) + 1, // 1-5
  });

  // http.post(url, body, params) — a single HTTP request.
  // localhost:8100 = the host-published port from docker-compose.yml
  // ("8100:8000"), so this works whether k6 runs natively on your Mac or in
  // Docker Desktop's default bridge network.
  const res = http.post('http://localhost:8100/orders/create_order/', payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  // check() is k6's assertion — it doesn't stop the test on failure, it just
  // records a pass/fail rate you see in the final summary.
  check(res, { 'status is 201': (r) => r.status === 201 });
}
