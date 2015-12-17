package httprouter

import (
	"fmt"
	"testing"
)

type testRoute struct {
	path     string
	conflict bool
}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		fmt.Println(route.path)
		recv := catchPanic(func() {
			tree.addRoute(route.path, nil)
		})
		if route.conflict {
			if recv == nil {
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		fmt.Println("捕获到panic")
		recv = recover()
	}()

	testFunc()
	return
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		{"/cmd/vet", false},
		{"/cmd/vet/:sub", false},
		{"/cmd/:tool/:sub", true},
		//{"/admin/:category/:page", false},
		//{"/src/AUTHORS", false},
		//{"/src/*filepath", true},
		//{"/user_x", false},
		//{"/user_:name", true},
		//{"/id/:id", false},
		//{"/id:id", true},
		//{"/:id", true},
		//{"/*/", true},
		//{"/*filepath", true},
	}
	testRoutes(t, routes)
}
