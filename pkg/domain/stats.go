package domain

type Stats struct {
	Best        []GuessExtended `json:"best"`
	TopUsers    []User          `json:"top_users"`
	Count24h    int             `json:"count_24h"`
	CountGlobal int             `json:"count_global"`
}
