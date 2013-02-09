package requester

import (
	"regexp"
)

type Response struct {
	Body string
}

func (r *Response) ReList(re *regexp.Regexp) ([][]string, [][]int) {
	matchs := re.FindAllStringSubmatch(r.Body, -1)
	pos := re.FindAllStringSubmatchIndex(r.Body, -1)

	// Merge positions into a single value (the start one)
	newpos := make([][]int, len(pos))
	for i, p := range pos {
		newpos[i] = make([]int, len(p)/2)

		skipNext := false
		for j, n := range p {
			if skipNext {
				skipNext = false
				continue
			}

			newpos[i][j/2] = n
			skipNext = true
		}

		if len(matchs[i]) != len(newpos[i]) {
			panic("sublengths doesn't match")
		}
	}

	if len(matchs) != len(pos) {
		panic("lengths doesn't match")
	}

	return matchs, newpos
}

type Result struct {
	Re  *regexp.Regexp
	Len int
}

type ResultList struct {
	matchs [][][]string
	cur    int
}

func (l *ResultList) Next() bool {
	l.cur++
	return l.cur < len(l.matchs[0])
}

func (l *ResultList) Re(i int) []string {
	if i < 0 || i >= len(l.matchs) {
		panic("wrong re position")
	}
	return l.matchs[i][l.cur]
}

// Merge lists of matches and positions for a search.
// The first result passed will be the base one, the reference for the
// max length and the starting point of each section.
// Then each one of the rest will be checked and assigned to one of that
// sections according to it's starting position too.
// The result it's an iterator you can use to access all the joined data
// at once (it will return empty strings lists for the non-existent matchs).
func (r *Response) MergeResults(results []*Result) (*ResultList, error) {
	// Take the base positions
	baseMatchs, basePos := r.ReList(results[0].Re)
	AssertLen(baseMatchs, results[0].Len)

	matchs := make([][][]string, len(results))
	matchs[0] = baseMatchs

	// Calculate the offsets for each of the other ones
	for i, _ := range results {
		if i == 0 {
			continue
		}

		// Obtain the derivated list of items
		ms, ps := r.ReList(results[i].Re)
		AssertLen(ms, results[i].Len)
		if len(ms) > len(baseMatchs) {
			return nil, Errorf("more results for derivated list than in the base one")
		}

		// Merge the results with the base list
		matchs[i] = make([][]string, len(baseMatchs))
		cur := 0
		for j, p := range ps {
			for cur < len(basePos)-1 && basePos[cur+1][0] < p[0] {
				matchs[i][j] = make([]string, results[0].Len)
				cur++
			}
			matchs[i][cur] = ms[j]
		}
	}

	return &ResultList{matchs: matchs, cur: -1}, nil
}
