package semver

import (
	"math"
	"strings"
)

type Sem struct {
	From struct {
		Ver []int
		Inc bool
	}
	To struct {
		Ver []int
		Inc bool
	}
}

func Build(req string) []*Sem {
	spl := strings.Split(req, "||")
	sems := make([]*Sem, len(spl))
	for index := range spl {
		sems[index] = semver(spl[index])
	}
	return sems
}

func semver(req string) *Sem {
	sem := new(Sem)
	sem.From.Ver, sem.From.Inc = []int{0, 0, 0}, true
	sem.To.Ver, sem.To.Inc = []int{math.MaxInt, 0, 0}, false
	var last int
	var prev byte
	var res []string
	for index := 0; index < len(req); index++ {
		if req[index] == ' ' && strings.IndexByte("<>=~^", prev) == -1 {
			if last < index-1 {
				res = append(res, strings.ReplaceAll(req[last:index], " ", ""))
			}
			last = index + 1
		} else if req[index] != ' ' {
			prev = req[index]
		}
	}
	if last < len(req)-1 {
		res = append(res, strings.ReplaceAll(req[last:], " ", ""))
	}
	for index := range res {
		switch {
		case strings.HasPrefix(res[index], "^"):
			ver := norm(strings.TrimPrefix(res[index], "^"))
			sem.From.Ver, sem.From.Inc = ver, true
			sem.To.Ver, sem.To.Inc = []int{ver[0] + 1, 0, 0}, false
		case strings.HasPrefix(res[index], "~"):
			ver := norm(strings.TrimPrefix(res[index], "~"))
			sem.From.Ver, sem.From.Inc = ver, true
			sem.To.Ver, sem.To.Inc = []int{ver[0], ver[1] + 1, 0}, false
		case strings.HasPrefix(res[index], "="):
			ver := norm(strings.TrimPrefix(res[index], "="))
			sem.From.Ver, sem.From.Inc = ver, true
			sem.To.Ver, sem.To.Inc = ver, true
		case strings.HasPrefix(res[index], ">="):
			ver := norm(strings.TrimPrefix(res[index], ">="))
			sem.From.Ver, sem.From.Inc = ver, true
		case strings.HasPrefix(res[index], "<="):
			ver := norm(strings.TrimPrefix(res[index], "<="))
			sem.To.Ver, sem.To.Inc = ver, true
		case strings.HasPrefix(res[index], ">"):
			ver := norm(strings.TrimPrefix(res[index], ">"))
			sem.From.Ver, sem.From.Inc = ver, false
		case strings.HasPrefix(res[index], "<"):
			ver := norm(strings.TrimPrefix(res[index], "<"))
			sem.To.Ver, sem.To.Inc = ver, false
		default:
			ver := norm(res[index])
			sem.From.Ver, sem.From.Inc = ver, true
			sem.To.Ver, sem.To.Inc = ver, true
		}
	}
	return sem
}

func (sem *Sem) Match(ver []int) bool {
	for index := range sem.From.Ver {
		if (ver[index] == sem.From.Ver[index] || sem.From.Ver[index] == -1) && index != 2 {
			continue
		} else if ver[index] < sem.From.Ver[index] {
			return false
		} else if (ver[index] == sem.From.Ver[index] || sem.From.Ver[index] == -1) && !sem.From.Inc {
			return false
		}
		break
	}
	for index := range sem.To.Ver {
		if (ver[index] == sem.To.Ver[index] || sem.To.Ver[index] == -1) && index != 2 {
			continue
		} else if ver[index] > sem.To.Ver[index] {
			return false
		} else if (ver[index] == sem.To.Ver[index] || sem.To.Ver[index] == -1) && !sem.To.Inc {
			return false
		}
		break
	}
	return true
}
