package sanitize

import (
	"testing"
)

func TestHtml(t *testing.T) {
	raw := `<h1>Test</h1>` + "\n" +
		`<a href="javascript:alert('Hello!');" style="color:#000;">Link</a>` + "\n" +
		`<script type="text/javascript">alert('Hello!');</script>`

	expected := `Test` + "\n" +
		`<a href="#js">Link</a>` + "\n" +
		`alert(&#39;Hello!&#39;);`

	sanitized := Html(raw, []string{"a[href]"}, HtmlOptions{
		JsHrefAllow:   false,
		JsHrefReplace: "#js",
	})

	if sanitized != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+sanitized)
	}
}
