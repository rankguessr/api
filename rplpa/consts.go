package rplpa

// All osu playmodes
const (
	OSU = iota
	TAIKO
	CTB
	MANIA
)

// ClickState
const (
	LEFTCLICK = 1 << iota
	RIGHTCLICK
	KEY1
	KEY2
	SMOKE
)

type LazerHitResult string

const (
	LazerNone                LazerHitResult = "none"
	LazerMiss                LazerHitResult = "miss"
	LazerMeh                 LazerHitResult = "meh"
	LazerOk                  LazerHitResult = "ok"
	LazerGood                LazerHitResult = "good"
	LazerGreat               LazerHitResult = "great"
	LazerPerfect             LazerHitResult = "perfect"
	LazerSmallTickMiss       LazerHitResult = "small_tick_miss"
	LazerSmallTickHit        LazerHitResult = "small_tick_hit"
	LazerLargeTickMiss       LazerHitResult = "large_tick_miss"
	LazerLargeTickHit        LazerHitResult = "large_tick_hit"
	LazerSmallBonus          LazerHitResult = "small_bonus"
	LazerLargeBonus          LazerHitResult = "large_bonus"
	LazerIgnoreMiss          LazerHitResult = "ignore_miss"
	LazerIgnoreHit           LazerHitResult = "ignore_hit"
	LazerComboBreak          LazerHitResult = "combo_break"
	LazerSliderTailHit       LazerHitResult = "slider_tail_hit"
	LazerLegacyComboIncrease LazerHitResult = "legacy_combo_increase"
)
