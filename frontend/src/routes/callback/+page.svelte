<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { AuthService } from '$lib/auth';

  let isLoading = true;
  let error: string | null = null;
  let status = 'Processing authentication...';

  onMount(async () => {
    try {
      const urlParams = new URLSearchParams(window.location.search);
      const shop = urlParams.get('shop');
      const token = urlParams.get('token');
      
      if (!shop || !token) {
        throw new Error('Missing shop or token parameter');
      }

      status = 'Processing authentication...';
      
      // Store the JWT token
      AuthService.setToken(token);
      
      status = 'Authentication successful! Redirecting...';
      
      // Redirect to dashboard
      setTimeout(() => {
        goto('/dashboard');
      }, 1000);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Authentication failed';
    } finally {
      isLoading = false;
    }
  });
</script>

<svelte:head>
  <title>Authentication - Forecasting App</title>
</svelte:head>

<div class="callback-container">
  {#if isLoading}
    <div class="spinner"></div>
    <h2>Authenticating...</h2>
    <p>{status}</p>
  {:else if error}
    <div class="error">
      <h2>Authentication Failed</h2>
      <p>{error}</p>
      <button on:click={() => goto('/')}>Try Again</button>
    </div>
  {:else}
    <div class="success">
      <h2>Authentication Successful!</h2>
      <p>You will be redirected to the dashboard shortly.</p>
    </div>
  {/if}
</div>

<style>
  .callback-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    text-align: center;
    padding: 20px;
  }

  .spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid #007cba;
    border-radius: 50%;
    width: 50px;
    height: 50px;
    animation: spin 1s linear infinite;
    margin-bottom: 2rem;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  h2 {
    color: #333;
    margin-bottom: 1rem;
  }

  p {
    color: #666;
    margin-bottom: 1rem;
  }

  .error {
    max-width: 400px;
  }

  .error h2 {
    color: #dc3545;
  }

  .success h2 {
    color: #28a745;
  }

  button {
    background-color: #007cba;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    margin-top: 1rem;
  }

  button:hover {
    background-color: #005a82;
  }
</style>
