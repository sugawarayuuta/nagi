package semver

import (
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Sort(vers [][]string) [][]string {
	sort.Slice(vers, func(left, right int) bool {
		lTime, _ := time.Parse(time.RFC3339, vers[left][1])
		rTime, _ := time.Parse(time.RFC3339, vers[right][1])
		return lTime.After(rTime)
	})
	return vers
}

func Split(req string) (string, string) {
	if mat, _ := path.Match("@*/*@*", req); mat {
		last := strings.LastIndex(req, "@")
		return req[:last], req[last+1:]
	}
	if mat, _ := path.Match("*@*", req); mat {
		spl := strings.SplitN(req, "@", 2)
		return spl[0], spl[1]
	}
	return req, "latest"
}

func Version(req string) []int {
	spl := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(req), "v"), ".", 3)
	ver := make([]int, 3)
	ver[0], _ = strconv.Atoi(spl[0])
	ver[1], _ = strconv.Atoi(spl[1])
	var err error
	if ver[2], err = strconv.Atoi(spl[2]); err == nil {
		return ver
	}
	if index := strings.Index(spl[2], "+"); index != -1 {
		ver[2], _ = strconv.Atoi(spl[2][:index])
	}
	if index := strings.Index(spl[2], "-"); index != -1 {
		ver[2], _ = strconv.Atoi(spl[2][:index])
	}
	return ver
}

func norm(req string) []int {
	spl := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(req), "v"), ".", 3)
	ver := make([]int, 3)

	if strings.Contains("xX*", spl[0]) {
		ver[0], ver[1], ver[2] = -1, -1, -1
		return ver
	} else if len(spl) == 1 || strings.Contains("xX*", spl[1]) {
		ver[0], _ = strconv.Atoi(spl[0])
		ver[1], ver[2] = -1, -1
		return ver
	} else if len(spl) == 2 || strings.Contains("xX*", spl[2]) {
		ver[0], _ = strconv.Atoi(spl[0])
		ver[1], _ = strconv.Atoi(spl[1])
		ver[2] = -1
		return ver
	}
	return Version(req)
}