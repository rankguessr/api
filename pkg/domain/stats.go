package domain

type Stats struct {
	Count24h    int             `json:"count_24h"`
	CountGlobal int             `json:"count_global"`
	Best        []GuessExtended `json:"best"`
	TopUsers    []UserExtended  `json:"top_users"`
}
