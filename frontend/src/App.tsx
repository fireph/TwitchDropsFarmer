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

// Import Socket.IO client
import { io, Socket } from 'socket.io-client';

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

  // Socket.IO connection
  let socket: Socket | null = null;
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000;

  const connectSocket = () => {
    if (socket && socket.connected) {
      return; // Already connected
    }

    console.log('Attempting to connect to Socket.IO server...');
    
    // Create Socket.IO connection
    socket = io('/', {
      transports: ['websocket', 'polling'], // Allow fallback to polling
      timeout: 10000,
      autoConnect: true,
      reconnection: true,
      reconnectionAttempts: maxReconnectAttempts,
      reconnectionDelay: reconnectDelay,
    });

    socket.on('connect', () => {
      console.log('Socket.IO connected successfully');
      reconnectAttempts = 0;
      
      // Send a ping to test the connection
      socket?.emit('ping');
    });

    socket.on('pong', () => {
      console.log('Socket.IO ping/pong successful');
    });

    socket.on('message', (data) => {
      try {
        console.log('Received Socket.IO message:', data);
        
        switch (data.type) {
          case 'status':
            setAppState('minerStatus', data.data);
            break;
          case 'log':
            setAppState('logs', (logs) => [...logs, data.data].slice(-1000));
            break;
          default:
            console.log('Unknown message type:', data.type);
        }
      } catch (error) {
        console.error('Failed to parse Socket.IO message:', error);
      }
    });

    socket.on('disconnect', (reason) => {
      console.log('Socket.IO disconnected:', reason);
      
      if (reason === 'io server disconnect') {
        // The disconnection was initiated by the server, reconnect manually
        socket?.connect();
      }
    });

    socket.on('connect_error', (error) => {
      console.error('Socket.IO connection error:', error);
      reconnectAttempts++;
      
      if (reconnectAttempts >= maxReconnectAttempts) {
        console.error('Max reconnection attempts reached');
      }
    });

    socket.on('error', (error) => {
      console.error('Socket.IO error:', error);
    });
  };

  const disconnectSocket = () => {
    if (socket) {
      console.log('Disconnecting Socket.IO...');
      socket.disconnect();
      socket = null;
    }
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
      console.log('Auth status response:', authStatus);
      setAppState('authenticated', authStatus.authenticated);
      
      // Also get debug info if not authenticated
      if (!authStatus.authenticated) {
        try {
          const debugInfo = await api.get('/debug/auth');
          console.log('Auth debug info:', debugInfo);
        } catch (debugError) {
          console.log('Could not get debug info:', debugError);
        }
      }
      
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
    connectSocket();
  });

  onCleanup(() => {
    disconnectSocket();
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