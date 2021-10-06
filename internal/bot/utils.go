package bot

import (
	"errors"
	"fmt"
	"github.com/jalaali/go-jalaali"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func StringHasSubmatch(s, sub string) bool {
	if res := regexp.MustCompile(fmt.Sprintf(`\s%s\s`, sub)).FindStringSubmatch(s); len(res) >= 1 {
		return true
	}
	return false
}

func GetParam(text string, flags ...string) (string, error) {
	for _, s := range flags {
		reg := fmt.Sprintf(`\s%s\s*(.*?)(\z|\s)`, s)
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

func formatDateTime(timeString string, date *time.Time, loc *time.Location) error {
	var sec, min, hour int
	var round = time.Minute

	hhMmSs := regexp.MustCompile(`(?P<hour>\d{1,2}):(?P<min>\d{1,2}):(?P<sec>\d{1,2})`)
	hhMm := regexp.MustCompile(`(?P<hour>\d{1,2}):(?P<min>\d{1,2})`)

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

	*date = time.Date(date.Year(), date.Month(), date.Day(), hour, min, sec, 0, loc).Round(round).UTC()
	return nil
}

func stringToJalaliDate(dateString string, loc *time.Location) (time.Time, error) {
	year, month, day, err := stringToParams(dateString)
	if err != nil {
		return time.Time{}, nil
	}
	year, month, day, err = jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
}

func stringToDate(dateString string, loc *time.Location) (time.Time, error) {
	year, month, day, err := stringToParams(dateString)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, month, day, 0, 0, 0, 0, loc), nil
}

func stringToParams(dateString string) (int, time.Month, int, error) {
	var year, day int

	yyyyMmDd := regexp.MustCompile(`(?P<year>\d{4})[-/,](?P<month>\d{1,2})[-/,](?P<day>\d{1,2})`)
	ddMmYyyy := regexp.MustCompile(`(?P<day>\d{1,2})[-/,](?P<month>\d{1,2})[-/,](?P<year>\d{4})`)

	hm, err := matchRegexAlt(dateString, yyyyMmDd, ddMmYyyy)
	if err != nil {
		return 0, 0, 0, errors.New("invalid date formatting")
	}

	year, _ = strconv.Atoi(hm["year"])
	month, _ := strconv.Atoi(hm["month"])
	day, _ = strconv.Atoi(hm["day"])

	return year, time.Month(month), day, nil
}

func DurationToString(d time.Duration) (string, bool) {
	var message string

	total := int(d.Seconds())
	days := total / (60 * 60 * 24)
	hours := total / (60 * 60) % 24
	minutes := total / 60 % 60
	seconds := total % 60

	switch {
	case days > 0:
		message = fmt.Sprintf("%s%dd ", message, days)
		fallthrough
	case hours > 0:
		message = fmt.Sprintf("%s%dh ", message, hours)
		fallthrough
	case minutes > 0:
		message = fmt.Sprintf("%s%02dm ", message, minutes)
		fallthrough
	case seconds > 0:
		message = fmt.Sprintf("%s%02ds", message, seconds)

	default:
		return "", true

	}

	return message, false
}

func selectEmoji(d time.Duration) rune {
	var emoji rune

	switch {
	case d.Seconds() > (72 * time.Hour).Seconds():
		emoji = '💠'
	case d.Seconds() > (48 * time.Hour).Seconds():
		emoji = '❇'
	case d.Seconds() > (24 * time.Hour).Seconds():
		emoji = '⚠'
	case d.Seconds() > (8 * time.Hour).Seconds():
		emoji = '🛑'
	default:
		emoji = '🆘'
	}

	return emoji
}

func ToPersianDigits(text string) string {
	return strings.NewReplacer(
		"0", "۰",
		"1", "۱",
		"2", "۲",
		"3", "۳",
		"4", "۴",
		"5", "۵",
		"6", "۶",
		"7", "۷",
		"8", "۸",
		"9", "۹",
	).Replace(text)
}
