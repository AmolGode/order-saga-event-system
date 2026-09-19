import { useEffect, useRef } from 'react';

function fmtDelta(ms) {
  return ms < 1000 ? `${Math.round(ms)}ms` : `${(ms / 1000).toFixed(1)}s`;
}

function severityOf(topic) {
  if (topic.endsWith('-dlq')) return 'dead';
  if (topic === 'payment-failed' || topic === 'inventory-failed') return 'critical';
  if (topic === 'payment-success') return 'success';
  return '';
}

function outcomeFor(hops) {
  const last = hops[hops.length - 1].topic;
  if (last === 'payment-success') return { cls: 'success', text: 'Order confirmed.' };
  if (last === 'payment-failed')
    return { cls: 'critical', text: 'Payment declined — order cancelled, reserved stock released back to inventory.' };
  if (last === 'inventory-failed')
    return { cls: 'warning', text: 'Cancelled — insufficient stock at the time of request.' };
  if (last.endsWith('-dlq'))
    return { cls: 'critical', text: `Stuck — exhausted 3 retries and landed on ${last}. Needs manual replay.` };
  return { cls: 'success', text: 'In progress.' };
}

export default function JourneyModal({ orderId, hops, onClose }) {
  const closeBtnRef = useRef(null);

  useEffect(() => {
    closeBtnRef.current?.focus();
    function onKeydown(ev) {
      if (ev.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', onKeydown);
    return () => document.removeEventListener('keydown', onKeydown);
  }, [onClose]);

  const hasHops = hops && hops.length > 0;
  const outcome = hasHops ? outcomeFor(hops) : null;

  return (
    <div
      className="modal-overlay"
      onClick={(ev) => {
        if (ev.target === ev.currentTarget) onClose();
      }}
    >
      <div className="modal-box" role="dialog" aria-modal="true">
        <div className="modal-head">
          <h2>Order journey</h2>
          <button className="modal-close" ref={closeBtnRef} onClick={onClose} aria-label="Close">
            &times;
          </button>
        </div>
        <p className="modal-sub">order_id: {orderId}</p>

        {!hasHops && <p>No events found.</p>}

        {hasHops && (
          <>
            <ul className="timeline">
              {hops.map((h, idx) => {
                const delta = idx > 0 ? new Date(h.timestamp) - new Date(hops[idx - 1].timestamp) : null;
                return (
                  <li key={h.event_id} className={severityOf(h.topic)}>
                    <div className="timeline-head">
                      <span className="timeline-topic">{h.topic}</span>
                      {delta !== null && <span className="timeline-delta">+{fmtDelta(delta)}</span>}
                    </div>
                    <span className="timeline-time">{new Date(h.timestamp).toLocaleString()}</span>
                    <span className="timeline-eid">event_id: {h.event_id}</span>
                    <pre className="raw-payload">{JSON.stringify(h.payload, null, 2)}</pre>
                  </li>
                );
              })}
            </ul>
            <div className={`timeline-outcome ${outcome.cls}`}>{outcome.text}</div>
          </>
        )}
      </div>
    </div>
  );
}
