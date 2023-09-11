package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	// read input from console in a long string format
	fmt.Println("Enter a cron parser expression string")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	str := strings.Fields(text)

	// validate and parse string to an expression format
	expression, err := parseStringArray(str)
	if err != nil {
		fmt.Println(err)
		return
	}

	// build response for each field
	res, err := expression.buildResponse()
	if err != nil {
		fmt.Println(err)
		return
	}

	// render table output
	res.renderOutput()
}

func parseStringArray(str []string) (*body, error) {
	if len(str) != 6 {
		return nil, errors.New("invalid expression")
	}
	b := &body{
		minutes:    str[0],
		hour:       str[1],
		dayOfMonth: str[2],
		month:      str[3],
		dayOfWeek:  str[4],
		command:    str[5],
	}
	return b, nil
}

func formatFieldValues(str string, allowedVals []string) int {
	for i, j := range allowedVals {
		if j == strings.ToLower(str) {
			return i + 1
		}
	}
	return -1
}

func (body *body) buildResponse() (*response, error) {
	res := &response{
		command: body.command,
	}
	minutes, minutesErr := prepare(body.minutes, "minute", isValidMinute, 59)
	hour, hourErr := prepare(body.hour, "hour", isValidHour, 23)
	dom, domErr := prepare(body.dayOfMonth, "day of month", isValidDOM, 31)
	month, monthErr := prepare(body.month, "month", isValidMonth, 12)
	dow, dowErr := prepare(body.dayOfWeek, "day of week", isValidDOW, 7)
	err := errors.Join(minutesErr, hourErr, domErr, monthErr, dowErr)
	res.minutes = minutes
	res.hour = hour
	res.dayOfMonth = dom
	res.month = month
	res.dayOfWeek = dow

	if err != nil {
		return nil, err
	}
	return res, nil
}

func prepareHyphenString(str []string, expField string, fn validate) ([]int, error) {
	arr := []int{}
	if expField == "month" {
		s0 := formatFieldValues(str[0], allowedEnglishMonthNames)
		s1 := formatFieldValues(str[1], allowedEnglishMonthNames)
		if s0 != -1 {
			str[0] = strconv.Itoa(s0)
		}
		if s1 != -1 {
			str[1] = strconv.Itoa(s1)
		}
	}
	if expField == "day of week" {
		s0 := formatFieldValues(str[0], allowedEnglishWeekNames)
		s1 := formatFieldValues(str[1], allowedEnglishWeekNames)
		if s0 != -1 {
			str[0] = strconv.Itoa(s0 - 1)
		}
		if s1 != -1 {
			str[1] = strconv.Itoa(s1 - 1)
		}
		// This would allow sunday to sunday
		if (s0 == s1) && (s1 == 1) {
			str[1] = strconv.Itoa(7)
		}
	}
	initial, iErr := strconv.Atoi(str[0])
	last, lErr := strconv.Atoi(str[1])
	if iErr != nil || lErr != nil {
		return []int{}, errors.New("invalid " + expField + " value")
	}
	if !fn(initial) || !fn(last) {
		return []int{}, errors.New("invalid " + expField + " value")
	}
	for j := initial; j <= last; j++ {
		arr = append(arr, j)
	}
	return arr, nil
}

func prepareSlashString(str []string, expField string, fn validate, max int) ([]int, error) {
	arr := []int{}
	if expField == "month" {
		s0 := formatFieldValues(str[0], allowedEnglishMonthNames)
		s1 := formatFieldValues(str[1], allowedEnglishMonthNames)
		if s0 != -1 {
			str[0] = strconv.Itoa(s0)
		}
		if s1 != -1 {
			str[1] = strconv.Itoa(s1)
		}
	}
	if expField == "day of week" {
		s0 := formatFieldValues(str[0], allowedEnglishWeekNames)
		s1 := formatFieldValues(str[1], allowedEnglishWeekNames)
		if s0 != -1 {
			str[0] = strconv.Itoa(s0 - 1)
		}
		if s1 != -1 {
			str[1] = strconv.Itoa(s1 - 1)
		}
	}

	// value is *
	start := 0

	subStringsHyphen := strings.Split(str[0], "-")
	if len(subStringsHyphen) == 2 {
		// value is a range
		vals, err := prepareHyphenString(subStringsHyphen, expField, fn)
		if err != nil {
			return []int{}, errors.New("invalid " + expField + " value")
		}
		start = vals[0]
		max = vals[len(vals)-1]
	} else {
		// value is a single numeric digit
		start, _ = strconv.Atoi(str[0])
	}
	// TODO: error check
	interval, _ := strconv.Atoi(str[1])
	if !fn(start) || !fn(max) {
		return []int{}, errors.New("invalid " + expField + " value")
	}
	for start <= max {
		if !fn(start) {
			return []int{}, errors.New("invalid " + expField + " value")
		}
		arr = append(arr, start)
		start = start + interval
	}
	return arr, nil
}

func prepare(str string, expField string, fn validate, max int) ([]int, error) {
	arr := []int{}
	subStringsComma := strings.Split(str, ",")
	for _, i := range subStringsComma {
		subStringsSlash := strings.Split(i, "/")
		if len(subStringsSlash) == 2 {
			val, err := prepareSlashString(subStringsSlash, expField, fn, max)
			if err != nil {
				return []int{}, err
			}
			arr = append(arr, val...)
		} else {
			subStringsHyphen := strings.Split(i, "-")
			if len(subStringsHyphen) == 2 {
				val, err := prepareHyphenString(subStringsHyphen, expField, fn)
				if err != nil {
					return []int{}, err
				}
				arr = append(arr, val...)
			} else {
				if expField == "month" {
					e1 := formatFieldValues(i, allowedEnglishMonthNames)
					if e1 != -1 {
						i = strconv.Itoa(e1)
					}
				}
				if expField == "day of week" {
					e1 := formatFieldValues(i, allowedEnglishWeekNames)
					if e1 != -1 {
						i = strconv.Itoa(e1 - 1)
					}
				}
				if i == "*" {
					arr = append(arr, addAll(expField)...)
				} else {
					ele, err := strconv.Atoi(i)
					if err != nil {
						return []int{}, errors.New("invalid " + expField + " value")
					}
					if !fn(ele) {
						return []int{}, errors.New("invalid " + expField + " value")
					}
					arr = append(arr, ele)
				}
			}
		}
	}
	return arr, nil
}

func addAll(expField string) []int {
	switch expField {
	case "minute":
		return allValues(0, 59)
	case "hour":
		return allValues(0, 23)
	case "day of month":
		return allValues(1, 31)
	case "month":
		return allValues(1, 12)
	case "day of week":
		return allValues(0, 6)
	}
	return []int{}
}
func allValues(min, max int) []int {
	arr := []int{}
	for i := min; i <= max; i++ {
		arr = append(arr, i)
	}
	return arr
}

func (res *response) renderOutput() {
	t := table.NewWriter()
	t.AppendRow(table.Row{"minute", res.minutes})
	t.AppendRow(table.Row{"hour", res.hour})
	t.AppendRow(table.Row{"day of month", res.dayOfMonth})
	t.AppendRow(table.Row{"month", res.month})
	t.AppendRow(table.Row{"day of week", res.dayOfWeek})
	t.AppendRow(table.Row{"command", res.command})
	fmt.Println(t.Render())
}

func isValidMinute(minute int) bool {
	return (minute >= 0 && minute < 60)
}
func isValidHour(hour int) bool {
	return hour >= 0 && hour < 24
}
func isValidDOM(dom int) bool {
	return dom >= 1 && dom < 32
}
func isValidDOW(dow int) bool {
	return dow >= 0 && dow < 8
}
func isValidMonth(month int) bool {
	return month >= 1 && month <= 12
}

var allowedEnglishMonthNames = []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}
var allowedEnglishWeekNames = []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

type body struct {
	minutes    string
	hour       string
	dayOfMonth string
	month      string
	dayOfWeek  string
	command    string
}

type validate func(int) bool

type response struct {
	minutes    []int
	hour       []int
	dayOfMonth []int
	month      []int
	dayOfWeek  []int
	command    string
}
