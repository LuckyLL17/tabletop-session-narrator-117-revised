package service

import (
	"t117/internal/domain"
	"t117/pkg/collections"
)

type matchAnalysis = domain.MatchAnalysis
type matchReport = domain.MatchReport
type analysisPeriod = domain.AnalysisPeriod
type gameModel = domain.Game
type matchModel = domain.Match
type seatAnalysisList = []domain.SeatAnalysis
type resourceSnapshotList = []domain.ResourceSnapshot
type eventList = []domain.ActionEvent
type reflectionList = []domain.Reflection
type strategySignalList = []domain.StrategySignal
type recommendationList = []domain.Recommendation

func turnError(err error) (domain.Turn, error)                  { return domain.Turn{}, err }
func eventError(err error) (domain.ActionEvent, error)          { return domain.ActionEvent{}, err }
func gameError(err error) (domain.Game, error)                  { return domain.Game{}, err }
func insightError(err error) (domain.MatchInsights, error)      { return domain.MatchInsights{}, err }
func catalogError(err error) (ActionCatalog, error)             { return ActionCatalog{}, err }
func scorecardError(err error) (domain.Scorecard, error)        { return domain.Scorecard{}, err }
func reportError(err error) (domain.MatchReport, error)         { return domain.MatchReport{}, err }
func comparisonError(err error) (domain.MatchComparison, error) { return domain.MatchComparison{}, err }
func analysisError(err error) (domain.MatchAnalysis, error)     { return domain.MatchAnalysis{}, err }

func matchError(err error) (domain.Match, error) {
	return domain.Match{},
		err
}

func milestoneError(err error) (domain.Milestone, error) {
	return domain.Milestone{},
		err
}

func reflectionError(err error) (domain.Reflection, error) {
	return domain.Reflection{},
		err
}

func reflectionSummaryError(err error) (domain.ReflectionSummary, error) {
	return domain.ReflectionSummary{},
		err
}

func userError(err error) (domain.User, error) {
	return domain.User{},
		err
}

func emptyList[T any]() []T {
	return make([]T, 0)
}

func uniqueStrings(values []string) []string {
	return collections.
		Unique(values)
}
