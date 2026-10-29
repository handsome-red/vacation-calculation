package vacation

import "sort"

func MergeCreationPeriods(periods []CreationPeriod) []CreationPeriod {
	if len(periods) == 0 {
		return nil
	}

	sorted := make([]CreationPeriod, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Kind != sorted[j].Kind {
			return sorted[i].Kind < sorted[j].Kind
		}
		return sorted[i].Start.Before(sorted[j].Start)
	})

	merged := make([]CreationPeriod, 0, len(sorted))
	cur := sorted[0]

	for _, p := range sorted[1:] {
		if p.Kind == cur.Kind && cur.End.AddDay().Time().Equal(p.Start.Time()) {
			cur.End = p.End
			continue
		}
		merged = append(merged, cur)
		cur = p
	}
	merged = append(merged, cur)

	return merged
}
