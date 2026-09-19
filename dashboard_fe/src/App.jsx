import { useState } from 'react';
import Sidebar from './components/Sidebar';
import ModelTable from './components/ModelTable';
import EventDashboard from './components/EventDashboard';

export default function App() {
  const [view, setView] = useState({ type: 'eventDashboard' });

  return (
    <div className="app">
      <Sidebar currentView={view} onSelect={setView} />
      <main className="main">
        {view.type === 'eventDashboard' && <EventDashboard />}
        {view.type === 'model' && (
          <ModelTable serviceKey={view.serviceKey} serviceName={view.serviceName} model={view.model} />
        )}
      </main>
    </div>
  );
}
