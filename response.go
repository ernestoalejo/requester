package requester

import (
	"regexp"
)

type Response struct {
	Body string
}

// Apply the regular expression and return the list of all sub-matches
// and a list of the positions. The positions are unique, and calculated
// doing an average of the positions of all sub-matches.
func (r *Response) ReList(re *regexp.Regexp) ([][]string, []int) {
	matchs := re.FindAllStringSubmatch(r.Body, -1)
	pos := re.FindAllStringSubmatchIndex(r.Body, -1)

	// Merge positions into a single value (the start one)
	newpos := make([]int, len(pos))
	for i, p := range pos {
		sum := 0
		items := 0
		for _, n := range p {
			sum += n
			items++
		}
		newpos[i] = sum / items
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

		matchs[i] = make([][]string, len(baseMatchs))
		cur := 0
		for j, _ := range matchs[i] {
			// If the next section it's still valid, skip this one with an empty
			// list; otherwise fill it with the matched contents
			if cur >= len(ps) || (j < len(basePos)-1 && basePos[j+1] < ps[cur]) {
				matchs[i][j] = make([]string, results[i].Len)
			} else {
				matchs[i][j] = ms[cur]
				cur++
			}
		}
	}

	return &ResultList{matchs: matchs, cur: -1}, nil
}
