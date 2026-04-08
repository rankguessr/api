package rplpa

import (
	"time"
)

// Replay is the Parsed replay.
type Replay struct {
	PlayMode     int8
	OsuVersion   int32
	BeatmapMD5   string
	Username     string
	ReplayMD5    string
	Count300     uint16
	Count100     uint16
	Count50      uint16
	CountGeki    uint16
	CountKatu    uint16
	CountMiss    uint16
	Score        int32
	MaxCombo     uint16
	Fullcombo    bool
	Mods         uint32
	LifebarGraph []LifeBarGraph
	Timestamp    time.Time
	ReplayData   []byte
	ScoreID      int64
	ScoreInfo    *ScoreInfo
}

type ScoreInfo struct {
	ScoreId           int64                    `json:"online_id"`
	Mods              []*ModInfo               `json:"mods"`
	Statistics        map[LazerHitResult]int64 `json:"statistics"`
	MaximumStatistics map[LazerHitResult]int64 `json:"maximum_statistics"`
}

type ModInfo struct {
	Acronym  string                 `json:"acronym"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// ReplayData is the Parsed Compressed Replay data.
type ReplayData struct {
	Time       float64 // Lazer is converting timestamps to int, but preparing just in case
	MouseX     float64
	MouseY     float64
	KeyPressed *KeyPressed
}

// KeyPressed is the Parsed Compressed KeyPressed.
type KeyPressed struct {
	LeftClick  bool
	RightClick bool
	Key1       bool
	Key2       bool
	Smoke      bool
}

// LifeBarGraph is the Bar under the Score stuff.
type LifeBarGraph struct {
	Time int32
	HP   float32
}
