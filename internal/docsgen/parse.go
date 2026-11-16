package docsgen

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	prettyprintRe = regexp.MustCompile(`(?i)^(GET|POST|PUT|PATCH|DELETE)\s+(/\S+)$`)
	pathParamRe   = regexp.MustCompile(`\{([^}]+)\}`)
	defaultRe     = regexp.MustCompile(`(?i)Default:\s*([^\n]*?)(?:\s+Available:|$)`)
	availableRe   = regexp.MustCompile(`(?i)Available:\s*([^\n]+)`)
	arrayTypeRe   = regexp.MustCompile(`(?i)^Array\[(.+)\]$`)
)

// ParseHTML scrapes Affise Slate docs HTML into operations.
func ParseHTML(raw string) ([]Operation, error) {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var ops []Operation
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "code" && hasClass(n, "prettyprint") {
			text := strings.TrimSpace(textContent(n))
			m := prettyprintRe.FindStringSubmatch(text)
			if m != nil {
				op := Operation{
					Method:  strings.ToUpper(m[1]),
					Path:    m[2],
					Summary: findPrecedingSummary(n),
				}
				op.Parameters = parseParamTable(findFollowingTable(n))
				op.Parameters = mergePathParams(op.Path, op.Parameters)
				ops = append(ops, op)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if len(ops) == 0 {
		return nil, fmt.Errorf("no Affise operations found (prettyprint METHOD /path)")
	}
	return ops, nil
}

// FilterPaths keeps operations matching the allowlist.
// Empty allowlist keeps all.
// Entries ending with "/" are prefix matches; others are exact path matches.
func FilterPaths(ops []Operation, allowlist []string) []Operation {
	if len(allowlist) == 0 {
		return ops
	}
	var out []Operation
	for _, op := range ops {
		for _, p := range allowlist {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if strings.HasSuffix(p, "/") {
				if strings.HasPrefix(op.Path, p) {
					out = append(out, op)
					break
				}
				continue
			}
			if op.Path == p {
				out = append(out, op)
				break
			}
		}
	}
	return out
}

func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key != "class" {
			continue
		}
		for _, c := range strings.Fields(a.Val) {
			if c == class {
				return true
			}
		}
	}
	return false
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func findPrecedingSummary(code *html.Node) string {
	// Prefer previous sibling <p> text; else walk up looking for prior p/h1/h2.
	for sib := previousElement(code); sib != nil; sib = previousElement(sib) {
		if sib.Data == "p" || sib.Data == "h1" || sib.Data == "h2" || sib.Data == "h3" {
			s := strings.TrimSpace(textContent(sib))
			if s != "" && !strings.HasPrefix(s, "curl") {
				return s
			}
		}
		if sib.Data == "code" && hasClass(sib, "prettyprint") {
			break
		}
	}
	parent := code.Parent
	for parent != nil {
		for sib := previousElement(parent); sib != nil; sib = previousElement(sib) {
			if sib.Data == "p" || sib.Data == "h1" || sib.Data == "h2" || sib.Data == "h3" {
				s := strings.TrimSpace(textContent(sib))
				if s != "" {
					return s
				}
			}
		}
		parent = parent.Parent
	}
	return ""
}

func previousElement(n *html.Node) *html.Node {
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.Type == html.ElementNode {
			return s
		}
	}
	return nil
}

func nextElement(n *html.Node) *html.Node {
	for s := n.NextSibling; s != nil; s = s.NextSibling {
		if s.Type == html.ElementNode {
			return s
		}
	}
	return nil
}

func findFollowingTable(code *html.Node) *html.Node {
	// Table often follows as sibling of parent <p>, or next elements.
	start := code.Parent
	if start == nil {
		start = code
	}
	for n := nextElement(start); n != nil; n = nextElement(n) {
		if n.Data == "table" {
			return n
		}
		if n.Data == "code" && hasClass(n, "prettyprint") {
			return nil
		}
		if n.Data == "h1" || n.Data == "h2" {
			return nil
		}
		// sometimes wrapped
		if t := findDescendantTable(n); t != nil {
			return t
		}
	}
	// also check siblings after the code itself inside parent
	for n := nextElement(code); n != nil; n = nextElement(n) {
		if n.Data == "table" {
			return n
		}
	}
	return nil
}

func findDescendantTable(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "table" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := findDescendantTable(c); t != nil {
			return t
		}
	}
	return nil
}

func parseParamTable(table *html.Node) []Parameter {
	if table == nil {
		return nil
	}
	var rows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cells = append(cells, strings.TrimSpace(textContent(c)))
				}
			}
			if len(cells) >= 2 {
				rows = append(rows, cells)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)

	var params []Parameter
	for i, row := range rows {
		if i == 0 && strings.EqualFold(row[0], "Parameter") {
			continue
		}
		name := strings.TrimSpace(row[0])
		if name == "" {
			continue
		}
		p := Parameter{
			Name: name,
			In:   "query",
			Type: "String",
		}
		if len(row) > 1 {
			p.Type = strings.TrimSpace(row[1])
		}
		availDefault := ""
		if len(row) > 2 {
			availDefault = row[2]
		}
		if len(row) > 3 {
			p.Description = row[3]
		} else if len(row) == 3 && !strings.Contains(strings.ToLower(row[2]), "default") && !strings.Contains(strings.ToLower(row[2]), "available") {
			// some tables might be Parameter/Type/Description
			p.Description = row[2]
			availDefault = ""
		}
		if m := defaultRe.FindStringSubmatch(availDefault); m != nil {
			p.Default = strings.TrimSpace(m[1])
		}
		if m := availableRe.FindStringSubmatch(availDefault); m != nil {
			p.Enum = splitCSV(m[1])
		}
		params = append(params, p)
	}
	return params
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".")
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func mergePathParams(path string, params []Parameter) []Parameter {
	seen := map[string]bool{}
	for i := range params {
		seen[strings.ToLower(params[i].Name)] = true
	}
	matches := pathParamRe.FindAllStringSubmatch(path, -1)
	for _, m := range matches {
		name := m[1]
		key := strings.ToLower(name)
		// Docs often list path params as query "id"; promote matching names to path.
		promoted := false
		for i := range params {
			if strings.EqualFold(params[i].Name, name) || (strings.EqualFold(name, "ID") && strings.EqualFold(params[i].Name, "id")) {
				params[i].In = "path"
				params[i].Required = true
				params[i].Name = name // keep path spelling
				promoted = true
				seen[key] = true
				break
			}
		}
		if !promoted && !seen[key] {
			params = append(params, Parameter{
				Name:     name,
				In:       "path",
				Type:     "String",
				Required: true,
			})
		}
	}
	return params
}

// MapAffiseType converts Affise docs type strings to OpenAPI type/format/items.
func MapAffiseType(affiseType string) (typ, format string, itemsType string) {
	t := strings.TrimSpace(affiseType)
	if m := arrayTypeRe.FindStringSubmatch(t); m != nil {
		inner, _, _ := MapAffiseType(m[1])
		return "array", "", inner
	}
	switch strings.ToLower(t) {
	case "integer", "int":
		return "integer", "int64", ""
	case "number", "float", "double":
		return "number", "double", ""
	case "boolean", "bool":
		return "boolean", "", ""
	default:
		return "string", "", ""
	}
}
