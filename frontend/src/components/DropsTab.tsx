import { Component, For, Show } from 'solid-js';
import { useAuth } from '../context/AuthContext';

const DropsTab: Component = () => {
  const { appState, api } = useAuth();

  const claimDrop = async (dropId: string) => {
    try {
      await api.post(`/drops/${dropId}/claim`);
      // Refresh drops data
      const dropsData = await api.get('/drops');
      appState.setAppState('drops', dropsData.drops || []);
    } catch (error) {
      console.error('Failed to claim drop:', error);
    }
  };

  const getProgressPercentage = (current: number, required: number) => {
    return Math.min((current / required) * 100, 100);
  };

  const formatTimeRemaining = (current: number, required: number) => {
    const remaining = Math.max(required - current, 0);
    const hours = Math.floor(remaining / 60);
    const minutes = remaining % 60;
    return `${hours}h ${minutes}m`;
  };

  return (
    <div class="space-y-6">
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-dark-text mb-6">
          Available Drops
        </h2>
        
        <Show 
          when={appState.authenticated}
          fallback={
            <div class="text-center py-12">
              <div class="text-gray-500 dark:text-dark-text-secondary mb-4">
                Please login with Twitch to view available drops
              </div>
            </div>
          }
        >
          <Show 
            when={appState.drops.length > 0}
            fallback={
              <div class="text-center py-12">
                <div class="text-gray-500 dark:text-dark-text-secondary mb-4">
                  No drops available for your current games
                </div>
                <p class="text-sm text-gray-400 dark:text-dark-text-secondary">
                  Add games to your watch list to see available drops
                </p>
              </div>
            }
          >
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <For each={appState.drops}>
                {(drop) => (
                  <div class="bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg p-6 border border-gray-200 dark:border-dark-border">
                    <div class="flex items-start space-x-4">
                      <img
                        src={drop.imageURL}
                        alt={drop.name}
                        class="w-16 h-16 rounded-lg object-cover shrink-0"
                        loading="lazy"
                      />
                      
                      <div class="flex-1 min-w-0">
                        <h3 class="font-semibold text-gray-900 dark:text-dark-text text-sm mb-1 truncate">
                          {drop.name}
                        </h3>
                        <p class="text-xs text-gray-600 dark:text-dark-text-secondary mb-2">
                          {drop.gameName}
                        </p>
                        
                        {/* Progress Bar */}
                        <div class="mb-3">
                          <div class="flex justify-between text-xs text-gray-600 dark:text-dark-text-secondary mb-1">
                            <span>{drop.currentMinutes} / {drop.requiredMinutes} minutes</span>
                            <span>{getProgressPercentage(drop.currentMinutes, drop.requiredMinutes).toFixed(0)}%</span>
                          </div>
                          <div class="w-full bg-gray-200 dark:bg-dark-bg rounded-full h-2">
                            <div 
                              class="bg-twitch-purple h-2 rounded-full transition-all duration-300"
                              style={`width: ${getProgressPercentage(drop.currentMinutes, drop.requiredMinutes)}%`}
                            />
                          </div>
                        </div>
                        
                        {/* Status and Actions */}
                        <div class="flex items-center justify-between">
                          <Show when={drop.isCompleted && !drop.isClaimed}>
                            <button
                              onClick={() => claimDrop(drop.id)}
                              class="px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-xs font-medium rounded-sm transition-colors"
                            >
                              Claim
                            </button>
                          </Show>
                          
                          <Show when={drop.isClaimed}>
                            <span class="px-3 py-1 bg-gray-500 text-white text-xs font-medium rounded-sm">
                              Claimed
                            </span>
                          </Show>
                          
                          <Show when={!drop.isCompleted && !drop.isClaimed}>
                            <span class="text-xs text-gray-500 dark:text-dark-text-secondary">
                              {formatTimeRemaining(drop.currentMinutes, drop.requiredMinutes)} remaining
                            </span>
                          </Show>
                        </div>
                      </div>
                    </div>
                    
                    <Show when={drop.description}>
                      <p class="text-xs text-gray-600 dark:text-dark-text-secondary mt-3 line-clamp-2">
                        {drop.description}
                      </p>
                    </Show>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Show>
      </div>
    </div>
  );
};

export default DropsTab;
