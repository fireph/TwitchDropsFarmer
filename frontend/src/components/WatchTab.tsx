import { Component, createSignal, For, Show } from 'solid-js';
import { useAuth } from '../context/AuthContext';

const WatchTab: Component = () => {
  const { appState, addGame, removeGame, reorderGames, startMiner, stopMiner } = useAuth();
  const [newGameName, setNewGameName] = createSignal('');
  const [isAddingGame, setIsAddingGame] = createSignal(false);
  const [draggedIndex, setDraggedIndex] = createSignal<number | null>(null);

  const handleAddGame = async (e: Event) => {
    e.preventDefault();
    if (!newGameName().trim()) return;

    setIsAddingGame(true);
    try {
      await addGame(newGameName().trim());
      setNewGameName('');
    } catch (error) {
      console.error('Failed to add game:', error);
    } finally {
      setIsAddingGame(false);
    }
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  return (
    <div class="space-y-6">
      {/* Header with Miner Controls */}
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-dark-text">
            Watch & Farm Drops
          </h2>
          
          <div class="flex items-center space-x-3">
            <Show when={appState.minerStatus?.isRunning}>
              <button
                onClick={stopMiner}
                class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-medium transition-colors"
              >
                Stop Mining
              </button>
            </Show>
            
            <Show when={!appState.minerStatus?.isRunning}>
              <button
                onClick={startMiner}
                disabled={!appState.authenticated || appState.games.length === 0}
                class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors"
              >
                Start Mining
              </button>
            </Show>
          </div>
        </div>

        {/* Current Status */}
        <Show when={appState.minerStatus}>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div class="bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg p-4">
              <div class="text-sm text-gray-600 dark:text-dark-text-secondary">Current Game</div>
              <div class="font-medium text-gray-900 dark:text-dark-text">
                {appState.minerStatus?.currentGame?.displayName || 'None'}
              </div>
            </div>
            
            <div class="bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg p-4">
              <div class="text-sm text-gray-600 dark:text-dark-text-secondary">Current Stream</div>
              <div class="font-medium text-gray-900 dark:text-dark-text">
                {appState.minerStatus?.currentStream?.userName || 'None'}
              </div>
            </div>
            
            <div class="bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg p-4">
              <div class="text-sm text-gray-600 dark:text-dark-text-secondary">Watch Duration</div>
              <div class="font-medium text-gray-900 dark:text-dark-text">
                {appState.minerStatus?.watchDuration 
                  ? formatDuration(appState.minerStatus.watchDuration) 
                  : '0h 0m'}
              </div>
            </div>
          </div>
        </Show>
      </div>

      {/* Add Game Form */}
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-4">
          Add Game to Watch List
        </h3>
        
        <form onSubmit={handleAddGame} class="flex space-x-3">
          <input
            type="text"
            value={newGameName()}
            onInput={(e) => setNewGameName(e.currentTarget.value)}
            placeholder="Enter game name (e.g., 'Valorant', 'League of Legends')"
            class="flex-1 px-4 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-bg-tertiary text-gray-900 dark:text-dark-text placeholder-gray-500 dark:placeholder-dark-text-secondary focus:ring-2 focus:ring-twitch-purple focus:border-transparent"
            disabled={isAddingGame()}
          />
          <button
            type="submit"
            disabled={isAddingGame() || !newGameName().trim()}
            class="px-6 py-2 bg-twitch-gradient hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-opacity"
          >
            {isAddingGame() ? 'Adding...' : 'Add Game'}
          </button>
        </form>
      </div>

      {/* Games List */}
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-4">
          Watch Priority List
        </h3>
        
        <Show 
          when={appState.games.length > 0}
          fallback={
            <div class="text-center py-8 text-gray-500 dark:text-dark-text-secondary">
              No games in your watch list. Add some games to start farming drops!
            </div>
          }
        >
          <div class="space-y-3">
            <For each={appState.games}>
              {(game, index) => (
                <div class="flex items-center space-x-4 p-4 bg-gray-50 dark:bg-dark-bg-tertiary rounded-lg hover:bg-gray-100 dark:hover:bg-dark-bg transition-colors">
                  <div class="shrink-0">
                    <span class="inline-flex items-center justify-center w-8 h-8 bg-twitch-purple text-white text-sm font-medium rounded-full">
                      {index() + 1}
                    </span>
                  </div>
                  
                  <div class="flex-1">
                    <h4 class="font-medium text-gray-900 dark:text-dark-text">
                      {game.displayName}
                    </h4>
                  </div>
                  
                  <button
                    onClick={() => removeGame(game.id)}
                    class="p-2 text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              )}
            </For>
          </div>
        </Show>
      </div>

      {/* Real-time Logs */}
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-4">
          Mining Logs
        </h3>
        
        <div class="bg-gray-900 rounded-lg p-4 h-64 overflow-y-auto font-mono text-sm">
          <Show 
            when={appState.logs.length > 0}
            fallback={
              <div class="text-gray-500 text-center py-8">
                No logs yet. Start mining to see activity.
              </div>
            }
          >
            <For each={appState.logs.slice(-50)}>
              {(log) => (
                <div class={`mb-1 ${
                  log.level === 'ERROR' ? 'text-red-400' :
                  log.level === 'WARNING' ? 'text-yellow-400' :
                  log.level === 'SUCCESS' ? 'text-green-400' :
                  log.level === 'INFO' ? 'text-blue-400' :
                  'text-gray-300'
                }`}>
                  <span class="text-gray-500">
                    [{formatTimestamp(log.timestamp)}]
                  </span>
                  <span class="ml-2 font-medium">
                    [{log.level}]
                  </span>
                  <span class="ml-2">
                    {log.message}
                  </span>
                </div>
              )}
            </For>
          </Show>
        </div>
      </div>
    </div>
  );
};

export default WatchTab;
