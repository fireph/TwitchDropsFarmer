import { createContext, useContext } from 'solid-js';

interface AuthContextType {
  appState: any;
  setAppState: any;
  api: any;
  addGame: (gameName: string) => Promise<any>;
  removeGame: (gameId: string) => Promise<void>;
  reorderGames: (gameIds: string[]) => Promise<void>;
  startMiner: () => Promise<void>;
  stopMiner: () => Promise<void>;
  updateSettings: (settings: any) => Promise<void>;
  logout: () => Promise<void>;
  setShowAuthModal: (show: boolean) => void;
}

const AuthContext = createContext<AuthContextType>();

export const AuthProvider = AuthContext.Provider;

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
