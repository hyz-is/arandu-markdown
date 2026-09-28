package markdown

import (
	"html/template"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/arandu-io/hesape/str"
	"golang.org/x/net/html"
)

// element is how the rewrite treats one element on the allowlist.
type element struct {
	// void elements have no content and no end tag.
	void bool
	// block elements close an open paragraph or heading, and the inline
	// elements still open inside it, before they start -- as a browser would,
	// so a block never lands inside a span of text.
	block bool
	// inline elements are spans of text.
	inline bool
}

// elements is the allowlist: exactly what str.Markdown writes, and nothing it
// does not. An element that is not here is shown as the text it was.
var elements = map[string]element{
	"p": {block: true}, "pre": {block: true}, "blockquote": {block: true},
	"ul": {block: true}, "ol": {block: true}, "table": {block: true},
	"h1": {block: true}, "h2": {block: true}, "h3": {block: true},
	"h4": {block: true}, "h5": {block: true}, "h6": {block: true},
	"hr": {void: true, block: true},
	"li": {}, "thead": {}, "tbody": {}, "tr": {}, "th": {}, "td": {},
	"strong": {inline: true}, "em": {inline: true}, "del": {inline: true},
	"code": {inline: true}, "a": {inline: true},
	"br": {void: true}, "img": {void: true}, "input": {void: true},
}

// lineEnds are the elements whose end starts a new line of Document.Text.
var lineEnds = map[string]bool{
	"p": true, "pre": true, "blockquote": true, "li": true, "tr": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// maxSlug is the longest a heading's id is before a number makes it unique.
const maxSlug = 80

var (
	languageClass = regexp.MustCompile(`^language-[A-Za-z0-9_+#.-]{1,32}$`)
	anchorHref    = regexp.MustCompile(`^#[A-Za-z0-9_-]{1,128}$`)
	listStart     = regexp.MustCompile(`^[0-9]{1,9}$`)
	emailAddress  = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	spaces        = regexp.MustCompile(`\s+`)
)

// Render renders a Markdown body.
//
// It never fails: a body is text somebody wrote, and every part of it that is
// not on the allowlist is still text. What comes back is the same for the same
// source and configuration, so a caller may cache it by the source.
//
// The cost is str.Markdown's plus one linear pass. str.Markdown is linear on
// ordinary prose and quadratic on a run of link openers that never close, so a
// field that takes a body from people the application does not trust caps its
// length, as a form field does anyway.
func (m *Module) Render(src string) Document {
	// Control characters other than the line breaks and the tab are dropped
	// before anything reads them: they have no place in a body, and a browser
	// ignores some of them inside a URL the checks below read with them in it.
	src = strings.Map(func(c rune) rune {
		if c == '\n' || c == '\r' || c == '\t' || (c >= 0x20 && c != 0x7f) {
			return c
		}
		return -1
	}, src)

	w := &writer{m: m, taken: map[string]bool{}}
	z := html.NewTokenizer(strings.NewReader(parse(src)))
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			// The reader is a string, so the only error is the end of it.
			break
		}
		// Raw is read before Token, which reuses the buffer it points into.
		raw := string(z.Raw())
		token := z.Token()
		switch kind {
		case html.TextToken:
			w.text(token.Data)
		case html.StartTagToken, html.SelfClosingTagToken:
			w.start(token, raw)
		case html.EndTagToken:
			w.end(token.Data, raw)
		}
		// A comment or a doctype is the author's note to themselves, not part
		// of the page, and is dropped.
	}
	for len(w.stack) > 0 {
		w.pop()
	}

	text := w.plain()
	words := len(strings.Fields(text))
	minutes := 0
	if words > 0 {
		minutes = (words + m.cfg.WordsPerMinute - 1) / m.cfg.WordsPerMinute
	}
	return Document{
		HTML:     template.HTML(w.root.String()),
		Headings: w.headings,
		Text:     text,
		Words:    words,
		Minutes:  minutes,
	}
}

// frame is one open element.
type frame struct {
	// name is the element as it came in, which is what its end tag names.
	name string
	// tag is the element as it went out: a heading's level after
	// TopHeading, or empty when the start tag was dropped and its end has to be
	// dropped with it.
	tag string
	// buf holds the content of a paragraph or a heading until its end decides
	// how it is written.
	buf *strings.Builder
	// heading is the heading's plain text.
	heading strings.Builder
	// image is the one image of a paragraph that may turn out to hold nothing
	// else; alone is false once anything else is found in it.
	image *picture
	alone bool
}

type picture struct {
	src, alt, title string
}

type writer struct {
	m        *Module
	stack    []*frame
	root     strings.Builder
	said     strings.Builder
	headings []Heading
	taken    map[string]bool
}

// out is where markup goes: the innermost paragraph or heading still being
// held, or the document.
func (w *writer) out() *strings.Builder {
	if f := w.held(); f != nil {
		return f.buf
	}
	return &w.root
}

// held is the innermost paragraph or heading still being held, if any.
func (w *writer) held() *frame {
	for i := len(w.stack) - 1; i >= 0; i-- {
		if w.stack[i].buf != nil {
			return w.stack[i]
		}
	}
	return nil
}

// open is the innermost open element with this name.
func (w *writer) open(name string) *frame {
	for i := len(w.stack) - 1; i >= 0; i-- {
		if w.stack[i].name == name {
			return w.stack[i]
		}
	}
	return nil
}

// notAlone records that the paragraph being held holds more than one image.
func (w *writer) notAlone() {
	if p := w.open("p"); p != nil {
		p.alone = false
	}
}

func (w *writer) text(s string) {
	w.out().WriteString(html.EscapeString(s))
	if w.open("pre") == nil {
		s = spaces.ReplaceAllString(s, " ")
	}
	w.said.WriteString(s)
	if f := w.held(); f != nil && f.name != "p" {
		f.heading.WriteString(s)
	}
	if strings.TrimSpace(s) != "" {
		w.notAlone()
	}
}

// literal shows a tag that is not on the allowlist as the text it was.
func (w *writer) literal(raw string) {
	w.out().WriteString(html.EscapeString(raw))
	w.notAlone()
}

func (w *writer) start(t html.Token, raw string) {
	el, ok := elements[t.Data]
	if !ok {
		w.literal(raw)
		return
	}
	if el.block {
		w.closeText()
		w.said.WriteString("\n")
	}
	if p := w.open("p"); p != nil && !(t.Data == "img" && w.stack[len(w.stack)-1] == p) {
		p.alone = false
	}

	name, tag := t.Data, t.Data
	var open string
	switch name {
	case "p":
		w.stack = append(w.stack, &frame{name: name, tag: tag, buf: &strings.Builder{}, alone: true})
		return
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := max(int(name[1]-'0'), w.m.cfg.TopHeading)
		w.stack = append(w.stack, &frame{name: name, tag: "h" + strconv.Itoa(level), buf: &strings.Builder{}})
		return
	case "a":
		href, ok := w.m.href(attr(t, "href"))
		if !ok || w.open("a") != nil {
			// A link to somewhere a body may not point, or a link inside a
			// link: the words stay and the link does not.
			w.stack = append(w.stack, &frame{name: name})
			return
		}
		open = `<a href="` + html.EscapeString(href) + `"` + titleAttr(t) + `>`
	case "img":
		w.image(t)
		return
	case "input":
		// A task list's box. It is drawn and never submitted, so it is
		// always disabled whatever the source said.
		if !strings.EqualFold(attr(t, "type"), "checkbox") {
			w.literal(raw)
			return
		}
		open = `<input type="checkbox" disabled`
		if has(t, "checked") {
			open += ` checked`
		}
		open += `>`
	case "code":
		open = `<code>`
		if class := attr(t, "class"); languageClass.MatchString(class) {
			open = `<code class="` + class + `">`
		}
	case "ol":
		open = `<ol>`
		if start := attr(t, "start"); listStart.MatchString(start) {
			open = `<ol start="` + start + `">`
		}
	case "th", "td":
		open = "<" + name + ">"
		switch align := attr(t, "align"); align {
		case "left", "center", "right":
			open = "<" + name + ` align="` + align + `">`
		}
	case "br", "hr":
		open = "<" + name + ">"
		w.said.WriteString("\n")
	default:
		open = "<" + name + ">"
	}
	w.out().WriteString(open)
	if !el.void {
		w.stack = append(w.stack, &frame{name: name, tag: tag})
	}
}

// image writes an image, or its description when its address is one the page
// could not load.
func (w *writer) image(t html.Token) {
	img := picture{alt: attr(t, "alt"), title: attr(t, "title")}
	src, ok := w.m.src(attr(t, "src"))
	if !ok {
		if img.alt != "" {
			w.text(img.alt)
		}
		return
	}
	img.src = src
	out := `<img src="` + html.EscapeString(img.src) + `" alt="` + html.EscapeString(img.alt) + `"`
	if img.title != "" {
		out += ` title="` + html.EscapeString(img.title) + `"`
	}
	w.out().WriteString(out + ` loading="lazy" decoding="async">`)
	w.said.WriteString(img.alt + " ")

	if top := w.top(); top != nil && top.name == "p" {
		if top.image != nil {
			top.alone = false
		}
		top.image = &img
	}
}

func (w *writer) top() *frame {
	if len(w.stack) == 0 {
		return nil
	}
	return w.stack[len(w.stack)-1]
}

// closeText closes the paragraph or heading and the spans of text still open,
// because a block is starting.
func (w *writer) closeText() {
	for len(w.stack) > 0 {
		top := w.top()
		if top.buf == nil && !elements[top.name].inline {
			return
		}
		w.pop()
	}
}

func (w *writer) end(name, raw string) {
	if _, ok := elements[name]; !ok {
		w.literal(raw)
		return
	}
	// An end tag with nothing open to close is dropped: a browser would make
	// an empty element of it, which says nothing.
	for i := len(w.stack) - 1; i >= 0; i-- {
		if w.stack[i].name == name {
			for len(w.stack) > i {
				w.pop()
			}
			return
		}
	}
}

// pop closes the innermost open element.
func (w *writer) pop() {
	f := w.stack[len(w.stack)-1]
	w.stack = w.stack[:len(w.stack)-1]
	if lineEnds[f.name] {
		w.said.WriteString("\n")
	} else if f.name == "th" || f.name == "td" {
		w.said.WriteString(" ")
	}

	switch {
	case f.tag == "":
		return
	case f.name == "p":
		content := f.buf.String()
		switch {
		case f.image != nil && f.alone:
			// A paragraph holding one image and nothing else is a figure, and
			// the image's title is its caption.
			out := `<figure><img src="` + html.EscapeString(f.image.src) + `" alt="` + html.EscapeString(f.image.alt) +
				`" loading="lazy" decoding="async">`
			if f.image.title != "" {
				out += `<figcaption>` + html.EscapeString(f.image.title) + `</figcaption>`
				w.said.WriteString(f.image.title + "\n")
			}
			w.out().WriteString(out + "</figure>")
		case strings.TrimSpace(content) != "":
			w.out().WriteString("<p>" + content + "</p>")
		}
	case f.buf != nil:
		w.heading(f)
	default:
		w.out().WriteString("</" + f.tag + ">")
	}
}

// heading writes a heading with its id, and lists it when it is a section of
// the page.
func (w *writer) heading(f *frame) {
	text := strings.TrimSpace(f.heading.String())
	if text == "" {
		// A heading that says nothing gives a table of contents an entry with
		// no words, and a page a landmark with no name.
		return
	}
	nested := len(w.stack) > 0
	if nested {
		w.out().WriteString("<" + f.tag + ">" + f.buf.String() + "</" + f.tag + ">")
		return
	}
	id := w.unique(str.Slug(text, "-"))
	w.out().WriteString("<" + f.tag + ` id="` + id + `">` + f.buf.String() + "</" + f.tag + ">")
	if level := int(f.tag[1] - '0'); level <= w.m.cfg.TopHeading+1 {
		w.headings = append(w.headings, Heading{Level: level, ID: id, Text: text})
	}
}

// unique is an id no heading and no reserved element holds yet.
//
// The base is the heading's slug, cut at a word to keep an address someone
// shares readable, or "section" for a heading with no letter a slug keeps.
func (w *writer) unique(base string) string {
	if len(base) > maxSlug {
		base = base[:maxSlug]
		if cut := strings.LastIndexByte(base, '-'); cut > 0 {
			base = base[:cut]
		}
	}
	if base = strings.Trim(base, "-"); base == "" {
		base = "section"
	}
	for n := 1; ; n++ {
		id := base
		if n > 1 {
			id += "-" + strconv.Itoa(n)
		}
		if !w.taken[id] && !w.m.reserved[id] {
			w.taken[id] = true
			return id
		}
	}
}

// plain is Document.Text: one line per block, with nothing but words in it.
func (w *writer) plain() string {
	var lines []string
	for _, line := range strings.Split(w.said.String(), "\n") {
		if line = strings.Join(strings.Fields(line), " "); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// href is a link destination a body may point at: an http(s) address, an
// e-mail, a path on this site or a heading of this page.
func (m *Module) href(dest string) (string, bool) {
	dest = strings.TrimSpace(dest)
	if dest == "" || strings.ContainsAny(dest, "\\\t\n\r") {
		return "", false
	}
	switch {
	case dest[0] == '#':
		return dest, anchorHref.MatchString(dest)
	case dest[0] == '/':
		return dest, samePath(dest)
	}
	u, err := url.Parse(dest)
	if err != nil {
		return "", false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return dest, u.Host != "" && u.User == nil
	case "mailto":
		address, _, _ := strings.Cut(u.Opaque, "?")
		return dest, emailAddress.MatchString(address)
	}
	return "", false
}

// src is an image address a page can load: a path on this site, or an
// address on an origin Config.ImageOrigins names.
func (m *Module) src(dest string) (string, bool) {
	dest = strings.TrimSpace(dest)
	if dest == "" || strings.ContainsAny(dest, "\\\t\n\r") {
		return "", false
	}
	if dest[0] == '/' {
		return dest, samePath(dest)
	}
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Host == "" {
		return "", false
	}
	return dest, m.origins["https://"+strings.ToLower(u.Host)]
}

// samePath is a path on this site: one slash, never two, and no way back up
// out of it.
func samePath(dest string) bool {
	if strings.HasPrefix(dest, "//") {
		return false
	}
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil {
		return false
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == ".." {
			return false
		}
	}
	return true
}

func attr(t html.Token, key string) string {
	for _, a := range t.Attr {
		if a.Namespace == "" && a.Key == key {
			return a.Val
		}
	}
	return ""
}

func has(t html.Token, key string) bool {
	for _, a := range t.Attr {
		if a.Namespace == "" && a.Key == key {
			return true
		}
	}
	return false
}

func titleAttr(t html.Token) string {
	if title := attr(t, "title"); title != "" {
		return ` title="` + html.EscapeString(title) + `"`
	}
	return ""
}

// parse is the source as str.Markdown renders it, or as escaped paragraphs
// when the parser cannot read it.
//
// Render promises an answer for any text, and a parser handed text somebody
// typed is the part most likely to meet an input nobody wrote a test for: a
// panic there would be one visitor's body failing every request for the page.
// The paragraphs go through the same rewrite as the parser's output, so the
// page, the text and the word count still agree.
func parse(src string) (out string) {
	defer func() {
		if recover() != nil {
			var b strings.Builder
			for _, block := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n\n") {
				if strings.TrimSpace(block) != "" {
					b.WriteString("<p>" + html.EscapeString(block) + "</p>\n")
				}
			}
			out = b.String()
		}
	}()
	return str.Markdown(src)
}
