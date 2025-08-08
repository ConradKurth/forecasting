import { apiClient } from './client';

export interface SyncStatus {
  integration_id: string;
  status: string;
  last_synced?: string | null;
  error?: string;
}

export interface TriggerSyncRequest {
  shop_domain: string;
  force?: boolean;
}

export class SyncApiService {
  /**
   * Get the sync status for a shop
   */
  async getSyncStatus(shopDomain: string): Promise<SyncStatus> {
    return apiClient.get<SyncStatus>(`/v1/sync/status/${shopDomain}`, true);
  }

  /**
   * Trigger a sync for a shop
   */
  async triggerSync(request: TriggerSyncRequest): Promise<SyncStatus> {
    return apiClient.post<SyncStatus>('/v1/sync/trigger', request, true);
  }
}

// Export a singleton instance
export const syncApi = new SyncApiService();
