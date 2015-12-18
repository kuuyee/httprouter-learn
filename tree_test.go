package httprouter

import (
	"fmt"
	"net/http"
	"testing"
)

type testRoute struct {
	path     string
	conflict bool
}

var h Handle = func(http.ResponseWriter, *http.Request, Params) {}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		fmt.Println(route.path)
		recv := catchPanic(func() {
			tree.addRoute(route.path, h)
		})
		if route.conflict {
			if recv == nil {
				fmt.Printf("捕获到panic no panic for conflicting route '%s'", route.path)
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			fmt.Printf("捕获到panic unexpected panic for route '%s': %v", route.path, recv)
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		//{"/search/", false},
		//{"/support/", false},
		//{"/blog/:post/", false},
		//{"/about-us/team/", false},
		//{"contact", false},
		{"/cmd/veta", false},
		{"cmd/vetb/:sub", false},
		//{"/cmd/:tool/:sub", true},
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
