import { useEffect, useState } from 'react';
import { fetchModel } from '../api';
import Pagination, { PAGE_SIZE } from './Pagination';

function formatCell(value) {
  if (value === null || value === undefined) return '—';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}

export default function ModelTable({ serviceKey, serviceName, model }) {
  const [rows, setRows] = useState(null);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);

  useEffect(() => {
    setRows(null);
    setError(null);
    setPage(1);
    fetchModel(serviceKey, model.path)
      .then(setRows)
      .catch((err) => setError(err.message));
  }, [serviceKey, model.path]);

  const columns = rows && rows.length ? Object.keys(rows[0]) : [];
  const pageRows = rows ? rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE) : rows;

  return (
    <section>
      <div className="view-head">
        <h2>{model.name}</h2>
        <span className="hint">{serviceName} &middot; read-only</span>
      </div>

      {error && (
        <div className="banner">
          Couldn't load this from <code>{serviceKey}</code>: {error}
        </div>
      )}

      <div className="table-wrap">
        <table className="data-table">
          {columns.length > 0 && (
            <thead>
              <tr>
                {columns.map((c) => (
                  <th key={c}>{c}</th>
                ))}
              </tr>
            </thead>
          )}
          <tbody>
            {rows === null && !error && (
              <tr>
                <td className="loading-state" colSpan={columns.length || 1}>
                  Loading&hellip;
                </td>
              </tr>
            )}
            {rows && rows.length === 0 && (
              <tr>
                <td className="empty-state" colSpan={columns.length || 1}>
                  No rows yet.
                </td>
              </tr>
            )}
            {pageRows &&
              pageRows.map((row, i) => (
                <tr key={row.id ?? row.event_id ?? i}>
                  {columns.map((c) => (
                    <td key={c}>{formatCell(row[c])}</td>
                  ))}
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {rows && <Pagination page={page} totalRows={rows.length} onChange={setPage} />}
    </section>
  );
}
