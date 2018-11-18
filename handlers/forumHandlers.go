package handlers

import (
	"errors"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
)

func CreateForum(ctx *fasthttp.RequestCtx) {
	action := ctx.UserValue("slug_action").(string)
	if action != "create" {
		err := errors.New("no such action")
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	var pendingForum models.Forum
	pendingForum.UnmarshalJSON(ctx.PostBody())

	payload, err := DB.ForumCreate(&pendingForum)
	if err != nil {
		if payload != nil {
			body, _ := payload.MarshalJSON()
			WriteResponseJSON(ctx, fasthttp.StatusConflict, body)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		}
		return
	}

	body, _ := payload.MarshalJSON()
	WriteResponseJSON(ctx, fasthttp.StatusCreated, body)
}

func GetForumDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)
	payload, err := DB.ForumGetOne(slug)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}
	body, _ := payload.MarshalJSON()

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func GetForumUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)
	limitStr := string(ctx.QueryArgs().Peek("limit"))

	var limit *int = nil
	if limitStr == "" {
		limit = new(int)
		*limit = 100
	} else {
		limit = new(int)
		*limit, _ = strconv.Atoi(limitStr)
	}

	sinceStr := string(ctx.QueryArgs().Peek("since"))
	var since *string
	if sinceStr != "" {
		since = new(string)
		*since = sinceStr
	} // todo valid?

	descStr := string(ctx.QueryArgs().Peek("desc"))
	var desc *bool = nil
	if descStr == "true" {
		desc = new(bool)
		*desc = true
	}

	payload, err := DB.ForumGetUsers(slug, desc, since, limit)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, err := payload.MarshalJSON()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func GetForumThreads(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)
	limitStr := string(ctx.QueryArgs().Peek("limit"))

	var limit *int = nil
	if limitStr == "" {
		limit = new(int)
		*limit = 100
	} else {
		limit = new(int)
		*limit, _ = strconv.Atoi(limitStr)
	}

	sinceStr := string(ctx.QueryArgs().Peek("since"))
	var since *string
	if sinceStr != "" {
		since = new(string)
		*since = sinceStr
	}
	// todo valid?

	descStr := string(ctx.QueryArgs().Peek("desc"))
	var desc *bool = nil
	if descStr == "true" {
		desc = new(bool)
		*desc = true
	}

	payload, err := DB.ForumGetThreads(slug, desc, since, limit)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, err := payload.MarshalJSON()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func CreateThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)

	var pendingThread models.Thread
	pendingThread.UnmarshalJSON(ctx.PostBody())

	payload, err := DB.ThreadCreate(slug, &pendingThread)
	if err != nil {
		if payload != nil {
			body, _ := payload.MarshalJSON()
			WriteResponseJSON(ctx, fasthttp.StatusConflict, body)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		}
		return
	}

	body, _ := payload.MarshalJSON()
	// todo часть новосозданные значений - копия переданных
	WriteResponseJSON(ctx, fasthttp.StatusCreated, body)
}
