package handlers

import (
	"encoding/json"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/valyala/fasthttp"
)

func CreateUser(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)
	var pendingUser models.User
	pendingUser.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.UserCreate(nickname, &pendingUser)
	if err != nil {
		if payload != nil {
			body, _ := json.Marshal(payload)
			WriteResponseJSON(ctx, fasthttp.StatusConflict, body)
		}
		return
	}

	body, _ := pendingUser.MarshalBinary()
	WriteResponseJSON(ctx, fasthttp.StatusCreated, body)
}

func GetUserDetails(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)

	payload, err := DB.UserGetOne(nickname)
	if err != nil {
		WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		return
	}

	body, _ := payload.MarshalBinary()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}

func UpdateUserDetails(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)
	var pendingUser models.UserUpdate
	pendingUser.UnmarshalBinary(ctx.PostBody())

	payload, err := DB.UserUpdate(nickname, &pendingUser)
	if err != nil {
		if err.Error() == "Can't find user by nickname" {
			WriteErrorJSON(ctx, fasthttp.StatusNotFound, err)
		} else {
			WriteErrorJSON(ctx, fasthttp.StatusConflict, err)
		}
		return
	}

	body, _ := payload.MarshalBinary()
	WriteResponseJSON(ctx, fasthttp.StatusOK, body)
}
