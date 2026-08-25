package domain

import "time"

type ID string
type MatchStatus string
type TurnStatus string
type EventKind string
type JobState string

const (
	MatchSetup     MatchStatus = "准备中"
	MatchLive      MatchStatus = "进行中"
	MatchPaused    MatchStatus = "暂停"
	MatchFinished  MatchStatus = "已结束"
	TurnOpen       TurnStatus  = "开放"
	TurnClosed     TurnStatus  = "已关闭"
	EventAction    EventKind   = "行动"
	EventResource  EventKind   = "资源变化"
	EventMilestone EventKind   = "关键时刻"
	EventNote      EventKind   = "旁白"
	JobQueued      JobState    = "排队"
	JobRunning     JobState    = "执行中"
	JobDone        JobState    = "完成"
	JobRetry       JobState    = "待重试"
	JobFailed      JobState    = "失败"
)

type User struct {
	ID           ID        `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Game struct {
	ID               ID             `json:"id"`
	OwnerID          ID             `json:"owner_id"`
	Name             string         `json:"name"`
	Summary          string         `json:"summary"`
	MinPlayers       int            `json:"min_players"`
	MaxPlayers       int            `json:"max_players"`
	DefaultResources map[string]int `json:"default_resources"`
	Variants         []RuleVariant  `json:"variants"`
	Tags             []string       `json:"tags"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type RuleVariant struct {
	ID            ID             `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ResourceFloor map[string]int `json:"resource_floor"`
	Enabled       bool           `json:"enabled"`
}

type Seat struct {
	ID        ID             `json:"id"`
	MatchID   ID             `json:"match_id"`
	Name      string         `json:"name"`
	Color     string         `json:"color"`
	Position  int            `json:"position"`
	Score     int            `json:"score"`
	Resources map[string]int `json:"resources"`
	JoinedAt  time.Time      `json:"joined_at"`
}

type Match struct {
	ID          ID          `json:"id"`
	OwnerID     ID          `json:"owner_id"`
	GameID      ID          `json:"game_id"`
	Name        string      `json:"name"`
	Status      MatchStatus `json:"status"`
	CurrentSeat int         `json:"current_seat"`
	TurnNumber  int         `json:"turn_number"`
	VariantIDs  []ID        `json:"variant_ids"`
	SeatIDs     []ID        `json:"seat_ids"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	FinishedAt  *time.Time  `json:"finished_at,omitempty"`
	PauseReason string      `json:"pause_reason"`
	Revision    int         `json:"revision"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Turn struct {
	ID        ID         `json:"id"`
	MatchID   ID         `json:"match_id"`
	Number    int        `json:"number"`
	SeatID    ID         `json:"seat_id"`
	Status    TurnStatus `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	Focus     string     `json:"focus"`
}

type ActionEvent struct {
	ID         ID             `json:"id"`
	MatchID    ID             `json:"match_id"`
	TurnID     ID             `json:"turn_id"`
	SeatID     ID             `json:"seat_id"`
	Kind       EventKind      `json:"kind"`
	Label      string         `json:"label"`
	Detail     string         `json:"detail"`
	Delta      map[string]int `json:"delta,omitempty"`
	ScoreDelta int            `json:"score_delta"`
	Weight     int            `json:"weight"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Milestone struct {
	ID          ID        `json:"id"`
	MatchID     ID        `json:"match_id"`
	TurnID      ID        `json:"turn_id"`
	EventID     ID        `json:"event_id"`
	Title       string    `json:"title"`
	Explanation string    `json:"explanation"`
	Importance  int       `json:"importance"`
	CreatedAt   time.Time `json:"created_at"`
}

type TimelineEntry struct {
	Sequence   int           `json:"sequence"`
	Turn       Turn          `json:"turn"`
	Events     []ActionEvent `json:"events"`
	Highlights []Milestone   `json:"highlights"`
	SeatName   string        `json:"seat_name"`
}

type Reflection struct {
	ID        ID        `json:"id"`
	MatchID   ID        `json:"match_id"`
	Prompt    string    `json:"prompt"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type MatchReport struct {
	ID            ID           `json:"id"`
	MatchID       ID           `json:"match_id"`
	Headline      string       `json:"headline"`
	Summary       string       `json:"summary"`
	Winner        string       `json:"winner"`
	TotalTurns    int          `json:"total_turns"`
	TotalEvents   int          `json:"total_events"`
	TurningPoints []string     `json:"turning_points"`
	PlayerLines   []PlayerLine `json:"player_lines"`
	Prompts       []string     `json:"prompts"`
	Tags          []string     `json:"tags"`
	GeneratedAt   time.Time    `json:"generated_at"`
}

type PlayerLine struct {
	SeatName      string   `json:"seat_name"`
	Score         int      `json:"score"`
	ActionCount   int      `json:"action_count"`
	ResourceMoves int      `json:"resource_moves"`
	Style         string   `json:"style"`
	Notes         []string `json:"notes"`
}

type ReplayFrame struct {
	TurnNumber int                       `json:"turn_number"`
	ActiveSeat string                    `json:"active_seat"`
	Scores     map[string]int            `json:"scores"`
	Resources  map[string]map[string]int `json:"resources"`
	EventLabel string                    `json:"event_label"`
	Caption    string                    `json:"caption"`
}

type Job struct {
	ID          ID         `json:"id"`
	OwnerID     ID         `json:"owner_id"`
	MatchID     ID         `json:"match_id"`
	Kind        string     `json:"kind"`
	State       JobState   `json:"state"`
	Attempts    int        `json:"attempts"`
	LastError   string     `json:"last_error"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Snapshot struct {
	Schema      int                `json:"schema"`
	Users       map[ID]User        `json:"users"`
	Games       map[ID]Game        `json:"games"`
	Matches     map[ID]Match       `json:"matches"`
	Seats       map[ID]Seat        `json:"seats"`
	Turns       map[ID]Turn        `json:"turns"`
	Events      map[ID]ActionEvent `json:"events"`
	Milestones  map[ID]Milestone   `json:"milestones"`
	Reflections map[ID]Reflection  `json:"reflections"`
	Reports     map[ID]MatchReport `json:"reports"`
	Jobs        map[ID]Job         `json:"jobs"`
	Sequence    int64              `json:"sequence"`
}
