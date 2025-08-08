<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { AuthService, type User, type DashboardData } from '$lib/auth';
  import { syncApi, type SyncStatus } from '$lib/api/sync';
  import { goto } from '$app/navigation';

  let user: User | null = null;
  let dashboardData: DashboardData | null = null;
  let isLoading = true;
  let error: string | null = null;
  
  // Sync state
  let syncStatus: SyncStatus | null = null;
  let isSyncing = false;
  let syncError: string | null = null;
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  onMount(async () => {
    // Check if user is authenticated
    if (!AuthService.isAuthenticated()) {
      goto('/');
      return;
    }

    try {
      // Load user profile and dashboard data
      [user, dashboardData] = await Promise.all([
        AuthService.getUserProfile(),
        AuthService.getDashboardData()
      ]);
      
      // Load initial sync status
      if (user?.shop) {
        await loadSyncStatus();
      }
    } catch (err) {
      error = 'Failed to load dashboard data';
      console.error(err);
    } finally {
      isLoading = false;
    }
  });

  function handleLogout() {
    AuthService.logout();
  }

  async function loadSyncStatus() {
    if (!user?.shop) return;
    
    try {
      syncStatus = await syncApi.getSyncStatus(user.shop);
      syncError = null;
    } catch (err) {
      syncError = 'Failed to load sync status';
      console.error(err);
    }
  }

  async function triggerSync() {
    if (!user?.shop || isSyncing) return;
    
    isSyncing = true;
    syncError = null;
    
    try {
      syncStatus = await syncApi.triggerSync({ shop_domain: user.shop, force: true });
      
      // Start polling for status updates
      startPolling();
    } catch (err) {
      syncError = 'Failed to trigger sync';
      console.error(err);
      isSyncing = false;
    }
  }

  function startPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
    }
    
    pollInterval = setInterval(async () => {
      if (!user?.shop) return;
      
      try {
        const status = await syncApi.getSyncStatus(user.shop);
        syncStatus = status;
        
        // Stop polling if sync is complete or failed
        if (status.status === 'completed' || status.status === 'failed' || status.status === 'never_synced') {
          stopPolling();
          isSyncing = false;
        }
      } catch (err) {
        console.error('Error polling sync status:', err);
        stopPolling();
        isSyncing = false;
      }
    }, 2000); // Poll every 2 seconds
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
  }

  function getSyncStatusText(status: string): string {
    switch (status) {
      case 'in_progress':
        return 'Syncing...';
      case 'completed':
        return 'Completed';
      case 'failed':
        return 'Failed';
      case 'never_synced':
        return 'Never synced';
      default:
        return status;
    }
  }

  function getSyncStatusColor(status: string): string {
    switch (status) {
      case 'in_progress':
        return '#007cba';
      case 'completed':
        return '#28a745';
      case 'failed':
        return '#dc3545';
      case 'never_synced':
        return '#6c757d';
      default:
        return '#6c757d';
    }
  }

  onDestroy(() => {
    stopPolling();
  });
</script>

<svelte:head>
  <title>Dashboard - Forecasting App</title>
</svelte:head>

{#if isLoading}
  <div class="loading-container">
    <div class="spinner"></div>
    <p>Loading dashboard...</p>
  </div>
{:else if error}
  <div class="error-container">
    <h1>Error</h1>
    <p>{error}</p>
    <button on:click={() => goto('/')}>Go Home</button>
  </div>
{:else if user}
  <div class="dashboard">
    <header class="dashboard-header">
      <h1>Welcome to your Dashboard</h1>
      <div class="user-info">
        <span>Shop: {user.shop}</span>
        <button class="logout-btn" on:click={handleLogout}>Logout</button>
      </div>
    </header>

    <main class="dashboard-content">
      <div class="card">
        <h2>Profile Information</h2>
        <p><strong>Shop:</strong> {user.shop}</p>
        <p><strong>User ID:</strong> {user.userId}</p>
      </div>

      <div class="card">
        <h2>Inventory Sync</h2>
        
        {#if syncStatus}
          <div class="sync-status">
            <p><strong>Status:</strong> 
              <span style="color: {getSyncStatusColor(syncStatus.status)}">
                {getSyncStatusText(syncStatus.status)}
              </span>
            </p>
            
            {#if syncStatus.last_synced}
              <p><strong>Last Synced:</strong> {new Date(syncStatus.last_synced).toLocaleString()}</p>
            {/if}
            
            {#if syncStatus.error}
              <p><strong>Error:</strong> <span class="error-text">{syncStatus.error}</span></p>
            {/if}
          </div>
        {/if}
        
        {#if syncError}
          <p class="error-text">{syncError}</p>
        {/if}
        
        <button 
          class="sync-btn" 
          on:click={triggerSync} 
          disabled={isSyncing || syncStatus?.status === 'in_progress'}
        >
          {#if isSyncing || syncStatus?.status === 'in_progress'}
            <span class="spinner-small"></span>
            Syncing...
          {:else}
            Start Sync
          {/if}
        </button>
      </div>

      {#if dashboardData}
        <div class="card">
          <h2>Dashboard Data</h2>
          <p>{dashboardData.message}</p>
          <ul>
            {#each dashboardData.data as item}
              <li>{item}</li>
            {/each}
          </ul>
        </div>
      {/if}
    </main>
  </div>
{:else}
  <div class="error-container">
    <h1>Authentication Required</h1>
    <p>Please log in to access the dashboard.</p>
    <button on:click={() => goto('/')}>Go Home</button>
  </div>
{/if}

<style>
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    text-align: center;
  }

  .spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid #007cba;
    border-radius: 50%;
    width: 50px;
    height: 50px;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .dashboard {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
  }

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 2px solid #eee;
  }

  .dashboard-header h1 {
    color: #333;
    margin: 0;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 15px;
  }

  .logout-btn {
    background-color: #dc3545;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .logout-btn:hover {
    background-color: #c82333;
  }

  .dashboard-content {
    display: grid;
    gap: 20px;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  }

  .card {
    background: white;
    border-radius: 8px;
    padding: 24px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    border: 1px solid #e1e5e9;
  }

  .card h2 {
    color: #333;
    margin-top: 0;
    margin-bottom: 16px;
  }

  .card p {
    margin-bottom: 8px;
    color: #666;
  }

  .card ul {
    margin: 16px 0 0 0;
    padding-left: 20px;
  }

  .card li {
    margin-bottom: 8px;
    color: #666;
  }

  .error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    text-align: center;
    padding: 20px;
  }

  .error-container h1 {
    color: #dc3545;
    margin-bottom: 16px;
  }

  .error-container button {
    background-color: #007cba;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    margin-top: 16px;
  }

  .error-container button:hover {
    background-color: #005a82;
  }

  .sync-status {
    margin-bottom: 20px;
  }

  .sync-btn {
    background-color: #007cba;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: background-color 0.2s;
  }

  .sync-btn:hover:not(:disabled) {
    background-color: #005a82;
  }

  .sync-btn:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }

  .spinner-small {
    border: 2px solid #f3f3f3;
    border-top: 2px solid #ffffff;
    border-radius: 50%;
    width: 16px;
    height: 16px;
    animation: spin 1s linear infinite;
  }

  .error-text {
    color: #dc3545;
    font-weight: 500;
  }
</style>
