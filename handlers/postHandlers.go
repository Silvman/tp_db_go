package handlers

import (
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
	"strings"
)

func GetPostDetails(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))
	relatedVal := string(ctx.QueryArgs().Peek("related"))

	related := strings.Split(relatedVal, ",")
	//
	//var related []string
	//for _, value := range relatedBytes {
	//	related = append(related, string(value))
	//	log.Println(string(value))
	//}

	payload, err := DB.PostGetOne(id, related)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, err := payload.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func UpdatePostDetails(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))

	var pendingUpdate models.PostUpdate
	pendingUpdate.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.PostUpdate(id, &pendingUpdate)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, err := payload.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}

	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}
