<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { shopifyApi } from '$lib/api';
  import { AuthService } from '$lib/auth';

  let shop: string = '';
  let isLoading: boolean = false;

  onMount(() => {
    // Check if user is already authenticated
    if (AuthService.isAuthenticated()) {
      goto('/dashboard');
    }
  });

  function handleInstall() {
    if (!shop.trim()) {
      alert('Please enter your shop name');
      return;
    }

    isLoading = true;
    const installUrl = shopifyApi.getInstallUrl(shop.trim());
    window.location.href = installUrl;
  }

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleInstall();
    }
  }
</script>

<style>
  .landing-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    text-align: center;
    padding: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .landing-content {
    max-width: 500px;
    background: rgba(255, 255, 255, 0.1);
    padding: 40px;
    border-radius: 12px;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  h1 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
    font-weight: 700;
  }

  .subtitle {
    font-size: 1.2rem;
    margin-bottom: 2rem;
    opacity: 0.9;
  }

  .install-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .input-group {
    display: flex;
    gap: 0.5rem;
  }

  input {
    flex: 1;
    padding: 12px 16px;
    border: none;
    border-radius: 6px;
    font-size: 16px;
    background: rgba(255, 255, 255, 0.9);
    color: #333;
  }

  input::placeholder {
    color: #666;
  }

  .install-btn {
    background-color: #28a745;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 16px;
    font-weight: 600;
    transition: background-color 0.2s;
    min-width: 120px;
  }

  .install-btn:hover:not(:disabled) {
    background-color: #218838;
  }

  .install-btn:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }

  .spinner {
    border: 2px solid transparent;
    border-top: 2px solid white;
    border-radius: 50%;
    width: 16px;
    height: 16px;
    animation: spin 1s linear infinite;
    display: inline-block;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .help-text {
    font-size: 0.9rem;
    opacity: 0.8;
    margin-top: 1rem;
  }
</style>

<svelte:head>
  <title>Forecasting App - Connect Your Shopify Store</title>
</svelte:head>

<div class="landing-container">
  <div class="landing-content">
    <h1>Forecasting App</h1>
    <p class="subtitle">Connect your Shopify store to get started with advanced forecasting</p>
    
    <div class="install-form">
      <div class="input-group">
        <input
          type="text"
          bind:value={shop}
          on:keypress={handleKeyPress}
          placeholder="your-shop-name.myshopify.com"
          disabled={isLoading}
        />
        <button
          class="install-btn"
          on:click={handleInstall}
          disabled={isLoading}
        >
          {#if isLoading}
            <span class="spinner"></span>
          {:else}
            Install
          {/if}
        </button>
      </div>
      
      <p class="help-text">
        Enter your shop name (e.g., "mystore" for mystore.myshopify.com)
      </p>
    </div>
  </div>
</div>
  