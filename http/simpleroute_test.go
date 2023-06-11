package http

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestSimpleRoute_Route(t *testing.T) {
	req := Request{
		request: &http.Request{},
	}

	req.Request().URL = &url.URL{}
	req.Request().URL.Path = "/auth/group/user/123/book/newbook/name"
	route := NewSimpleRoute()
	route.SetEndpoint("/auth/group/user", new(DefaultHandlerTask))
	route.SetEndpoint("/auth/group/user/book", new(DefaultHandlerTask))
	route.SetEndpoint("/auth/group/user/book/name", new(DefaultHandlerTask))
	route.SetEndpoint("/auth/group/user/profile/info", new(DefaultHandlerTask))
	point, m, isLast := route.RouteNode(req.Url().Path)
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if !isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user/123"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.False(t, isLast)
	assert.Equal(t, "123", m["user_id"])
	assert.Equal(t, "user", point.Name())
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user/123/book"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.True(t, isLast)
	assert.Equal(t, "123", m["user_id"])
	assert.Equal(t, "book", point.Name())
	assert.Equal(t, "user", point.Parent().Name())
	assert.Equal(t, "group", point.Parent().Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().Parent().RouteType())
	assert.Equal(t, "auth", point.Parent().Parent().Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().Parent().Parent().RouteType())
	assert.Equal(t, RouteTypeRootEndPoint, point.Parent().Parent().Parent().Parent().RouteType())
	assert.Nil(t, point.Parent().Parent().Parent().Parent().Parent())
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if !isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user/123/book/newbook"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.False(t, isLast)
	assert.Equal(t, "newbook", m["book_id"])
	assert.Equal(t, "123", m["user_id"])
	assert.Equal(t, "book", point.Name())
	assert.Equal(t, "user", point.Parent().Name())
	assert.Equal(t, "group", point.Parent().Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().Parent().RouteType())
	assert.Equal(t, "auth", point.Parent().Parent().Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().Parent().Parent().RouteType())
	assert.Equal(t, RouteTypeRootEndPoint, point.Parent().Parent().Parent().Parent().RouteType())
	assert.Nil(t, point.Parent().Parent().Parent().Parent().Parent())
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.True(t, isLast)
	assert.Equal(t, "user", point.Name())
	assert.Equal(t, "group", point.Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().RouteType())
	assert.Equal(t, "auth", point.Parent().Parent().Name())
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if !isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user/123/profile/info/myname"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.False(t, isLast)
	assert.Equal(t, "myname", m["info_id"])
	assert.Equal(t, "123", m["user_id"])
	assert.Equal(t, "profile", point.Parent().Name())
	assert.Equal(t, "user", point.Parent().Parent().Name())
	assert.Equal(t, RouteTypeEndPoint, point.Parent().Parent().RouteType())
	assert.Equal(t, "group", point.Parent().Parent().Parent().Name())
	assert.Equal(t, RouteTypeGroup, point.Parent().Parent().Parent().RouteType())
	println(point.Name())
	for k, v := range m {
		println(fmt.Sprintf("%s: %s", k, v))
	}

	if isLast {
		t.Error("")
	}

	println("----")
	req.Request().URL.Path = "/auth/group/user/123/book/newbook/dasdqwe"
	point, m, isLast = route.RouteNode(req.Url().Path)
	assert.Nil(t, point)
	assert.False(t, isLast)
	assert.Nil(t, m)
}
