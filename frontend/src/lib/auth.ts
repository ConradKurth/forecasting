import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authApi, type User, type DashboardData } from '$lib/api/auth';

export class AuthService {
  private static readonly TOKEN_KEY = 'auth_token';

  static getToken(): string | null {
    if (!browser) return null;
    return localStorage.getItem(this.TOKEN_KEY);
  }

  static setToken(token: string): void {
    if (!browser) return;
    localStorage.setItem(this.TOKEN_KEY, token);
  }

  static removeToken(): void {
    if (!browser) return;
    localStorage.removeItem(this.TOKEN_KEY);
  }

  static isAuthenticated(): boolean {
    const token = this.getToken();
    if (!token) return false;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Date.now() / 1000;
      return payload.exp > now;
    } catch {
      return false;
    }
  }

  static async getUserProfile(): Promise<User | null> {
    if (!this.isAuthenticated()) return null;

    try {
      return await authApi.getUserProfile();
    } catch (error) {
      console.error('Error fetching user profile:', error);
      this.removeToken();
      return null;
    }
  }

  static async getDashboardData(): Promise<DashboardData | null> {
    if (!this.isAuthenticated()) return null;

    try {
      return await authApi.getDashboardData();
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
      throw error;
    }
  }

  static async validateToken(): Promise<boolean> {
    if (!this.isAuthenticated()) return false;

    try {
      return await authApi.validateToken();
    } catch {
      this.removeToken();
      return false;
    }
  }

  static logout(): void {
    this.removeToken();
    goto('/');
  }

  static requireAuth(): void {
    if (!this.isAuthenticated()) {
      goto('/');
    }
  }
}

// Re-export types for convenience
export type { User, DashboardData } from '$lib/api/auth';
