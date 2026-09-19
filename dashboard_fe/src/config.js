// Every service's base URL comes from its own env var — change these when
// deploying, nothing else in the app needs to know about it.
export const BASE_URLS = {
  order: import.meta.env.VITE_ORDER_SERVICE_URL || 'http://localhost:8100',
  inventory: import.meta.env.VITE_INVENTORY_SERVICE_URL || 'http://localhost:8200',
  payment: import.meta.env.VITE_PAYMENT_SERVICE_URL || 'http://localhost:8300',
  analytics: import.meta.env.VITE_ANALYTICS_SERVICE_URL || 'http://localhost:8400',
};

// Sidebar structure: service -> its own read-only models. Adding a model
// here is all it takes to get a new browsable table.
export const SERVICES = [
  {
    key: 'order',
    name: 'Order Service',
    models: [
      { name: 'Orders', path: '/api/orders/' },
      { name: 'Order Items', path: '/api/order-items/' },
      { name: 'Processed Events', path: '/api/processed-events/' },
    ],
  },
  {
    key: 'inventory',
    name: 'Inventory Service',
    models: [
      { name: 'Products', path: '/api/products/' },
      { name: 'Inventory', path: '/api/inventory/' },
      { name: 'Processed Events', path: '/api/processed-events/' },
    ],
  },
  {
    key: 'payment',
    name: 'Payment Service',
    models: [
      { name: 'Payments', path: '/api/payments/' },
      { name: 'Processed Events', path: '/api/processed-events/' },
    ],
  },
  {
    key: 'analytics',
    name: 'Analytics Service',
    models: [
      { name: 'Analytics Events', path: '/api/events/' },
      { name: 'Dead Letters', path: '/api/dead-letters/' },
    ],
  },
];
