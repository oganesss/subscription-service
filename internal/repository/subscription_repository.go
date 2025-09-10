package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"subscription-service/internal/models"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s models.Subscription) (models.Subscription, error) {
	const q = `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,now(),now()) RETURNING created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	if err := row.Scan(&s.CreatedAt, &s.UpdatedAt); err != nil {
		return s, fmt.Errorf("insert subscription: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Subscription, error) {
	const q = `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	var s models.Subscription
	if err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows { return s, fmt.Errorf("not found") }
		return s, fmt.Errorf("get subscription: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s models.Subscription) (models.Subscription, error) {
	const q = `UPDATE subscriptions SET service_name=$2, price=$3, user_id=$4, start_date=$5, end_date=$6, updated_at=now() WHERE id=$1 RETURNING created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	if err := row.Scan(&s.CreatedAt, &s.UpdatedAt); err != nil {
		return s, fmt.Errorf("update subscription: %w", err)
	}
	return s, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM subscriptions WHERE id=$1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil { return fmt.Errorf("delete subscription: %w", err) }
	if ct.RowsAffected() == 0 { return fmt.Errorf("not found") }
	return nil
}

type ListFilters struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

func (r *SubscriptionRepository) List(ctx context.Context, f ListFilters) ([]models.Subscription, int, error) {
	base := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE 1=1`
	countBase := `SELECT count(1) FROM subscriptions WHERE 1=1`
	args := []any{}
	idx := 1

	if f.UserID != nil { base += fmt.Sprintf(" AND user_id = $%d", idx); countBase += fmt.Sprintf(" AND user_id = $%d", idx); args = append(args, *f.UserID); idx++ }
	if f.ServiceName != nil { base += fmt.Sprintf(" AND service_name = $%d", idx); countBase += fmt.Sprintf(" AND service_name = $%d", idx); args = append(args, *f.ServiceName); idx++ }
	if f.From != nil { base += fmt.Sprintf(" AND (end_date IS NULL OR end_date >= $%d)", idx); countBase += fmt.Sprintf(" AND (end_date IS NULL OR end_date >= $%d)", idx); args = append(args, *f.From); idx++ }
	if f.To != nil { base += fmt.Sprintf(" AND start_date <= $%d", idx); countBase += fmt.Sprintf(" AND start_date <= $%d", idx); args = append(args, *f.To); idx++ }

	base += " ORDER BY created_at DESC"
	base += fmt.Sprintf(" LIMIT %d OFFSET %d", f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, base, args...)
	if err != nil { return nil, 0, fmt.Errorf("list subscriptions: %w", err) }
	defer rows.Close()

	var items []models.Subscription
	for rows.Next() {
		var s models.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan subscription: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil { return nil, 0, fmt.Errorf("rows err: %w", err) }

	var total int
	if err := r.pool.QueryRow(ctx, countBase, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count subscriptions: %w", err)
	}
	return items, total, nil
}


