package center_test

import (
	"fmt"
	"github.com/goalong/center"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Index(w http.ResponseWriter, r *http.Request, params url.Values) {
	fmt.Fprintf(w, "index")

}



func TestAddRoute(t *testing.T) {
	mux := center.NewRouter()
	method := "GET"
	path := "/index"
	mux.AddRoute(method, path, Index)
	node, _ := mux.Tree.FindNode(strings.Split(path, "/")[1:], nil)
	if handler, ok := node.HandlerMap[method]; ok {
		if reflect.ValueOf(handler).Pointer() == reflect.ValueOf(Index).Pointer() {
			t.Log("ok")
		} else {
			t.Error("wrong")
		}
	} else {
		t.Error("wrong")
	}


}


func TestAddNode(t *testing.T) {


}


func TestFindNode(t *testing.T) {

}