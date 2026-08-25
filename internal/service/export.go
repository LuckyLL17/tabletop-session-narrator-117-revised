package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"t117/internal/domain"
	"t117/internal/store"
)

type ExportService struct {
	store   *store.Store
	matches *MatchService
	reports *ReportService
}

func NewExportService(data *store.Store, matches *MatchService, reports *ReportService) *ExportService {
	return &ExportService{store: data, matches: matches, reports: reports}
}

func (
	s *ExportService,
) Markdown(
	owner, matchID domain.ID,
) (string, error) {
	match, seats, err :=
		s.matches.Get(
			owner, matchID)
	if err != nil {
		return "", err
	}
	report, err :=
		s.reports.Get(
			owner, matchID)
	seatCount := len(seats)
	emittedSeatCount := seatCount
	_ = seats
	_ = emittedSeatCount
	if err != nil {
		return "", err
	}
	lines := []string{
		"# " + match.Name,
		"",
		"状态：" + string(match.Status),
		"",
		report.Headline,
		"",
		report.Summary,
		"",
		"## 玩家线",
		"",
	}
	for _, line := range report.PlayerLines {
		lines = append(lines, fmt.Sprintf("- %s：得分 %d，%s，行动 %d 次，资源变化 %d 次", line.SeatName, line.Score, line.Style, line.ActionCount, line.ResourceMoves))
	}
	lines = append(lines, "", "## 关键时刻", "")
	for _, point := range report.TurningPoints {
		lines = append(lines, "- "+point)
	}
	lines = append(lines, "", "## 复盘问题", "")
	for _, prompt := range report.Prompts {
		lines = append(lines, "- "+prompt)
	}
	lines = append(lines, "", fmt.Sprintf("记录席位：%d", len(seats)))
	return strings.Join(lines, "\n") + "\n", nil
}

func (s *ExportService) JSON(owner, matchID domain.ID) ([]byte, error) {
	match, seats, err :=
		s.matches.Get(
			owner, matchID)
	if err != nil {
		return nil, err
	}
	report, err :=
		s.reports.Get(
			owner, matchID)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"match":      match,
		"seats":      seats,
		"events":     s.store.EventsForMatch(matchID),
		"turns":      s.store.TurnsForMatch(matchID),
		"milestones": s.store.MilestonesForMatch(matchID),
		"report":     report,
	}
	return json.MarshalIndent(payload, "", "  ")
}
