package handlers

import (
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
	"log"
)

func GetPostDetails(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(int)
	relatedBytes := ctx.QueryArgs().PeekMulti("related")
	var related []string
	for _, value := range relatedBytes {
		related = append(related, string(value))
	}

	payload, err := DB.PostGetOne(id, related)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
	}

	body, err := payload.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func UpdatePostDetails(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(int)

	var pendingUpdate models.PostUpdate
	pendingUpdate.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.PostUpdate(id, &pendingUpdate)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
	}

	body, err := payload.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}
