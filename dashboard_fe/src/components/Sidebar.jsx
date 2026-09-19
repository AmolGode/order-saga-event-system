import { SERVICES } from '../config';

export default function Sidebar({ currentView, onSelect }) {
  const isEventDashboard = currentView.type === 'eventDashboard';

  return (
    <nav className="sidebar" aria-label="Dashboard navigation">
      <div className="brand">
        <span className="live-dot" aria-hidden="true"></span>
        <h1>Saga Dashboard</h1>
      </div>

      <div className="nav-group">
        <button
          type="button"
          className={`nav-btn${isEventDashboard ? ' active' : ''}`}
          onClick={() => onSelect({ type: 'eventDashboard' })}
        >
          Event Dashboard
        </button>
      </div>

      <hr className="nav-divider" />

      {SERVICES.map((service) => (
        <div className="nav-group" key={service.key}>
          <div className="nav-group-label">{service.name}</div>
          {service.models.map((model) => {
            const active =
              currentView.type === 'model' &&
              currentView.serviceKey === service.key &&
              currentView.model.path === model.path;
            return (
              <button
                key={model.path}
                type="button"
                className={`nav-btn${active ? ' active' : ''}`}
                onClick={() =>
                  onSelect({ type: 'model', serviceKey: service.key, serviceName: service.name, model })
                }
              >
                {model.name}
              </button>
            );
          })}
        </div>
      ))}
    </nav>
  );
}
