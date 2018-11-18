package handlers

import (
	"encoding/json"
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
	pendingForum.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.ForumCreate(&pendingForum)
	if err != nil {
		if payload != nil {
			body, _ := payload.MarshalBinary()
			WriteResponseJSON(ctx, fasthttp.StatusConflict, body)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		}
		return
	}

	//body, _ := payload.MarshalBinary()
	// новосозданный форум имеет ровно те значения, которые мы ему передали
	WriteResponseJSON(ctx, fasthttp.StatusCreated, ctx.PostBody())
}

func GetForumDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)
	payload, err := DB.ForumGetOne(slug)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
	}
	body, _ := payload.MarshalBinary()

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func GetForumUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)
	limitStr := string(ctx.QueryArgs().Peek("limit"))

	var limit *int = nil
	if limitStr == "" {
		*limit = 100
	} else {
		*limit, _ = strconv.Atoi(limitStr)
	}

	since := string(ctx.QueryArgs().Peek("since"))
	// todo valid?

	descStr := string(ctx.QueryArgs().Peek("desc"))
	var desc *bool = nil
	if descStr == "true" {
		*desc = true
	}

	payload, err := DB.ForumGetUsers(slug, desc, &since, limit)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
	}

	body, err := json.Marshal(payload)
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
		*limit = 100
	} else {
		*limit, _ = strconv.Atoi(limitStr)
	}

	since := string(ctx.QueryArgs().Peek("since"))
	// todo valid?

	descStr := string(ctx.QueryArgs().Peek("desc"))
	var desc *bool = nil
	if descStr == "true" {
		*desc = true
	}

	payload, err := DB.ForumGetThreads(slug, desc, &since, limit)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func CreateThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_action").(string)

	var pendingThread models.Thread
	pendingThread.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.ThreadCreate(slug, &pendingThread)
	if err != nil {
		if payload != nil {
			body, _ := payload.MarshalBinary()
			WriteResponseJSON(ctx, fasthttp.StatusConflict, body)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		}
		return
	}

	body, _ := payload.MarshalBinary()
	// todo часть новосозданные значений - копия переданных
	WriteResponseJSON(ctx, fasthttp.StatusCreated, body)
}
