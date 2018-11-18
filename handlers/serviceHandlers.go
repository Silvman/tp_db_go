package handlers

import "github.com/valyala/fasthttp"

func Clear(ctx *fasthttp.RequestCtx) {
	DB.Clear()
	WriteBlankJSON(ctx, fasthttp.StatusOK)
}

func Status(ctx *fasthttp.RequestCtx) {
	payload := DB.Status()
	body, _ := payload.MarshalBinary()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}
