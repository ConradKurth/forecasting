import { apiClient } from './client';

export interface User {
  shop: string;
  userId: string;
}

export interface DashboardData {
  message: string;
  shop: string;
  data: string[];
}

export class AuthApiService {
  /**
   * Get the current user's profile
   */
  async getUserProfile(): Promise<User> {
    return apiClient.get<User>('/v1/dashboard/profile', true);
  }

  /**
   * Get dashboard data for the authenticated user
   */
  async getDashboardData(): Promise<DashboardData> {
    return apiClient.get<DashboardData>('/v1/dashboard/data', true);
  }

  /**
   * Check if a token is valid by making a test API call
   */
  async validateToken(): Promise<boolean> {
    try {
      await this.getUserProfile();
      return true;
    } catch {
      return false;
    }
  }
}

// Export a singleton instance
export const authApi = new AuthApiService();
