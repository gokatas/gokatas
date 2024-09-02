// Stolen from https://github.com/knbr13/gitcs
package main

import (
	"fmt"
	"strings"
	"time"
)

type Boundary struct {
	Since time.Time
	Until time.Time
}

var sixEmptySpaces = strings.Repeat(" ", 6)

func buildHeader(start, end time.Time) string {
	s := strings.Builder{}
	for current := start; current.Before(end) || current.Equal(end); current = current.AddDate(0, 1, 0) {
		s.WriteString(fmt.Sprintf("%-16s", current.Month().String()[:3]))
	}
	return s.String()
}

func daysAgo(t time.Time) int {
	return int(time.Since(t).Round(time.Hour*24).Hours() / 24)
}

func getDay(i int) string {
	switch i {
	case 1:
		return "Mon"
	case 3:
		return "Wed"
	case 5:
		return "Fri"
	}
	return strings.Repeat(" ", 3)
}

func parseKatasForReport(katas Katas) map[int]int {
	doneKatas := make(map[int]int)
	for _, kata := range katas {
		for _, done := range kata.done {
			days := daysAgo(done)
			doneKatas[days]++
		}
	}
	return doneKatas
}

func printReport(katas Katas, b Boundary) {
	doneKatas := parseKatasForReport(katas)

	for b.Since.Weekday() != time.Sunday {
		b.Since = b.Since.AddDate(0, 0, -1)
	}
	for b.Until.Weekday() != time.Saturday {
		b.Until = b.Until.AddDate(0, 0, 1)
	}

	fmt.Printf("%s     %s\n", sixEmptySpaces, buildHeader(b.Since, b.Until))

	s := strings.Builder{}
	s1 := b.Since

	for i := 0; i < 7; i++ {
		s.WriteString(fmt.Sprintf("%-5s", getDay(i)))
		sn2 := s1
		for !sn2.After(b.Until) {
			d := daysAgo(sn2)
			s.WriteString(printCell(doneKatas[d]))
			sn2 = sn2.AddDate(0, 0, 7)
		}
		s1 = s1.AddDate(0, 0, 1)
		fmt.Println(s.String())
		s.Reset()
	}
}

func printCell(val int) string {
	if val == 0 {
		return "  - "
	}
	return fmt.Sprintf(" %2d ", val)
}
