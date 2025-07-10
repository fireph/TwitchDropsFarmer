package util

import (
	"context"

	"twitchdropsfarmer/internal/drops"
	"twitchdropsfarmer/internal/twitch"
)

// GenerateActiveDrops creates a list of ActiveDrop objects with real-time progress data
// This function uses DropCurrentSessionContext to get accurate progress information
func GenerateActiveDrops(ctx context.Context, twitchClient *twitch.Client, campaign *twitch.Campaign, currentStream *twitch.Stream) ([]drops.ActiveDrop, error) {
	var activeDrops []drops.ActiveDrop

	// Get real-time progress data using DropCurrentSessionContext
	currentDropInfo, err := twitchClient.GetCurrentDropProgress(ctx, currentStream.UserID)
	if err != nil {
		// If we can't get current progress, we'll still return drops with zero progress
		// This ensures the UI still shows the drops even if real-time data is unavailable
		currentDropInfo = nil
	}

	// Filter out subscription/gift sub drops (RequiredMinutesWatched = 0)
	// These drops require subscriptions or gift subs and cannot be farmed through watching
	var farmableDrops []twitch.TimeBased
	for _, drop := range campaign.TimeBasedDrops {
		if drop.RequiredMinutesWatched > 0 {
			farmableDrops = append(farmableDrops, drop)
		}
	}

	// Create a sorted copy of the farmable drops
	sortedDrops := make([]twitch.TimeBased, len(farmableDrops))
	copy(sortedDrops, farmableDrops)

	// Sort drops by required minutes (30, 90, 180, etc.)
	// This ensures we process them in the correct order for status inference
	for i := 0; i < len(sortedDrops)-1; i++ {
		for j := i + 1; j < len(sortedDrops); j++ {
			if sortedDrops[i].RequiredMinutesWatched > sortedDrops[j].RequiredMinutesWatched {
				sortedDrops[i], sortedDrops[j] = sortedDrops[j], sortedDrops[i]
			}
		}
	}

	// Process each drop to determine its current status and progress
	for i, drop := range sortedDrops {
		currentMinutes := 0
		isClaimed := false

		if currentDropInfo != nil && currentDropInfo.DropID == drop.ID {
			// This is the currently active drop - use real-time progress from DropCurrentSessionContext
			currentMinutes = currentDropInfo.CurrentMinutesWatched
			isClaimed = currentMinutes >= drop.RequiredMinutesWatched
		} else {
			// This is not the currently active drop - infer status based on sequence
			if currentDropInfo != nil {
				// Find which drop is currently active to determine this drop's status
				for j, checkDrop := range sortedDrops {
					if checkDrop.ID == currentDropInfo.DropID {
						if j > i {
							// The active drop comes after this one in sequence,
							// so this drop must be completed
							currentMinutes = drop.RequiredMinutesWatched
							isClaimed = true
						} else {
							// The active drop is this one or comes before this one,
							// so this drop hasn't been started yet
							currentMinutes = 0
							isClaimed = false
						}
						break
					}
				}
			}
			// If no currentDropInfo available, defaults remain: currentMinutes = 0, isClaimed = false
		}

		// Calculate progress percentage
		progress := 0.0
		if drop.RequiredMinutesWatched > 0 {
			progress = float64(currentMinutes) / float64(drop.RequiredMinutesWatched)
		}

		// Create the ActiveDrop object
		activeDrop := drops.ActiveDrop{
			ID:              drop.ID,
			Name:            drop.Name,
			GameName:        campaign.Game.Name,
			RequiredMinutes: drop.RequiredMinutesWatched,
			CurrentMinutes:  currentMinutes,
			Progress:        progress,
			IsClaimed:       isClaimed,
		}
		activeDrops = append(activeDrops, activeDrop)
	}

	return activeDrops, nil
}
