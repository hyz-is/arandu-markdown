# Working on Arandu Markdown

This is an Arandu package that renders a Markdown body for a page: HTML safe to
write unescaped, the headings a table of contents lists, the plain text search
and generative engines read, and the reading time. It parses nothing itself —
`hesape/str.Markdown` does — and what it adds is the closed allowlist that
output is rewritten through.
It is a Go module somebody `go get`s and registers by hand in their own
`bootstrap/app.go`, which is the whole difference from working in an
application. There is no service provider, no container and no discovery — if a
line of wiring is not written in the installer's repository, it does not happen.

Read `.agents/skills/` before writing code. Each skill is a procedure, and the
one you need is named by the situation you are in.

## The gates

Nothing is finished until all four exit zero.

```sh
export GOWORK=off
gofmt -l $(find . -name '*.go' -not -path '*/testdata/*' -not -name '*.kyse.go')
go build ./...
go vet ./...
go test -race ./...
```

`GOWORK=off` is not borrowed from somewhere else, and here it is not a
preference either. This checkout may sit beside a Go workspace that lists the
framework repositories and does not list this one; when it does, every command
above fails before it compiles anything:

```
pattern ./...: directory prefix . does not contain modules listed in go.work
or their selected dependencies
```

With the workspace off, the module resolves the framework version in `go.mod` —
which is what CI compiles against, and what somebody's `go get` will get.

Both filters on `gofmt` are load-bearing in the toolchain even where this
repository has nothing for them to skip: `gofmt` is the only tool in the chain
that ignores build tags, and `testdata/` is where a fixture is allowed to be
invalid on purpose.

`aru doctor` is not one of the gates, and running it here costs a minute and
answers nothing:

```
this is not an Arandu project: no go.mod, main.go and arandu.toml together.
Run it from inside a project, or create one with `aru new`
```

It exits 1. It reads applications, and this is a library.

It does not read this one after it is installed either, and that is why
`tests/Unit/audit_test.go` exists. The doctor walks the application's own tree,
skips `vendor/`, and opens the one `arandu.mod.toml` at its root; it never loads
a dependency. So `tenant-from-request`, `system-grant-without-tenant` and
`permission-not-declared` never see a line of an installed package, and whatever
this one must prove about itself it proves in its own suite or nowhere.


## What this repository holds

| | measured with |
| --- | --- |
| 3 Go files, one per role, all in one package at the root | `grep -l '^package markdown' *.go` |
| 5 test files, one of them a fuzz target | `find tests -name '*_test.go'` |
| 0 routes, 0 tables, 0 capabilities | `module.go`, `arandu.mod.toml` |
| 3 direct dependencies: `framework`, `hesape`, `golang.org/x/net` | `go list -m -f '{{if and (not .Indirect) (not .Main)}}{{.Path}} {{.Version}}{{end}}' all` |

```
module.go   registration, Document and Heading, and the one entry point
config.go   what the application passes in
render.go   the allowlist and the rewrite
```

## What does not exist here, on purpose

| A model reaches for | What is here instead |
| --- | --- |
| a Markdown parser of its own, or goldmark | `str.Markdown`, the collection's one renderer. A second parser is a second way to spell the same document |
| a sanitizer library | `golang.org/x/net/html`'s tokenizer and a closed list written out in `render.go`. The list is what `str.Markdown` emits and nothing else, so it is short enough to read whole |
| stripping a tag that is not on the list | showing it as the text it was. An author who pasted an iframe sees the tag in the preview instead of wondering where it went |
| an option per element, per attribute, per scheme | four settings in `Config`. The allowlist is not configurable: a list each installer can widen is a list nobody can audit |
| a route that renders a body | nothing. A body reaches a page through the controller that authorized its entry; a route here would render text for anybody |
| a cache | the caller's. `Render` is a pure function of the source and the configuration |

## The four properties

A change that breaks one of them is not merged, whatever else it improves.
`tests/Unit/render_test.go` holds all four against the code.

1. **Only the allowlist leaves.** Every element and attribute in
   `Document.HTML` is one `render.go` wrote itself; text is escaped.
   `FuzzRenderWritesOnlyTheAllowlist` reads the output back with the same
   tokenizer and fails on anything else — its seeds run on every `go test`.
2. **A destination is checked, not trusted.** Links keep http(s) with no
   userinfo, mailto with an address, same-site paths with no `..`, and in-page
   anchors; images keep same-site paths and the origins `Config.ImageOrigins`
   names, which is the list the application's content security policy allows.
3. **Nothing is silently lost.** A refused link keeps its words, a refused image
   its description, a refused tag its text.
4. **`arandu.mod.toml` matches the code.** `tests/Unit/audit_test.go` compares
   the declaration with what the code *calls* — the day a socket, a file or a
   table appears, the manifest says so in the same commit.

## Writing code

Everything in the source is in English: identifiers, doc comments, internal
comments, error messages, log messages, and the names and messages of tests.
`pkg.go.dev` publishes the doc comments and its readers are users of this
package.

Every exported symbol carries a doc comment, and the comment documents the
symbol and nothing else. Why a signature is what it is belongs there when it is
a fact about the code. A date, an issue number, a version in progress or the
name of another repository does not.

Tests go under `tests/`, in a capitalized category directory declaring a
lowercase external package: `tests/Unit` holds `package unit_test`,
`tests/Feature` holds `package feature_test`. Both import the package by its
module path, which is what makes them see exactly what a caller sees.
