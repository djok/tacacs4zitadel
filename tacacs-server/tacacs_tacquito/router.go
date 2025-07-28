package tacacs_tacquito

import (
	tq "github.com/facebookincubator/tacquito"
)

type RouterHandler struct {
	server *TacacsServer
	authHandler *AuthHandler
	authorHandler *AuthorHandler
	acctHandler *AcctHandler
}

func NewRouterHandler(server *TacacsServer) *RouterHandler {
	return &RouterHandler{
		server: server,
		authHandler: NewAuthHandler(server),
		authorHandler: NewAuthorHandler(server),
		acctHandler: NewAcctHandler(server),
	}
}

func (r *RouterHandler) Handle(response tq.Response, request tq.Request) {
	switch request.Header.Type {
	case tq.Authenticate:
		r.authHandler.Handle(response, request)
	case tq.Authorize:
		r.authorHandler.Handle(response, request)
	case tq.Accounting:
		r.acctHandler.Handle(response, request)
	default:
		r.server.logger.Errorf(request.Context, "Unknown TACACS+ packet type: %v", request.Header.Type)
		// Just close the connection for unknown types
		return
	}
}