// Dashboard component
class Dashboard {
    constructor(api) {
        this.api = api;
        this.currentStatus = null;
        this.refreshInterval = null;
    }

    async init() {
        try {
            // Load initial data
            await this.loadMinerStatus();
            
            // Set up periodic refresh for real-time progress updates
            this.refreshInterval = setInterval(() => {
                this.loadMinerStatus();
            }, 5000); // Refresh every 5 seconds for real-time progress
        } catch (error) {
            console.error('Failed to initialize dashboard:', error);
        }
    }

    async loadMinerStatus() {
        try {
            const [status, progress] = await Promise.all([
                this.api.getMinerStatus(),
                this.api.getDropProgress()
            ]);
            
            // Merge progress data into status
            status.active_drops = progress.active_drops || [];
            status.total_progress = progress.total_progress || {};
            
            this.updateStatus(status);
        } catch (error) {
            console.error('Failed to load miner status:', error);
        }
    }

    updateStatus(status) {
        this.currentStatus = status;
        
        // Update status cards
        this.updateStatusCards(status);
        
        // Update active drops
        this.updateActiveDrops(status.active_drops || []);
        
        // Update current stream
        this.updateCurrentStream(status.current_stream);
        
        // Update button states
        this.updateButtonStates(status.is_running);
    }

    updateStatusCards(status) {
        // Miner status
        const minerStatus = document.getElementById('miner-status');
        if (minerStatus) {
            minerStatus.textContent = status.is_running ? 'Running' : 'Stopped';
            minerStatus.className = `text-lg font-semibold ${status.is_running ? 'text-green-600' : 'text-red-600'}`;
        }

        // Active drops count - show watching drops vs total
        const activeDropsCount = document.getElementById('active-drops-count');
        if (activeDropsCount) {
            const totalDrops = status.active_drops ? status.active_drops.length : 0;
            const watchingDrops = status.current_campaign && status.is_running ? 
                status.active_drops?.filter(drop => drop.game_name === status.current_campaign.game.name).length || 0 : 0;
            activeDropsCount.textContent = watchingDrops > 0 ? `${watchingDrops}/${totalDrops}` : totalDrops;
        }

        // Claimed drops count
        const claimedDropsCount = document.getElementById('claimed-drops-count');
        if (claimedDropsCount) {
            claimedDropsCount.textContent = status.claimed_drops || 0;
        }

        // Current stream and campaign
        const currentStreamElement = document.getElementById('current-stream');
        if (currentStreamElement) {
            if (status.current_stream && status.current_campaign) {
                currentStreamElement.innerHTML = `
                    <div class="flex flex-col">
                        <span class="font-semibold">${status.current_stream.user_name}</span>
                        <span class="text-xs text-gray-500 dark:text-gray-400">${status.current_campaign.game.name}</span>
                    </div>
                `;
            } else {
                currentStreamElement.textContent = 'None';
            }
        }
    }

    updateActiveDrops(drops) {
        const container = document.getElementById('active-drops-list');
        if (!container) return;

        if (!drops || drops.length === 0) {
            container.innerHTML = `
                <div class="text-center text-gray-500 dark:text-gray-400">
                    <p>No active drops found. Start the miner to begin farming!</p>
                </div>
            `;
            return;
        }

        // Sort drops by current campaign first, then by progress
        const currentCampaign = this.currentStatus?.current_campaign;
        const sortedDrops = drops.sort((a, b) => {
            // Show current campaign's drops first
            if (currentCampaign) {
                const aIsCurrentCampaign = a.game_name === currentCampaign.game.name;
                const bIsCurrentCampaign = b.game_name === currentCampaign.game.name;
                if (aIsCurrentCampaign && !bIsCurrentCampaign) return -1;
                if (!aIsCurrentCampaign && bIsCurrentCampaign) return 1;
            }
            // Then sort by progress (highest first)
            return b.progress - a.progress;
        });

        container.innerHTML = sortedDrops.map(drop => this.createDropCard(drop, currentCampaign)).join('');
    }

    createDropCard(drop, currentCampaign) {
        const progress = Math.min(100, (drop.current_minutes / drop.required_minutes) * 100);
        const isCompleted = drop.current_minutes >= drop.required_minutes;
        const estimatedTime = this.getEstimatedTime(drop);
        const isCurrentCampaign = currentCampaign && drop.game_name === currentCampaign.game.name;
        const isWatching = isCurrentCampaign && this.currentStatus?.is_running;

        return `
            <div class="drop-card bg-gray-50 dark:bg-gray-700 rounded-lg p-4 mb-4 ${isCurrentCampaign ? 'ring-2 ring-green-500' : ''}">
                <div class="flex items-start space-x-4">
                    <div class="game-icon relative">
                        <div class="w-12 h-12 bg-gray-300 dark:bg-gray-600 rounded-lg flex items-center justify-center">
                            <svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                            </svg>
                        </div>
                        ${isWatching ? '<div class="absolute -top-1 -right-1 w-4 h-4 bg-green-500 rounded-full border-2 border-white dark:border-gray-700 pulse-dot"></div>' : ''}
                    </div>
                    
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center justify-between">
                            <h4 class="text-sm font-medium text-gray-900 dark:text-white truncate">
                                ${drop.name}
                                ${isWatching ? '<span class="ml-2 text-xs bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100 px-2 py-1 rounded-full">WATCHING</span>' : ''}
                            </h4>
                            <span class="status-indicator ${isCompleted ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100' : isWatching ? 'bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100'}">
                                ${isCompleted ? 'Ready to claim' : isWatching ? 'In progress' : 'Waiting'}
                            </span>
                        </div>
                        
                        <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            ${drop.game_name}
                        </p>
                        
                        <div class="mt-3">
                            <div class="flex items-center justify-between text-sm">
                                <span class="text-gray-600 dark:text-gray-400">
                                    ${drop.current_minutes}/${drop.required_minutes} minutes
                                </span>
                                <span class="text-gray-600 dark:text-gray-400">
                                    ${Math.round(progress)}%
                                </span>
                            </div>
                            
                            <div class="mt-1 w-full bg-gray-200 dark:bg-gray-600 rounded-full h-2">
                                <div class="progress-bar h-2 rounded-full ${isWatching ? 'bg-green-500' : isCompleted ? 'bg-green-500' : 'bg-gray-400'}" style="width: ${progress}%"></div>
                            </div>
                        </div>
                        
                        ${!isCompleted ? `
                            <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                Estimated time: ${estimatedTime}
                            </p>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    updateCurrentStream(stream) {
        const container = document.getElementById('current-stream-info');
        if (!container) return;

        if (!stream) {
            container.innerHTML = `
                <div class="text-center text-gray-500 dark:text-gray-400">
                    <p>No stream currently being watched</p>
                </div>
            `;
            return;
        }

        const thumbnailUrl = stream.thumbnail_url ? 
            stream.thumbnail_url.replace('{width}', '320').replace('{height}', '180') : 
            null;

        container.innerHTML = `
            <div class="flex items-start space-x-4">
                ${thumbnailUrl ? `
                    <div class="stream-thumbnail w-32">
                        <img src="${thumbnailUrl}" alt="${stream.title}" class="w-full h-auto rounded-lg">
                    </div>
                ` : ''}
                
                <div class="flex-1 min-w-0">
                    <div class="flex items-center space-x-2">
                        <h4 class="text-lg font-medium text-gray-900 dark:text-white truncate">
                            ${stream.user_name}
                        </h4>
                        <div class="pulse-dot"></div>
                        <span class="text-sm text-red-600 font-medium">LIVE</span>
                    </div>
                    
                    <p class="text-sm text-gray-600 dark:text-gray-400 mt-1 line-clamp-2">
                        ${stream.title}
                    </p>
                    
                    <div class="flex items-center space-x-4 mt-2">
                        <span class="text-sm text-gray-600 dark:text-gray-400">
                            ${stream.game_name}
                        </span>
                        <span class="text-sm text-gray-600 dark:text-gray-400">
                            ${stream.viewer_count ? stream.viewer_count.toLocaleString() : 0} viewers
                        </span>
                        <span class="text-sm text-gray-600 dark:text-gray-400">
                            ${stream.language?.toUpperCase() || 'N/A'}
                        </span>
                    </div>
                </div>
            </div>
        `;
    }

    updateCurrentCampaign(campaign, isRunning) {
        const section = document.getElementById('current-campaign-section');
        const container = document.getElementById('current-campaign-info');
        
        if (!section || !container) return;

        if (!campaign || !isRunning) {
            section.classList.add('hidden');
            return;
        }

        section.classList.remove('hidden');
        
        // Find drops for this campaign
        const campaignDrops = this.currentStatus?.active_drops?.filter(drop => 
            drop.game_name === campaign.game.name
        ) || [];
        
        const totalProgress = campaignDrops.length > 0 ? 
            campaignDrops.reduce((sum, drop) => sum + drop.progress, 0) / campaignDrops.length * 100 : 0;
        
        const currentSessionMinutes = this.currentStatus?.current_progress || 0;
        
        container.innerHTML = `
            <div class="space-y-4">
                <div class="flex items-start space-x-4">
                    <div class="w-16 h-16 bg-gray-300 dark:bg-gray-600 rounded-lg flex items-center justify-center">
                        <svg class="w-8 h-8 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h8m-9-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                    </div>
                    
                    <div class="flex-1">
                        <h4 class="text-xl font-semibold text-gray-900 dark:text-white">
                            ${campaign.name}
                        </h4>
                        <p class="text-lg text-gray-600 dark:text-gray-400 mt-1">
                            ${campaign.game.name}
                        </p>
                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">
                            ${campaign.description || 'No description available'}
                        </p>
                    </div>
                </div>
                
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                        <div class="text-sm text-gray-600 dark:text-gray-400">Session Progress</div>
                        <div class="text-2xl font-bold text-gray-900 dark:text-white">
                            ${currentSessionMinutes} min
                        </div>
                        <div class="text-xs text-gray-500 dark:text-gray-400">This session</div>
                    </div>
                    
                    <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                        <div class="text-sm text-gray-600 dark:text-gray-400">Campaign Drops</div>
                        <div class="text-2xl font-bold text-gray-900 dark:text-white">
                            ${campaignDrops.length}
                        </div>
                        <div class="text-xs text-gray-500 dark:text-gray-400">Available drops</div>
                    </div>
                    
                    <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                        <div class="text-sm text-gray-600 dark:text-gray-400">Overall Progress</div>
                        <div class="text-2xl font-bold text-gray-900 dark:text-white">
                            ${Math.round(totalProgress)}%
                        </div>
                        <div class="text-xs text-gray-500 dark:text-gray-400">Average completion</div>
                    </div>
                </div>
                
                ${campaignDrops.length > 0 ? `
                    <div class="space-y-2">
                        <h5 class="text-sm font-medium text-gray-700 dark:text-gray-300">Campaign Drops:</h5>
                        ${campaignDrops.map(drop => `
                            <div class="flex items-center justify-between text-sm">
                                <span class="text-gray-600 dark:text-gray-400">${drop.name}</span>
                                <div class="flex items-center space-x-2">
                                    <div class="w-20 bg-gray-200 dark:bg-gray-600 rounded-full h-2">
                                        <div class="bg-green-500 h-2 rounded-full" style="width: ${Math.round(drop.progress * 100)}%"></div>
                                    </div>
                                    <span class="text-gray-500 dark:text-gray-400 w-12 text-right">${Math.round(drop.progress * 100)}%</span>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                ` : ''}
            </div>
        `;
    }

    updateButtonStates(isRunning) {
        const startBtn = document.getElementById('start-miner');
        const stopBtn = document.getElementById('stop-miner');

        if (startBtn) {
            startBtn.disabled = isRunning;
            startBtn.classList.toggle('opacity-50', isRunning);
            startBtn.classList.toggle('cursor-not-allowed', isRunning);
        }

        if (stopBtn) {
            stopBtn.disabled = !isRunning;
            stopBtn.classList.toggle('opacity-50', !isRunning);
            stopBtn.classList.toggle('cursor-not-allowed', !isRunning);
        }
    }

    getEstimatedTime(drop) {
        const remainingMinutes = drop.required_minutes - drop.current_minutes;
        
        if (remainingMinutes <= 0) {
            return 'Completed';
        }
        
        const hours = Math.floor(remainingMinutes / 60);
        const minutes = remainingMinutes % 60;
        
        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        } else {
            return `${minutes}m`;
        }
    }

    destroy() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }
    }
}

// Add CSS for pulse animation
if (!document.getElementById('dashboard-styles')) {
    const style = document.createElement('style');
    style.id = 'dashboard-styles';
    style.textContent = `
        .pulse-dot {
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
            0%, 100% {
                opacity: 1;
            }
            50% {
                opacity: 0.5;
            }
        }
        
        .drop-card {
            transition: all 0.3s ease;
        }
        
        .drop-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
        }
        
        .status-indicator {
            padding: 4px 8px;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 500;
        }
    `;
    document.head.appendChild(style);
}