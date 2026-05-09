package utils

import "github.com/wieku/rplpa"

func AnonymizeReplay(data []byte) (int, []byte, error) {
	r, err := rplpa.ParseReplay(data)
	if err != nil {
		return 0, nil, err
	}

	scoreId := r.ScoreID
	if r.ScoreInfo != nil {
		scoreId = r.ScoreInfo.ScoreId
	}

	r.Username = "rankguessr"
	r.ScoreID = 0
	if r.ScoreInfo != nil {
		r.ScoreInfo.ScoreId = 0
	}

	anonymized, err := rplpa.WriteReplay(r)
	if err != nil {
		return 0, nil, err
	}

	return int(scoreId), anonymized, nil
}
