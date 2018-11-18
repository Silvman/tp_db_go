package main

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/handlers"
	"github.com/Silvman/tech-db-forum/mapper"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
)

type MyRouter struct {
	*fasthttprouter.Router
	BaseUrl string
}

func NewRouter() *MyRouter {
	return &MyRouter{fasthttprouter.New(), ""}
}

func (r *MyRouter) POST(path string, request fasthttp.RequestHandler) *MyRouter {
	r.Router.POST(r.BaseUrl+path, request)
	return r
}

func (r *MyRouter) GET(path string, request fasthttp.RequestHandler) *MyRouter {
	r.Router.GET(r.BaseUrl+path, request)
	return r
}

func HelloApi(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello api")
	log.Println("HelloApi")
}

func main() {
	var err error
	handlers.DB, err = mapper.NewHandler()
	if err != nil {
		log.Fatalln(err)
		return
	}

	router := NewRouter()
	router.BaseUrl = "/api"

	router.GET("/", HelloApi)

	router.
		POST("/forum/:slug_action", handlers.CreateForum).
		GET("/forum/:slug_action/details", handlers.GetForumDetails).
		POST("/forum/:slug_action/create", handlers.CreateThread).
		GET("/forum/:slug_action/users", handlers.GetForumUsers).
		GET("/forum/:slug_action/threads", handlers.GetForumThreads)

	router.
		POST("/post/:id/details", handlers.UpdatePostDetails).
		GET("/post/:id/details", handlers.GetPostDetails)

	router.
		POST("/service/clear", handlers.Clear).
		GET("/service/status", handlers.Status)

	router.
		POST("/thread/:slug_id/create", handlers.CreatePosts).
		GET("/thread/:slug_id/details", handlers.GetThreadDetails).
		POST("/thread/:slug_id/details", handlers.UpdateThreadDetails).
		GET("/thread/:slug_id/posts", handlers.GetPosts).
		POST("/thread/:slug_id/vote", handlers.CreateThreadVote)

	router.
		POST("/user/:nickname/create", handlers.CreateUser).
		GET("/user/:nickname/profile", handlers.GetUserDetails).
		POST("/user/:nickname/profile", handlers.UpdateUserDetails)

	log.Println("starting server at :5000")
	if err := fasthttp.ListenAndServe(":5000", router.Handler); err != nil {
		log.Fatalln(err)
	}

	// todo gshtd
}
