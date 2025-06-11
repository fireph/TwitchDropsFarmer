import { render } from 'solid-js/web';
import App from './App';
import './App.css';

const root = document.getElementById('app');

if (!root) {
  throw new Error('Root element not found');
}

render(() => <App />, root);