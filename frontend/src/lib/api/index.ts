// Re-export all API services for easy importing
export { apiClient, ApiClient } from './client';
export { shopifyApi, ShopifyApiService } from './shopify';
export { authApi, AuthApiService } from './auth';
export type { ShopifyInstallParams, ShopifyCallbackParams } from './shopify';
export type { User, DashboardData } from './auth'; 