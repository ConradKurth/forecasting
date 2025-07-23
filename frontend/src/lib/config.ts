import { 
  PUBLIC_BACKEND_API_URL, 
  PUBLIC_FRONTEND_URL, 
  PUBLIC_SHOPIFY_API_KEY 
} from '$env/static/public';

export const config = {
  api: {
    baseUrl: PUBLIC_BACKEND_API_URL || 'http://localhost:3001',
  },
  frontend: {
    url: PUBLIC_FRONTEND_URL || 'http://localhost:5173',
  },
  shopify: {
    apiKey: PUBLIC_SHOPIFY_API_KEY || '',
  },
} as const;

export type Config = typeof config; 