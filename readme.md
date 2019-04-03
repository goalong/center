### 简介
这是一个简单的Go web框架，具有如下特性：
* 动态路由，支持如/post/:id这种带命名参数的形式
* 中间件支持，自带了log和返回json形式等中间件，自定义中间件也极为简单
* 完全使用标准库的接口，无第三方库的依赖
* 代码量少，便于理解

### 示例

```go
package main

import (
	"center/center"
	"encoding/json"
	"fmt"
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