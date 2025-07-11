package twitch

import "time"

type GameGQL struct {
	Typename    *string `json:"__typename,omitempty"`
	BoxArtURL   *string `json:"boxArtURL,omitempty"`
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	Streams     *struct {
		Typename *string `json:"__typename,omitempty"`
		Banners  *any    `json:"banners,omitempty"`
		Edges    *[]struct {
			Typename   *string    `json:"__typename,omitempty"`
			Cursor     *string    `json:"cursor,omitempty"`
			Node       *StreamGQL `json:"node,omitempty"`
			TrackingID *string    `json:"trackingID,omitempty"`
		} `json:"edges,omitempty"`
		PageInfo *struct {
			Typename    *string `json:"__typename,omitempty"`
			HasNextPage *bool   `json:"hasNextPage,omitempty"`
		} `json:"pageInfo,omitempty"`
	} `json:"streams,omitempty"`
}

type ChannelGQL struct {
	Typename    *string `json:"__typename,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
}

type StreamGQL struct {
	Typename    *string `json:"__typename,omitempty"`
	Broadcaster *struct {
		Typename        *string `json:"__typename,omitempty"`
		DisplayName     *string `json:"displayName,omitempty"`
		ID              *string `json:"id,omitempty"`
		Login           *string `json:"login,omitempty"`
		PrimaryColorHex *string `json:"primaryColorHex,omitempty"`
		ProfileImageURL *string `json:"profileImageURL,omitempty"`
		Roles           *struct {
			Typename  *string `json:"__typename,omitempty"`
			IsPartner *bool   `json:"isPartner,omitempty"`
		} `json:"roles,omitempty"`
	} `json:"broadcaster,omitempty"`
	FreeformTags *[]struct {
		Typename *string `json:"__typename,omitempty"`
		ID       *string `json:"id,omitempty"`
		Name     *string `json:"name,omitempty"`
	} `json:"freeformTags,omitempty"`
	Game                       *GameGQL `json:"game,omitempty"`
	ID                         *string  `json:"id,omitempty"`
	PreviewImageURL            *string  `json:"previewImageURL,omitempty"`
	PreviewThumbnailProperties *struct {
		Typename   *string `json:"__typename,omitempty"`
		BlurReason *string `json:"blurReason,omitempty"`
	} `json:"previewThumbnailProperties,omitempty"`
	Title        *string `json:"title,omitempty"`
	Type         *string `json:"type,omitempty"`
	ViewersCount *int    `json:"viewersCount,omitempty"`
}

type UserDropRewardGQL struct {
	Typename            *string    `json:"__typename,omitempty"`
	Game                *GameGQL   `json:"game,omitempty"`
	ID                  *string    `json:"id,omitempty"`
	ImageURL            *string    `json:"imageURL,omitempty"`
	IsConnected         *bool      `json:"isConnected,omitempty"`
	LastAwardedAt       *time.Time `json:"lastAwardedAt,omitempty"`
	Name                *string    `json:"name,omitempty"`
	RequiredAccountLink *string    `json:"requiredAccountLink,omitempty"`
	TotalCount          *int       `json:"totalCount,omitempty"`
}

type TimeBasedDropGQL struct {
	Typename     *string `json:"__typename,omitempty"`
	BenefitEdges *[]struct {
		Typename *string `json:"__typename,omitempty"`
		Benefit  *struct {
			Typename          *string    `json:"__typename,omitempty"`
			CreatedAt         *time.Time `json:"createdAt,omitempty"`
			DistributionType  *string    `json:"distributionType,omitempty"`
			EntitlementLimit  *int       `json:"entitlementLimit,omitempty"`
			Game              *GameGQL   `json:"game,omitempty"`
			ID                *string    `json:"id,omitempty"`
			ImageAssetURL     *string    `json:"imageAssetURL,omitempty"`
			IsIosAvailable    *bool      `json:"isIosAvailable,omitempty"`
			Name              *string    `json:"name,omitempty"`
			OwnerOrganization *struct {
				Typename *string `json:"__typename,omitempty"`
				ID       *string `json:"id,omitempty"`
				Name     *string `json:"name,omitempty"`
			} `json:"ownerOrganization,omitempty"`
		} `json:"benefit,omitempty"`
		ClaimCount       *int `json:"claimCount,omitempty"`
		EntitlementLimit *int `json:"entitlementLimit,omitempty"`
	} `json:"benefitEdges,omitempty"`
	Campaign               *DropCampaignGQL `json:"campaign,omitempty"`
	StartAt                *time.Time       `json:"startAt,omitempty"`
	EndAt                  *time.Time       `json:"endAt,omitempty"`
	ID                     *string          `json:"id,omitempty"`
	Name                   *string          `json:"name,omitempty"`
	PreconditionDrops      *any             `json:"preconditionDrops,omitempty"`
	RequiredMinutesWatched *int             `json:"requiredMinutesWatched,omitempty"`
	RequiredSubs           *int             `json:"requiredSubs,omitempty"`
	Self                   *struct {
		Typename              *string `json:"__typename,omitempty"`
		CurrentMinutesWatched *int    `json:"currentMinutesWatched,omitempty"`
		CurrentSubs           *int    `json:"currentSubs,omitempty"`
		DropInstanceID        *any    `json:"dropInstanceID,omitempty"`
		HasPreconditionsMet   *bool   `json:"hasPreconditionsMet,omitempty"`
		IsClaimed             *bool   `json:"isClaimed,omitempty"`
	} `json:"self,omitempty"`
}

type DropCampaignGQL struct {
	Typename       *string `json:"__typename,omitempty"`
	AccountLinkURL *string `json:"accountLinkURL,omitempty"`
	Allow          *struct {
		Typename  *string       `json:"__typename,omitempty"`
		Channels  *[]ChannelGQL `json:"channels,omitempty"`
		IsEnabled *bool         `json:"isEnabled,omitempty"`
	} `json:"allow,omitempty"`
	Description *string    `json:"description,omitempty"`
	DetailsURL  *string    `json:"detailsURL,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
	Game        *GameGQL   `json:"game,omitempty"`
	ID          *string    `json:"id,omitempty"`
	ImageURL    *string    `json:"imageURL,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Owner       *struct {
		Typename *string `json:"__typename,omitempty"`
		ID       *string `json:"id,omitempty"`
		Name     *string `json:"name,omitempty"`
	} `json:"owner,omitempty"`
	Self *struct {
		Typename           *string `json:"__typename,omitempty"`
		IsAccountConnected *bool   `json:"isAccountConnected,omitempty"`
	} `json:"self,omitempty"`
	StartAt        *time.Time          `json:"startAt,omitempty"`
	Status         *string             `json:"status,omitempty"`
	TimeBasedDrops *[]TimeBasedDropGQL `json:"timeBasedDrops,omitempty"`
}

type InventoryGQL struct {
	Typename                *string              `json:"__typename,omitempty"`
	DropCampaignsInProgress *[]DropCampaignGQL   `json:"dropCampaignsInProgress,omitempty"`
	GameEventDrops          *[]UserDropRewardGQL `json:"gameEventDrops,omitempty"`
}

type DropCurrentSessionGQL struct {
	Typename               *string     `json:"__typename,omitempty"`
	Channel                *ChannelGQL `json:"channel,omitempty"`
	CurrentMinutesWatched  *int        `json:"currentMinutesWatched,omitempty"`
	DropID                 *string     `json:"dropID,omitempty"`
	Game                   *GameGQL    `json:"game,omitempty"`
	RequiredMinutesWatched *int        `json:"requiredMinutesWatched,omitempty"`
}

type PlaybackAccessTokenGQL struct {
	Typename      *string `json:"__typename,omitempty"`
	Authorization *struct {
		Typename            *string `json:"__typename,omitempty"`
		ForbiddenReasonCode *string `json:"forbiddenReasonCode,omitempty"`
		IsForbidden         *bool   `json:"isForbidden,omitempty"`
	} `json:"authorization,omitempty"`
	Signature *string `json:"signature,omitempty"`
	Value     *string `json:"value,omitempty"`
}

type OpInventoryResponse struct {
	CurrentUser *struct {
		Typename  *string       `json:"__typename,omitempty"`
		ID        *string       `json:"id,omitempty"`
		Inventory *InventoryGQL `json:"inventory,omitempty"`
	} `json:"currentUser,omitempty"`
}

type OpCurrentDropResponse struct {
	CurrentUser *struct {
		Typename           *string                `json:"__typename,omitempty"`
		DropCurrentSession *DropCurrentSessionGQL `json:"dropCurrentSession,omitempty"`
		ID                 *string                `json:"id,omitempty"`
	} `json:"currentUser,omitempty"`
}

type OpCampaignsResponse struct {
	CurrentUser *struct {
		Typename      *string            `json:"__typename,omitempty"`
		DropCampaigns *[]DropCampaignGQL `json:"dropCampaigns,omitempty"`
		ID            *string            `json:"id,omitempty"`
		Login         *string            `json:"login,omitempty"`
	} `json:"currentUser,omitempty"`
}

type OpCampaignDetailsResponse struct {
	User *struct {
		Typename     *string          `json:"__typename,omitempty"`
		DropCampaign *DropCampaignGQL `json:"dropCampaign,omitempty"`
		ID           *string          `json:"id,omitempty"`
	} `json:"user,omitempty"`
}

type OpPlaybackAccessTokenResponse struct {
	StreamPlaybackAccessToken *PlaybackAccessTokenGQL `json:"streamPlaybackAccessToken,omitempty"`
}

type OpGameDirectoryResponse struct {
	Game *GameGQL `json:"game,omitempty"`
}

type OpSlugRedirectResponse struct {
	Game *GameGQL `json:"game,omitempty"`
}
