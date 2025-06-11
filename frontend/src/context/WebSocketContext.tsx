import { createContext, useContext, ParentComponent } from 'solid-js';

interface WebSocketContextType {
  // Add WebSocket context methods here if needed
}

const WebSocketContext = createContext<WebSocketContextType>();

export const WebSocketProvider: ParentComponent = (props) => {
  const contextValue = {
    // WebSocket context implementation
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {props.children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};
