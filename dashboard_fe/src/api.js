import { BASE_URLS } from './config';

async function getJson(url) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`${url} responded with ${res.status}`);
  }
  return res.json();
}

export function fetchModel(serviceKey, path) {
  return getJson(BASE_URLS[serviceKey] + path);
}

export function fetchRecentEvents() {
  return getJson(BASE_URLS.analytics + '/api/events/');
}

// Resolves either id to the same thing: the full ordered list of events for
// one order. event_id is unique per hop (a fresh one is generated at every
// republish), so searching by it only works because the backend resolves it
// to that hop's order_id first, then returns the whole chain.
export function fetchJourney({ orderId, eventId }) {
  const params = orderId
    ? `order_id=${encodeURIComponent(orderId)}`
    : `event_id=${encodeURIComponent(eventId)}`;
  return getJson(`${BASE_URLS.analytics}/api/events/?${params}`);
}
