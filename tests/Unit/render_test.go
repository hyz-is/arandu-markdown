package unit_test

import (
	"runtime"
	"strings"
	"sync"
	"testing"

	"golang.org/x/net/html"

	markdown "github.com/hyz-is/arandu-markdown"
)

// renderer builds the module the way an application does, over the settings
// the tests below read: a layout that owns #content, and one CDN.
func renderer(t testing.TB) *markdown.Module {
	t.Helper()
	m, err := markdown.New(markdown.Config{
		Reserved:     []string{"content"},
		ImageOrigins: []string{"https://cdn.example.com"},
	})
	if err != nil {
		t.Fatalf("building the module: %v", err)
	}
	return m
}

func render(t testing.TB, src string) markdown.Document {
	t.Helper()
	return renderer(t).Render(src)
}

// allowed is every element the renderer may write, with the attributes it may
// write on each. It is spelled out here rather than read from the package, so
// a change to the allowlist is a change somebody made to this test too.
var allowed = map[string]map[string]bool{
	"p": {}, "pre": {}, "blockquote": {}, "ul": {}, "ol": {"start": true},
	"table": {}, "thead": {}, "tbody": {}, "tr": {}, "th": {"align": true}, "td": {"align": true},
	"h1": {"id": true}, "h2": {"id": true}, "h3": {"id": true},
	"h4": {"id": true}, "h5": {"id": true}, "h6": {"id": true},
	"hr": {}, "li": {}, "strong": {}, "em": {}, "del": {}, "br": {},
	"code":       {"class": true},
	"a":          {"href": true, "title": true},
	"img":        {"src": true, "alt": true, "title": true, "loading": true, "decoding": true},
	"input":      {"type": true, "disabled": true, "checked": true},
	"figure":     {},
	"figcaption": {},
}

// markupOf is every tag in the output, read by the same tokenizer a browser's
// parser starts from.
func markupOf(out string) []html.Token {
	var tags []html.Token
	z := html.NewTokenizer(strings.NewReader(out))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return tags
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken, html.CommentToken, html.DoctypeToken:
			tags = append(tags, z.Token())
		}
	}
}

// onlyAllowlisted fails for any element or attribute the renderer does not
// write, and for any destination that is not one a body may point at.
func onlyAllowlisted(t testing.TB, src string, doc markdown.Document) {
	t.Helper()
	for _, tag := range markupOf(string(doc.HTML())) {
		if tag.Type == html.CommentToken || tag.Type == html.DoctypeToken {
			t.Fatalf("source %q wrote a %v into the page", src, tag.Type)
		}
		attrs, ok := allowed[tag.Data]
		if !ok {
			t.Fatalf("source %q wrote <%s>, which is not on the allowlist:\n%s", src, tag.Data, doc.HTML())
		}
		for _, a := range tag.Attr {
			if !attrs[a.Key] {
				t.Fatalf("source %q wrote %s on <%s>:\n%s", src, a.Key, tag.Data, doc.HTML())
			}
			if a.Key == "href" || a.Key == "src" {
				lower := strings.ToLower(strings.TrimSpace(a.Val))
				for _, scheme := range []string{"javascript:", "vbscript:", "data:", "file:"} {
					if strings.HasPrefix(lower, scheme) {
						t.Fatalf("source %q kept %s=%q", src, a.Key, a.Val)
					}
				}
			}
		}
	}
}

func TestRawHTMLIsShownAsTheTextItWas(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		"<script>alert(1)</script>",
		"<iframe src=\"https://example.com\"></iframe>",
		"<div onclick=\"steal()\">hello</div>",
		"<style>body{display:none}</style>",
		"<svg><script>alert(1)</script></svg>",
		"Some <span style=\"color:red\">inline</span> HTML.",
		"<img src=x onerror=alert(1)>",
		"<a href=\"javascript:alert(1)\">x</a>",
		"<form action=\"/steal\"><input type=\"password\"></form>",
		"<plaintext>everything after",
	} {
		doc := render(t, src)
		onlyAllowlisted(t, src, doc)
	}

	doc := render(t, "<div onclick=\"steal()\">hello</div>")
	if !strings.Contains(string(doc.HTML()), "&lt;div onclick=&#34;steal()&#34;&gt;hello&lt;/div&gt;") {
		t.Fatalf("a div was not shown as its own text: %s", doc.HTML())
	}
}

func TestLinksKeepOnlyDestinationsABodyMayPointAt(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		dest string
		kept bool
	}{
		{"https://example.com/a?b=c#d", true},
		{"http://example.com", true},
		{"/blog/first-post", true},
		{"#a-section", true},
		{"mailto:someone@example.com", true},
		{"mailto:someone@example.com?subject=Hi", true},
		{"javascript:alert(1)", false},
		{"JaVaScRiPt:alert(1)", false},
		{"&#106;avascript:alert(1)", false},
		{"vbscript:msgbox(1)", false},
		{"data:text/html,<script>alert(1)</script>", false},
		{"//evil.example", false},
		{"/\\evil.example", false},
		{"https://user@evil.example", false},
		{"relative/path", false},
		{"../up", false},
		{"/a/../../etc", false},
		{"#not an id", false},
		{"mailto:javascript:alert(1)", false},
		{"ftp://example.com", false},
	} {
		src := "[words](<" + c.dest + ">)"
		doc := render(t, src)
		onlyAllowlisted(t, src, doc)
		linked := strings.Contains(string(doc.HTML()), "<a ")
		if linked != c.kept {
			t.Errorf("%q: linked = %v, want %v\n%s", c.dest, linked, c.kept, doc.HTML())
		}
		if !strings.Contains(string(doc.HTML()), "words") {
			t.Errorf("%q: the link's words were lost: %s", c.dest, doc.HTML())
		}
	}
}

func TestALinkInsideALinkKeepsItsWordsAndNotItsLink(t *testing.T) {
	t.Parallel()

	doc := render(t, "[outer [inner](/in) text](/out)")
	if got := strings.Count(string(doc.HTML()), "<a "); got != 1 {
		t.Fatalf("links = %d, want 1: %s", got, doc.HTML())
	}
	if !strings.Contains(string(doc.HTML()), "inner") {
		t.Fatalf("the inner link's words were lost: %s", doc.HTML())
	}
}

func TestImagesLoadOnlyFromThisSiteAndTheNamedOrigins(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		src  string
		kept bool
	}{
		{"/media/cover.png", true},
		{"https://cdn.example.com/prod/cover.png", true},
		{"https://CDN.example.com/prod/cover.png", true},
		{"https://elsewhere.example/cover.png", false},
		{"http://cdn.example.com/cover.png", false},
		{"https://user@cdn.example.com/cover.png", false},
		{"//cdn.example.com/cover.png", false},
		{"/../secret.png", false},
		{"data:image/png;base64,AAAA", false},
		{"javascript:alert(1)", false},
	} {
		src := "Look: ![a cover](" + c.src + ") here."
		doc := render(t, src)
		onlyAllowlisted(t, src, doc)
		if kept := strings.Contains(string(doc.HTML()), "<img "); kept != c.kept {
			t.Errorf("%q: kept = %v, want %v\n%s", c.src, kept, c.kept, doc.HTML())
		}
		if !strings.Contains(string(doc.HTML()), "a cover") {
			t.Errorf("%q: the description was lost: %s", c.src, doc.HTML())
		}
	}

	doc := render(t, "![a cover](/media/cover.png)")
	for _, want := range []string{`loading="lazy"`, `decoding="async"`} {
		if !strings.Contains(string(doc.HTML()), want) {
			t.Errorf("an image is written without %s: %s", want, doc.HTML())
		}
	}
}

func TestAParagraphHoldingOnlyAnImageIsAFigure(t *testing.T) {
	t.Parallel()

	doc := render(t, "![The panel](/media/panel.png \"The panel, in May\")")
	want := `<figure><img src="/media/panel.png" alt="The panel" loading="lazy" decoding="async"><figcaption>The panel, in May</figcaption></figure>`
	if !strings.Contains(string(doc.HTML()), want) {
		t.Fatalf("got %s\nwant %s", doc.HTML(), want)
	}

	inline := render(t, "Before ![icon](/media/icon.png) after.")
	if strings.Contains(string(inline.HTML()), "<figure>") {
		t.Fatalf("an image inside a sentence became a figure: %s", inline.HTML())
	}
	two := render(t, "![a](/media/a.png) ![b](/media/b.png)")
	if strings.Contains(string(two.HTML()), "<figure>") {
		t.Fatalf("two images became one figure: %s", two.HTML())
	}
}

func TestHeadingsCarryUniqueAddressesAndTheTopTwoLevelsAreListed(t *testing.T) {
	t.Parallel()

	doc := render(t, strings.Join([]string{
		"# Content",
		"## Visão *geral*",
		"### Details",
		"#### Too deep for the contents",
		"## Visão geral",
		"> ## A heading in a quote",
		"##",
	}, "\n\n"))

	want := []markdown.Heading{
		{Level: 2, ID: "content-2", Text: "Content"},
		{Level: 2, ID: "visao-geral", Text: "Visão geral"},
		{Level: 3, ID: "details", Text: "Details"},
		{Level: 2, ID: "visao-geral-2", Text: "Visão geral"},
	}
	if len(doc.Headings) != len(want) {
		t.Fatalf("headings = %+v, want %+v", doc.Headings, want)
	}
	for i := range want {
		if doc.Headings[i] != want[i] {
			t.Errorf("heading %d = %+v, want %+v", i, doc.Headings[i], want[i])
		}
	}
	for _, want := range []string{
		`<h2 id="content-2">Content</h2>`,
		`<h2 id="visao-geral">Visão <em>geral</em></h2>`,
		`<h4 id="too-deep-for-the-contents">`,
		`<h2>A heading in a quote</h2>`,
	} {
		if !strings.Contains(string(doc.HTML()), want) {
			t.Errorf("missing %s in\n%s", want, doc.HTML())
		}
	}
	if strings.Contains(string(doc.HTML()), "<h2></h2>") || strings.Contains(string(doc.HTML()), `id="section"`) {
		t.Errorf("an empty heading was written: %s", doc.HTML())
	}
}

func TestTopHeadingLowersAndNeverShifts(t *testing.T) {
	t.Parallel()

	m, err := markdown.New(markdown.Config{TopHeading: 3})
	if err != nil {
		t.Fatal(err)
	}
	doc := m.Render("# One\n\n## Two\n\n### Three\n\n#### Four")
	for _, want := range []string{`<h3 id="one">`, `<h3 id="two">`, `<h3 id="three">`, `<h4 id="four">`} {
		if !strings.Contains(string(doc.HTML()), want) {
			t.Errorf("missing %s in\n%s", want, doc.HTML())
		}
	}
	if len(doc.Headings) != 4 {
		t.Errorf("headings = %+v, want the four at levels 3 and 4", doc.Headings)
	}
}

func TestTextIsWhatTheBodySaysOneLinePerBlock(t *testing.T) {
	t.Parallel()

	doc := render(t, "## A title\n\nSome **bold** words and a [link](/x).\n\n- one\n- two\n\n```go\nfmt.Println(1 < 2)\n```\n\n<b>raw</b>")
	want := "A title\nSome bold words and a link.\none\ntwo\nfmt.Println(1 < 2)\nraw"
	if doc.Text != want {
		t.Fatalf("Text =\n%q\nwant\n%q", doc.Text, want)
	}
	if doc.Words != 14 {
		t.Errorf("Words = %d, want 14", doc.Words)
	}
	if doc.Minutes != 1 {
		t.Errorf("Minutes = %d, want 1", doc.Minutes)
	}

	long := render(t, strings.Repeat("word ", 401))
	if long.Minutes != 3 {
		t.Errorf("401 words at 200 a minute = %d minutes, want 3", long.Minutes)
	}
	if empty := render(t, " \n\n "); empty.Minutes != 0 || empty.Words != 0 || empty.HTML() != "" {
		t.Errorf("an empty body = %+v, want nothing", empty)
	}
}

func TestCodeKeepsItsLanguageAndEscapesItsContents(t *testing.T) {
	t.Parallel()

	doc := render(t, "```html\n<script>alert(1)</script>\n```\n\n```bad language\nx\n```")
	onlyAllowlisted(t, "code", doc)
	if !strings.Contains(string(doc.HTML()), `<code class="language-html">&lt;script&gt;alert(1)&lt;/script&gt;`) {
		t.Fatalf("code was not kept as text: %s", doc.HTML())
	}
}

func TestATaskListBoxIsAlwaysDisabled(t *testing.T) {
	t.Parallel()

	doc := render(t, "- [ ] open\n- [x] done\n\n<input type=\"checkbox\" name=\"x\"><input type=\"text\">")
	onlyAllowlisted(t, "tasks", doc)
	if got := strings.Count(string(doc.HTML()), `<input type="checkbox" disabled`); got != 3 {
		t.Fatalf("disabled boxes = %d, want 3: %s", got, doc.HTML())
	}
	if !strings.Contains(string(doc.HTML()), `&lt;input type=&#34;text&#34;&gt;`) {
		t.Fatalf("a text input was not shown as text: %s", doc.HTML())
	}
}

func TestAnUnclosedSpanDoesNotSwallowTheBlocksAfterIt(t *testing.T) {
	t.Parallel()

	doc := render(t, "<em>never closed\n\n## A section\n\nA paragraph.")
	if !strings.Contains(string(doc.HTML()), `<h2 id="a-section">`) {
		t.Fatalf("the heading after an unclosed span lost its address: %s", doc.HTML())
	}
	if len(doc.Headings) != 1 {
		t.Fatalf("headings = %+v, want the section", doc.Headings)
	}
}

func TestControlCharactersAreDropped(t *testing.T) {
	t.Parallel()

	doc := render(t, "[x](java\x00script:alert(1)) a\x01b\x7f")
	onlyAllowlisted(t, "controls", doc)
	if strings.ContainsAny(string(doc.HTML())+doc.Text, "\x00\x01\x7f") {
		t.Fatalf("a control character reached the page: %q", doc.HTML())
	}
}

func TestNewRefusesAConfigurationThatCannotWork(t *testing.T) {
	t.Parallel()

	for _, cfg := range []markdown.Config{
		{TopHeading: -1},
		{TopHeading: 7},
		{WordsPerMinute: -1},
		{WordsPerMinute: markdown.MaxWordsPerMinute + 1},
		{Reserved: []string{"1st"}},
		{Reserved: []string{"has space"}},
		{ImageOrigins: []string{"http://cdn.example.com"}},
		{ImageOrigins: []string{"https://cdn.example.com/prod"}},
		{ImageOrigins: []string{"https://user@cdn.example.com"}},
		{ImageOrigins: []string{"cdn.example.com"}},
		{ImageOrigins: []string{"https://cdn.example.com?x"}},
	} {
		if _, err := markdown.New(cfg); err == nil {
			t.Errorf("New(%+v) was accepted", cfg)
		}
	}
	if _, err := markdown.New(markdown.Config{}); err != nil {
		t.Errorf("the zero configuration was refused: %v", err)
	}
}

func TestOneModuleRendersForManyRequestsAtOnce(t *testing.T) {
	t.Parallel()

	m := renderer(t)
	want := m.Render("## Same\n\nbody")
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if got := m.Render("## Same\n\nbody"); got.HTML() != want.HTML() || got.Headings[0] != want.Headings[0] {
				t.Errorf("a concurrent render differed: %s", got.HTML())
			}
		})
	}
	wg.Wait()
}

// FuzzRenderWritesOnlyTheAllowlist runs its seeds on every `go test`, and
// searches past them under `go test -fuzz`.
func FuzzRenderWritesOnlyTheAllowlist(f *testing.F) {
	for _, seed := range []string{
		"<script>alert(1)</script>",
		"[x](javascript:alert(1))",
		"![x](javascript:alert(1))",
		"<a href=\"/ok\" onclick=\"x\">y</a>",
		"<img src=/x onerror=y>",
		"<svg/onload=alert(1)>",
		"<math><mi xlink:href=\"javascript:alert(1)\">x</mi></math>",
		"<!--><img src=x onerror=alert(1)>-->",
		"<p title=\"</p><script>alert(1)</script>\">",
		"`<script>` and ``<b>``",
		"| a |\n|---|\n| <script> |",
		"> <iframe>\n> - [ ] <b>",
		"<textarea><script>alert(1)</script></textarea>",
		"<noscript><p title=\"</noscript><img src=x onerror=alert(1)>\">",
		"<em><strong>a</em></strong>",
	} {
		f.Add(seed)
	}
	m := renderer(f)
	f.Fuzz(func(t *testing.T, src string) {
		onlyAllowlisted(t, src, m.Render(src))
	})
}

// TestABodyTheParserCannotReadIsShownAsItsText holds the answer for a source
// the parser once panicked on: a link destination ending on a lone backslash.
func TestABodyTheParserCannotReadIsShownAsItsText(t *testing.T) {
	t.Parallel()

	doc := render(t, "Before.\n\n[](\\")
	onlyAllowlisted(t, "[](\\", doc)
	if !strings.Contains(doc.Text, "[](\\") || !strings.Contains(doc.Text, "Before.") {
		t.Fatalf("Text = %q, want the source's words", doc.Text)
	}
}

// TestAHostileNestedBodyIsHeldOnce holds the memory a render takes for a body
// built to make the parser copy: a paragraph opened under a hundred nested
// quotes and continued by lazy lines, which every one of the hundred levels
// carries down to the paragraph at the bottom.
//
// A parser that copies those lines at every level allocates about 580 bytes for
// each byte of this body, 23 MB for 40 KB of text. One that holds a level as a
// range of the lines it already has allocates 36, the same with the race
// detector on and off, and a body of ordinary prose of the same length takes
// 54 to 58 through this package's rewrite. The bound sits above both and well
// below the copy.
//
// It does not call t.Parallel: the memory profile it reads belongs to the whole
// process, and a top-level test that stays sequential runs while every parallel
// one is paused.
func TestAHostileNestedBodyIsHeldOnce(t *testing.T) {
	const bound = 100

	m := renderer(t)
	src := strings.Repeat("> ", 100) + "x\n" + strings.Repeat("y\n", 20000)
	var doc markdown.Document
	perByte := allocatedPerByte(func() { doc = m.Render(src) }, len(src))

	if n := strings.Count(string(doc.HTML()), "<blockquote>"); n != 100 {
		t.Fatalf("the body rendered %d nested quotes, want 100", n)
	}
	if doc.Words != 20001 {
		t.Fatalf("the body rendered %d words, want the 20001 it holds", doc.Words)
	}
	if perByte > bound {
		t.Fatalf("a render of %d bytes allocated %.0f bytes for each, want at most %d", len(src), perByte, bound)
	}
	t.Logf("a render of %d bytes allocated %.1f bytes for each", len(src), perByte)
}

// allocatedPerByte is what render allocates under this package, divided by
// size, read from a memory profile that records every allocation.
//
// The profile is read rather than the heap's running total because of what
// package regexp allocates under the race detector. A matcher keeps its
// backtracking state, 32 KB of it, in a sync.Pool, and with the detector on the
// pool drops a quarter of what is put back, so a render that matches short
// lines many times allocates far more than a build without the detector ever
// does. That is the detector's cost and not the renderer's; the allocations
// made under regexp are left out, and what remains is the same with the
// detector on and off.
func allocatedPerByte(render func(), size int) float64 {
	defer func(rate int) { runtime.MemProfileRate = rate }(runtime.MemProfileRate)
	runtime.MemProfileRate = 1
	before := renderAllocations()
	render()
	return float64(renderAllocations()-before) / float64(size)
}

// renderAllocations is the total the memory profile has recorded allocated
// under a function of this package and outside package regexp. The parser's
// allocations count: it runs under Render. The collection it runs first is what
// publishes the allocations made since the last one.
func renderAllocations() int64 {
	runtime.GC()
	var records []runtime.MemProfileRecord
	n, ok := runtime.MemProfile(nil, true)
	for !ok {
		records = make([]runtime.MemProfileRecord, n+64)
		n, ok = runtime.MemProfile(records, true)
	}
	var total int64
	for _, r := range records[:n] {
		inPackage, inRegexp := false, false
		frames := runtime.CallersFrames(r.Stack())
		for {
			frame, more := frames.Next()
			inPackage = inPackage || strings.HasPrefix(frame.Function, "github.com/hyz-is/arandu-markdown.")
			inRegexp = inRegexp || strings.HasPrefix(frame.Function, "regexp.")
			if !more {
				break
			}
		}
		if inPackage && !inRegexp {
			total += r.AllocBytes
		}
	}
	return total
}
