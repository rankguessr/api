package ranking

import "testing"

func TestBasic(t *testing.T) {
	testData := []struct {
		guess         int
		actual        int
		expectedScore int
	}{
		{guess: 10000, actual: 10000, expectedScore: 5000},
		{guess: 15000, actual: 20000, expectedScore: 2500},
		{guess: 50000, actual: 100000, expectedScore: -5000},
		{guess: 50000, actual: 30000, expectedScore: 1000},
	}

	for _, td := range testData {
		score, err := Calculate(td.guess, td.actual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if score != td.expectedScore {
			t.Fatalf("expected score to be %d, got %d", td.expectedScore, score)
		}
	}
}
