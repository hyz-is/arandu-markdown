<p align="center">
  <img src=".github/logo.svg" alt="Arandu" width="180">
</p>

# Arandu Markdown

An Arandu package that renders a Markdown body for a page: HTML that is safe to
write unescaped, the headings a table of contents lists, the plain text search
and generative engines read, and the reading time.

It does not parse Markdown. `hesape/str.Markdown` is the collection's renderer —
CommonMark with GitHub tables, task lists, strikethrough and autolinks — and it
passes raw HTML through untouched, leaving the sanitizing to its caller. This
package is that caller: it reads the output back with the HTML5 tokenizer and
writes it again through a closed allowlist.

## Install

```bash
go get github.com/hyz-is/arandu-markdown
```

## Wire it

An Arandu application registers a module explicitly. There is no service
provider, no container and no discovery, so these are the lines to paste into
`bootstrap/app.go` and there are no others.

The import, with the other module imports:

```go
import (
	markdown "github.com/hyz-is/arandu-markdown"
)
```

The construction, in `Build`, before `k.Register`:

```go
	markdownModule, err := markdown.New(markdown.Config{
		// The ids the layout already uses, so no heading takes one.
		Reserved: []string{"content", "comments"},
		// The origins img-src allows besides 'self' -- the CDN of the
		// media library, if there is one.
		ImageOrigins: imageOrigins,
	})
	if err != nil {
		return App{}, err
	}
```

And the registration, inside the `k.Register(...)` call already there:

```go
		markdownModule,
```

`New` returns an error rather than starting half-wired: a heading level out of
range, a reserved id no heading could take, or an image origin that is not a
bare `https://` origin is refused where it is written.

Registering it adds no route, no table and no background loop. It serves no
route on purpose: a body reaches a page through the controller that authorized
the entry it belongs to, and a route here would render text for anybody who
sent some.

## Render

The module value is shared: hand it to whatever builds the page.

```go
doc := markdownModule.Render(post.Body)

doc.HTML()   // template.HTML, written unescaped by the view: {!! .Body.HTML() !!}
doc.Headings // []Heading{Level, ID, Text}, for the table of contents
doc.Text     // what the body says, one line per block, for search and llms-full.txt
doc.Words    // len(strings.Fields(doc.Text))
doc.Minutes  // reading time at Config.WordsPerMinute, 0 for an empty body
```

`Render` never fails and is a pure function of the source and the
configuration, so the result may be cached by the source.

## What comes out

| Markdown | HTML |
| --- | --- |
| paragraphs, hard breaks, `---` | `p`, `br`, `hr` |
| `#` to `######` | `h1`–`h6`, never above `Config.TopHeading`, with an `id` at the top level |
| `**strong**`, `*em*`, `~~del~~`, `` `code` `` | `strong`, `em`, `del`, `code` |
| fenced and indented code | `pre` and `code class="language-…"` |
| lists, task lists, quotes, tables | `ul`, `ol start`, `li`, disabled `input type="checkbox"`, `blockquote`, `table` with `align` |
| links | `a href title` — http(s) with no userinfo, `mailto:` with an address, same-site paths, in-page anchors |
| images | `img src alt title loading="lazy" decoding="async"` — same-site paths and `Config.ImageOrigins` |
| an image alone in its paragraph | `figure`, with the image's title as its `figcaption` |

Everything else is shown as the text it was. A pasted `<iframe>` appears on the
page as `<iframe>`, a refused link keeps its words and a refused image its
description, so an author reading the preview sees what did not make it instead
of wondering where it went.

## Configuration

| Field | Default | What it does |
| --- | --- | --- |
| `TopHeading` | `2` | The highest level a heading is drawn at. A higher one is lowered to it, never shifted, so `#` and `##` both become sections. The contents list this level and the one below it |
| `Reserved` | none | Ids the page already uses. A heading that would take one gets `-2` |
| `ImageOrigins` | none | Bare `https://` origins an image may load from besides the site itself — the same list as the content security policy's `img-src` |
| `WordsPerMinute` | `200` | The reading speed `Minutes` is measured at |

## Cost

One linear pass over what `str.Markdown` returns, and `str.Markdown` is linear
in the length of the source on hostile input too: on a run of openers that never
close, and on quotes and lists nested to the hundred levels it allows, whose
lines it holds once rather than copying them at every level. Linear is not cheap
on every shape, since a line under a hundred nested lists is read again at each
of them, so a field that takes a body from people the application does not
trust still caps its length, as its form validation does anyway.

## Layout

```
module.go   registration, Document and Heading, and the one entry point
config.go   what the application passes in
render.go   the allowlist and the rewrite
```

## Tests

```bash
go test -race ./...
```

`FuzzRenderWritesOnlyTheAllowlist` reads every output back with the same
tokenizer and fails on any element, attribute or destination the allowlist does
not write. Its seeds run on every `go test`; `go test -fuzz` searches past them.

`arandu.mod.toml` declares no capability, and `tests/Unit/audit_test.go`
compares that with what the code calls.

## Licence

MIT. See [LICENSE.md](LICENSE.md). Copyright HYZIS.
