import { Component, createSignal, onMount, onCleanup } from 'solid-js';
import { createStore } from 'solid-js/store';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import WatchTab from './components/WatchTab';
import DropsTab from './components/DropsTab';
import SettingsTab from './components/SettingsTab';
import AuthModal from './components/AuthModal';
import { AuthProvider } from './context/AuthContext';
import { WebSocketProvider } from './context/WebSocketContext';
import './App.css';

export interface Game {
  id: string;
  name: string;
  displayName: string;
  boxArtURL: string;
}

export interface Drop {
  id: string;
  name: string;
  description: string;
  imageURL: string;
  startAt: string;
  endAt: string;
  requiredMinutes: number;
  currentMinutes: number;
  gameID: string;
  gameName: string;
  isClaimed: boolean;
  isCompleted: boolean;
}

export interface MinerStatus {
  isRunning: boolean;
  currentGame?: Game;
  currentStream?: {
    id: string;
    userLogin: string;
    userName: string;
    title: string;
    viewerCount: number;
  };
  watchDuration: number;
  totalWatched: number;
  gamesQueue: Game[];
  lastUpdate: string;
}

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  gameId?: string;
  streamId?: string;
}

export interface Settings {
  games: string[];
  watchInterval: number;
  autoClaimDrops: boolean;
  notificationsEnabled: boolean;
  theme: string;
  language: string;
  updatedAt: string;
}

const App: Component = () => {
  const [activeTab, setActiveTab] = createSignal('watch');
  const [sidebarOpen, setSidebarOpen] = createSignal(false);
  const [showAuthModal, setShowAuthModal] = createSignal(false);
  
  const [appState, setAppState] = createStore({
    games: [] as Game[],
    drops: [] as Drop[],
    minerStatus: null as MinerStatus | null,
    logs: [] as LogEntry[],
    settings: null as Settings | null,
    authenticated: false,
    loading: true,
  });

  // WebSocket connection for real-time updates
  let ws: WebSocket | null = null;

  const connectWebSocket = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`;
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
    };
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        
        switch (message.type) {
          case 'status':
            setAppState('minerStatus', message.data);
            break;
          case 'log':
            setAppState('logs', (logs) => [...logs, message.data].slice(-1000));
            break;
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected, attempting to reconnect...');
      setTimeout(connectWebSocket, 3000);
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  };

  // API helper functions
  const api = {
    async get(endpoint: string) {
      const response = await fetch(`/api/v1${endpoint}`);
      if (!response.ok) throw new Error(`API Error: ${response.statusText}`);
      return response.json();
    },
    
    async post(endpoint: string, data?: any) {
      const response = await fetch(`/api/v1${endpoint}`, {
        method: 'POST',
        headers: data ? { 'Content-Type': 'application/json' } : {},
        body: data ? JSON.stringify(data) : undefined,
      });
      if (!response.ok) throw new Error(`API Error: ${response.statusText}`);
      return response.json();
    },
    
    async put(endpoint: string, data: any) {
      const response = await fetch(`/api/v1${endpoint}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error(`API Error: ${response.statusText}`);
      return response.json();
    },
    
    async delete(endpoint: string) {
      const response = await fetch(`/api/v1${endpoint}`, {
        method: 'DELETE',
      });
      if (!response.ok) throw new Error(`API Error: ${response.statusText}`);
      return response.json();
    },
  };

  // Load initial data
  const loadData = async () => {
    try {
      setAppState('loading', true);
      
      // Check authentication status
      const authStatus = await api.get('/auth/status');
      setAppState('authenticated', authStatus.authenticated);
      
      if (authStatus.authenticated) {
        // Load games, drops, and settings
        const [gamesData, dropsData, settingsData, minerStatusData, logsData] = await Promise.all([
          api.get('/games'),
          api.get('/drops'),
          api.get('/settings'),
          api.get('/miner/status'),
          api.get('/miner/logs'),
        ]);
        
        setAppState({
          games: gamesData.games || [],
          drops: dropsData.drops || [],
          settings: settingsData,
          minerStatus: minerStatusData,
          logs: logsData.logs || [],
        });
      }
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setAppState('loading', false);
    }
  };

  // Game management functions
  const addGame = async (gameName: string) => {
    try {
      const result = await api.post('/games', { name: gameName });
      const updatedGames = await api.get('/games');
      setAppState('games', updatedGames.games || []);
      return result.game;
    } catch (error) {
      console.error('Failed to add game:', error);
      throw error;
    }
  };

  const removeGame = async (gameId: string) => {
    try {
      await api.delete(`/games/${gameId}`);
      const updatedGames = await api.get('/games');
      setAppState('games', updatedGames.games || []);
    } catch (error) {
      console.error('Failed to remove game:', error);
      throw error;
    }
  };

  const reorderGames = async (gameIds: string[]) => {
    try {
      await api.put('/games/reorder', { gameIds });
      const updatedGames = await api.get('/games');
      setAppState('games', updatedGames.games || []);
    } catch (error) {
      console.error('Failed to reorder games:', error);
      throw error;
    }
  };

  // Miner control functions
  const startMiner = async () => {
    try {
      await api.post('/miner/start');
      const status = await api.get('/miner/status');
      setAppState('minerStatus', status);
    } catch (error) {
      console.error('Failed to start miner:', error);
      throw error;
    }
  };

  const stopMiner = async () => {
    try {
      await api.post('/miner/stop');
      const status = await api.get('/miner/status');
      setAppState('minerStatus', status);
    } catch (error) {
      console.error('Failed to stop miner:', error);
      throw error;
    }
  };

  // Settings management
  const updateSettings = async (newSettings: Settings) => {
    try {
      const result = await api.put('/settings', newSettings);
      setAppState('settings', result);
    } catch (error) {
      console.error('Failed to update settings:', error);
      throw error;
    }
  };

  // Authentication
  const logout = async () => {
    try {
      await api.delete('/auth/logout');
      setAppState({
        authenticated: false,
        games: [],
        drops: [],
        minerStatus: null,
        logs: [],
        settings: null,
      });
    } catch (error) {
      console.error('Failed to logout:', error);
    }
  };

  onMount(() => {
    loadData();
    connectWebSocket();
  });

  onCleanup(() => {
    if (ws) {
      ws.close();
    }
  });

  const contextValue = {
    appState,
    setAppState,
    api,
    addGame,
    removeGame,
    reorderGames,
    startMiner,
    stopMiner,
    updateSettings,
    logout,
    setShowAuthModal,
  };

  // Render current tab content
  const renderTabContent = () => {
    switch (activeTab()) {
      case 'drops':
        return <DropsTab />;
      case 'settings':
        return <SettingsTab />;
      default:
        return <WatchTab />;
    }
  };

  return (
    <AuthProvider value={contextValue}>
      <WebSocketProvider>
        <div class="min-h-screen bg-gray-50 dark:bg-dark-bg transition-colors duration-200">
          <Header 
            activeTab={activeTab()}
            onTabChange={setActiveTab}
            onMenuClick={() => setSidebarOpen(!sidebarOpen())}
          />
          
          <div class="flex">
            <Sidebar 
              isOpen={sidebarOpen()}
              onClose={() => setSidebarOpen(false)}
              activeTab={activeTab()}
              onTabChange={setActiveTab}
            />
            
            <main class="flex-1 lg:ml-64">
              <div class="p-6">
                {renderTabContent()}
              </div>
            </main>
          </div>
          
          <AuthModal 
            isOpen={showAuthModal()}
            onClose={() => setShowAuthModal(false)}
          />
        </div>
      </WebSocketProvider>
    </AuthProvider>
  );
};

export default App;