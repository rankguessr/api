package utils

import "github.com/wieku/rplpa"

func AnonymizeReplay(data []byte) ([]byte, error) {
	r, err := rplpa.ParseReplay(data)
	if err != nil {
		return nil, err
	}

	r.Username = "rankguessr"
	r.ScoreID = 0
	if r.ScoreInfo != nil {
		r.ScoreInfo.ScoreId = 0
	}

	anonymized, err := rplpa.WriteReplay(r)
	if err != nil {
		return nil, err
	}

	return anonymized, nil
}
