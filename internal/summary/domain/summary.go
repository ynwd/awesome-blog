package domain

type Summary struct {
	Username string `json:"username"`
}

type SummaryData struct {
	Likes    map[string]int64 `json:"likes"`
	Comments map[string]int64 `json:"comments"`
	Posts    map[string]int64 `json:"posts"`
}
