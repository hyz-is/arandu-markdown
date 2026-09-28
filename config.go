package markdown

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// The defaults for the optional settings. They are constants rather than
// literals inside Config.withDefaults, so the value a reader finds here is the
// value the package uses.
const (
	// DefaultTopHeading is the level the highest heading of a body is drawn at
	// when Config leaves TopHeading at zero. A page has one h1, its own title,
	// so a body starts one level below it.
	DefaultTopHeading = 2
	// DefaultWordsPerMinute is the reading speed Document.Minutes is measured
	// at when Config leaves WordsPerMinute at zero.
	DefaultWordsPerMinute = 200
	// MaxWordsPerMinute is the ceiling WordsPerMinute is refused above. A
	// speed nobody reads at is a reading time that says every body takes a
	// minute.
	MaxWordsPerMinute = 1000
)

// Config is what the application passes when it builds the renderer.
//
// A typed struct rather than a map: a misspelled key in a map is a setting that
// silently keeps its default, and the failure shows up as a page nobody asked
// for rather than as an error. Here a field that does not exist does not
// compile.
type Config struct {
	// TopHeading is the level the highest heading of a body is drawn at, from
	// 1 to 6. Zero means DefaultTopHeading.
	//
	// A heading above it is lowered to it rather than shifted: a body written
	// with # and ## for its sections comes out with both at this level, and
	// the ### under them one below. Shifting every level by one would turn the
	// ## an author meant as a section into a subsection whenever somebody else
	// in the same collection used #.
	TopHeading int

	// Reserved are the element ids the page around the body already uses --
	// the skip link's target, the comment form, the share bar. A heading never
	// takes one, so an anchor the layout owns keeps pointing where it did.
	Reserved []string

	// ImageOrigins are the origins an image in a body may load from besides the
	// page's own: bare https origins, such as the CDN a media library serves
	// from. Empty means same-site paths only.
	//
	// It is the list the application's content security policy names in
	// img-src, and it is asked for here rather than read from there because a
	// renderer that emitted an image the policy refuses would draw a broken
	// picture instead of its description.
	ImageOrigins []string

	// WordsPerMinute is the reading speed Document.Minutes is measured at.
	// Zero means DefaultWordsPerMinute.
	WordsPerMinute int
}

// idPattern is what a reserved id has to look like to be one a heading could
// otherwise have taken: the characters a slug is made of.
var idPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

// Validate reports what the configuration cannot be used with.
//
// It is called by New, so an application with a setting that cannot work fails
// where it is wired rather than on the first body that needed it.
func (c Config) Validate() error {
	if c.TopHeading < 0 || c.TopHeading > 6 {
		return fmt.Errorf("markdown: Config.TopHeading is %d, and has to be between 1 and 6, or 0 for %d", c.TopHeading, DefaultTopHeading)
	}
	if c.WordsPerMinute < 0 || c.WordsPerMinute > MaxWordsPerMinute {
		return fmt.Errorf("markdown: Config.WordsPerMinute is %d, and has to be between 0 and %d, where 0 means %d", c.WordsPerMinute, MaxWordsPerMinute, DefaultWordsPerMinute)
	}
	for _, id := range c.Reserved {
		if !idPattern.MatchString(id) {
			return fmt.Errorf("markdown: Config.Reserved holds %q, which no heading could take: an id is a letter followed by letters, digits, - and _", id)
		}
	}
	for _, origin := range c.ImageOrigins {
		if _, err := bareOrigin(origin); err != nil {
			return err
		}
	}
	return nil
}

// bareOrigin is an https origin with nothing after it, lowercased.
//
// The same shape a content security policy source takes. A path would read as
// a prefix it is not -- an origin is allowed whole or not at all -- and plain
// http would let a network in the middle choose the picture.
func bareOrigin(origin string) (string, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Opaque != "" ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.ForceQuery ||
		strings.HasSuffix(origin, "#") {
		return "", fmt.Errorf("markdown: Config.ImageOrigins holds %q, which is not a bare https origin such as https://cdn.example.com", origin)
	}
	return "https://" + strings.ToLower(u.Host), nil
}

// withDefaults returns the configuration with the optional fields filled in.
//
// It runs after Validate and never before: filling a default in first would
// hide the value somebody actually wrote from the check that would have refused
// it.
func (c Config) withDefaults() Config {
	if c.TopHeading == 0 {
		c.TopHeading = DefaultTopHeading
	}
	if c.WordsPerMinute == 0 {
		c.WordsPerMinute = DefaultWordsPerMinute
	}
	return c
}
