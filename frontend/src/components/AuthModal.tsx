// src/components/AuthModal.tsx
import { Component, createSignal, Show } from 'solid-js';
import { useAuth } from '../context/AuthContext';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const AuthModal: Component<AuthModalProps> = (props) => {
  const { api, setAppState } = useAuth();
  const [authStep, setAuthStep] = createSignal<'start' | 'waiting' | 'success' | 'error'>('start');
  const [authData, setAuthData] = createSignal<any>(null);
  const [isLoading, setIsLoading] = createSignal(false);
  const [errorMessage, setErrorMessage] = createSignal('');

  const startAuth = async () => {
    setIsLoading(true);
    setErrorMessage('');
    
    try {
      // Get the auth URL and device code
      const data = await api.get('/auth/url');
      setAuthData(data);
      setAuthStep('waiting');
      
      // Open Twitch auth page in a new window
      const authWindow = window.open(data.verification_uri, '_blank', 'width=600,height=700');
      
      // Start the callback process with the device code
      await api.post('/auth/callback', {
        deviceCode: data.device_code,  // Fix: pass the deviceCode from the auth data
        interval: data.interval,
      });
      
      // Poll for authentication completion by checking auth status
      const pollForCompletion = async () => {
        try {
          const status = await api.get('/auth/status');
          console.log('Auth status check:', status);
          
          if (status.authenticated) {
            setAuthStep('success');
            setAppState('authenticated', true);
            
            // Close the auth window if it's still open
            if (authWindow && !authWindow.closed) {
              authWindow.close();
            }
            
            // Close modal after a short delay and reload data
            setTimeout(() => {
              handleClose();
              window.location.reload();
            }, 2000);
            return true; // Stop polling
          }
          return false; // Continue polling
        } catch (error) {
          console.error('Error checking auth status:', error);
          return false; // Continue polling
        }
      };
      
      // Start polling every 2 seconds
      const maxPollingTime = 300000; // 5 minutes
      const pollingInterval = 2000; // 2 seconds
      const startTime = Date.now();
      
      const poll = async () => {
        const elapsed = Date.now() - startTime;
        
        if (elapsed > maxPollingTime) {
          setAuthStep('error');
          setErrorMessage('Authentication timed out. Please try again.');
          return;
        }
        
        const completed = await pollForCompletion();
        if (!completed) {
          setTimeout(poll, pollingInterval);
        }
      };
      
      // Start polling
      setTimeout(poll, pollingInterval);
      
    } catch (error: any) {
      console.error('Auth failed:', error);
      setAuthStep('error');
      setErrorMessage(error.message || 'Failed to start authentication process');
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    // Reset state
    setAuthStep('start');
    setAuthData(null);
    setErrorMessage('');
    setIsLoading(false);
    
    props.onClose();
  };

  const handleRetry = () => {
    setAuthStep('start');
    setErrorMessage('');
    setAuthData(null);
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (error) {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
  };

  return (
    <Show when={props.isOpen}>
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50 backdrop-blur-sm">
        <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xl border border-gray-200 dark:border-dark-border max-w-md w-full p-6 animate-slide-up">
          {/* Header */}
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-bold text-gray-900 dark:text-dark-text">
              Login with Twitch
            </h2>
            <button
              onClick={handleClose}
              class="text-gray-400 hover:text-gray-600 dark:hover:text-dark-text transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Start Step */}
          <Show when={authStep() === 'start'}>
            <div class="text-center">
              <div class="w-16 h-16 bg-twitch-gradient rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg">
                <svg class="w-8 h-8 text-white" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.64 5.93L13.07 4.5A2 2 0 0115.9 4.5L19.5 8.1A2 2 0 0119.5 10.93L18.07 12.36L11.64 5.93Z"/>
                  <path d="M11.64 18.07L13.07 19.5A2 2 0 0015.9 19.5L19.5 15.9A2 2 0 0019.5 13.07L18.07 11.64L11.64 18.07Z"/>
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-2">
                Connect Your Twitch Account
              </h3>
              <p class="text-gray-600 dark:text-dark-text-secondary mb-6 text-sm">
                This uses the same SmartTV login method as TwitchDropsMiner for secure authentication without passwords.
              </p>
              <button
                onClick={startAuth}
                disabled={isLoading()}
                class="w-full px-6 py-3 bg-twitch-gradient hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-opacity shadow-lg"
              >
                {isLoading() ? (
                  <div class="flex items-center justify-center space-x-2">
                    <div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    <span>Connecting...</span>
                  </div>
                ) : (
                  'Connect with Twitch'
                )}
              </button>
            </div>
          </Show>

          {/* Waiting Step */}
          <Show when={authStep() === 'waiting'}>
            <div class="text-center">
              <div class="w-16 h-16 bg-yellow-500 rounded-full flex items-center justify-center mx-auto mb-4 animate-pulse shadow-lg">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-2">
                Waiting for Authentication
              </h3>
              <p class="text-gray-600 dark:text-dark-text-secondary mb-4 text-sm">
                Please complete the login process in the opened Twitch window.
              </p>
              
              <Show when={authData()}>
                <div class="bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg p-4 mb-4 border border-gray-200 dark:border-dark-border">
                  <p class="text-sm text-gray-600 dark:text-dark-text-secondary mb-2">
                    Enter this code on Twitch:
                  </p>
                  <div class="flex items-center justify-between">
                    <div class="text-2xl font-mono font-bold text-twitch-purple">
                      {authData()?.user_code}
                    </div>
                    <button
                      onClick={() => copyToClipboard(authData()?.user_code)}
                      class="p-2 text-gray-500 hover:text-gray-700 dark:text-dark-text-secondary dark:hover:text-dark-text transition-colors"
                      title="Copy to clipboard"
                    >
                      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </button>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-dark-text-secondary mt-2">
                    Visit: {authData()?.verification_uri}
                  </p>
                </div>
              </Show>
              
              <div class="flex items-center justify-center space-x-2 text-sm text-gray-500 dark:text-dark-text-secondary">
                <div class="w-3 h-3 bg-twitch-purple rounded-full animate-bounce" />
                <span>This window will close automatically when complete</span>
              </div>
            </div>
          </Show>

          {/* Success Step */}
          <Show when={authStep() === 'success'}>
            <div class="text-center">
              <div class="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-2">
                Authentication Successful!
              </h3>
              <p class="text-gray-600 dark:text-dark-text-secondary text-sm">
                You can now start farming Twitch drops. The page will refresh automatically.
              </p>
            </div>
          </Show>

          {/* Error Step */}
          <Show when={authStep() === 'error'}>
            <div class="text-center">
              <div class="w-16 h-16 bg-red-500 rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg">
                <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-2">
                Authentication Failed
              </h3>
              <p class="text-red-600 dark:text-red-400 text-sm mb-4">
                {errorMessage()}
              </p>
              <div class="flex space-x-3">
                <button
                  onClick={handleRetry}
                  class="flex-1 px-4 py-2 bg-twitch-gradient hover:opacity-90 text-white rounded-lg font-medium transition-opacity"
                >
                  Try Again
                </button>
                <button
                  onClick={handleClose}
                  class="flex-1 px-4 py-2 bg-gray-500 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </Show>

          {/* Footer Info */}
          <div class="mt-6 pt-4 border-t border-gray-200 dark:border-dark-border">
            <p class="text-xs text-gray-500 dark:text-dark-text-secondary text-center">
              🔒 Secure SmartTV OAuth flow • No passwords stored • Same method as TwitchDropsMiner
            </p>
          </div>
        </div>
      </div>
    </Show>
  );
};

export default AuthModal;