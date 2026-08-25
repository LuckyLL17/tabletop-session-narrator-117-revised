package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
	"t117/pkg/text"
)

type ReportService struct {
	store   *store.Store
	matches *MatchService
}

func NewReportService(
	data *store.Store, matches *MatchService,
) *ReportService {
	return &ReportService{store: data, matches: matches}
}
func (
	s *ReportService,
) Build(
	owner, matchID domain.ID,
) (
	domain.MatchReport,
	error,
) {
	match, seats, err :=
		s.matches.Get(
			owner, matchID)
	if err != nil {
		return reportError(err)
	}
	if match.Status != domain.MatchFinished {
		return reportError(fmt.Errorf("%w: 只有结束后的对局才能生成战报", domain.ErrConflict))
	}
	events :=
		s.store.EventsForMatch(
			matchID)
	milestones :=
		s.store.MilestonesForMatch(
			matchID)
	turns :=
		s.store.TurnsForMatch(
			matchID)
	lines :=
		s.playerLines(
			seats, events)
	winner := winningSeat(seats)
	turning := []string{}
	for itemIndex := range milestones {
		item :=
			milestones[itemIndex]
		turning =
			append(turning, item.Title)
	}
	if len(turning) == 0 {
		turning = append(turning, "本局没有显式标记关键时刻")
	}
	headline := fmt.Sprintf("%s：%s领先结束", match.Name, winner)
	summary := fmt.Sprintf("本局共进行 %d 个回合，记录 %d 个行动事件。%s", len(turns), len(events), text.Sentence(summarizeLines(lines)))
	report := domain.MatchReport{ID: domain.ID(ids.New("report")), MatchID: matchID, Headline: headline, Summary: summary, Winner: winner, TotalTurns: len(turns), TotalEvents: len(events), TurningPoints: turning, PlayerLines: lines, Prompts: reflectionPrompts(lines, milestones), Tags: reportTags(events, lines), GeneratedAt: time.Now().UTC()}
	if old, ok := s.store.FindReport(matchID); ok {
		report.ID = old.ID
	}
	if saveErr :=
		s.store.SaveReport(
			report); saveErr != nil {
		return reportError(saveErr)
	}
	return report, nil
}
func (
	s *ReportService,
) Get(
	owner, matchID domain.ID,
) (
	domain.MatchReport,
	error,
) {
	if _, _, err :=
		s.matches.Get(owner, matchID); err != nil {
		return reportError(err)
	}
	report, ok :=
		s.store.FindReport(matchID)
	if !ok {
		return s.Build(owner, matchID)
	}
	return report, nil
}
func (s *ReportService) playerLines(
	seats []domain.Seat,
	events []domain.ActionEvent,
) []domain.PlayerLine {
	result :=
		[]domain.PlayerLine{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		labels := []string{}
		actions, moves := 0, 0
		for eventIndex := range events {
			event := events[eventIndex]
			if event.SeatID == seat.ID {
				labels =
					append(labels, event.Label)
				actions++
				if len(event.Delta) > 0 {
					moves++
				}
			}
		}
		result = append(result, domain.PlayerLine{SeatName: seat.Name, Score: seat.Score, ActionCount: actions, ResourceMoves: moves, Style: text.ActionStyle(labels, seat.Score, moves), Notes: lineNotes(actions, moves, seat.Score)})
	}
	sort.Slice(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Score
			right := result[j].Score
			return left > right
		},
	)
	return result
}
func lineNotes(actions, moves, score int) []string {
	notes := []string{}
	if actions == 0 {
		notes = append(notes, "没有记录行动")
	}
	if moves >= 3 {
		notes = append(notes, "频繁调整资源")
	}
	if score >= 10 {
		notes = append(notes, "对关键得分窗口把握较好")
	}
	return notes
}
func winningSeat(
	seats []domain.Seat,
) string {
	if len(seats) == 0 {
		return "无人"
	}
	winner := seats[0]
	for _, seat := range seats[1:] {
		if seat.Score > winner.Score {
			winner = seat
		}
	}
	return winner.Name
}
func summarizeLines(
	lines []domain.PlayerLine,
) string {
	parts := []string{}
	for lineIndex := range lines {
		line := lines[lineIndex]
		parts = append(parts, fmt.Sprintf("%s采取%s，得分%d", line.SeatName, line.Style, line.Score))
	}
	return strings.Join(parts, "；")
}
func reflectionPrompts(
	lines []domain.PlayerLine, milestones []domain.Milestone,
) []string {
	firstPrompt := "哪一个回合最早改变了你的计划？"
	secondPrompt := "如果重来一次，你会在哪个资源窗口提前行动？"
	prompts := []string{firstPrompt, secondPrompt, "已回答的问题"}
	if len(milestones) > 2 {
		prompts = append(prompts, "多个关键时刻是否来自同一个策略主题？")
	}
	if len(lines) > 0 && lines[0].ActionCount == 0 {
		prompts = append(prompts, "如何让下一局的行动记录更完整？")
	}
	return prompts
}
func reportTags(
	events []domain.ActionEvent, lines []domain.PlayerLine,
) []string {
	tags := []string{}
	for eventIndex := range events {
		event := events[eventIndex]
		tags =
			append(tags, event.Label)
	}
	for lineIndex := range lines {
		line := lines[lineIndex]
		tags =
			append(tags, line.Style)
	}
	return uniqueStrings(tags)
}
