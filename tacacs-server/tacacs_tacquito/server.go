package tacacs_tacquito

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"tacacs-zitadel-server/auth"
	"tacacs-zitadel-server/config"
	"tacacs-zitadel-server/zitadel"

	tq "github.com/facebookincubator/tacquito"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
}

func (l *Logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *Logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *Logger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *Logger) Record(ctx context.Context, r map[string]string, obscure ...string) {
	fields := logrus.Fields{}
	for k, v := range r {
		fields[k] = v
	}
	l.logger.WithFields(fields).Info("TACACS+ Record")
}

type SecretProvider struct {
	secret  string
	handler tq.Handler
}

func (sp *SecretProvider) Get(ctx context.Context, remote net.Addr) ([]byte, tq.Handler, error) {
	return []byte(sp.secret), sp.handler, nil
}

type TacacsServer struct {
	config         *config.Config
	logger         *Logger
	authProvider   auth.AuthProvider
	db             *sql.DB
	server         *tq.Server
	listener       net.Listener
	wg             sync.WaitGroup
	stopChan       chan struct{}
	sessions       map[string]*Session
	sessionsMutex  sync.RWMutex
}

type Session struct {
	ID        string
	Username  string
	ClientIP  string
	Roles     []string
	StartTime time.Time
	Commands  []Command
	Active    bool
}

type Command struct {
	Command   string
	Timestamp time.Time
	Allowed   bool
}

func NewTacacsServer(cfg *config.Config, logger *logrus.Logger) (*TacacsServer, error) {
	// Create Zitadel auth provider
	authProvider, err := zitadel.NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zitadel client: %w", err)
	}
	logger.Info("Using Zitadel as authentication provider")

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	tqLogger := &Logger{logger: logger}
	
	ts := &TacacsServer{
		config:       cfg,
		logger:       tqLogger,
		authProvider: authProvider,
		db:           db,
		stopChan:     make(chan struct{}),
		sessions:     make(map[string]*Session),
	}

	// Create router handler
	routerHandler := NewRouterHandler(ts)
	
	secretProvider := &SecretProvider{
		secret:  cfg.TACACSSecret,
		handler: routerHandler,
	}

	// Create server
	server := tq.NewServer(tqLogger, secretProvider)
	ts.server = server

	if err := ts.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	go ts.cleanupRoutine()

	return ts, nil
}

func (ts *TacacsServer) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tacacs_sessions (
			id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			client_ip VARCHAR(45) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tacacs_commands (
			id SERIAL PRIMARY KEY,
			session_id VARCHAR(255) REFERENCES tacacs_sessions(id),
			command TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			allowed BOOLEAN NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_username ON tacacs_sessions(username)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON tacacs_sessions(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_commands_session_id ON tacacs_commands(session_id)`,
	}

	for _, query := range queries {
		if _, err := ts.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (ts *TacacsServer) Start() error {
	listener, err := net.Listen("tcp", ts.config.TACACSListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ts.config.TACACSListenAddress, err)
	}

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("listener must be a TCP listener")
	}

	ts.listener = listener
	ts.logger.Infof(context.Background(), "TACACS+ server started on %s", ts.config.TACACSListenAddress)

	ctx := context.Background()
	return ts.server.Serve(ctx, tcpListener)
}

func (ts *TacacsServer) Stop() error {
	close(ts.stopChan)
	
	if ts.listener != nil {
		ts.listener.Close()
	}
	
	ts.wg.Wait()
	
	if ts.db != nil {
		ts.db.Close()
	}
	
	return nil
}

func (ts *TacacsServer) recordSession(session *Session) {
	query := `INSERT INTO tacacs_sessions (id, username, client_ip, start_time, status) 
			  VALUES ($1, $2, $3, $4, $5)`
	
	_, err := ts.db.Exec(query, session.ID, session.Username, session.ClientIP, session.StartTime, "active")
	if err != nil {
		ts.logger.Errorf(context.Background(), "Failed to record session: %v", err)
	}
}

func (ts *TacacsServer) recordCommand(username, command string, allowed bool) {
	sessionID := ts.findActiveSession(username)
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_unknown_%d", username, time.Now().Unix())
	}

	query := `INSERT INTO tacacs_commands (session_id, command, timestamp, allowed) 
			  VALUES ($1, $2, $3, $4)`
	
	_, err := ts.db.Exec(query, sessionID, command, time.Now(), allowed)
	if err != nil {
		ts.logger.Errorf(context.Background(), "Failed to record command: %v", err)
	}
}

func (ts *TacacsServer) findActiveSession(username string) string {
	ts.sessionsMutex.RLock()
	defer ts.sessionsMutex.RUnlock()
	
	for sessionID, session := range ts.sessions {
		if session.Username == username && session.Active {
			return sessionID
		}
	}
	return ""
}

func (ts *TacacsServer) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ts.stopChan:
			return
		case <-ticker.C:
			ts.authProvider.CleanupCache()
			ts.cleanupExpiredSessions()
		}
	}
}

func (ts *TacacsServer) cleanupExpiredSessions() {
	ts.sessionsMutex.Lock()
	defer ts.sessionsMutex.Unlock()
	
	timeout := time.Duration(ts.config.SessionTimeout) * time.Second
	cutoff := time.Now().Add(-timeout)
	
	for sessionID, session := range ts.sessions {
		if session.StartTime.Before(cutoff) {
			session.Active = false
			delete(ts.sessions, sessionID)
			
			query := `UPDATE tacacs_sessions SET end_time = $1, status = $2 WHERE id = $3`
			ts.db.Exec(query, time.Now(), "expired", sessionID)
		}
	}
}

// Helper function to get client IP from connection
func getClientIP(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	if strings.Contains(addr, ":") {
		return strings.Split(addr, ":")[0]
	}
	return addr
}