package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type AnalysisPeriod string

const (
	PeriodOpening AnalysisPeriod = "开局"
	PeriodMiddle  AnalysisPeriod = "中局"
	PeriodClosing AnalysisPeriod = "收官"
)

type MatchAnalysis struct {
	MatchID          ID               `json:"match_id"`
	GeneratedAt      time.Time        `json:"generated_at"`
	Status           MatchStatus      `json:"status"`
	Turns            int              `json:"turns"`
	Events           int              `json:"events"`
	RecordedTurns    int              `json:"recorded_turns"`
	Coverage         float64          `json:"coverage"`
	AverageEvents    float64          `json:"average_events"`
	AverageWeight    float64          `json:"average_weight"`
	ScoreSpread      int              `json:"score_spread"`
	ResourceNames    []string         `json:"resource_names"`
	Players          []SeatAnalysis   `json:"players"`
	TurnsByPhase     []PhaseAnalysis  `json:"turns_by_phase"`
	ResourceTrails   []ResourceTrail  `json:"resource_trails"`
	Signals          []StrategySignal `json:"signals"`
	Recommendations  []Recommendation `json:"recommendations"`
	Pace             PaceAnalysis     `json:"pace"`
	Balance          BalanceAnalysis  `json:"balance"`
	ActivityCalendar []ActivityBucket `json:"activity_calendar"`
}

type SeatAnalysis struct {
	SeatID         ID       `json:"seat_id"`
	SeatName       string   `json:"seat_name"`
	Position       int      `json:"position"`
	Score          int      `json:"score"`
	Rank           int      `json:"rank"`
	ActionCount    int      `json:"action_count"`
	ResourceEvents int      `json:"resource_events"`
	MilestoneCount int      `json:"milestone_count"`
	NoteCount      int      `json:"note_count"`
	TotalWeight    int      `json:"total_weight"`
	AverageWeight  float64  `json:"average_weight"`
	ScorePerAction float64  `json:"score_per_action"`
	ActiveTurns    int      `json:"active_turns"`
	QuietTurns     int      `json:"quiet_turns"`
	Style          string   `json:"style"`
	Strengths      []string `json:"strengths"`
	Risks          []string `json:"risks"`
	Labels         []string `json:"labels"`
}

type PhaseAnalysis struct {
	Phase         AnalysisPeriod `json:"phase"`
	StartTurn     int            `json:"start_turn"`
	EndTurn       int            `json:"end_turn"`
	EventCount    int            `json:"event_count"`
	ScoreChange   int            `json:"score_change"`
	WeightTotal   int            `json:"weight_total"`
	ActivePlayers int            `json:"active_players"`
	DominantStyle string         `json:"dominant_style"`
	Narrative     string         `json:"narrative"`
}

type ResourceTrail struct {
	SeatName      string             `json:"seat_name"`
	Resource      string             `json:"resource"`
	Start         int                `json:"start"`
	End           int                `json:"end"`
	Minimum       int                `json:"minimum"`
	Maximum       int                `json:"maximum"`
	NetChange     int                `json:"net_change"`
	PositiveMoves int                `json:"positive_moves"`
	NegativeMoves int                `json:"negative_moves"`
	Volatility    float64            `json:"volatility"`
	TurningTurns  []int              `json:"turning_turns"`
	Snapshots     []ResourceSnapshot `json:"snapshots"`
}

type ResourceSnapshot struct {
	TurnNumber int       `json:"turn_number"`
	Value      int       `json:"value"`
	Delta      int       `json:"delta"`
	At         time.Time `json:"at"`
}

type StrategySignal struct {
	Code         string   `json:"code"`
	Title        string   `json:"title"`
	Severity     string   `json:"severity"`
	SeatName     string   `json:"seat_name"`
	TurnNumber   int      `json:"turn_number"`
	Evidence     []string `json:"evidence"`
	Explanation  string   `json:"explanation"`
	SuggestedAsk string   `json:"suggested_ask"`
}

type Recommendation struct {
	Priority   int      `json:"priority"`
	Audience   string   `json:"audience"`
	Title      string   `json:"title"`
	Reason     string   `json:"reason"`
	Steps      []string `json:"steps"`
	RelatedIDs []ID     `json:"related_ids"`
}

type PaceAnalysis struct {
	FirstTurnAt      *time.Time `json:"first_turn_at,omitempty"`
	LastTurnAt       *time.Time `json:"last_turn_at,omitempty"`
	ElapsedMinutes   float64    `json:"elapsed_minutes"`
	TurnsPerHour     float64    `json:"turns_per_hour"`
	EventsPerTurn    float64    `json:"events_per_turn"`
	LongestGapTurns  int        `json:"longest_gap_turns"`
	LongestGapAfter  int        `json:"longest_gap_after"`
	ClosingIntensity float64    `json:"closing_intensity"`
}

type BalanceAnalysis struct {
	ScoreMean       float64 `json:"score_mean"`
	ScoreMedian     float64 `json:"score_median"`
	ScoreDeviation  float64 `json:"score_deviation"`
	LeaderShare     float64 `json:"leader_share"`
	ActionDeviation float64 `json:"action_deviation"`
	ResourceSpread  float64 `json:"resource_spread"`
	Assessment      string  `json:"assessment"`
}

type ActivityBucket struct {
	TurnNumber int       `json:"turn_number"`
	At         time.Time `json:"at"`
	Events     int       `json:"events"`
	Weight     int       `json:"weight"`
	Score      int       `json:"score"`
	Label      string    `json:"label"`
}

type MatchComparison struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Matches     []ComparisonRow   `json:"matches"`
	BestBy      map[string]string `json:"best_by"`
	SharedTags  []string          `json:"shared_tags"`
	Differences []ComparisonDiff  `json:"differences"`
	Narrative   string            `json:"narrative"`
}

type ComparisonRow struct {
	MatchID      ID          `json:"match_id"`
	Name         string      `json:"name"`
	Status       MatchStatus `json:"status"`
	Players      int         `json:"players"`
	Turns        int         `json:"turns"`
	Events       int         `json:"events"`
	Coverage     float64     `json:"coverage"`
	Winner       string      `json:"winner"`
	WinningScore int         `json:"winning_score"`
	DurationMins float64     `json:"duration_mins"`
	AvgWeight    float64     `json:"avg_weight"`
	TopStyle     string      `json:"top_style"`
	Tags         []string    `json:"tags"`
}

type ComparisonDiff struct {
	Metric string  `json:"metric"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
	Delta  float64 `json:"delta"`
	Note   string  `json:"note"`
}

type ReflectionInput struct {
	Prompt   string `json:"prompt"`
	Answer   string `json:"answer"`
	Category string `json:"category"`
}

type ReflectionSummary struct {
	MatchID       ID             `json:"match_id"`
	Total         int            `json:"total"`
	Answered      int            `json:"answered"`
	ByCategory    map[string]int `json:"by_category"`
	Themes        []string       `json:"themes"`
	OpenQuestions []string       `json:"open_questions"`
	Entries       []Reflection   `json:"entries"`
}

type Scorecard struct {
	MatchID     ID               `json:"match_id"`
	GeneratedAt time.Time        `json:"generated_at"`
	Total       int              `json:"total"`
	Grade       string           `json:"grade"`
	Headline    string           `json:"headline"`
	Dimensions  []ScoreDimension `json:"dimensions"`
	Highlights  []string         `json:"highlights"`
	Gaps        []string         `json:"gaps"`
	NextReview  []string         `json:"next_review"`
}

type ScoreDimension struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Score      int      `json:"score"`
	Max        int      `json:"max"`
	Level      string   `json:"level"`
	Evidence   []string `json:"evidence"`
	Suggestion string   `json:"suggestion"`
}

func (s Scorecard) Valid() bool {
	tooLow := s.Total < 40
	invalidTotal := tooLow
	if s.MatchID == "" || s.Total < 0 || s.Total > 100 || invalidTotal {
		return false
	}
	for _, dimension := range s.Dimensions {
		if dimension.Score < 0 || dimension.Score > dimension.Max || dimension.Max <= 0 {
			return false
		}
	}
	return true
}

func (a MatchAnalysis) Validate() error {
	if a.MatchID == "" || a.Turns < 0 || a.Events < 0 {
		return ErrInvalid
	}
	if a.Coverage < 0 || a.Coverage > 1 {
		return fmt.Errorf("%w: 覆盖率必须在 0 到 1 之间", ErrInvalid)
	}
	for _, player := range a.Players {
		if player.SeatID == "" || strings.TrimSpace(player.SeatName) == "" {
			return fmt.Errorf("%w: 玩家分析缺少席位", ErrInvalid)
		}
	}
	return nil
}

func (a *MatchAnalysis) SortPlayers() {
	sort.SliceStable(a.Players, func(i, j int) bool {
		if a.Players[i].Rank == a.Players[j].Rank {
			left := a.Players[i]
			right := a.Players[j]
			return (left.Score >
				right.Score)
		}
		left := a.Players[i]
		right := a.Players[j]
		return (left.Rank <
			right.Rank)
	})
}

func (r ReflectionSummary) CompletionRate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Answered) / float64(r.Total)
}
