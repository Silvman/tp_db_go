// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/Silvman/tech-db-forum/modules/service"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
)

//go:generate swagger generate server --target .. --name Forum --spec ../swagger.yml

func configureFlags(api *operations.ForumAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ForumAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.BinConsumer = runtime.ByteStreamConsumer()

	api.JSONProducer = runtime.JSONProducer()

	handler, err := service.NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	api.ClearHandler = operations.ClearHandlerFunc(handler.Clear)
	api.StatusHandler = operations.StatusHandlerFunc(handler.Status)

	api.ForumCreateHandler = operations.ForumCreateHandlerFunc(handler.ForumCreate)
	api.ForumGetOneHandler = operations.ForumGetOneHandlerFunc(handler.ForumGetOne)
	api.ForumGetThreadsHandler = operations.ForumGetThreadsHandlerFunc(handler.ForumGetThreads)
	api.ForumGetUsersHandler = operations.ForumGetUsersHandlerFunc(handler.ForumGetUsers)

	api.PostGetOneHandler = operations.PostGetOneHandlerFunc(handler.PostGetOne)
	api.PostUpdateHandler = operations.PostUpdateHandlerFunc(handler.PostUpdate)
	api.PostsCreateHandler = operations.PostsCreateHandlerFunc(handler.PostsCreate)

	api.UserCreateHandler = operations.UserCreateHandlerFunc(handler.UserCreate)
	api.UserGetOneHandler = operations.UserGetOneHandlerFunc(handler.UserGetOne)
	api.UserUpdateHandler = operations.UserUpdateHandlerFunc(handler.UserUpdate)

	api.ThreadCreateHandler = operations.ThreadCreateHandlerFunc(handler.ThreadCreate)
	api.ThreadGetOneHandler = operations.ThreadGetOneHandlerFunc(handler.ThreadGetOne)
	api.ThreadUpdateHandler = operations.ThreadUpdateHandlerFunc(handler.ThreadUpdate)
	api.ThreadVoteHandler = operations.ThreadVoteHandlerFunc(handler.ThreadVote)
	api.ThreadGetPostsHandler = operations.ThreadGetPostsHandlerFunc(handler.ThreadGetPosts)

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
