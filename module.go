// Package markdown is an Arandu module that turns a Markdown body into what a
// page draws from it: HTML that is safe to write unescaped, the headings a
// table of contents lists, the plain text search and generative engines read,
// and the reading time.
//
// It does not parse Markdown. hesape/str.Markdown is the collection's one
// renderer -- CommonMark with the GitHub tables, task lists, strikethrough and
// autolinks -- and a second parser here would be a second way to spell the same
// document. What str.Markdown leaves to its caller is the reason this package
// exists: raw HTML passes through it untouched, so rendering a body somebody
// typed renders whatever HTML they typed.
//
// This package reads that output back as HTML and writes it again through a
// closed list of elements and attributes. Everything on the list is what the
// renderer itself emits; everything else is shown as the text it was, so an
// author who pasted an iframe sees the tag on the page instead of wondering
// where it went. Links keep http(s), mailto, same-site paths and in-page
// anchors. Images keep same-site paths and the origins the application names,
// which is what its content security policy lets a page load.
//
// The files are laid out by role:
//
//	module.go  -> registration and the one entry point, Render
//	config.go  -> what the application passes in
//	render.go  -> the allowlist and the rewrite
//
// An application builds it once in bootstrap/app.go and registers it there.
// There is no service provider, no container and no discovery.
package markdown

import (
	"html/template"

	"github.com/arandu-io/framework/foundation"
	fhttp "github.com/arandu-io/framework/http"
)

// Module is what the application registers, and what renders a body.
//
// It holds nothing that changes after New, so one value is safe to share
// between every request that renders.
type Module struct {
	cfg      Config
	origins  map[string]bool
	reserved map[string]bool
}

// Compile-time proof that the module honors the contract it claims.
var _ foundation.Module = (*Module)(nil)

// New returns the module, or the reason it cannot be built.
//
// It returns an error rather than panicking or carrying on, because everything
// it refuses is a wiring mistake, and a wiring mistake found at boot costs one
// restart. The same mistake found later is a page drawn with a broken picture.
func New(cfg Config) (*Module, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg = cfg.withDefaults()
	m := &Module{cfg: cfg, origins: map[string]bool{}, reserved: map[string]bool{}}
	for _, origin := range cfg.ImageOrigins {
		bare, _ := bareOrigin(origin)
		m.origins[bare] = true
	}
	for _, id := range cfg.Reserved {
		m.reserved[id] = true
	}
	return m, nil
}

// Name is the module identifier: a lowercase slug, stable, no spaces.
func (m *Module) Name() string { return "markdown" }

// Routes registers nothing.
//
// A body reaches a page through the application's own controller, which already
// holds the Grant that decided whether the reader may see the entry it belongs
// to. A route here would render text for anybody who sent some, with no entry
// and no Grant behind it.
func (m *Module) Routes(*fhttp.Router) {}

// Document is a rendered body.
type Document struct {
	// html is the body as markup; HTML is how a view reaches it.
	html template.HTML
	// Headings are the top two levels of the body's own headings, in order,
	// for a table of contents. A heading inside a quote or a list is part of
	// that block, not a section of the page, and is not listed.
	Headings []Heading
	// Text is what the body says, without markup: one line per block, for a
	// search index, a description, llms-full.txt and the word count. It is
	// never HTML, and a view that prints it escapes it like any other string.
	Text string
	// Words is the number of words in Text.
	Words int
	// Minutes is the reading time at Config.WordsPerMinute, rounded up: at
	// least one for a body with any words, and zero for one without.
	Minutes int
}

// HTML is the body as markup. Every element and attribute in it is one the
// allowlist wrote, and every character of text in it was escaped, so a view
// writes it unescaped: {!! .Body.HTML() !!} in Kyse.
//
// It is a method and not a field because that is the shape a view's raw
// output is entitled to. `aru doctor` accepts {!! !!} around a call -- markup
// something produced by escaping -- and warns on a value, which is what a
// field reads as; the renderer is the thing that escaped this one.
func (d Document) HTML() template.HTML { return d.html }

// Heading is one entry of a table of contents.
type Heading struct {
	// Level is the level the heading was drawn at, after Config.TopHeading.
	Level int
	// ID is the element id the heading carries, unique in the body and never
	// one of Config.Reserved, so "#" + ID is its address.
	ID string
	// Text is the heading without markup.
	Text string
}
