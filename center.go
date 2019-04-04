package center

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Handler func(http.ResponseWriter, *http.Request, url.Values)
type Middleware func(Handler) Handler

// 表示字典树的一个节点
type Node struct {
	Data       string // 该节点表示的路径
	isVar      bool   // 该节点的路径是否是命名参数形式的
	Children   []*Node
	HandlerMap map[string]Handler //存储http方法到Handler的映射
}

// 路由, 支持全局的中间件
type Router struct {
	Tree        *Node
	middlewares []Middleware
}

func NewRouter() *Router {
	node := Node{Data: "/", isVar: false, HandlerMap: make(map[string]Handler)}
	return &Router{Tree: &node}
}

// 使用哪些中间件
func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// 增加一个路由
func (r *Router) AddRoute(method, path string, handler Handler) {
	if len(path) < 1 || path[0] != '/' {
		panic("invalid path")
	}
	r.Tree.AddNode(method, path, handler)

}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	params := req.Form
	node, _ := r.Tree.FindNode(strings.Split(req.URL.Path, "/")[1:], params)
	if handler, ok := node.HandlerMap[req.Method]; ok {
		handler = Chain(handler, r.middlewares...)
		handler(w, req, params)
	} else {
		NotFoundHanler(w, req)
	}
}

// 遍历字典树，寻找URL路径对应的节点，找到的节点要么正好是URL对应的节点，要么是url前边若干部分对应的节点
// 可以根据找到的节点存储的data值和URL路径的最后一部分进行对比来判断是否正好是完美的对应，还是只是前边若干部分的对应

func (n *Node) FindNode(parts []string, params url.Values) (*Node, string) {
	if len(n.Children) > 0 {
		for _, child := range n.Children {
			if child.Data == parts[0] || child.isVar {
				if child.isVar && params != nil {
					params.Add(child.Data[1:], parts[0]) // 将URL中的变量参数提取出来
				}
				left := parts[1:]
				if len(left) > 0 {
					return child.FindNode(left, params)
				} else {
					return child, parts[0]
				}
			}
		}
	}
	return n, parts[0]
}

// 将新的URL按/分成各个部分，然后往字典树上增加还不存在的节点，注意到最后一部分时，需要设置该节点的HandlerMap
func (n *Node) AddNode(method, path string, handler Handler) {
	parts := strings.Split(path, "/")[1:]
	total := len(parts)
	for i := 0; i < total; i++ {
		pNode, _ := n.FindNode(parts, nil)
		current := parts[i]
		if pNode.Data == current && i == total-1 {
			pNode.HandlerMap[method] = handler
			return
		}
		newNode := Node{Data: current, isVar: false, HandlerMap: make(map[string]Handler)}
		if len(current) > 0 && current[0] == ':' {
			newNode.isVar = true
		}
		if i == total-1 {
			newNode.HandlerMap[method] = handler
		}
		pNode.Children = append(pNode.Children, &newNode)
	}

}

func NotFoundHanler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(404)
	log.Println(req.RequestURI, "404 Not Found")
	fmt.Fprint(w, "Not Found")
}
