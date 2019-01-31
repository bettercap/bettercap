package tui

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansi = regexp.MustCompile("\033\\[(?:[0-9]{1,3}(?:;[0-9]{1,3})*)?[m|K]")

func viewLen(s string) int {
	for _, m := range ansi.FindAllString(s, -1) {
		s = strings.Replace(s, m, "", -1)
	}
	return utf8.RuneCountInString(s)
}

func maxLen(strings []string) int {
	maxLen := 0
	for _, s := range strings {
		len := viewLen(s)
		if len > maxLen {
			maxLen = len
		}
	}
	return maxLen
}

type alignment int

const (
	alignLeft   = alignment(0)
	alignCenter = alignment(1)
	alignRight  = alignment(2)
)

func getPads(s string, maxLen int, align alignment) (lPad int, rPad int) {
	len := viewLen(s)
	diff := maxLen - len

	if align == alignLeft {
		lPad = 0
		rPad = diff - lPad + 1
	} else if align == alignCenter {
		lPad = diff / 2
		rPad = diff - lPad + 1
	} /* else {
		TODO
	} */

	return
}

func padded(s string, maxLen int, align alignment) string {
	lPad, rPad := getPads(s, maxLen, align)
	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", lPad), s, strings.Repeat(" ", rPad))
}

// Table accepts a slice of column labels and a 2d slice of rows
// and prints on the writer an ASCII based datagrid of such
// data.
func Table(w io.Writer, columns []string, rows [][]string) {
	for i, col := range columns {
		columns[i] = fmt.Sprintf(" %s", col)
	}

	for i, row := range rows {
		for j, cell := range row {
			rows[i][j] = fmt.Sprintf(" %s", cell)
		}
	}

	colPaddings := make([]int, 0)
	lineSep := ""
	for colIndex, colHeader := range columns {
		column := []string{colHeader}
		for _, row := range rows {
			column = append(column, row[colIndex])
		}
		mLen := maxLen(column)
		colPaddings = append(colPaddings, mLen)
		lineSep += fmt.Sprintf("+%s", strings.Repeat("-", mLen+1))
	}
	lineSep += "+"

	table := ""

	// header
	table += fmt.Sprintf("%s\n", lineSep)
	for colIndex, colHeader := range columns {
		table += fmt.Sprintf("|%s", padded(colHeader, colPaddings[colIndex], alignCenter))
	}
	table += fmt.Sprintf("|\n")
	table += fmt.Sprintf("%s\n", lineSep)

	// rows
	for _, row := range rows {
		for colIndex, cell := range row {
			table += fmt.Sprintf("|%s", padded(cell, colPaddings[colIndex], alignLeft))
		}
		table += fmt.Sprintf("|\n")
	}

	// footer
	table += lineSep

	fmt.Fprintf(w, "\n%s\n", table)
}
