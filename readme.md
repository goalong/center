### 简介
这是一个简单的Go web框架，具有如下特性：
* 动态路由，支持如/post/:id这种带命名参数的形式
* 中间件支持，自带了log和返回json形式等中间件，支持全局以及单个接口绑定的中间件，自定义中间件也极为简单
* 完全使用标准库的接口，无第三方库的依赖
* 代码量少，便于理解

### 核心结构
路由使用的是类似字典树的数据结构，每个节点用如下的Node结构表示：
```
type Node struct {
	Data       string // 该节点表示的路径
	isVar      bool   // 该节点的路径是否是命名参数形式的
	Children   []*Node // 子节点
	HandlerMap map[string]Handler //存储http方法到Handler的映射
}
```

### 注册路由
使用的是AddRoute方法：
```
func (r *Router) AddRoute(method, path string, handler Handler)
```

例如给/articles/:id 这个URL注册路由，首先会将URL按/分成几个部分，这个例子中
会分成articles和:id两个部分，对每一部分，在树上进行查找，如果没有就需要添加一个节点，
如果某一部分是以冒号开头的，例如:id, 这种节点会将isVar字段设为true,
对于每个URL的最后一部分，需要将Handler放到对应节点的HandlerMap中，HandlerMap的键为HTTP方法，
值为给该URL注册的Handler.

### 路由分发
主要由FindNode方法来完成：
```
func (n *Node) FindNode(parts []string, params url.Values) (*Node, string)
```

会根据URL，按照/分成几个部分，然后逐层在树上进行查找，直到找到完全契合的节点或者发现找不到，
找不到就交给NotFoundHanler处理，找到的话就可以在目标节点的HandlerMap中根据请求的方法找到对应的Handler


### 使用
详见下面使用mongo、mgo以及此框架构建的Restful API, 代码在examples/demo.go文件中。

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/goalong/center"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"time"
)

type Book struct {
	Id         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Title      string        `json:"title"`
	Author     string        `json:"author"`
	CreateTime int64         `json:"create_time"`
}

const URL string = "127.0.0.1:27017"

var collection *mgo.Collection
var session *mgo.Session

func init() {
	session, _ = mgo.Dial(URL)
	db := session.DB("test")
	collection = db.C("book")
}

// 添加一本书
func AddBook(w http.ResponseWriter, r *http.Request, params url.Values) {
	var book Book
	json.NewDecoder(r.Body).Decode(&book)
	book.CreateTime = time.Now().Unix()
	err := collection.Insert(book)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	json.NewEncoder(w).Encode(book)

}

// 获取书籍列表
func GetBooks(w http.ResponseWriter, r *http.Request, params url.Values) {
	var books []Book
	iter := collection.Find(nil).Limit(10).Iter()
	err := iter.All(&books)
	if err != nil {
		panic(err)
	}
	ret := make(map[string]interface{})
	ret["items"] = books
	json.NewEncoder(w).Encode(ret)

}

// 获取一本书的详情
func GetBook(w http.ResponseWriter, r *http.Request, params url.Values) {
	var book Book
	id := params["id"][0]
	ok := bson.IsObjectIdHex(id)
	if !ok {
		w.WriteHeader(404)
	}
	err := collection.FindId(bson.ObjectIdHex(id)).One(&book)
	if err != nil {
		if err.Error() == "not found" {
			w.WriteHeader(404)
		} else {
			panic(err.Error())
		}

	}
	json.NewEncoder(w).Encode(book)

}

// 更新一本书的信息
func UpdateBook(w http.ResponseWriter, r *http.Request, params url.Values) {
	var newBook, originBook Book
	id := params["id"][0]
	err := collection.FindId(bson.ObjectIdHex(id)).One(&originBook)
	if err != nil {
		if err.Error() == "not found" {
			w.WriteHeader(404)
		} else {
			panic(err.Error())
		}
	}
	json.NewDecoder(r.Body).Decode(&newBook)
	originBook.Title = newBook.Title
	originBook.Author = newBook.Author
	err = collection.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{
		"$set": bson.M{"title": newBook.Title, "author": newBook.Author}})
	if err != nil {
		panic(err.Error())
	}
	json.NewEncoder(w).Encode(map[string]interface{}{})

}

// 删除一本书
func DeleteBook(w http.ResponseWriter, r *http.Request, params url.Values) {
	id := params["id"][0]
	err := collection.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		panic(err.Error())
	}
	json.NewEncoder(w).Encode(map[string]interface{}{})

}

func main() {

	r := center.NewRouter()
	// 使用中间件
	r.Use(center.Logging(), center.Recover(), center.ReturnJson())
	r.AddRoute("POST", "/books", AddBook)
	r.AddRoute("GET", "/books", GetBooks)
	r.AddRoute("GET", "/books/:id", GetBook)
	r.AddRoute("PUT", "/books/:id", UpdateBook)
	r.AddRoute("DELETE", "/books/:id", DeleteBook)
	http.ListenAndServe(":8000", r)
}
```