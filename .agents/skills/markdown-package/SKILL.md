---
name: markdown-package
description: Install, wire and use the Arandu Markdown package (Go, Arandu) in an application. Use when the request is to "install arandu-markdown", "render Markdown", "render a blog post", "sanitize Markdown", "table of contents from headings", "reading time", "plain text for llms.txt", "go get github.com/hyz-is/arandu-markdown", "wire it into bootstrap/app.go", "images from the CDN do not show in posts", "my iframe shows as text", "a heading id changed to -2", or when a project's go.mod already requires github.com/hyz-is/arandu-markdown. Covers the wiring and where it goes, the four Config fields, the content security policy's img-src, caching, and what each Document field is for.
license: MIT
metadata:
  audience: app
---

# Using Arandu Markdown

A module that renders a Markdown body into HTML safe to write unescaped, with
the headings, the plain text and the reading time. It is registered by hand —
there is no service provider, no container and no discovery.

## Install

```bash
go get github.com/hyz-is/arandu-markdown
```

## The wiring, and where it goes

All of it is in `bootstrap/app.go`.

```go
import (
	markdown "github.com/hyz-is/arandu-markdown"
)
```

In `Build`, before `k.Register`:

```go
	markdownModule, err := markdown.New(markdown.Config{
		Reserved:     []string{"content"},  // ids the layout already uses
		ImageOrigins: imageOrigins,         // the same list as img-src
	})
	if err != nil {
		return App{}, err
	}
```

Inside the `k.Register(...)` call:

```go
		markdownModule,
```

Then hand `markdownModule` to whatever builds pages — a service, a presenter —
as a parameter. It holds nothing that changes after `New`, so one value serves
every request.

## Rendering

```go
doc := markdownModule.Render(source)
```

| Field | Use |
| --- | --- |
| `HTML()` | the view writes it unescaped: `{!! .Body.HTML() !!}` in Kyse. A method, so `aru doctor` reads it as markup something escaped |
| `Headings` | the table of contents: `<a href="#{{ h.ID }}">{{ h.Text }}</a>` |
| `Text` | search, meta descriptions, `llms-full.txt`, JSON-LD `articleBody`. Plain text: escape it like any string |
| `Words`, `Minutes` | reading time, `timeRequired` as `PT{{Minutes}}M` |

`Render` is pure: cache by the source when a page renders the same body often.

## Symptoms to recognise

- **An image in a post shows its description instead of the picture.** Its
  address is on an origin `ImageOrigins` does not name. Add the origin — and it
  must be the one the content security policy allows in `img-src`, or the
  browser refuses the picture instead.
- **A pasted `<iframe>`, `<div>` or `<script>` shows up as text.** That is the
  allowlist. Raw HTML is shown as what it was, so the author sees it; there is
  no switch to let it through.
- **A heading's id ends in `-2`.** Another heading has the same words, or the
  id is in `Reserved`.
- **`New` returns an error at boot.** The message names the field: a heading
  level outside 1–6, a reserved id that is not an id, or an image origin that
  is not a bare `https://` origin.
