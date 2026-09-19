import { useEffect, useState } from 'react';
import { fetchJourney, fetchRecentEvents } from '../api';
import JourneyModal from './JourneyModal';
import Pagination, { PAGE_SIZE } from './Pagination';

export default function EventDashboard() {
  const [events, setEvents] = useState(null);
  const [error, setError] = useState(null);
  const [searchValue, setSearchValue] = useState('');
  const [searchError, setSearchError] = useState(null);
  const [journey, setJourney] = useState(null); // { orderId, hops }
  const [page, setPage] = useState(1);

  useEffect(() => {
    fetchRecentEvents().then(setEvents).catch((err) => setError(err.message));
  }, []);

  const pageEvents = events ? events.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE) : events;

  function openJourneyFor(orderId) {
    fetchJourney({ orderId })
      .then((hops) => setJourney({ orderId, hops }))
      .catch((err) => setError(err.message));
  }

  function runSearch() {
    const value = searchValue.trim();
    setSearchError(null);
    if (!value) return;

    fetchJourney({ eventId: value })
      .then((hops) => {
        if (!hops.length) {
          setSearchError(`No events found for event_id “${value}”.`);
          return;
        }
        setJourney({ orderId: hops[0].order_id, hops });
      })
      .catch((err) => setSearchError(`Search failed: ${err.message}`));
  }

  return (
    <section>
      <div className="view-head">
        <h2>Event Dashboard</h2>
        <span className="hint">Paste the event_id from an order API response to trace its journey.</span>
      </div>

      <div className="event-search-panel">
        <div className="search-box">
          <label htmlFor="event-search">event_id</label>
          <div className="search-row">
            <input
              id="event-search"
              type="text"
              autoComplete="off"
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && runSearch()}
              placeholder="e.g. 956a8955-41f8-4da7-b14c-4615fe70cc95"
            />
            <button type="button" onClick={runSearch}>
              Search
            </button>
          </div>
          {searchError && <div className="search-error">{searchError}</div>}
        </div>
      </div>

      {error && (
        <div className="banner">
          Couldn't load events from <code>analytics-service</code>: {error}
        </div>
      )}

      <div className="table-wrap">
        <table className="data-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>event_id</th>
              <th>order_id</th>
              <th>Topic</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {events === null && !error && (
              <tr>
                <td className="loading-state" colSpan={5}>
                  Loading&hellip;
                </td>
              </tr>
            )}
            {events && events.length === 0 && (
              <tr>
                <td className="empty-state" colSpan={5}>
                  No events yet.
                </td>
              </tr>
            )}
            {pageEvents &&
              pageEvents.map((e) => (
                <tr
                  key={e.event_id}
                  className="clickable"
                  tabIndex={0}
                  role="button"
                  onClick={() => openJourneyFor(e.order_id)}
                  onKeyDown={(ev) => {
                    if (ev.key === 'Enter' || ev.key === ' ') {
                      ev.preventDefault();
                      openJourneyFor(e.order_id);
                    }
                  }}
                >
                  <td>{new Date(e.timestamp).toLocaleString()}</td>
                  <td>{e.event_id}</td>
                  <td>{e.order_id}</td>
                  <td>{e.topic}</td>
                  <td>{e.status}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {events && <Pagination page={page} totalRows={events.length} onChange={setPage} />}

      {journey && <JourneyModal orderId={journey.orderId} hops={journey.hops} onClose={() => setJourney(null)} />}
    </section>
  );
}
