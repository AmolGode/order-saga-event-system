export const PAGE_SIZE = 50;

export default function Pagination({ page, totalRows, onChange }) {
  const totalPages = Math.max(1, Math.ceil(totalRows / PAGE_SIZE));
  if (totalRows === 0) return null;

  const start = (page - 1) * PAGE_SIZE + 1;
  const end = Math.min(page * PAGE_SIZE, totalRows);

  return (
    <div className="pagination">
      <span className="pagination-info">
        {start}&ndash;{end} of {totalRows}
      </span>
      <div className="pagination-controls">
        <button type="button" disabled={page <= 1} onClick={() => onChange(page - 1)}>
          Prev
        </button>
        <span className="pagination-page">
          Page {page} of {totalPages}
        </span>
        <button type="button" disabled={page >= totalPages} onClick={() => onChange(page + 1)}>
          Next
        </button>
      </div>
    </div>
  );
}
