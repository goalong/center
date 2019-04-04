package center_test

import (
	"fmt"
	"github.com/goalong/center"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

const indexString = "Hey, you, Welcome!"

func Index(w http.ResponseWriter, r *http.Request, params url.Values) {
	fmt.Fprintf(w, indexString)
}

func Hello(w http.ResponseWriter, r *http.Request, params url.Values) {
	name := ""
	if _, ok := params["name"]; ok {
		name = params["name"][0]
	}
	fmt.Fprintf(w, name)
}

func TestAddNode(t *testing.T) {
	mux := center.NewRouter()
	method := "GET"
	path := "/index"
	mux.AddRoute(method, path, Index)
	success := false
	for _, child := range mux.Tree.Children {
		if child.Data == path[1:] {
			success = true
		}
	}
	if !success {
		t.Error("failed")
	} else {
		t.Log("success")
	}

}

func TestFindNode(t *testing.T) {
	mux := center.NewRouter()
	method := "GET"
	path := "/index"
	mux.AddRoute(method, path, Index)
	node, _ := mux.Tree.FindNode(strings.Split(path, "/")[1:], nil)
	if handler, ok := node.HandlerMap[method]; ok {
		if reflect.ValueOf(handler).Pointer() == reflect.ValueOf(Index).Pointer() {
			t.Log("success")
		} else {
			t.Error("failed")
		}
	} else {
		t.Error("failed")
	}
}

func TestAddRoute(t *testing.T) {
	req, err := http.NewRequest("GET", "/index", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	//直接使用HealthCheckHandler，传入参数rr,req
	Index(rr, req, nil)
	if status := rr.Code; status != http.StatusOK {
		t.Error("failed")
	}
	if rr.Body.String() == indexString {
		t.Log("success")

	} else {
		t.Error("failed")
	}
}

func TestAddRoute2(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello/:name", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	userName := "Monica"
	params := url.Values{"name": []string{userName}}
	Hello(rr, req, params)
	if status := rr.Code; status != http.StatusOK {
		t.Error("failed")
	}
	if rr.Body.String() == userName {
		t.Log("success")

	} else {
		t.Error("failed")
	}
}
