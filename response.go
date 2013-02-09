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
