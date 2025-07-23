<style>
  .container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    text-align: center;
    padding: 2rem;
  }

  .error-code {
    font-size: 6rem;
    font-weight: bold;
    color: #e74c3c;
    margin-bottom: 1rem;
  }

  .error-message {
    font-size: 1.5rem;
    color: #333;
    margin-bottom: 1rem;
  }

  .error-description {
    font-size: 1rem;
    color: #666;
    max-width: 600px;
    line-height: 1.6;
  }

  .back-button {
    margin-top: 2rem;
    padding: 0.75rem 1.5rem;
    background-color: #007cba;
    color: white;
    text-decoration: none;
    border-radius: 4px;
    font-weight: 500;
    transition: background-color 0.2s;
  }

  .back-button:hover {
    background-color: #005a87;
  }
</style>

<div class="container">
  <div class="error-code">404</div>
  <div class="error-message">Page Not Found</div>
  <div class="error-description">
    This Shopify app requires a valid shop parameter to function properly. 
    Please install the app from your Shopify admin panel or contact your administrator.
  </div>
  <a href="/" class="back-button">Go Back</a>
</div> 