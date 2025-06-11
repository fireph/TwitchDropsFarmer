import { Component, createSignal, onMount } from 'solid-js';
import { useAuth } from '../context/AuthContext';

const SettingsTab: Component = () => {
  const { appState, updateSettings } = useAuth();
  const [settings, setSettings] = createSignal(appState.settings || {
    games: [],
    watchInterval: 20,
    autoClaimDrops: true,
    notificationsEnabled: true,
    theme: 'dark',
    language: 'en',
    updatedAt: new Date().toISOString(),
  });
  
  const [isSaving, setIsSaving] = createSignal(false);

  onMount(() => {
    if (appState.settings) {
      setSettings(appState.settings);
    }
  });

  const handleSave = async () => {
    setIsSaving(true);
    try {
      await updateSettings(settings());
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const updateSetting = <K extends keyof typeof settings>(key: K, value: any) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  return (
    <div class="space-y-6">
      <div class="bg-white dark:bg-dark-bg-secondary rounded-xl shadow-xs border border-gray-200 dark:border-dark-border p-6">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-dark-text mb-6">
          Settings
        </h2>
        
        <div class="space-y-6">
          {/* Mining Settings */}
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-4">
              Mining Settings
            </h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-dark-text-secondary mb-2">
                  Watch Interval (seconds)
                </label>
                <input
                  type="number"
                  min="10"
                  max="120"
                  value={settings().watchInterval}
                  onInput={(e) => updateSetting('watchInterval', parseInt(e.currentTarget.value))}
                  class="w-32 px-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-bg-tertiary text-gray-900 dark:text-dark-text focus:ring-2 focus:ring-twitch-purple focus:border-transparent"
                />
                <p class="text-xs text-gray-500 dark:text-dark-text-secondary mt-1">
                  How often to send watch requests (default: 20 seconds)
                </p>
              </div>
              
              <div class="flex items-center space-x-3">
                <input
                  type="checkbox"
                  id="autoClaim"
                  checked={settings().autoClaimDrops}
                  onChange={(e) => updateSetting('autoClaimDrops', e.currentTarget.checked)}
                  class="w-4 h-4 text-twitch-purple border-gray-300 dark:border-dark-border rounded-sm focus:ring-twitch-purple dark:bg-dark-bg-tertiary"
                />
                <label for="autoClaim" class="text-sm font-medium text-gray-700 dark:text-dark-text-secondary">
                  Automatically claim completed drops
                </label>
              </div>
              
              <div class="flex items-center space-x-3">
                <input
                  type="checkbox"
                  id="notifications"
                  checked={settings().notificationsEnabled}
                  onChange={(e) => updateSetting('notificationsEnabled', e.currentTarget.checked)}
                  class="w-4 h-4 text-twitch-purple border-gray-300 dark:border-dark-border rounded-sm focus:ring-twitch-purple dark:bg-dark-bg-tertiary"
                />
                <label for="notifications" class="text-sm font-medium text-gray-700 dark:text-dark-text-secondary">
                  Enable notifications
                </label>
              </div>
            </div>
          </div>

          {/* Appearance Settings */}
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-dark-text mb-4">
              Appearance
            </h3>
            
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-dark-text-secondary mb-2">
                  Theme
                </label>
                <select
                  value={settings().theme}
                  onChange={(e) => updateSetting('theme', e.currentTarget.value)}
                  class="w-40 px-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-bg-tertiary text-gray-900 dark:text-dark-text focus:ring-2 focus:ring-twitch-purple focus:border-transparent"
                >
                  <option value="light">Light</option>
                  <option value="dark">Dark</option>
                  <option value="auto">Auto</option>
                </select>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-dark-text-secondary mb-2">
                  Language
                </label>
                <select
                  value={settings().language}
                  onChange={(e) => updateSetting('language', e.currentTarget.value)}
                  class="w-40 px-3 py-2 border border-gray-300 dark:border-dark-border rounded-lg bg-white dark:bg-dark-bg-tertiary text-gray-900 dark:text-dark-text focus:ring-2 focus:ring-twitch-purple focus:border-transparent"
                >
                  <option value="en">English</option>
                  <option value="es">Español</option>
                  <option value="fr">Français</option>
                  <option value="de">Deutsch</option>
                  <option value="pt">Português</option>
                </select>
              </div>
            </div>
          </div>

          {/* Save Button */}
          <div class="pt-4 border-t border-gray-200 dark:border-dark-border">
            <button
              onClick={handleSave}
              disabled={isSaving()}
              class="px-6 py-2 bg-twitch-gradient hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-opacity"
            >
              {isSaving() ? 'Saving...' : 'Save Settings'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SettingsTab;
