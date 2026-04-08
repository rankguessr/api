package osuapi

import "time"

type Ruleset string

var (
	ModeStandard Ruleset = "osu"
	ModeTaiko    Ruleset = "taiko"
	ModeCatch    Ruleset = "fruits"
	ModeMania    Ruleset = "mania"
)

type ExchangeTokenResponse struct {
	// always "Bearer"
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (r ExchangeTokenResponse) ExpiresAt() time.Time {
	return time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
}

type UserStatistics struct {
	PP         float32 `json:"pp"`
	GlobalRank int     `json:"global_rank"`
}

type User struct {
	ID          int            `json:"id"`
	Username    string         `json:"username"`
	AvatarURL   string         `json:"avatar_url"`
	CountryCode string         `json:"country_code"`
	IsActive    bool           `json:"is_active"`
	IsBot       bool           `json:"is_bot"`
	IsDeleted   bool           `json:"is_deleted"`
	IsOnline    bool           `json:"is_online"`
	IsSupporter bool           `json:"is_supporter"`
	LastVisit   time.Time      `json:"last_visit"`
	Statistics  UserStatistics `json:"statistics"`
}

type UserExtended struct {
	User
	// deprecated, use cover.url instead
	CoverURL string `json:"cover_url"`

	Discord      string    `json:"discord"`
	HasSupported bool      `json:"has_supported"`
	Interests    string    `json:"interests"`
	JoinDate     time.Time `json:"join_date"`
	Location     *string   `json:"location"`
	MaxBlocks    int       `json:"max_blocks"`
	MaxFriends   int       `json:"max_friends"`
	Occupation   *string   `json:"occupation"`
	Playmode     string    `json:"playmode"`
	Playstyle    []string  `json:"playstyle"`
	PostCount    int       `json:"post_count"`
	ProfileHue   *int      `json:"profile_hue"`
	// TODO: dont really care about these, maybe add them later if needed
	ProfileOrder []string `json:"profile_order"`
	Title        *string  `json:"title"`
	TitleURL     *string  `json:"title_url"`
	Twitter      *string  `json:"twitter"`
	Website      *string  `json:"website"`
}

type ClientToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type MultiRoom struct {
	RecentParticipants []User `json:"recent_participants"`
}

type Beatmap struct {
	DifficultyRating float32 `json:"difficulty_rating"`
	TotalLength      int     `json:"total_length"`
	MaxCombo         int     `json:"max_combo"`
}

type BeatmapExtended struct {
	Beatmap
	AR           float32 `json:"ar"`
	BPM          float32 `json:"bpm"`
	BeatmapSetId int     `json:"beatmapset_id"`
}

type Covers struct {
	SlimCoverURL string `json:"slimcover"`
}

type BeatmapSet struct {
	Artist         string `json:"artist"`
	Title          string `json:"title"`
	PreviewURL     string `json:"preview_url"`
	FavouriteCount int    `json:"favourite_count"`
	Covers         Covers `json:"covers"`
}

type Score struct {
	ID              int             `json:"id"`
	PP              float32         `json:"pp"`
	Mods            []string        `json:"mods"`
	Accuracy        float32         `json:"accuracy"`
	BeatmapID       int             `json:"beatmap_id"`
	HasReplay       bool            `json:"has_replay"`
	ReplayLegacy    bool            `json:"replay"`
	BeatmapSet      BeatmapSet      `json:"beatmapset"`
	BeatmapExtended BeatmapExtended `json:"beatmap"`
}

func (s Score) Replay() bool {
	return s.HasReplay || s.ReplayLegacy
}
