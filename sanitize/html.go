// Пакет sanitize предоставляет функцию для удаления нежелательных тегов из HTML.
package sanitize

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	htmlBodyRegexp = regexp.MustCompile(`(?is)<body(\s+.*?)?>\s*(.*)\s*</body(\s+.*?)?>`)
	allowTagRegexp = regexp.MustCompile(`(\w+)(\[((\w+,?)*)\])?`)
	jsHrefRegexp   = regexp.MustCompile(`(?is)^\s*javascript\s*:`)
)

type HtmlOptions struct {
	JsHrefAllow   bool
	JsHrefReplace string
}

func Html(s string, allowTags []string, options HtmlOptions) string {
	b := bytes.NewBufferString(s)

	n, err := html.Parse(b)
	if err != nil {
		return ""
	}

	allowTagsMap := map[string][]string{
		"html": {},
		"body": {},
	}
	for _, allowTag := range allowTags {
		m := allowTagRegexp.FindStringSubmatch(allowTag)
		tag := m[1]
		attrs := strings.Split(m[3], ",")

		allowTagsMap[tag] = attrs
	}

	htmlNode(n, allowTagsMap, &options)

	err = html.Render(b, n)
	if err != nil {
		return ""
	}
	s = b.String()

	m := htmlBodyRegexp.FindStringSubmatch(s)
	if len(m) > 1 {
		s = m[2]
	}

	return s
}

func htmlNode(n *html.Node, allowTags map[string][]string, options *HtmlOptions) {
	if n.Type == html.ElementNode {
		allowAttrs, ok := allowTags[n.Data]
		if ok {
			htmlNodeAttr(n, allowAttrs, options)
		} else {
			n.Type = html.DocumentNode
		}
	}

	var children []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	for _, c := range children {
		htmlNode(c, allowTags, options)
	}
}

func htmlNodeAttr(n *html.Node, allowAttrs []string, options *HtmlOptions) {
	var a []html.Attribute

	for _, attr := range n.Attr {
		if n.Data == "a" {
			if attr.Key == "href" {
				if !options.JsHrefAllow && jsHrefRegexp.MatchString(attr.Val) {
					if len(options.JsHrefReplace) > 0 {
						attr.Val = options.JsHrefReplace
					} else {
						continue
					}
				}
			}
		}

		for _, allowAttr := range allowAttrs {
			if attr.Key == allowAttr {
				a = append(a, attr)
				break
			}
		}
	}

	n.Attr = a
}
