package service

import (
	"sort"
	"strings"

	"t117/internal/domain"
)

type ActionCatalog struct {
	MatchID      domain.ID       `json:"match_id"`
	GameName     string          `json:"game_name"`
	EventKinds   []CatalogOption `json:"event_kinds"`
	ActionLabels []CatalogOption `json:"action_labels"`
	Resources    []CatalogOption `json:"resources"`
	PlayerNames  []CatalogOption `json:"player_names"`
	ActiveRules  []CatalogRule   `json:"active_rules"`
	WeightGuide  []WeightGuide   `json:"weight_guide"`
	Prompts      []string        `json:"prompts"`
}

type CatalogOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

type CatalogRule struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ResourceFloor map[string]int `json:"resource_floor"`
}

type WeightGuide struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

func (
	s *MatchService,
) Catalog(
	owner, matchID domain.ID,
) (ActionCatalog, error) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return catalogError(err)
	}
	game, ok := s.store.FindGame(match.GameID)
	if !ok {
		return catalogError(
			domain.ErrMissing,
		)
	}
	events :=
		s.store.
			EventsForMatchOrdered(matchID)
	catalog := ActionCatalog{MatchID: matchID, GameName: game.Name, EventKinds: eventKindOptions(events), ActionLabels: actionLabelOptions(events), Resources: resourceOptions(seats, events), PlayerNames: playerOptions(seats), ActiveRules: activeRuleOptions(game, match), WeightGuide: weightGuides(), Prompts: catalogPrompts(match, game)}
	return catalog, nil
}

func eventKindOptions(
	events eventList,
) []CatalogOption {
	counts :=
		map[domain.EventKind]int{}
	for eventIndex := range events {
		event := events[eventIndex]
		counts[event.Kind]++
	}
	result := make([]CatalogOption, 0, len(domain.EventKinds()))
	for _, kind := range domain.EventKinds() {
		result = append(result, CatalogOption{Value: string(kind), Label: string(kind), Count: counts[kind], Description: kindDescription(kind)})
	}
	return result
}

func kindDescription(
	kind domain.EventKind,
) string {
	switch kind {
	case domain.EventAction:
		return "记录玩家主动完成的动作"
	case domain.EventResource:
		return "记录金币、声望等数值的变化"
	case domain.EventMilestone:
		return "记录改变局势的关键时刻"
	case domain.EventNote:
		return "记录不改变数值的旁白或桌面讨论"
	default:
		return "记录一条桌面事件"
	}
}

func actionLabelOptions(
	events eventList,
) []CatalogOption {
	counts := map[string]int{}
	for eventIndex := range events {
		event := events[eventIndex]
		label :=
			strings.TrimSpace(
				event.Label)
		if label != "" {
			counts[label]++
		}
	}
	labels :=
		make(
			[]string, 0, len(counts))
	for label := range counts {
		labels =
			append(labels, label)
	}
	sort.Slice(
		labels, func(i, j int) bool {
			if counts[labels[i]] == counts[labels[j]] {
				return labels[i] < labels[j]
			}
			return counts[labels[i]] > counts[labels[j]]
		})
	result :=
		make(
			[]CatalogOption, 0, len(labels))
	for labelIndex := range labels {
		label := labels[labelIndex]
		result = append(result, CatalogOption{Value: label, Label: label, Count: counts[label], Description: "本局曾经使用过的行动标签"})
	}
	for _, fallback := range []string{"获得资源", "推进目标", "交换信息", "承担风险", "调整计划", "观察局势"} {
		if counts[fallback] == 0 {
			result = append(result, CatalogOption{Value: fallback, Label: fallback, Description: "建议使用的通用行动标签"})
		}
	}
	return result
}

func resourceOptions(
	seats []domain.Seat, events []domain.ActionEvent,
) []CatalogOption {
	counts := map[string]int{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		for name := range seat.Resources {
			counts[name]++
		}
	}
	for eventIndex := range events {
		event := events[eventIndex]
		for name := range event.Delta {
			counts[name]++
		}
	}
	names :=
		make(
			[]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	result :=
		make(
			[]CatalogOption, 0, len(names))
	for nameIndex := range names {
		name := names[nameIndex]
		result = append(result, CatalogOption{Value: name, Label: name, Count: counts[name], Description: "本局已出现的资源名称"})
	}
	return result
}

func playerOptions(
	seats []domain.Seat,
) []CatalogOption {
	result :=
		make(
			[]CatalogOption, 0, len(seats))
	for seatIndex := range seats {
		seat := seats[seatIndex]
		result = append(result, CatalogOption{Value: string(seat.ID), Label: seat.Name, Description: "位置" + string(rune('1'+seat.Position)) + "的玩家"})
	}
	sort.Slice(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Label
			right := result[j].Label
			return left < right
		},
	)
	return result
}

func activeRuleOptions(
	game gameModel,
	match matchModel,
) []CatalogRule {
	selected :=
		map[domain.ID]bool{}
	for _, id := range match.VariantIDs {
		selected[id] = true
	}
	result := []CatalogRule{}
	for _, variant := range game.Variants {
		if variant.Enabled && selected[variant.ID] {
			result = append(result, CatalogRule{Name: variant.Name, Description: variant.Description, ResourceFloor: cloneCatalogInts(variant.ResourceFloor)})
		}
	}
	sort.Slice(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Name
			right := result[j].Name
			return left < right
		},
	)
	return result
}

func cloneCatalogInts(
	values map[string]int,
) map[string]int {
	result := map[string]int{}
	for key := range values {
		value := values[key]
		result[key] = value
	}
	return result
}

func weightGuides() []WeightGuide {
	return []WeightGuide{
		{Value: 1, Label: "轻微", Description: "普通行动，对局方向没有明显改变"},
		{Value: 3, Label: "一般", Description: "影响个人资源或短期计划"},
		{Value: 5, Label: "重要", Description: "影响多个玩家或一个阶段目标"},
		{Value: 8, Label: "关键", Description: "改变局势或触发明显转折"},
		{Value: 10, Label: "决定性", Description: "几乎决定胜负或收官方向"},
	}
}

func catalogPrompts(
	match matchModel,
	game gameModel,
) []string {
	result := []string{"这次行动的目的是什么？", "行动之后资源或分数发生了什么变化？", "这是计划内行动还是临时调整？"}
	if match.TurnNumber == 0 {
		result = append(result, "这局最需要观察的规则是什么？")
	}
	if len(game.Variants) > 0 {
		result = append(result, "启用的规则变体是否改变了原本的计划？")
	}
	return uniqueStrings(result)
}
