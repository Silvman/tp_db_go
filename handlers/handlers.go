package handlers

import (
	"github.com/Silvman/tech-db-forum/mapper"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
)

var DB *mapper.HandlerDB

func WriteResponseJSON(ctx *fasthttp.RequestCtx, status int, body []byte) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.Write(body)
}

func WriteBlankJSON(ctx *fasthttp.RequestCtx, status int) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
}

func WriteErrorJSON(ctx *fasthttp.RequestCtx, status int, err error) {
	msg := models.Error{Message: err.Error()}
	body, _ := msg.MarshalBinary()
	WriteResponseJSON(ctx, status, body)
}
