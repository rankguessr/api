package ranking

import (
	"errors"
)

type Range struct {
	From int
	To   int
}

var ranges = []Range{
	{From: 1, To: 100},
	{From: 100, To: 1000},
	{From: 1000, To: 3000},
	{From: 3000, To: 10_000},
	{From: 10_000, To: 20_000},
	{From: 20_000, To: 50_000},
	{From: 50_000, To: 80_000},
	{From: 80_000, To: 100_000},
	{From: 100_000, To: 150_000},
	{From: 150_000, To: 250_000},
	{From: 250_000, To: 400_000},
	{From: 400_000, To: 700_000},
	{From: 700_000, To: 1_000_000},
	{From: 1_000_000, To: 1_500_000},
	{From: 1_500_000, To: 3_000_000},
}

const multiplier = 5000

func getRange(actual int) (Range, error) {
	for _, r := range ranges {
		if actual >= r.From && actual < r.To {
			return r, nil
		}
	}

	return Range{}, errors.New("unexpected rank value, out of range")
}

func Calculate(guess, actual int) (int, error) {
	r, err := getRange(actual)
	if err != nil {
		return 0, err
	}

	// the range of the guess when its counted as successful
	// e.g. 10_000 - 30_000, limit = 10_000
	limit := (r.To - r.From) / 2

	// difference between the guess and the actual rank
	diff := abs(guess - actual)

	if diff > limit {
		score := float32(diff-limit) / float32(limit) * multiplier
		return -min(int(score), multiplier), nil
	}

	return int(float32(limit-diff) / float32(limit) * multiplier), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
