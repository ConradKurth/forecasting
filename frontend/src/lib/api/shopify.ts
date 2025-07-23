import { apiClient } from './client';

export interface ShopifyInstallParams {
  shop: string;
}

export interface ShopifyCallbackParams {
  shop: string;
  code: string;
  [key: string]: string;
}

export class ShopifyApiService {
  /**
   * Get the Shopify install URL for redirecting users to OAuth
   */
  getInstallUrl(shop: string): string {
    return apiClient.buildUrl('/v1/shopify/install', { shop });
  }

  /**
   * Handle Shopify OAuth callback (if needed for API calls in the future)
   */
  async handleCallback(params: ShopifyCallbackParams): Promise<any> {
    return apiClient.get(`/v1/shopify/callback?${new URLSearchParams(params).toString()}`);
  }

  /**
   * Future: Get shop information after installation
   */
  async getShopInfo(shop: string): Promise<any> {
    return apiClient.get(`/v1/shopify/shop/${encodeURIComponent(shop)}`);
  }

  /**
   * Future: Make authenticated Shopify API calls
   */
  async makeShopifyApiCall<T>(shop: string, endpoint: string, method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET', data?: any): Promise<T> {
    const apiEndpoint = `/v1/shopify/api/${encodeURIComponent(shop)}${endpoint}`;
    
    switch (method) {
      case 'POST':
        return apiClient.post<T>(apiEndpoint, data);
      case 'PUT':
        return apiClient.put<T>(apiEndpoint, data);
      case 'DELETE':
        return apiClient.delete<T>(apiEndpoint);
      default:
        return apiClient.get<T>(apiEndpoint);
    }
  }
}

// Export a singleton instance
export const shopifyApi = new ShopifyApiService(); 