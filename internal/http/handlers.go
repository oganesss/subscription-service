package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"subscription-service/internal/models"
	"subscription-service/internal/service"
)

type HandlersImpl struct {
	log *zap.Logger
	svc *service.SubscriptionService
}

func NewHandlers(log *zap.Logger, svc *service.SubscriptionService) *HandlersImpl {
	return &HandlersImpl{log: log, svc: svc}
}

type CreateRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

type UpdateRequest = CreateRequest

type SubscriptionDTO struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SubscriptionResponse struct {
	Subscription SubscriptionDTO `json:"subscription"`
}

type CreateResponse = SubscriptionResponse

type ListQuery struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *string
	To          *string
	Limit       int
	Offset      int
}

type ListResponse struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
	Total         int               `json:"total"`
}

type TotalQuery struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        string
	To          string
}

type TotalResponse struct {
	Amount int `json:"amount"`
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *HandlersImpl) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "invalid json"}})
		return
	}
	sub, err := h.svc.Create(service.CreateInput{ServiceName: req.ServiceName, Price: req.Price, UserID: req.UserID, StartDate: req.StartDate, EndDate: req.EndDate})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusCreated, SubscriptionResponse{Subscription: toDTO(sub)})
}

func (h *HandlersImpl) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "invalid id"}})
		return
	}
	sub, err := h.svc.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": map[string]any{"code": 404, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, SubscriptionResponse{Subscription: toDTO(sub)})
}

func (h *HandlersImpl) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "invalid id"}})
		return
	}
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "invalid json"}})
		return
	}
	sub, err := h.svc.Update(id, service.CreateInput{ServiceName: req.ServiceName, Price: req.Price, UserID: req.UserID, StartDate: req.StartDate, EndDate: req.EndDate})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, SubscriptionResponse{Subscription: toDTO(sub)})
}

func (h *HandlersImpl) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "invalid id"}})
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"errors": map[string]any{"code": 404, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *HandlersImpl) List(w http.ResponseWriter, r *http.Request) {
	q := ListQuery{Limit: 50, Offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parseInt(v); err == nil { q.Limit = n }
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := parseInt(v); err == nil { q.Offset = n }
	}
	if v := r.URL.Query().Get("service_name"); v != "" { q.ServiceName = &v }
	if v := r.URL.Query().Get("user_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil { q.UserID = &id }
	}
	if v := r.URL.Query().Get("from"); v != "" { q.From = &v }
	if v := r.URL.Query().Get("to"); v != "" { q.To = &v }

	list, total, err := h.svc.List(service.ListQuery{UserID: q.UserID, ServiceName: q.ServiceName, From: q.From, To: q.To, Limit: q.Limit, Offset: q.Offset})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": err.Error()}})
		return
	}
	dtos := make([]SubscriptionDTO, 0, len(list))
	for _, m := range list { dtos = append(dtos, toDTO(m)) }
	writeJSON(w, http.StatusOK, ListResponse{Subscriptions: dtos, Total: total})
}

func (h *HandlersImpl) Total(w http.ResponseWriter, r *http.Request) {
	var q TotalQuery
	if v := r.URL.Query().Get("from"); v != "" { q.From = v } else {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "from required"}}); return
	}
	if v := r.URL.Query().Get("to"); v != "" { q.To = v } else {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": "to required"}}); return
	}
	if v := r.URL.Query().Get("service_name"); v != "" { q.ServiceName = &v }
	if v := r.URL.Query().Get("user_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil { q.UserID = &id }
	}
	amount, err := h.svc.Total(service.TotalQuery{UserID: q.UserID, ServiceName: q.ServiceName, From: q.From, To: q.To})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": map[string]any{"code": 400, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, TotalResponse{Amount: amount})
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func toDTO(m models.Subscription) SubscriptionDTO {
	return SubscriptionDTO{
		ID: m.ID,
		ServiceName: m.ServiceName,
		Price: m.Price,
		UserID: m.UserID,
		StartDate: m.StartDate,
		EndDate: m.EndDate,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}


