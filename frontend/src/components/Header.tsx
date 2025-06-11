import { Component } from 'solid-js';
import { useAuth } from '../context/AuthContext';

interface HeaderProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  onMenuClick: () => void;
}

const Header: Component<HeaderProps> = (props) => {
  const { appState, logout, setShowAuthModal } = useAuth();

  return (
    <header class="bg-white dark:bg-dark-bg-secondary border-b border-gray-200 dark:border-dark-border shadow-xs">
      <div class="px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <button
              onClick={props.onMenuClick}
              class="lg:hidden p-2 rounded-md text-gray-500 hover:text-gray-700 dark:text-dark-text-secondary dark:hover:text-dark-text"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-twitch-gradient rounded-lg flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.64 5.93L13.07 4.5A2 2 0 0115.9 4.5L19.5 8.1A2 2 0 0119.5 10.93L18.07 12.36L11.64 5.93Z"/>
                  <path d="M11.64 18.07L13.07 19.5A2 2 0 0015.9 19.5L19.5 15.9A2 2 0 0019.5 13.07L18.07 11.64L11.64 18.07Z"/>
                </svg>
              </div>
              <h1 class="text-xl font-bold text-gray-900 dark:text-dark-text">
                TwitchDropsFarmer
              </h1>
            </div>
          </div>

          <div class="flex items-center space-x-4">
            {/* Miner Status Indicator */}
            <div class="flex items-center space-x-2">
              <div class={`w-3 h-3 rounded-full ${
                appState.minerStatus?.isRunning 
                  ? 'bg-green-500 animate-pulse' 
                  : 'bg-gray-400'
              }`} />
              <span class="text-sm text-gray-600 dark:text-dark-text-secondary">
                {appState.minerStatus?.isRunning ? 'Mining' : 'Stopped'}
              </span>
            </div>

            {/* Auth Status */}
            {appState.authenticated ? (
              <button
                onClick={logout}
                class="px-4 py-2 text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400"
              >
                Logout
              </button>
            ) : (
              <button
                onClick={() => setShowAuthModal(true)}
                class="px-4 py-2 bg-twitch-gradient text-white text-sm font-medium rounded-lg hover:opacity-90 transition-opacity"
              >
                Login with Twitch
              </button>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
