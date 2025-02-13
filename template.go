package main

import (
	"html/template"
	"regexp"
	"strings"
)

const tmplHtml = `
<html>
<head>
<style>
	body { background-color: #2d2d2d; color: #ccc; padding: 20px; }
	body, pre { font-family: Consolas, Monaco, 'Andale Mono', 'Ubuntu Mono', monospace; }
	pre { font-size: 16px; line-height: 1.6; white-space: pre-wrap; tab-size: 3; }
	h1 { font-size: 20px; }
	.info { color: #f08d49; }
	.quote, .quote span { color: #67cdcc; }
	.warning { color: #d4af37; }
	.error { color: #e2777a; }
	.success { color: #4CAF50; }
	.code { color: #e83e8c; border: 1px solid #e83e8c; font-size: 85%; padding: 3px; border-radius: 6px; }
</style>
</head>
<body>
<h1>{{.err}}</h1>
<pre>{{.output | highlight}}</pre>
</body>
</html>
`

var (
	tmpl            = template.Must(template.New("").Funcs(template.FuncMap{"highlight": highlight}).Parse(tmplHtml))
	htmlRegex       = regexp.MustCompile(`<[^\s>]+>`)
	numberRegex     = regexp.MustCompile(`(\b\d+)`)
	quoteRegex      = regexp.MustCompile(`"[^"]*"|'[^']*'`)
	unexpectedRegex = regexp.MustCompile(`unexpected ([^\s,]+)`)
	expectedRegex   = regexp.MustCompile(`\bexpected ([^\s,]+)`)
	missingRegex    = regexp.MustCompile(`missing|unbalanced ([^\s:]+)`)
	typeRegex       = regexp.MustCompile(`\(value of type (\S+)\) as (\S+)`)
	errorRegex      = regexp.MustCompile(`(?i)((?:syntax )?errors?|failed|fail|undefined|wrong|invalid|closed|nil|unexported|✗)`)
	infoRegex       = regexp.MustCompile(`untyped|types?|const(?:ant)?|return|select|struct|range|implement`)
	highlightRegex  = regexp.MustCompile(`\\x00;(quote|error|info|warning|success|code)`)
)

func highlight(text string) template.HTML {
	text = replace(text, "warning", htmlRegex)
	text = replace(text, "info", numberRegex)
	text = replace(text, "success", typeRegex)
	text = replace(text, "code", unexpectedRegex)
	text = replace(text, "code", expectedRegex)
	text = replace(text, "code", missingRegex)
	text = replace(text, "quote", quoteRegex)
	text = replace(text, "error", errorRegex)
	text = replace(text, "info", infoRegex)

	text = template.HTMLEscapeString(text)
	text = highlightRegex.ReplaceAllString(text, `<span class="$1">`)
	text = strings.ReplaceAll(text, `\x00;end`, `</span>`)

	return template.HTML(indent(text))
}

func indent(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		indent := ""
		for i, part := range parts {
			if i < len(parts)-1 {
				result = append(result, indent+part+":")
				indent += "\t"
			} else {
				result = append(result, indent+part)
			}
		}

		if len(parts) > 1 {
			result = append(result, "\n")
		}
	}

	return strings.Join(result, "\n")
}

func replace(text string, color string, re *regexp.Regexp) string {
	matches := re.FindAllStringSubmatchIndex(text, -1)
	end := 0
	out := ""

	for _, m := range matches {
		if len(m) > 2 {
			m = m[2:]
		}

		for i := 0; i < len(m); i += 2 {
			ms, me := m[i], m[i+1]
			if ms < 0 {
				continue
			}
			replacement := "\\x00;" + color + text[ms:me] + "\\x00;end"
			out += text[end:ms] + replacement
			end = me
		}
	}

	if end < len(text) {
		out += text[end:]
	}

	return out
}
