// Package console exposes the ConfigHub HTTP API and embedded control pages.
package console

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"confighub/internal/audit"
	"confighub/internal/web"
)

// Backend is the cluster surface the console renders.
type Backend interface {
	Namespaces() []NamespaceInfo
	Items(namespace string, prefix string) []ItemInfo
	Item(namespace string, key string) (ItemInfo, bool)
	PutItem(namespace string, key string, value string, ttl int64) error
	DeleteItem(namespace string, key string) error
	Publish(namespace string, key string, value string) error
	Rollback(namespace string, key string) error
	Version(namespace string) VersionInfo
	Snapshots(namespace string) []SnapshotInfo
	Pull(namespace string) (uint64, error)
	Subscriptions() []SubscriptionInfo
	DraftList(namespace string) []DraftInfo
	EnsureDraft(namespace string, key string, value string) error
	ValidateDraft(namespace string, key string) error
	Search(namespace string, term string) []string
	Push(clientID string) (uint64, error)
	Retry(clientID string) (uint64, bool, error)
	Deliver(clientID string, version uint64) error
	AuditEntries(limit int) []audit.Event
	AuditCount() int
	AuditCounts() map[string]int
	Quota(namespace string) QuotaInfo
}

// Server serves the console pages and JSON API.
type Server struct {
	backend Backend
}

// NewServer creates a console server backed by a cluster.
func NewServer(backend Backend) *Server {
	return &Server{backend: backend}
}

// Handler builds the chi router.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/items", s.handlePage(web.ItemsHTML))
	r.Get("/releases", s.handlePage(web.ReleasesHTML))
	r.Get("/subscriptions", s.handlePage(web.SubscriptionsHTML))
	r.Get("/audit", s.handlePage(web.AuditHTML))
	r.Get("/api/health", s.handleHealth)
	r.Get("/api/namespaces", s.handleNamespaces)
	r.Get("/api/items", s.handleItems)
	r.Get("/api/search", s.handleSearch)
	r.Get("/api/items/{key}", s.handleItem)
	r.Put("/api/items/{key}", s.handlePutItem)
	r.Delete("/api/items/{key}", s.handleDeleteItem)
	r.Post("/api/publish", s.handlePublish)
	r.Post("/api/rollback", s.handleRollback)
	r.Get("/api/version", s.handleVersion)
	r.Get("/api/snapshots", s.handleSnapshots)
	r.Post("/api/pull", s.handlePull)
	r.Get("/api/subscriptions", s.handleSubscriptions)
	r.Get("/api/drafts", s.handleDrafts)
	r.Post("/api/drafts", s.handleCreateDraft)
	r.Post("/api/drafts/{key}/validate", s.handleValidateDraft)
	r.Post("/api/push", s.handlePush)
	r.Post("/api/retry", s.handleRetry)
	r.Post("/api/deliver", s.handleDeliver)
	r.Get("/api/audit", s.handleAudit)
	r.Get("/api/quota", s.handleQuota)
	return r
}

// Start serves the console until the process exits.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
