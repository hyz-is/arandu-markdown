package feature_test

import (
	"testing"

	"github.com/arandu-io/framework/foundation"
	fhttp "github.com/arandu-io/framework/http"

	markdown "github.com/hyz-is/arandu-markdown"
)

// These tests drive the module the way an application does: build it and
// register it on a router.

func TestTheModuleIsRegisteredAsAModule(t *testing.T) {
	t.Parallel()

	m, err := markdown.New(markdown.Config{})
	if err != nil {
		t.Fatalf("building the module: %v", err)
	}
	var module foundation.Module = m
	if module.Name() != "markdown" {
		t.Fatalf("Name() = %q, want markdown: route names and the boot log are keyed by it", module.Name())
	}
}

// TestTheModuleServesNoRoute holds the reason Routes is empty: a body reaches
// a page through the controller that authorized its entry, and a route here
// would render text for anybody, with no entry behind it.
func TestTheModuleServesNoRoute(t *testing.T) {
	t.Parallel()

	m, err := markdown.New(markdown.Config{})
	if err != nil {
		t.Fatalf("building the module: %v", err)
	}
	router := fhttp.NewRouter()
	m.Routes(router.ForModule(m.Name()))
	if routes := router.Routes(); len(routes) != 0 {
		t.Fatalf("the module registered %d routes, want none: %+v", len(routes), routes)
	}
}
