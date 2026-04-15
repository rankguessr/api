package utils

import (
	"strconv"
	"strings"
)

func ParseScoreURL(scoreURL string) (int, error) {
	id := strings.Replace(scoreURL, "https://osu.ppy.sh/scores/", "", 1)
	return strconv.Atoi(id)
}
