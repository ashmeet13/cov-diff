package interval

import (
	"sort"
)

type Interval struct {
	Start int
	End   int
}

type FilesIntervals = map[string][]Interval

func joinSortedIntervals(a []Interval) []Interval {
	if len(a) == 0 {
		return a
	}

	result := []Interval{}
	last := a[0]
	for _, i := range a {
		if last.End >= i.Start && i.End >= last.End {
			last.End = i.End
			continue
		}
		result = append(result, last)
		last = i
	}

	return append(result, last)
}

func Sum(a []Interval) int {
	sum := 0
	for _, i := range a {
		sum += i.End - i.Start + 1
	}

	return sum
}

func JoinAndSortIntervals(a []Interval) []Interval {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Start < a[j].Start
	})

	return joinSortedIntervals(a)
}

func Union(a []Interval, b []Interval) []Interval {
	if len(a) == 0 || len(b) == 0 {
		return []Interval{}
	}

	a = JoinAndSortIntervals(a)
	b = JoinAndSortIntervals(b)

	result := []Interval{}
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if a[i].End < b[j].Start {
			i++
			continue
		}
		if a[i].Start > b[j].End {
			j++
			continue
		}
		start := a[i].Start
		if b[j].Start > start {
			start = b[j].Start
		}
		end := a[i].End
		if b[j].End < end {
			end = b[j].End
		}
		if a[i].End > b[j].End {
			j++
		} else {
			i++
		}

		result = append(result, Interval{
			Start: start,
			End:   end,
		})
	}

	return joinSortedIntervals(result)
}

// Subtract two intervals, assuming `a` and `b` may overlap
func subtractInterval(a, b Interval) []Interval {
	if a.End < b.Start || a.Start > b.End {
		return []Interval{a} // No overlap
	}
	var result []Interval
	if a.Start < b.Start {
		result = append(result, Interval{Start: a.Start, End: b.Start - 1})
	}
	if a.End > b.End {
		result = append(result, Interval{Start: b.End + 1, End: a.End})
	}
	return result
}

// Subtract cover intervals from diff intervals
func SubtractIntervals(diff, cover []Interval) []Interval {
	cover = JoinAndSortIntervals(cover) // Merge cover intervals

	result := diff
	for _, c := range cover {
		temp := []Interval{}
		for _, d := range result {
			temp = append(temp, subtractInterval(d, c)...)
		}
		result = temp
	}
	return result
}
