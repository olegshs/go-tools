// Package sanitize removes unwanted HTML elements and attributes from
// user-supplied HTML.
package sanitize

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	// htmlBodyRegexp extracts the body content from the rendered document.
	htmlBodyRegexp = regexp.MustCompile(`(?is)<body(\s+.*?)?>\s*(.*)\s*</body(\s+.*?)?>`)
	// allowTagRegexp parses allowed tag specs such as "a[href,title]".
	allowTagRegexp = regexp.MustCompile(`(\w+)(\[((\w+,?)*)\])?`)
	// jsHrefRegexp matches javascript: URLs.
	jsHrefRegexp = regexp.MustCompile(`(?is)^\s*javascript\s*:`)
)

// HtmlOptions controls the handling of javascript: links. Unless
// JsHrefAllow is set, an <a href="javascript:..."> attribute is replaced
// with JsHrefReplace, or removed if JsHrefReplace is empty.
type HtmlOptions struct {
	JsHrefAllow   bool
	JsHrefReplace string
}

// Html parses s as an HTML fragment and keeps only the allowed elements and
// attributes. Each entry of allowTags is a tag name with an optional list
// of allowed attributes, such as "b", "p" or "a[href,title]".
//
// Elements that are not allowed are unwrapped: the tags are removed but
// their content is kept, so the text inside <script> or <style> remains as
// escaped text. Only the href of <a> is checked for javascript: URLs, so do
// not allow other URL attributes, such as src, for untrusted input.
// It returns "" if s cannot be parsed.
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

// htmlNode filters n and its descendants. Disallowed elements are turned
// into document nodes, which html.Render renders as their children only.
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

// htmlNodeAttr removes the attributes of n that are not allowed.
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
