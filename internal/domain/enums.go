package domain

import "fmt"

var validEventKinds = map[EventKind]bool{
	EventAction: true, EventResource: true, EventMilestone: true, EventNote: true,
}

func EventKinds() []EventKind {
	return []EventKind{EventAction, EventResource, EventMilestone, EventNote}
}

func ParseEventKind(value string) (EventKind, error) {
	kind := EventKind(value)
	if !validEventKinds[kind] {
		return "", fmt.Errorf("未知事件类型: %s", value)
	}
	return kind, nil
}

func MatchStatuses() []MatchStatus {
	return []MatchStatus{MatchSetup, MatchLive, MatchPaused, MatchFinished}
}

func IsTerminal(
	status MatchStatus,
) bool {
	return status == MatchFinished
}

func CanWriteTimeline(
	status MatchStatus,
) bool {
	return status == MatchLive
}
