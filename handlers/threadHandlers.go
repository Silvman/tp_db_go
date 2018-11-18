package handlers

import (
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
)

func CreatePosts(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_id").(string)
	var pendingPosts models.Posts
	pendingPosts.UnmarshalJSON(ctx.PostBody())

	payload, err := DB.PostsCreate(slug, pendingPosts)
	if err != nil {
		if err.Error() == "Parent post was created in another thread" {
			WriteErrorJSON(ctx, fasthttp.StatusConflict, err)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)

		}
		return
	}

	body, _ := payload.MarshalJSON()
	WriteResponseJSON(ctx, fasthttp.StatusCreated, body)
}

func GetThreadDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_id").(string)
	payload, err := DB.ThreadGetOne(slug)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, _ := payload.MarshalJSON()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)

}

func UpdateThreadDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_id").(string)
	var pendingThread models.ThreadUpdate
	pendingThread.UnmarshalJSON(ctx.PostBody())

	payload, err := DB.ThreadUpdate(slug, &pendingThread)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, _ := payload.MarshalJSON()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func GetPosts(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_id").(string)

	sort := string(ctx.QueryArgs().Peek("sort"))

	limitStr := string(ctx.QueryArgs().Peek("limit"))
	var limit *int = nil
	if limitStr == "" {
		limit = new(int)
		*limit = 100
	} else {
		limit = new(int)
		*limit, _ = strconv.Atoi(limitStr)
	}

	var since *int = nil
	sinceInt, err := strconv.Atoi(string(ctx.QueryArgs().Peek("since")))
	if err == nil {
		since = new(int)
		*since = sinceInt
	}
	// todo valid?

	descStr := string(ctx.QueryArgs().Peek("desc"))
	var desc *bool = nil
	if descStr == "true" {
		desc = new(bool)
		*desc = true
	}

	payload, err := DB.ThreadGetPosts(slug, &sort, since, desc, limit)
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

func CreateThreadVote(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug_id").(string)
	var pendingVote models.Vote
	pendingVote.UnmarshalJSON(ctx.PostBody())

	payload, err := DB.ThreadVote(slug, &pendingVote)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, _ := payload.MarshalJSON()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}
