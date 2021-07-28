package bot

import (
	"errors"
	"fmt"
	"github.com/jalaali/go-jalaali"
	"github.com/psyg1k/remindertelbot/internal"
	"regexp"
	"strconv"
	"time"
)

func StringHasSubmatch(s, sub string) bool {
	if res := regexp.MustCompile(fmt.Sprintf(`\s%s\s`, sub)).FindStringSubmatch(s); len(res) == 0 {
		return false
	}
	return true
}

func GetParam(text string, flags ...string) (string, error) {
	for _, s := range flags {
		reg := fmt.Sprintf(`\s%s\s*(.*?)(\z|\s)`, s)
		fmt.Println(reg)
		re := regexp.MustCompile(reg)
		match := re.FindStringSubmatch(text)
		if len(match) > 2 {
			return match[len(match)-2], nil
		}
	}
	return "", errors.New("string not matched")

}

func getRegexGroups(compRegEx *regexp.Regexp, url string) (paramsMap map[string]string) {
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func matchRegexAlt(s string, args ...*regexp.Regexp) (paramsMap map[string]string, err error) {
	for _, r := range args {
		paramsMap = getRegexGroups(r, s)
		if len(paramsMap) > 0 {
			return paramsMap, nil
		}
	}
	return paramsMap, errors.New("regexes not matched")
}

func formatDateTime(timeString string, date *time.Time, offset internal.Offset) error {
	var sec, min, hour int
	var round = time.Minute

	hhMmSs := regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2}):(?P<sec>\d{2})`)
	hhMm := regexp.MustCompile(`(?P<hour>\d{2}):(?P<min>\d{2})`)

	hm, err := matchRegexAlt(timeString, hhMmSs, hhMm)
	if err != nil {
		return InvalidTimeFormat
	}

	hour, _ = strconv.Atoi(hm["hour"])
	min, _ = strconv.Atoi(hm["min"])
	sec, _ = strconv.Atoi(hm["sec"])
	if sec > 0 {
		round = time.Second
	}

	*date = time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, 0, nil).Round(round).Add(-1 * time.Duration(offset))
	return nil
}

func stringToJalaliDate(dateString string) (time.Time, error) {
	year, month, day, err := stringToParams(dateString)
	if err != nil {
		return time.Time{}, nil
	}
	year, month, day, err = jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, month, day, 0, 0, 0, 0, nil), nil
}

func stringToDate(dateString string) (time.Time, error) {
	year, month, day, err := stringToParams(dateString)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, month, day, 0, 0, 0, 0, nil), nil
}

func stringToParams(dateString string) (int, time.Month, int, error) {
	var year, day int

	yyyyMmDd := regexp.MustCompile(`(?P<year>\d{4})[-/,](?P<month>\d{2})[-/,](?P<day>\d{2})`)
	ddMmYyyy := regexp.MustCompile(`(?P<day>\d{2})[-/,](?P<month>\d{2})[-/,](?P<year>\d{4})`)

	hm, err := matchRegexAlt(dateString, yyyyMmDd, ddMmYyyy)
	if err != nil {
		return 0, 0, 0, errors.New("invalid date formatting")
	}

	year, _ = strconv.Atoi(hm["year"])
	month, _ := strconv.Atoi(hm["month"])
	day, _ = strconv.Atoi(hm["day"])

	return year, time.Month(month), day, nil
}
