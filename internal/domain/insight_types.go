package domain

import "time"

type MatchInsights struct {
	MatchID       ID                `json:"match_id"`
	GeneratedAt   time.Time         `json:"generated_at"`
	Title         string            `json:"title"`
	Subtitle      string            `json:"subtitle"`
	Status        MatchStatus       `json:"status"`
	Winner        string            `json:"winner"`
	LeadMessage   string            `json:"lead_message"`
	HeadlineStats []InsightStat     `json:"headline_stats"`
	Story         []InsightChapter  `json:"story"`
	PlayerCards   []PlayerInsight   `json:"player_cards"`
	ResourceCards []ResourceInsight `json:"resource_cards"`
	Questions     []string          `json:"questions"`
	Actions       []Recommendation  `json:"actions"`
}

type InsightStat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint"`
}

type InsightChapter struct {
	Order      int      `json:"order"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Evidence   []string `json:"evidence"`
	TurnNumber int      `json:"turn_number"`
	Tone       string   `json:"tone"`
}

type PlayerInsight struct {
	SeatName   string   `json:"seat_name"`
	Rank       int      `json:"rank"`
	Score      int      `json:"score"`
	Style      string   `json:"style"`
	OneLine    string   `json:"one_line"`
	Highlights []string `json:"highlights"`
	Watchouts  []string `json:"watchouts"`
}

type ResourceInsight struct {
	SeatName  string `json:"seat_name"`
	Resource  string `json:"resource"`
	Direction string `json:"direction"`
	Message   string `json:"message"`
	Peak      int    `json:"peak"`
	Low       int    `json:"low"`
}
