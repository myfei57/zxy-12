package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NamespaceInfo describes one namespace for the console.
type NamespaceInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// ItemInfo describes one configuration item for the console.
type ItemInfo struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version uint64 `json:"version"`
	Expires int64  `json:"expires"`
}

// VersionInfo describes the version ledger of a namespace.
type VersionInfo struct {
	Revision  uint64 `json:"revision"`
	Published uint64 `json:"published"`
	State     string `json:"state"`
}

// SnapshotInfo describes one snapshot generation for the console.
type SnapshotInfo struct {
	Generation string `json:"generation"`
	Path       string `json:"path"`
}

// SubscriptionInfo describes one client subscription for the console.
type SubscriptionInfo struct {
	ClientID  string `json:"client_id"`
	Namespace string `json:"namespace"`
	Cursor    uint64 `json:"cursor"`
}

// DraftInfo describes one pending configuration draft for the console.
type DraftInfo struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	State     string `json:"state"`
}

// QuotaInfo describes the current capacity usage.
type QuotaInfo struct {
	Capacity int64 `json:"capacity"`
	Used     int64 `json:"used"`
}

type publishRequest struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

type rollbackRequest struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type pullRequest struct {
	Namespace string `json:"namespace"`
}

type pushRequest struct {
	ClientID string `json:"client_id"`
}

type deliverRequest struct {
	ClientID string `json:"client_id"`
	Version  uint64 `json:"version"`
}

type draftRequest struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.backend.Namespaces())
}

func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	prefix := r.URL.Query().Get("prefix")
	writeJSON(w, http.StatusOK, s.backend.Items(namespace, prefix))
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	term := r.URL.Query().Get("term")
	writeJSON(w, http.StatusOK, s.backend.Search(namespace, term))
}

func (s *Server) handleItem(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	key := chi.URLParam(r, "key")
	info, ok := s.backend.Item(namespace, key)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handlePutItem(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	key := chi.URLParam(r, "key")
	var req putItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.PutItem(namespace, key, req.Value, req.TTL); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stored"})
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	key := chi.URLParam(r, "key")
	if err := s.backend.DeleteItem(namespace, key); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	var req publishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Publish(req.Namespace, req.Key, req.Value); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func (s *Server) handleRollback(w http.ResponseWriter, r *http.Request) {
	var req rollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Rollback(req.Namespace, req.Key); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rolled back"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	writeJSON(w, http.StatusOK, s.backend.Version(namespace))
}

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	writeJSON(w, http.StatusOK, s.backend.Snapshots(namespace))
}

func (s *Server) handlePull(w http.ResponseWriter, r *http.Request) {
	var req pullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	generation, err := s.backend.Pull(req.Namespace)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]uint64{"generation": generation})
}

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.backend.Subscriptions())
}

func (s *Server) handleDrafts(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	writeJSON(w, http.StatusOK, s.backend.DraftList(namespace))
}

func (s *Server) handleCreateDraft(w http.ResponseWriter, r *http.Request) {
	var req draftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.EnsureDraft(req.Namespace, req.Key, req.Value); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draft created"})
}

func (s *Server) handleValidateDraft(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	key := chi.URLParam(r, "key")
	if err := s.backend.ValidateDraft(namespace, key); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "validated"})
}

func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	var req pushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	version, err := s.backend.Push(req.ClientID)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]uint64{"version": version})
}

func (s *Server) handleRetry(w http.ResponseWriter, r *http.Request) {
	var req pushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	version, pushed, err := s.backend.Retry(req.ClientID)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"version": version, "pushed": pushed})
}

func (s *Server) handleDeliver(w http.ResponseWriter, r *http.Request) {
	var req deliverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.backend.Deliver(req.ClientID, req.Version); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "delivered"})
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   s.backend.AuditCount(),
		"by_type": s.backend.AuditCounts(),
		"entries": s.backend.AuditEntries(100),
	})
}

func (s *Server) handleQuota(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	writeJSON(w, http.StatusOK, s.backend.Quota(namespace))
}

type putItemRequest struct {
	Value string `json:"value"`
	TTL   int64  `json:"ttl"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
