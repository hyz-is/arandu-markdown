# Contributing

Thank you for considering a contribution.

## Before you open a pull request

Run the four gates. They are what CI runs, and a pull request that fails one of
them fails for everybody who pulls after it.

```bash
gofmt -l $(find . -name '*.go' -not -path '*/testdata/*' -not -name '*.kyse.go')
go build ./...
go vet ./...
go test -race ./...
```

The two filters on `gofmt` are not preference. `gofmt` is the only tool in the
chain that ignores build tags, so it reads files the compiler deliberately
excludes; `testdata/` holds fixtures that are invalid on purpose.

## What a change has to keep true

Four properties are the reason this package is what it is. A change that breaks
one of them will not be merged, whatever else it improves.

1. **Only the allowlist leaves.** An element or attribute is added to
   `render.go` only when `hesape/str.Markdown` writes it, and in the same commit
   as the line in `tests/Unit/render_test.go` that allows it.
2. **A destination is checked, not trusted.** A new scheme or origin is a new
   way for a body to reach something, and needs the argument in the pull
   request.
3. **Nothing is silently lost.** What the allowlist refuses is shown as text.
4. **`arandu.mod.toml` matches the code.** Adding an outbound call, a file
   read or write, a process or a table means declaring it there in the same
   commit.

This package does not parse Markdown. A rendering bug belongs to
`hesape/str.Markdown`, and is fixed there.

## Style

Everything in the source is in English: identifiers, doc comments, internal
comments, error messages, log messages, and the names and messages of tests.
`pkg.go.dev` publishes the doc comments, and its readers are users of this
package.

Every exported symbol carries a doc comment saying what it does. The comment
documents the symbol and nothing else: no dates, no issue numbers, no
references to other repositories. Why a signature is what it is belongs there
when it is a fact about the code; anything else belongs in the pull request.

## Commits

One commit per logical change. Formatting is not mixed with behaviour, because
a diff that does both is a diff nobody reviews.

## Reporting a vulnerability

Do not open an issue. See [SECURITY.md](SECURITY.md).

## Licence

By contributing you agree that your contribution is licensed under the MIT
licence, the same as the rest of github.com/hyz-is/arandu-markdown.
