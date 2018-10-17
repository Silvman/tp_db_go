package service

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/Silvman/tech-db-forum/restapi/operations"
)

func (self HandlerDB) Clear(params operations.ClearParams) middleware.Responder {

	return operations.NewClearOK()
}
func (self HandlerDB) Status(params operations.StatusParams) middleware.Responder {

	return operations.NewStatusOK()
}
