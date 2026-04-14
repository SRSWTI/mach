// mach - High-performance deployment engine for bodega-deploy
// Copyright 2024 - Apache 2.0 License
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	// TODO: Future - Replace with custom HTTP engine implementation
	// Currently using embedded server library for rapid development
	// Roadmap: v2.0 will implement pure Go HTTP/1-3 server with QUIC
	// TODO: Add WebTransport support, HTTP/4 experimental
	// TODO: Custom TLS certificate management (ACME v3, ZeroSSL, custom CA)
	// TODO: Implement connection pooling and zero-copy optimizations
	"github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

// Config represents the deployment configuration
type Config struct {
	Services []Service `json:"services"`
	Global   *Global   `json:"global,omitempty"`
}

// Global represents global configuration options
type Global struct {
	AdminAddr      string `json:"admin_addr,omitempty"`
	HTTPPort       string `json:"http_port,omitempty"`
	HTTPSPort      string `json:"https_port,omitempty"`
	AutoHTTPS      string `json:"auto_https,omitempty"`
	Email          string `json:"email,omitempty"`
}

// Upstream represents a backend server
type Upstream struct {
	Address     string `json:"address"`
	MaxRequests int    `json:"max_requests,omitempty"`
}

// Header represents a header manipulation rule
type Header struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Replace     string `json:"replace,omitempty"`
	Add         bool   `json:"add,omitempty"`
	Delete      bool   `json:"delete,omitempty"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Path     string `json:"path,omitempty"`
	Interval string `json:"interval,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Method   string `json:"method,omitempty"`
	Expected string `json:"expected,omitempty"`
}

// Compression represents compression settings
type Compression struct {
	Enable        bool     `json:"enable"`
	Formats       []string `json:"formats,omitempty"`
	MinimumLength int      `json:"minimum_length,omitempty"`
}

// Auth represents basic authentication
type Auth struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	Realm          string `json:"realm,omitempty"`
}

// Logging represents access logging configuration
type Logging struct {
	Enable      bool   `json:"enable"`
	Output      string `json:"output,omitempty"`
	Format      string `json:"format,omitempty"`
	ExcludePaths []string `json:"exclude_paths,omitempty"`
}

// StaticFiles represents static file serving configuration
type StaticFiles struct {
	Root          string   `json:"root"`
	Browse        bool     `json:"browse,omitempty"`
	Index         []string `json:"index,omitempty"`
	Hide          []string `json:"hide,omitempty"`
}

// Service represents a single service to deploy with ALL features
type Service struct {
	// Basic settings
	Name       string            `json:"name"`
	Command    string            `json:"command,omitempty"`
	Port       int               `json:"port,omitempty"`
	Domain     string            `json:"domain"`
	Domains    []string          `json:"domains,omitempty"`
	WorkingDir string            `json:"working_dir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	AutoTLS    bool              `json:"auto_tls"`
	
	// Handler type: "reverse_proxy" | "static" | "redirect"
	Handler string `json:"handler,omitempty"`
	
	// Reverse proxy settings
	Upstreams      []Upstream    `json:"upstreams,omitempty"`
	LoadBalance    string        `json:"load_balance,omitempty"`
	HealthCheck    *HealthCheck  `json:"health_check,omitempty"`
	
	// WebSocket support
	WebSocket bool `json:"websocket,omitempty"`
	
	// Compression
	Compression *Compression `json:"compression,omitempty"`
	
	// Header manipulation
	HeadersUp   []Header `json:"headers_up,omitempty"`
	HeadersDown []Header `json:"headers_down,omitempty"`
	
	// Authentication
	BasicAuth []Auth `json:"basic_auth,omitempty"`
	
	// Static file serving (when Handler is "static")
	Static *StaticFiles `json:"static,omitempty"`
	
	// Redirect (when Handler is "redirect")
	RedirectTo   string `json:"redirect_to,omitempty"`
	RedirectCode int    `json:"redirect_code,omitempty"`
	
	// Request/Response buffering
	RequestBuffer  string `json:"request_buffer,omitempty"`
	ResponseBuffer string `json:"response_buffer,omitempty"`
	
	// Timeout settings
	ReadTimeout     string `json:"read_timeout,omitempty"`
	WriteTimeout    string `json:"write_timeout,omitempty"`
	IdleTimeout     string `json:"idle_timeout,omitempty"`
	
	// Advanced settings
	MaxRequestBodySize string `json:"max_request_body_size,omitempty"`
	BufferRequests     bool   `json:"buffer_requests,omitempty"`
	
	// TODO: Future - Additional features planned:
	// TODO: RateLimit - Requests per minute/hour per IP or user
	// TODO: CacheConfig - Response caching with Redis/Local
	// TODO: FirewallRules - IP allowlist/blocklist, geo-blocking
	// TODO: CanaryConfig - Percentage-based traffic splitting
	// TODO: CircuitBreaker - Fail fast when backend is unhealthy
	// TODO: RequestMirroring - Send traffic to shadow backend
	// TODO: GraphQLConfig - GraphQL-specific routing and caching
	// TODO: GRPCConfig - gRPC proxy with reflection support
	// TODO: WAFConfig - Web Application Firewall rules
	BufferResponses    bool   `json:"buffer_responses,omitempty"`
	
	// Logging
	Logging *Logging `json:"logging,omitempty"`
	
	// Error handling
	ErrorPages map[string]string `json:"error_pages,omitempty"`
}

// Status represents the current deployment status
type Status struct {
	Running   bool      `json:"running"`
	Services  []Service `json:"services"`
	Uptime    string    `json:"uptime"`
	Timestamp time.Time `json:"timestamp"`
}

type Server struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
	startTime  time.Time
}

func NewServer(configPath string) *Server {
	return &Server{
		configPath: configPath,
		startTime:  time.Now(),
	}
}

func (s *Server) loadConfig() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	s.mu.Lock()
	s.config = &config
	s.mu.Unlock()
	return nil
}

func (s *Server) saveConfig() error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (s *Server) generateServerConfig() ([]byte, error) {
	// Build server JSON config with all features
	apps := map[string]interface{}{
		"http": s.buildHTTPApp(),
	}

	if s.needsTLS() {
		apps["tls"] = s.buildTLSApp()
	}

	adminAddr := "localhost:2019"
	if s.config.Global != nil && s.config.Global.AdminAddr != "" {
		adminAddr = s.config.Global.AdminAddr
	}

	config := map[string]interface{}{
		"admin": map[string]interface{}{
			"listen": adminAddr,
		},
		"apps": apps,
		"logging": s.buildLogging(),
	}

	return json.MarshalIndent(config, "", "  ")
}

func (s *Server) buildHTTPApp() map[string]interface{} {
	s.mu.RLock()
	services := s.config.Services
	s.mu.RUnlock()

	routes := []map[string]interface{}{}

	for _, svc := range services {
		route := s.buildRoute(svc)
		routes = append(routes, route)
	}

	httpPort := ":80"
	httpsPort := ":443"
	
	if s.config.Global != nil {
		if s.config.Global.HTTPPort != "" {
			httpPort = s.config.Global.HTTPPort
		}
		if s.config.Global.HTTPSPort != "" {
			httpsPort = s.config.Global.HTTPSPort
		}
	}

	servers := map[string]interface{}{}
	if len(routes) > 0 {
		servers["srv0"] = map[string]interface{}{
			"listen": []string{httpPort, httpsPort},
			"routes": routes,
			"automatic_https": map[string]interface{}{
				"disable": s.config.Global != nil && s.config.Global.AutoHTTPS == "off",
			},
		}
	}

	return map[string]interface{}{
		"servers": servers,
	}
}

func (s *Server) buildRoute(svc Service) map[string]interface{} {
	// Build host matchers
	hosts := []string{svc.Domain}
	if len(svc.Domains) > 0 {
		hosts = append(hosts, svc.Domains...)
	}

	route := map[string]interface{}{
		"match": []map[string]interface{}{
			{
				"host": hosts,
			},
		},
	}

	handlers := []map[string]interface{}{}

	// 1. Logging (if enabled) - simplified, handled at global level
	if svc.Logging != nil && svc.Logging.Enable {
		handlers = append(handlers, s.buildLogHandler(svc))
	}

	// 2. Basic authentication (if configured)
	if len(svc.BasicAuth) > 0 {
		handlers = append(handlers, s.buildAuthHandler(svc))
	}

	// 3. Compression (if enabled)
	if svc.Compression != nil && svc.Compression.Enable {
		handlers = append(handlers, s.buildEncodeHandler(svc))
	}

	// 4. Header manipulation (upstream/requests)
	if len(svc.HeadersUp) > 0 {
		handlers = append(handlers, s.buildHeadersHandler(svc.HeadersUp, "request"))
	}

	// 5. Main handler (reverse_proxy, static, redirect)
	switch svc.Handler {
	case "static":
		handlers = append(handlers, s.buildStaticHandler(svc))
	case "redirect":
		handlers = append(handlers, s.buildRedirectHandler(svc))
	default:
		handlers = append(handlers, s.buildReverseProxyHandler(svc))
	}

	// 6. Header manipulation (downstream/responses)
	if len(svc.HeadersDown) > 0 {
		handlers = append(handlers, s.buildHeadersHandler(svc.HeadersDown, "response"))
	}

	route["handle"] = handlers

	if svc.AutoTLS {
		route["terminal"] = true
	}

	return route
}

func (s *Server) buildErrorHandler(svc Service) map[string]interface{} {
	// Build error page routes using handle_errors
	routes := []map[string]interface{}{}
	
	for code, path := range svc.ErrorPages {
		routes = append(routes, map[string]interface{}{
			"handle": []map[string]interface{}{
				{
					"handler": "rewrite",
					"uri": path,
				},
				{
					"handler": "reverse_proxy",
					"upstreams": []map[string]interface{}{
						{"dial": "localhost:8080"}, // Fallback, should serve static
					},
				},
			},
			"match": []map[string]interface{}{
				{
					"expression": fmt.Sprintf(`{http.error.status_code} == %s`, code),
				},
			},
		})
	}

	return map[string]interface{}{
		"handler": "subroute",
		"routes": routes,
	}
}

func (s *Server) buildLogHandler(svc Service) map[string]interface{} {
	// Logging is configured at server level, not handler level
	// This is a placeholder - the engine handles logging via global config
	return map[string]interface{}{
		"handler": "subroute",
		"routes": []map[string]interface{}{
			{
				"handle": []map[string]interface{}{
					{
						"handler": "copy_response",
					},
				},
			},
		},
	}
}

func (s *Server) buildAuthHandler(svc Service) map[string]interface{} {
	accounts := []map[string]string{}
	for _, auth := range svc.BasicAuth {
		accounts = append(accounts, map[string]string{
			"username": auth.Username,
			"password": auth.HashedPassword,
		})
	}

	realm := "Restricted"
	if len(svc.BasicAuth) > 0 && svc.BasicAuth[0].Realm != "" {
		realm = svc.BasicAuth[0].Realm
	}

	return map[string]interface{}{
		"handler": "authentication",
		"providers": map[string]interface{}{
			"http_basic": map[string]interface{}{
				"accounts": accounts,
				"realm":    realm,
			},
		},
	}
}

func (s *Server) buildEncodeHandler(svc Service) map[string]interface{} {
	encodings := map[string]interface{}{}
	
	if svc.Compression.Formats == nil || len(svc.Compression.Formats) == 0 {
		svc.Compression.Formats = []string{"zstd", "gzip"}
	}
	
	for _, format := range svc.Compression.Formats {
		switch format {
		case "gzip":
			encodings["gzip"] = map[string]interface{}{}
		case "zstd":
			encodings["zstd"] = map[string]interface{}{}
		}
	}

	minLength := 512
	if svc.Compression.MinimumLength > 0 {
		minLength = svc.Compression.MinimumLength
	}

	return map[string]interface{}{
		"handler": "encode",
		"encodings": encodings,
		"minimum_length": minLength,
	}
}

func (s *Server) buildHeadersHandler(headers []Header, direction string) map[string]interface{} {
	setOps := map[string][]string{}
	addOps := map[string][]string{}
	deleteOps := []string{}
	
	for _, h := range headers {
		if h.Delete {
			deleteOps = append(deleteOps, h.Name)
		} else if h.Add {
			if _, exists := addOps[h.Name]; !exists {
				addOps[h.Name] = []string{}
			}
			addOps[h.Name] = append(addOps[h.Name], h.Value)
		} else {
			setOps[h.Name] = []string{h.Value}
		}
	}

	reqConfig := map[string]interface{}{}
	respConfig := map[string]interface{}{}
	
	if len(setOps) > 0 {
		if direction == "request" {
			reqConfig["set"] = setOps
		} else {
			respConfig["set"] = setOps
		}
	}
	if len(addOps) > 0 {
		if direction == "request" {
			reqConfig["add"] = addOps
		} else {
			respConfig["add"] = addOps
		}
	}
	if len(deleteOps) > 0 {
		if direction == "request" {
			reqConfig["delete"] = deleteOps
		} else {
			respConfig["delete"] = deleteOps
		}
	}

	handler := map[string]interface{}{
		"handler": "headers",
	}
	
	if len(reqConfig) > 0 {
		handler["request"] = reqConfig
	}
	if len(respConfig) > 0 {
		handler["response"] = respConfig
	}
	
	return handler
}

func (s *Server) buildStaticHandler(svc Service) map[string]interface{} {
	if svc.Static == nil {
		svc.Static = &StaticFiles{}
	}

	handler := map[string]interface{}{
		"handler": "file_server",
		"root": svc.Static.Root,
	}

	if svc.Static.Browse {
		handler["browse"] = map[string]interface{}{}
	}

	if len(svc.Static.Index) > 0 {
		handler["index_names"] = svc.Static.Index
	}

	if len(svc.Static.Hide) > 0 {
		handler["hide"] = svc.Static.Hide
	}

	return handler
}

func (s *Server) buildRedirectHandler(svc Service) map[string]interface{} {
	code := 302
	if svc.RedirectCode > 0 {
		code = svc.RedirectCode
	}

	return map[string]interface{}{
		"handler": "static_response",
		"status_code": code,
		"headers": map[string]interface{}{
			"Location": []string{svc.RedirectTo},
		},
	}
}

func (s *Server) buildReverseProxyHandler(svc Service) map[string]interface{} {
	upstreams := []map[string]interface{}{}

	// Build upstream list
	if len(svc.Upstreams) > 0 {
		for _, u := range svc.Upstreams {
			upstream := map[string]interface{}{
				"dial": u.Address,
			}
			if u.MaxRequests > 0 {
				upstream["max_requests"] = u.MaxRequests
			}
			upstreams = append(upstreams, upstream)
		}
	} else if svc.Port > 0 {
		upstreams = append(upstreams, map[string]interface{}{
			"dial": fmt.Sprintf("localhost:%d", svc.Port),
		})
	}

	handler := map[string]interface{}{
		"handler": "reverse_proxy",
		"upstreams": upstreams,
	}

	// Load balancing
	if svc.LoadBalance != "" {
		handler["load_balancing"] = map[string]interface{}{
			"selection_policy": map[string]interface{}{
				"policy": svc.LoadBalance,
			},
		}
	}

	// Health checks
	if svc.HealthCheck != nil {
		hc := map[string]interface{}{}
		
		if svc.HealthCheck.Path != "" {
			hc["active"] = map[string]interface{}{
				"path": svc.HealthCheck.Path,
			}
			if svc.HealthCheck.Interval != "" {
				hc["active"].(map[string]interface{})["interval"] = svc.HealthCheck.Interval
			}
		if svc.HealthCheck.Timeout != "" {
			hc["active"].(map[string]interface{})["timeout"] = svc.HealthCheck.Timeout
		}
		}
		
		handler["health_checks"] = hc
	}

	// WebSocket support (just enable in reverse_proxy - the engine auto-handles this)
	if svc.WebSocket {
		handler["handle_response"] = []map[string]interface{}{}
	}

	// Buffering - use sensible defaults
	if svc.BufferRequests || svc.BufferResponses {
		handler["flush_interval"] = "10s"
	}

	// Timeout settings
	if svc.ReadTimeout != "" {
		handler["transport"].(map[string]interface{})["read_timeout"] = svc.ReadTimeout
	}

	return handler
}

func (s *Server) buildTLSApp() map[string]interface{} {
	s.mu.RLock()
	services := s.config.Services
	s.mu.RUnlock()

	automation := map[string]interface{}{
		"policies": []map[string]interface{}{},
	}

	email := ""
	if s.config.Global != nil {
		email = s.config.Global.Email
	}

	for _, svc := range services {
		if !svc.AutoTLS {
			continue
		}

		subjects := []string{svc.Domain}
		subjects = append(subjects, svc.Domains...)

		issuer := map[string]interface{}{
			"module": "acme",
		}
		
		if email != "" {
			issuer["email"] = email
		}

		policy := map[string]interface{}{
			"subjects": subjects,
			"issuers": []map[string]interface{}{issuer},
		}
		automation["policies"] = append(automation["policies"].([]map[string]interface{}), policy)
	}

	return map[string]interface{}{
		"automation": automation,
	}
}

func (s *Server) buildLogging() map[string]interface{} {
	return map[string]interface{}{
		"logs": map[string]interface{}{
			"default": map[string]interface{}{
				"level": "INFO",
			},
		},
	}
}

func (s *Server) needsTLS() bool {
	s.mu.RLock()
	services := s.config.Services
	s.mu.RUnlock()

	for _, svc := range services {
		if svc.AutoTLS {
			return true
		}
	}
	return false
}

func (s *Server) startEngine() error {
	config, err := s.generateServerConfig()
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// TODO: Future - Replace with native server lifecycle management
	// Current: Using embedded library for config reloading
	// Future: Implement hot-reload without process restart
	// TODO: Add graceful shutdown with connection draining
	// TODO: Implement zero-downtime config updates (SO_REUSEPORT)
	// TODO: Add health-check based traffic shifting (blue/green deploys)
	if err := caddy.Load(config, true); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

func (s *Server) stopEngine() error {
	// TODO: Future - Implement graceful shutdown
	// Current: Immediate stop via library
	// Future: Drain connections, finish requests, then shutdown
	return caddy.Stop()
}

// HTTP Handlers

func (s *Server) handleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, fmt.Sprintf("Invalid body: %v", err), http.StatusBadRequest)
		return
	}

	if service.Name == "" || (service.Domain == "" && len(service.Domains) == 0) {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Default to reverse_proxy if not specified
	if service.Handler == "" && service.Port > 0 {
		service.Handler = "reverse_proxy"
	}

	s.mu.Lock()
	
	// Check for existing service with same name
	found := false
	for i, svc := range s.config.Services {
		if svc.Name == service.Name {
			s.config.Services[i] = service
			found = true
			break
		}
	}
	
	if !found {
		s.config.Services = append(s.config.Services, service)
	}
	s.mu.Unlock()

	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Save failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.stopEngine(); err != nil {
		log.Printf("Warning: stop failed: %v", err)
	}

	if err := s.startEngine(); err != nil {
		http.Error(w, fmt.Sprintf("Start failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "deployed",
		"service": service.Name,
		"domain":  service.Domain,
		"handler": service.Handler,
	})
}

func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	newServices := []Service{}
	found := false
	for _, svc := range s.config.Services {
		if svc.Name != name {
			newServices = append(newServices, svc)
		} else {
			found = true
		}
	}

	if !found {
		s.mu.Unlock()
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	s.config.Services = newServices
	s.mu.Unlock()

	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Save failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.stopEngine(); err != nil {
		log.Printf("Warning: stop failed: %v", err)
	}

	s.mu.RLock()
	servicesCount := len(s.config.Services)
	s.mu.RUnlock()

	if servicesCount > 0 {
		if err := s.startEngine(); err != nil {
			http.Error(w, fmt.Sprintf("Restart failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "removed",
		"service": name,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	services := s.config.Services
	s.mu.RUnlock()

	status := Status{
		Running:   true,
		Services:  services,
		Uptime:    time.Since(s.startTime).String(),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	services := s.config.Services
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.loadConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Load failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.stopEngine(); err != nil {
		log.Printf("Warning: stop failed: %v", err)
	}

	s.mu.RLock()
	servicesCount := len(s.config.Services)
	s.mu.RUnlock()

	if servicesCount > 0 {
		if err := s.startEngine(); err != nil {
			http.Error(w, fmt.Sprintf("Restart failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "reloaded",
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		config := s.config
		s.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	
	case http.MethodPut:
		var config Config
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, fmt.Sprintf("Invalid body: %v", err), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.config = &config
		s.mu.Unlock()
		
		if err := s.saveConfig(); err != nil {
			http.Error(w, fmt.Sprintf("Save failed: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func parseStatusCode(code string) int {
	c, _ := strconv.Atoi(code)
	if c == 0 {
		return 500
	}
	return c
}

func (s *Server) Run(addr string) error {
	if err := s.loadConfig(); err != nil {
		if os.IsNotExist(err) {
			s.config = &Config{Services: []Service{}}
			if err := s.saveConfig(); err != nil {
				return fmt.Errorf("failed to create initial config: %w", err)
			}
		} else {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/deploy", s.handleDeploy)
	mux.HandleFunc("/remove", s.handleRemove)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/list", s.handleList)
	mux.HandleFunc("/reload", s.handleReload)
	mux.HandleFunc("/config", s.handleConfig)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		s.stopEngine()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	s.mu.RLock()
	servicesCount := len(s.config.Services)
	s.mu.RUnlock()

	if servicesCount > 0 {
		go func() {
			if err := s.startEngine(); err != nil {
				log.Printf("Failed to start: %v", err)
			}
		}()
	}

	log.Printf("mach server running on %s", addr)
	return server.ListenAndServe()
}

func main() {
	configPath := os.Getenv("MACH_CONFIG")
	if configPath == "" {
		configDir, _ := os.UserConfigDir()
		configPath = filepath.Join(configDir, "mach", "config.json")
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	addr := os.Getenv("MACH_ADDR")
	if addr == "" {
		addr = "localhost:8765"
	}

	server := NewServer(configPath)
	if err := server.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
