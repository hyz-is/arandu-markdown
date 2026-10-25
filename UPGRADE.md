# Upgrade Guide

## Unreleased

Go 1.26.10 is the floor: `go.mod` declares `go 1.26.10`, because every
vulnerability an Arandu application reaches is in the standard library and is
fixed there (crypto/x509, net/http HTTP/2 and CONNECT, html/template,
net/textproto, crypto/tls). A build on an older toolchain fetches 1.26.10 or
fails when `GOTOOLCHAIN=local`.

This release selects Framework v0.66.0 and Hesape v0.85.0; read their upgrade
guides from the versions the application required before (Framework v0.56.0
and Hesape v0.54.0 in v0.3.0). Rendering, the configuration and the sanitizer
are untouched.

`apidiff` reports `(*Module).Routes` as changed. Its signature reads as before;
what moved is underneath the name it takes: Framework v0.63.0 declares
`fhttp.Router` as an alias of hesape's `controllers.Router`, where v0.56.0
declared a struct of its own. An application on Framework v0.66.0 passes the
router it already builds and changes nothing; one that stays on an older
Framework cannot take this release.

## v0.3.0

Nothing to change in code that calls this package: every new setting is
optional, and the zero value of each writes what the previous release wrote.

### Links to other sites can carry a `rel`

An application that renders bodies visitors wrote sets

```go
markdown.Config{LinkRel: "noopener nofollow ugc"}
```

and every link to an http(s) address comes out with that `rel`. Links to the
site's own paths, `#` anchors and `mailto:` addresses never carry it. A link
type outside `external`, `nofollow`, `noopener`, `noreferrer`, `sponsored` and
`ugc` is refused by `New`.

### Images can be turned off

```go
markdown.Config{Images: markdown.ImagesNone}
```

draws no image at all, same-site paths included: every image is written as its
description, as text. `ImageOrigins` has to be empty with it, or `New` refuses
the configuration. Leaving `Images` empty keeps `ImagesAllowedOrigins`, the
behaviour every earlier release had.

## v0.2.4

Nothing to change in code that calls this package. The release raises its
floors, so `go get github.com/hyz-is/arandu-markdown@v0.2.4` brings `framework`
v0.56.0 and `hesape` v0.54.0 into an application that required less. One that
built a session store from `config.Config.SessionTTL` follows the framework's
upgrade guide for v0.56.0 first: the field is removed, and the lifetime is
`SESSION_LIFETIME` in minutes, read by `bootstrap.LoadConfiguration`.

## v0.2.3

Nothing to change in code that calls this package. The release raises its
floors, so `go get github.com/hyz-is/arandu-markdown@v0.2.3` brings `framework`
v0.55.1 and `hesape` v0.52.0 into an application that required less. One on an
older `framework` follows the framework's upgrade guide from its version up to
v0.55.0 first: the session is configured by `bootstrap.Session`, and an unread
`SESSION_*` setting or a malformed boolean stops the boot. One that calls
`hesape` directly follows its guide for v0.52.0, which removes the names it had
deprecated.

## v0.2.2

Nothing to change in code that calls this package. The release raises its
floors, so `go get github.com/hyz-is/arandu-markdown@v0.2.2` brings `framework`
v0.50.0 and `hesape` v0.50.3 into an application that required less. One still
on `framework` v0.49 follows the framework's upgrade guide for v0.50.0 first:
`middleware.KeyBySession` takes the session store, and `CSRFProtect` issues a
guest's CSRF token.

## v0.2.1

Nothing to change in an application: this release touches skills, not Go code.
`aru skills:sync` now offers `markdown-package` to a project that requires this
version.

## v0.2.0

### `Document.HTML` is a method

`Document.HTML` was a field and is now `Document.HTML()`. Add the parentheses
wherever it is read:

```go
doc.HTML   // v0.1.0
doc.HTML() // v0.2.0
```

and in a Kyse view, `{!! .Body.HTML !!}` becomes `{!! .Body.HTML() !!}`. The
markup is the same; the call is the shape `aru doctor` accepts around raw
output.

## v0.1.0

The first release. Nothing to upgrade from.
