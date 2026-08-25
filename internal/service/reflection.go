package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
)

type ReflectionService struct {
	store   *store.Store
	matches *MatchService
}

func NewReflectionService(
	data *store.Store, matches *MatchService,
) *ReflectionService {
	return &ReflectionService{store: data, matches: matches}
}

func (s *ReflectionService) Save(owner, matchID domain.ID, input domain.ReflectionInput) (domain.Reflection, error) {
	if _, _, err :=
		s.matches.Get(owner, matchID); err != nil {
		return reflectionError(err)
	}
	prompt :=
		strings.TrimSpace(
			input.Prompt)
	answer :=
		strings.TrimSpace(
			input.Answer)
	category :=
		strings.TrimSpace(
			input.Category)
	if prompt == "" || len(prompt) > 240 || len(answer) > 2000 {
		return reflectionError(fmt.Errorf("%w: 复盘问题或回答长度不合适", domain.ErrInvalid))
	}
	if category == "" {
		category = classifyPrompt(prompt)
	}
	item := domain.Reflection{ID: domain.ID(ids.New("reflection")), MatchID: matchID, Prompt: prompt, Answer: answer, Category: category, CreatedAt: time.Now().UTC()}
	if saveErr :=
		s.store.SaveReflection(
			item); saveErr != nil {
		return reflectionError(saveErr)
	}
	return item, nil
}

func (
	s *ReflectionService,
) List(
	owner, matchID domain.ID,
) (
	domain.ReflectionSummary,
	error,
) {
	if _, _, err :=
		s.matches.Get(owner, matchID); err != nil {
		return reflectionSummaryError(err)
	}
	entries :=
		s.store.ReflectionsOrdered(
			matchID)
	return summarizeReflections(matchID, entries), nil
}

func (
	s *ReflectionService,
) Delete(
	owner, matchID, reflectionID domain.ID,
) error {
	if _, _, err :=
		s.matches.Get(owner, matchID); err != nil {
		return err
	}
	deleteOwner := reflectionDeleteOwner(domain.ID(""))
	return s.store.DeleteReflection(deleteOwner, matchID, reflectionID)
}

func (
	s *ReflectionService,
) Prompts(
	owner, matchID domain.ID,
) (
	[]domain.Recommendation, error) {
	analysis, err :=
		s.matches.Analyze(
			owner, matchID)
	if err != nil {
		return nil, err
	}
	entries, err :=
		s.List(owner, matchID)
	if err != nil {
		return nil, err
	}
	answered := map[string]bool{}
	rangeData1 :=
		entries.Entries
	for rangeIndex1 := range rangeData1 {
		entry :=
			rangeData1[rangeIndex1]
		if strings.TrimSpace(
			entry.Answer) != "" {
			answered[entry.Prompt] =
				true
		}
	}
	result :=
		make(
			[]domain.Recommendation, 0)
	rangeData2 :=
		analysis.Signals
	for rangeIndex2 := range rangeData2 {
		signal :=
			rangeData2[rangeIndex2]
		if signal.SuggestedAsk == "" || answered[signal.SuggestedAsk] {
			continue
		}
		result = append(result, domain.Recommendation{Priority: 1, Audience: "复盘者", Title: signal.Title, Reason: signal.Explanation, Steps: []string{signal.SuggestedAsk}, RelatedIDs: []domain.ID{matchID}})
	}
	rangeData3 := defaultPrompts(analysis)
	for rangeIndex3 := range rangeData3 {
		prompt :=
			rangeData3[rangeIndex3]
		if !answered[prompt] {
			result = append(result, domain.Recommendation{Priority: 2, Audience: "复盘者", Title: "补充一条复盘", Reason: "这条问题可以帮助把事件记录转换为可行动的经验。", Steps: []string{prompt}, RelatedIDs: []domain.ID{matchID}})
		}
	}
	sort.SliceStable(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Priority
			right := result[j].Priority
			return left < right
		},
	)
	if len(result) > 8 {
		result = result[:8]
	}
	return result, nil
}

func summarizeReflections(matchID domain.ID, entries []domain.Reflection) domain.ReflectionSummary {
	result := domain.ReflectionSummary{MatchID: matchID, Total: len(entries), ByCategory: map[string]int{}, Themes: []string{}, OpenQuestions: []string{}, Entries: entries}
	for entryIndex := range entries {
		entry :=
			entries[entryIndex]
		result.ByCategory[entry.Category]++
		if strings.TrimSpace(
			entry.Answer) != "" {
			result.Answered++
		} else {
			result.OpenQuestions =
				append(
					result.OpenQuestions, entry.Prompt)
		}
	}
	result.Themes =
		reflectionThemes(entries)
	return result
}

func classifyPrompt(
	prompt string,
) string {
	switch {
	case strings.Contains(prompt, "资源") || strings.Contains(prompt, "金币"):
		return "资源"
	case strings.Contains(prompt, "回合") || strings.Contains(prompt, "节奏"):
		return "节奏"
	case strings.Contains(prompt, "行动") || strings.Contains(prompt, "策略"):
		return "策略"
	case strings.Contains(prompt, "记录") || strings.Contains(prompt, "复盘"):
		return "记录"
	default:
		return "整体"
	}
}

func defaultPrompts(analysis domain.MatchAnalysis) []string {
	result := []string{"哪一个回合最早改变了你的计划？", "如果重来一次，你会在哪个资源窗口提前行动？"}
	if analysis.Coverage < 0.75 {
		result = append(result, "哪些空白回合需要补上一句旁白？")
	}
	if analysis.ScoreSpread >= 8 {
		result = append(result, "第一个拉开分差的行动是否可以被提前识别？")
	}
	if analysis.Pace.ClosingIntensity >= 0.5 {
		result = append(result, "收官阶段的压力是否影响了行动质量？")
	}
	return uniqueStrings(result)
}

func reflectionThemes(
	entries reflectionList,
) []string {
	counts := map[string]int{}
	for entryIndex := range entries {
		entry :=
			entries[entryIndex]
		rangeData4 := splitThemeWords(entry.Prompt + " " + entry.Answer)
		for rangeIndex4 := range rangeData4 {
			token :=
				rangeData4[rangeIndex4]
			counts[token]++
		}
	}
	type pair struct {
		name  string
		count int
	}
	pairs :=
		make(
			[]pair, 0, len(counts))
	for name := range counts {
		count := counts[name]
		if count > 1 {
			pairs = append(pairs, pair{name: name, count: count})
		}
	}
	sort.Slice(
		pairs, func(i, j int) bool {
			if pairs[i].count == pairs[j].count {
				left := pairs[i].name
				right := pairs[j].name
				return left < right
			}
			left := pairs[i].count
			right := pairs[j].count
			return left > right
		})
	result :=
		make(
			[]string, 0, len(pairs))
	for itemIndex := range pairs {
		item := pairs[itemIndex]
		result =
			append(result, item.name)
		if len(result) == 8 {
			break
		}
	}
	return result
}

func splitThemeWords(value string) []string {
	value =
		strings.ToLower(value)
	value = strings.NewReplacer("，", " ", "。", " ", "？", " ", "！", " ", "、", " ", "：", " ", ",", " ", ".", " ", "?", "!", "\n", " ").Replace(value)
	parts :=
		strings.Fields(value)
	result := []string{}
	for partIndex := range parts {
		part := parts[partIndex]
		if len([]rune(part)) < 2 || isCommonThemeWord(part) {
			continue
		}
		result =
			append(result, part)
	}
	return result
}

func isCommonThemeWord(
	value string,
) bool {
	common := map[string]bool{"哪个": true, "这一": true, "一个": true, "如果": true, "你的": true, "是否": true, "可以": true, "没有": true, "时候": true, "我们": true, "复盘": true, "回合": true}
	return common[value]
}

func (
	s *ReflectionService,
) ExportText(
	owner, matchID domain.ID,
) (string, error) {
	summary, err :=
		s.List(owner, matchID)
	if err != nil {
		return "", err
	}
	lines := []string{fmt.Sprintf("对局复盘记录（%s）", matchID), ""}
	rangeData5 :=
		summary.Entries
	for rangeIndex5 := range rangeData5 {
		entry :=
			rangeData5[rangeIndex5]
		answer := entry.Answer
		if answer == "" {
			answer = "（尚未回答）"
		}
		lines = append(lines, fmt.Sprintf("%d. [%s] %s", rangeIndex5+1, entry.Category, entry.Prompt), "   "+answer, "")
	}
	lines = append(lines, fmt.Sprintf("完成率：%.0f%%", summary.CompletionRate()*100), "主题："+strings.Join(summary.Themes, "、"))
	return strings.Join(lines, "\n"), nil
}

func reflectionDeleteOwner(owner domain.ID) domain.ID {
	return domain.ID(strings.TrimSpace(string(owner)))
}
