package gerritchange

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	shortHostChangeNumRegexp = regexp.MustCompile("^[a-z\\-]+:[0-9]+$")
)

type ChangeList struct {
	ShortHost string
	ChangeNum int64
}

func (cl ChangeList) LongHost() string {
	return fmt.Sprintf("https://%v-review.googlesource.com", cl.ShortHost)
}

func (cl ChangeList) String() string {
	return fmt.Sprintf("%v:%v", cl.ShortHost, cl.ChangeNum)
}

// ParseChangeListsSorted takes a string of the form
// shorthost:num,shorthost:num,...
// and converts it into a slice of ChangeLists.
func ParseChangeListsSorted(cls string) ([]ChangeList, error) {
	var result []ChangeList
	for _, s := range strings.Split(cls, ",") {
		if !shortHostChangeNumRegexp.MatchString(s) {
			return nil, fmt.Errorf("got unexpected cl string: %v. Should look something like chromium:123", s)
		}
		subsplit := strings.Split(s, ":")
		changeNum, err := strconv.ParseInt(subsplit[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ChangeNum %v", subsplit[1])
		}
		result = append(result, ChangeList{ShortHost: subsplit[0], ChangeNum: changeNum})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ShortHost < result[j].ShortHost {
			return true
		}
		return result[i].ChangeNum < result[j].ChangeNum
	})
	return result, nil
}

func ChangeListsString(cls []ChangeList) string {
	var s []string
	for _, cl := range cls {
		s = append(s, cl.String())
	}
	return strings.Join(s, ",")
}
