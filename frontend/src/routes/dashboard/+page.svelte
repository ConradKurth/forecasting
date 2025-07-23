<script lang="ts">
  import { onMount } from 'svelte';
  import { AuthService, type User, type DashboardData } from '$lib/auth';
  import { goto } from '$app/navigation';

  let user: User | null = null;
  let dashboardData: DashboardData | null = null;
  let isLoading = true;
  let error: string | null = null;

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
</style>
