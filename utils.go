package requester

func AssertLen(lst [][]string, n int) {
	for _, x := range lst {
		if len(x) != n {
			panic("lengths doesn't match")
		}
	}
}
