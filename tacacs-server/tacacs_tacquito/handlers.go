package tacacs_tacquito

import (
	"fmt"
	"time"

	tq "github.com/facebookincubator/tacquito"
)

type AuthHandler struct {
	server *TacacsServer
}

func NewAuthHandler(server *TacacsServer) *AuthHandler {
	return &AuthHandler{server: server}
}

func (h *AuthHandler) Handle(response tq.Response, request tq.Request) {
	var body tq.AuthenStart
	if err := tq.Unmarshal(request.Body, &body); err != nil {
		h.server.logger.Errorf(request.Context, "Failed to unmarshal authentication start: %v", err)
		response.Reply(tq.NewAuthenReply(
			tq.SetAuthenReplyStatus(tq.AuthenStatusError),
			tq.SetAuthenReplyServerMsg("Invalid authentication request"),
		))
		return
	}

	username := string(body.User)
	password := string(body.Data)

	h.server.logger.Infof(request.Context, "Authentication request for user: %s", username)

	// Authenticate with auth provider
	userInfo, err := h.server.authProvider.AuthenticateUser(request.Context, username, password)
	if err != nil {
		h.server.logger.Errorf(request.Context, "Authentication failed for user %s: %v", username, err)
		response.Reply(tq.NewAuthenReply(
			tq.SetAuthenReplyStatus(tq.AuthenStatusFail),
			tq.SetAuthenReplyServerMsg("Authentication failed"),
		))
		return
	}

	// Create session
	sessionID := fmt.Sprintf("%s_%d", username, time.Now().Unix())
	session := &Session{
		ID:        sessionID,
		Username:  username,
		ClientIP:  "unknown", // TODO: extract from context
		Roles:     userInfo.Roles,
		StartTime: time.Now(),
		Commands:  []Command{},
		Active:    true,
	}

	h.server.sessionsMutex.Lock()
	h.server.sessions[sessionID] = session
	h.server.sessionsMutex.Unlock()

	h.server.recordSession(session)

	h.server.logger.Infof(request.Context, "User %s authenticated successfully with roles: %v", userInfo.Username, userInfo.Roles)

	response.Reply(tq.NewAuthenReply(
		tq.SetAuthenReplyStatus(tq.AuthenStatusPass),
		tq.SetAuthenReplyServerMsg("Authentication successful"),
	))
}

type AuthorHandler struct {
	server *TacacsServer
}

func NewAuthorHandler(server *TacacsServer) *AuthorHandler {
	return &AuthorHandler{server: server}
}

func (h *AuthorHandler) Handle(response tq.Response, request tq.Request) {
	var body tq.AuthorRequest
	if err := tq.Unmarshal(request.Body, &body); err != nil {
		h.server.logger.Errorf(request.Context, "Failed to unmarshal authorization request: %v", err)
		response.Reply(tq.NewAuthorReply(
			tq.SetAuthorReplyStatus(tq.AuthorStatusError),
			tq.SetAuthorReplyServerMsg("Invalid authorization request"),
		))
		return
	}

	username := string(body.User)
	
	// Extract command from args field
	var command string
	if len(body.Args) > 0 {
		command = string(body.Args[0])
	} else {
		command = "unknown"
	}

	h.server.logger.Infof(request.Context, "Authorization request for user %s, command: %s", username, command)

	// Get user roles from active session
	h.server.sessionsMutex.RLock()
	var userRoles []string
	for _, session := range h.server.sessions {
		if session.Username == username && session.Active {
			userRoles = session.Roles
			break
		}
	}
	h.server.sessionsMutex.RUnlock()

	// If no active session found, deny access
	if len(userRoles) == 0 {
		h.server.logger.Errorf(request.Context, "No active session found for user %s", username)
		response.Reply(tq.NewAuthorReply(
			tq.SetAuthorReplyStatus(tq.AuthorStatusFail),
			tq.SetAuthorReplyServerMsg("No active session"),
		))
		return
	}

	// Check authorization using auth provider
	allowed := h.server.authProvider.IsAuthorized(userRoles, command)
	h.server.recordCommand(username, command, allowed)

	if allowed {
		h.server.logger.Infof(request.Context, "Authorization granted for user %s, command: %s", username, command)
		response.Reply(tq.NewAuthorReply(
			tq.SetAuthorReplyStatus(tq.AuthorStatusPassAdd),
			tq.SetAuthorReplyServerMsg("Authorization granted"),
		))
	} else {
		h.server.logger.Infof(request.Context, "Authorization denied for user %s, command: %s", username, command)
		response.Reply(tq.NewAuthorReply(
			tq.SetAuthorReplyStatus(tq.AuthorStatusFail),
			tq.SetAuthorReplyServerMsg("Authorization denied"),
		))
	}
}

type AcctHandler struct {
	server *TacacsServer
}

func NewAcctHandler(server *TacacsServer) *AcctHandler {
	return &AcctHandler{server: server}
}

func (h *AcctHandler) Handle(response tq.Response, request tq.Request) {
	var body tq.AcctRequest
	if err := tq.Unmarshal(request.Body, &body); err != nil {
		h.server.logger.Errorf(request.Context, "Failed to unmarshal accounting request: %v", err)
		response.Reply(tq.NewAcctReply(
			tq.SetAcctReplyStatus(tq.AcctReplyStatusError),
			tq.SetAcctReplyServerMsg("Invalid accounting request"),
		))
		return
	}

	username := string(body.User)
	h.server.logger.Infof(request.Context, "Accounting request for user %s", username)

	// For accounting, we just log the session info
	switch {
	case body.Flags.Has(tq.AcctFlagStart):
		h.server.logger.Infof(request.Context, "Session started for user %s", username)
	case body.Flags.Has(tq.AcctFlagStop):
		h.server.logger.Infof(request.Context, "Session stopped for user %s", username)
		
		// Mark session as inactive
		h.server.sessionsMutex.Lock()
		for sessionID, session := range h.server.sessions {
			if session.Username == username && session.Active {
				session.Active = false
				query := `UPDATE tacacs_sessions SET end_time = $1, status = $2 WHERE id = $3`
				h.server.db.Exec(query, time.Now(), "completed", sessionID)
				break
			}
		}
		h.server.sessionsMutex.Unlock()
	case body.Flags.Has(tq.AcctFlagWatchdog):
		h.server.logger.Debugf(request.Context, "Watchdog update for user %s", username)
	}

	response.Reply(tq.NewAcctReply(
		tq.SetAcctReplyStatus(tq.AcctReplyStatusSuccess),
		tq.SetAcctReplyServerMsg("Accounting recorded"),
	))
}