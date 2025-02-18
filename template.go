package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

//go:embed template.gohtml
var htmlTemplate string

const (
	openDelimiter         = "{{"
	closeDelimiter        = "}}"
	escapedOpenDelimiter  = "\\{\\{"
	escapedCloseDelimiter = "\\}\\}"
	styleTag              = "style"
	closeTag              = "{{/style}}"
)

var (
	tmpl = template.Must(template.New("").
		Funcs(template.FuncMap{"highlight": highlight, "breakLongLine": breakLongLine, "emphasize": emphasize}).
		Parse(htmlTemplate))
	htmlRegex        = regexp.MustCompile(`<[^\s>]+>`)
	numberRegex      = regexp.MustCompile(`(\b\d+)`)
	quoteRegex       = regexp.MustCompile(`"[^"]*"|'[^']*'`)
	unexpectedRegex  = regexp.MustCompile(`unexpected ([^\s,]+)`)
	expectedRegex    = regexp.MustCompile("\\bexpected (`[^']+`|[^\\s,]+)")
	missingRegex     = regexp.MustCompile(`missing|unbalanced ([^\s:]+)`)
	typeRegex        = regexp.MustCompile(`\(value of type (\S+)\) as (\S+)`)
	errorRegex       = regexp.MustCompile(`(?i)((?:syntax )?errors?|failed|fail|undefined|wrong|invalid|closed|nil|unexported|✗)`)
	infoRegex        = regexp.MustCompile(`untyped|types?|const(?:ant)?|return|select|struct|range|implement`)
	placeholderRegex = regexp.MustCompile("{{style:(.+?)}}")
)

// highlight replaces certain words within text with style tags.
// NOTE: the original text is HTML escaped.
func highlight(text string) template.HTML {
	text = escapeDelimiter(text)

	text = replace(text, "warning", htmlRegex)
	text = replace(text, "info", numberRegex)
	text = replace(text, "success", typeRegex)
	text = replace(text, "code", unexpectedRegex)
	text = replace(text, "code", expectedRegex)
	text = replace(text, "code", missingRegex)
	text = replace(text, "quote", quoteRegex)
	text = replace(text, "error", errorRegex)
	text = replace(text, "info", infoRegex)

	text = styles(text)

	return template.HTML(unescapeDelimiter(text))
}

// emphasize wraps given col with an emphasize style.
func emphasize(col int, text string) template.HTML {
	if col < 0 {
		col = 0
	}

	if col >= len(text) {
		col = len(text)
		text += " "
	}

	text = text[:col] + placeholder("emphasize", text[col:col+1]) + text[col+1:]
	return template.HTML(`<span style="position: relative;">` + styles(text) + `</span>`)
}

// escapeDelimiter escapes style tag delimiters to retain their original text.
func escapeDelimiter(text string) string {
	text = strings.ReplaceAll(text, openDelimiter, escapedOpenDelimiter)
	text = strings.ReplaceAll(text, closeDelimiter, escapedCloseDelimiter)
	return text
}

// unescapeDelimiter replaces original delimiters that were escaped back to their original text.
func unescapeDelimiter(text string) string {
	text = strings.ReplaceAll(text, escapedOpenDelimiter, openDelimiter)
	text = strings.ReplaceAll(text, escapedCloseDelimiter, closeDelimiter)
	return text
}

// styles given text with <span class="style">text</span>.
// Note: The given text is HTML escaped first.
func styles(text string) string {
	text = template.HTMLEscapeString(text)
	text = placeholderRegex.ReplaceAllString(text, `<span class="$1">`)
	text = strings.ReplaceAll(text, closeTag, `</span>`)
	return text
}

// breakLongLine breaks up long log lines into multi-line for easier reading.
func breakLongLine(text string) string {
	if templeRegx.MatchString(text) {
		return strings.Replace(text, ": ", ":\n", 2)
	}
	return text
}

// replace all matches from re with the placeholder color tags in given text.
func replace(text string, color string, re *regexp.Regexp) string {
	var out strings.Builder

	end := 0
	matches := re.FindAllStringSubmatchIndex(text, -1)
	for _, m := range matches {
		if len(m) > 2 {
			m = m[2:]
		}

		for i := 0; i < len(m); i += 2 {
			ms, me := m[i], m[i+1]
			if ms < 0 {
				continue
			}
			replacement := placeholder(color, text[ms:me])
			out.WriteString(text[end:ms])
			out.WriteString(replacement)
			end = me
		}
	}

	if end < len(text) {
		out.WriteString(text[end:])
	}

	return out.String()
}

// placeholder wraps given text in {{style:style}}text{{/style}}.
func placeholder(style, text string) string {
	return fmt.Sprintf("%s%s:%s%s%s%s", openDelimiter, styleTag, style, closeDelimiter, text, closeTag)
}
