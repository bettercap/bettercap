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

func lineSeparator(num int, columns []string, rows [][]string) string {
	lineSep := ""
	first := ""
	div := ""
	end := ""

	if num == 0 {
		first = "┌"
		div = "┬"
		end = "┐"
	} else if num == 1 {
		first = "├"
		div = "┼"
		end = "┤"
	} else if num == 2 {
		first = "└"
		div = "┴"
		end = "┘"
	}

	for colIndex, colHeader := range columns {
		column := []string{colHeader}
		for _, row := range rows {
			column = append(column, row[colIndex])
		}
		mLen := maxLen(column)
		if colIndex == 0 {
			lineSep += fmt.Sprintf(first+"%s", strings.Repeat("─", mLen+1))
		} else {
			lineSep += fmt.Sprintf(div+"%s", strings.Repeat("─", mLen+1))
		}
	}
	lineSep += end

	return lineSep
}

// Table accepts a slice of column labels and a 2d slice of rows
// and prints on the writer an ASCII based datagrid of such
// data.
func Table(w io.Writer, columns []string, rows [][]string) {
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = fmt.Sprintf(" %s", col)
	}

	cells := make([][]string, len(rows))
	for i, row := range rows {
		cells[i] = make([]string, len(row))
		for j, cell := range row {
			cells[i][j] = fmt.Sprintf(" %s", cell)
		}
	}

	colPaddings := make([]int, 0)
	for colIndex, colHeader := range headers {
		column := []string{colHeader}
		for _, row := range cells {
			column = append(column, row[colIndex])
		}
		mLen := maxLen(column)
		colPaddings = append(colPaddings, mLen)
	}

	table := "\n"

	// header
	table += fmt.Sprintf("%s\n", lineSeparator(0, headers, cells))
	for colIndex, colHeader := range headers {
		table += fmt.Sprintf("│%s", padded(colHeader, colPaddings[colIndex], alignCenter))
	}
	table += fmt.Sprintf("│\n")

	table += fmt.Sprintf("%s\n", lineSeparator(1, headers, cells))

	// rows
	for _, row := range cells {
		for colIndex, cell := range row {
			table += fmt.Sprintf("│%s", padded(cell, colPaddings[colIndex], alignLeft))
		}
		table += fmt.Sprintf("│\n")
	}

	// footer
	table += fmt.Sprintf("%s\n", lineSeparator(2, headers, cells))

	fmt.Fprintf(w, "%s", table)
}
