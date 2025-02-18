package main

import (
	"bufio"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	templeRegx = regexp.MustCompile(`\[ error=[^:]+: (.+?\.templ) .+?: line (\d+), col (\d+) ]`)
	goRegx     = regexp.MustCompile(`(.+?):(\d+):(\d+): (.+(?:\n\t.+)?)`)
)

type outputLine struct {
	Content   string
	Codeblock *codeblock
}

type codeLine struct {
	LineNum int
	Content string
}

type codeblock struct {
	LineNum      int
	ColNum       int
	StartLineNum int
	Code         []codeLine
}

// parse text for line/col references and extracts code block for references.
func parse(text string, path string) []outputLine {
	var lines []outputLine

	for _, l := range strings.Split(text, "\n") {
		if l == "# command-line-arguments" || l == "" {
			continue
		}

		lines = append(lines, outputLine{
			Content:   l,
			Codeblock: extractCodeBlock(path, l),
		})
	}

	return lines
}

// extractCodeBlock finds error lines in output.
func extractCodeBlock(appPath string, output string) *codeblock {
	templMatches := templeRegx.FindStringSubmatch(output)
	if len(templMatches) > 0 {
		return newCodeBlock(templMatches[1], templMatches[2], templMatches[3], false)
	}

	goMatches := goRegx.FindStringSubmatch(output)
	if len(goMatches) > 0 {
		if fullPath, err := filepath.Abs(filepath.Join(appPath, goMatches[1])); err != nil {
			slog.Warn(err.Error())
			return nil
		} else {
			return newCodeBlock(fullPath, goMatches[2], goMatches[3], true)
		}
	}

	return nil
}

// newCodeBlock creates a new codeblock.
func newCodeBlock(file string, line string, col string, isGoError bool) *codeblock {
	lineNum, err := strconv.Atoi(line)
	if err != nil {
		slog.Warn(err.Error())
		return nil
	}

	colNum, err := strconv.Atoi(col)
	if err != nil {
		slog.Warn(err.Error())
		return nil
	}

	if isGoError {
		// Go errors are actual column number, whereas templ is char index
		// TODO: clean this up
		colNum -= 1
	}

	cb := &codeblock{
		LineNum: lineNum,
		ColNum:  colNum,
	}

	cb.StartLineNum = lineNum - 5
	if cb.StartLineNum < 1 {
		cb.StartLineNum = 1
	}

	code, err := readLinesInRange(file, cb.StartLineNum, lineNum+5)
	if err != nil {
		code = []codeLine{
			{Content: file},
			{Content: err.Error()},
			{Content: "`path` may not be configured correctly"},
		}
	}

	cb.Code = code
	return cb
}

// readLinesInRange reads +/-5 lines of code from the error reference file and line.
func readLinesInRange(filePath string, startLine, endLine int) ([]codeLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []codeLine
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	for scanner.Scan() {
		if lineNumber >= startLine && lineNumber <= endLine {
			lines = append(lines, codeLine{
				LineNum: lineNumber,
				Content: scanner.Text(),
			})
		}
		if lineNumber > endLine {
			break
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
