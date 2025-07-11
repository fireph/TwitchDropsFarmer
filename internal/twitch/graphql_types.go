package twitch

import "time"

type GameGQL struct {
	Typename    string  `json:"__typename"`
	BoxArtURL   *string `json:"boxArtURL,omitempty"`
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	Streams     *struct {
		Typename string `json:"__typename"`
		Banners  *any   `json:"banners,omitempty"`
		Edges    []struct {
			Typename   string     `json:"__typename"`
			Cursor     string     `json:"cursor"`
			Node       *StreamGQL `json:"node,omitempty"`
			TrackingID string     `json:"trackingID"`
		} `json:"edges"`
		PageInfo struct {
			Typename    string `json:"__typename"`
			HasNextPage bool   `json:"hasNextPage"`
		} `json:"pageInfo"`
	} `json:"streams,omitempty"`
}

type ChannelGQL struct {
	Typename    string `json:"__typename"`
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

type StreamGQL struct {
	Typename    string `json:"__typename"`
	Broadcaster struct {
		Typename        string `json:"__typename"`
		DisplayName     string `json:"displayName"`
		ID              string `json:"id"`
		Login           string `json:"login"`
		PrimaryColorHex string `json:"primaryColorHex"`
		ProfileImageURL string `json:"profileImageURL"`
		Roles           struct {
			Typename  string `json:"__typename"`
			IsPartner bool   `json:"isPartner"`
		} `json:"roles"`
	} `json:"broadcaster"`
	FreeformTags []struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
		Name     string `json:"name"`
	} `json:"freeformTags"`
	Game                       GameGQL `json:"game"`
	ID                         string  `json:"id"`
	PreviewImageURL            string  `json:"previewImageURL"`
	PreviewThumbnailProperties struct {
		Typename   string `json:"__typename"`
		BlurReason string `json:"blurReason"`
	} `json:"previewThumbnailProperties"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	ViewersCount int    `json:"viewersCount"`
}

type UserDropRewardGQL struct {
	Typename            string     `json:"__typename"`
	Game                GameGQL    `json:"game"`
	ID                  string     `json:"id"`
	ImageURL            string     `json:"imageURL"`
	IsConnected         bool       `json:"isConnected"`
	LastAwardedAt       *time.Time `json:"lastAwardedAt,omitempty"`
	Name                string     `json:"name"`
	RequiredAccountLink *string    `json:"requiredAccountLink,omitempty"`
	TotalCount          int        `json:"totalCount"`
}

type TimeBasedDropGQL struct {
	Typename     string `json:"__typename"`
	BenefitEdges []struct {
		Typename string `json:"__typename"`
		Benefit  struct {
			Typename          string     `json:"__typename"`
			CreatedAt         *time.Time `json:"createdAt,omitempty"`
			DistributionType  string     `json:"distributionType"`
			EntitlementLimit  *int       `json:"entitlementLimit,omitempty"`
			Game              *GameGQL   `json:"game,omitempty"`
			ID                string     `json:"id"`
			ImageAssetURL     string     `json:"imageAssetURL"`
			IsIosAvailable    *bool      `json:"isIosAvailable,omitempty"`
			Name              string     `json:"name"`
			OwnerOrganization *struct {
				Typename string `json:"__typename"`
				ID       string `json:"id"`
				Name     string `json:"name"`
			} `json:"ownerOrganization,omitempty"`
		} `json:"benefit"`
		ClaimCount       *int `json:"claimCount,omitempty"`
		EntitlementLimit int  `json:"entitlementLimit"`
	} `json:"benefitEdges"`
	Campaign               *DropCampaignGQL `json:"campaign,omitempty"`
	StartAt                time.Time        `json:"startAt"`
	EndAt                  time.Time        `json:"endAt"`
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	PreconditionDrops      *any             `json:"preconditionDrops,omitempty"`
	RequiredMinutesWatched int              `json:"requiredMinutesWatched"`
	RequiredSubs           int              `json:"requiredSubs"`
	Self                   *struct {
		Typename              string `json:"__typename"`
		CurrentMinutesWatched int    `json:"currentMinutesWatched"`
		CurrentSubs           int    `json:"currentSubs"`
		DropInstanceID        *any   `json:"dropInstanceID,omitempty"`
		HasPreconditionsMet   bool   `json:"hasPreconditionsMet"`
		IsClaimed             bool   `json:"isClaimed"`
	} `json:"self,omitempty"`
}

type DropCampaignGQL struct {
	Typename       string  `json:"__typename"`
	AccountLinkURL *string `json:"accountLinkURL,omitempty"`
	Allow          *struct {
		Typename  string        `json:"__typename"`
		Channels  *[]ChannelGQL `json:"channels,omitempty"`
		IsEnabled *bool         `json:"isEnabled,omitempty"`
	} `json:"allow,omitempty"`
	Description *string   `json:"description,omitempty"`
	DetailsURL  *string   `json:"detailsURL,omitempty"`
	EndAt       time.Time `json:"endAt"`
	Game        *GameGQL  `json:"game,omitempty"`
	ID          string    `json:"id"`
	ImageURL    *string   `json:"imageURL,omitempty"`
	Name        string    `json:"name"`
	Owner       *struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
		Name     string `json:"name"`
	} `json:"owner,omitempty"`
	Self *struct {
		Typename           string `json:"__typename"`
		IsAccountConnected bool   `json:"isAccountConnected"`
	} `json:"self,omitempty"`
	StartAt        time.Time           `json:"startAt"`
	Status         string              `json:"status"`
	TimeBasedDrops *[]TimeBasedDropGQL `json:"timeBasedDrops,omitempty"`
}

type InventoryGQL struct {
	Typename                string              `json:"__typename"`
	DropCampaignsInProgress []DropCampaignGQL   `json:"dropCampaignsInProgress"`
	GameEventDrops          []UserDropRewardGQL `json:"gameEventDrops"`
}

type DropCurrentSessionGQL struct {
	Typename               string     `json:"__typename"`
	Channel                ChannelGQL `json:"channel"`
	CurrentMinutesWatched  int        `json:"currentMinutesWatched"`
	DropID                 string     `json:"dropID"`
	Game                   GameGQL    `json:"game"`
	RequiredMinutesWatched int        `json:"requiredMinutesWatched"`
}

type PlaybackAccessTokenGQL struct {
	Typename      string `json:"__typename"`
	Authorization struct {
		Typename            string `json:"__typename"`
		ForbiddenReasonCode string `json:"forbiddenReasonCode"`
		IsForbidden         bool   `json:"isForbidden"`
	} `json:"authorization"`
	Signature string `json:"signature"`
	Value     string `json:"value"`
}

type OpInventoryResponse struct {
	CurrentUser struct {
		Typename  string       `json:"__typename"`
		ID        string       `json:"id"`
		Inventory InventoryGQL `json:"inventory"`
	} `json:"currentUser"`
}

type OpCurrentDropResponse struct {
	CurrentUser struct {
		Typename           string                 `json:"__typename"`
		DropCurrentSession *DropCurrentSessionGQL `json:"dropCurrentSession,omitempty"`
		ID                 string                 `json:"id"`
	} `json:"currentUser"`
}

type OpCampaignsResponse struct {
	CurrentUser struct {
		Typename      string            `json:"__typename"`
		DropCampaigns []DropCampaignGQL `json:"dropCampaigns"`
		ID            string            `json:"id"`
		Login         *string           `json:"login,omitempty"`
	} `json:"currentUser"`
}

type OpCampaignDetailsResponse struct {
	User struct {
		Typename     string           `json:"__typename"`
		DropCampaign *DropCampaignGQL `json:"dropCampaign,omitempty"`
		ID           string           `json:"id"`
	} `json:"user"`
}

type OpPlaybackAccessTokenResponse struct {
	StreamPlaybackAccessToken PlaybackAccessTokenGQL `json:"streamPlaybackAccessToken"`
}

type OpGameDirectoryResponse struct {
	Game GameGQL `json:"game"`
}

type OpSlugRedirectResponse struct {
	Game GameGQL `json:"game"`
}
