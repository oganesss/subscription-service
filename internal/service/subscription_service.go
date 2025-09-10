package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"subscription-service/internal/models"
	"subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func parseMonthYear(s string) (time.Time, error) {
	var mm, yyyy int
	_, err := fmt.Sscanf(s, "%02d-%04d", &mm, &yyyy)
	if err != nil { return time.Time{}, fmt.Errorf("invalid month-year: %s", s) }
	if mm < 1 || mm > 12 { return time.Time{}, fmt.Errorf("invalid month: %d", mm) }
	return time.Date(yyyy, time.Month(mm), 1, 0, 0, 0, 0, time.UTC), nil
}

type CreateInput struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   string
	EndDate     *string
}

type ListQuery struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *string
	To          *string
	Limit       int
	Offset      int
}

type TotalQuery struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        string
	To          string
}

func (s *SubscriptionService) Create(req CreateInput) (models.Subscription, error) {
	start, err := parseMonthYear(req.StartDate)
	if err != nil { return models.Subscription{}, err }
	var endPtr *time.Time
	if req.EndDate != nil {
		end, err := parseMonthYear(*req.EndDate)
		if err != nil { return models.Subscription{}, err }
		endPtr = &end
	}
	m := models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   start,
		EndDate:     endPtr,
	}
	ctx := context.Background()
	created, err := s.repo.Create(ctx, m)
	if err != nil { return models.Subscription{}, err }
	return created, nil
}

func (s *SubscriptionService) GetByID(id uuid.UUID) (models.Subscription, error) {
	ctx := context.Background()
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(id uuid.UUID, req CreateInput) (models.Subscription, error) {
	start, err := parseMonthYear(req.StartDate)
	if err != nil { return models.Subscription{}, err }
	var endPtr *time.Time
	if req.EndDate != nil {
		end, err := parseMonthYear(*req.EndDate)
		if err != nil { return models.Subscription{}, err }
		endPtr = &end
	}
	ctx := context.Background()
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil { return models.Subscription{}, err }
	existing.ServiceName = req.ServiceName
	existing.Price = req.Price
	existing.UserID = req.UserID
	existing.StartDate = start
	existing.EndDate = endPtr
	updated, err := s.repo.Update(ctx, existing)
	if err != nil { return models.Subscription{}, err }
	return updated, nil
}

func (s *SubscriptionService) Delete(id uuid.UUID) error {
	ctx := context.Background()
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(q ListQuery) ([]models.Subscription, int, error) {
	ctx := context.Background()
	var from, to *time.Time
	if q.From != nil { t, err := parseMonthYear(*q.From); if err != nil { return nil, 0, err }; from = &t }
	if q.To != nil { t, err := parseMonthYear(*q.To); if err != nil { return nil, 0, err }; to = &t }
	return s.repo.List(ctx, repository.ListFilters{UserID: q.UserID, ServiceName: q.ServiceName, From: from, To: to, Limit: q.Limit, Offset: q.Offset})
}

func monthsBetweenInclusive(a, b time.Time) int {
	return (int(b.Year())-int(a.Year()))*12 + int(b.Month()) - int(a.Month()) + 1
}

func (s *SubscriptionService) Total(q TotalQuery) (int, error) {
	ctx := context.Background()
	from, err := parseMonthYear(q.From)
	if err != nil { return 0, err }
	to, err := parseMonthYear(q.To)
	if err != nil { return 0, err }
	items, _, err := s.repo.List(ctx, repository.ListFilters{UserID: q.UserID, ServiceName: q.ServiceName, From: &from, To: &to, Limit: 100000, Offset: 0})
	if err != nil { return 0, err }
	var sum int
	for _, sbs := range items {
		start := sbs.StartDate
		end := to
		if sbs.EndDate != nil && sbs.EndDate.Before(to) { end = *sbs.EndDate }
		if end.Before(from) || start.After(to) { continue }
		if start.Before(from) { start = from }
		months := monthsBetweenInclusive(start, end)
		if months < 0 { months = 0 }
		sum += sbs.Price * months
	}
	return sum, nil
}
