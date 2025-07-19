package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"taskTestEffectMobile/internal/models/json_models"
	"taskTestEffectMobile/internal/models/sql_models"
	"time"
)

type SubscriptionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSubscriptionRepository(db *sql.DB, logger *zap.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger.With(zap.String("layer", "repository")),
	}
}

func (subscriptionRepository SubscriptionRepository) InsertSubscription(ctx context.Context, serviceName string, price int, userID string, startDate time.Time, endTime *time.Time) (string, error) {
	subscriptionRepository.logger.Debug("Inserting new subscription",
		zap.String("userID", userID),
		zap.String("service", serviceName))

	id := uuid.New().String()
	query := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := subscriptionRepository.db.ExecContext(ctx, query, id, serviceName, price, userID, startDate, endTime, time.Now())
	if err != nil {
		subscriptionRepository.logger.Error("Failed to insert subscription",
			zap.String("query", query),
			zap.String("userID", userID),
			zap.String("service", serviceName),
			zap.Error(err))
		return "", fmt.Errorf("failed to insert subscription: %w", err)
	}

	subscriptionRepository.logger.Info("Subscription created successfully",
		zap.String("subscriptionID", id))
	return id, nil
}

func (subscriptionRepository SubscriptionRepository) GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]sql_models.Subscription, error) {
	subscriptionRepository.logger.Debug("Getting user subscriptions",
		zap.String("userID", userID.String()))

	var subscriptions []sql_models.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at FROM subscriptions WHERE user_id = $1`

	rows, err := subscriptionRepository.db.QueryContext(ctx, query, userID)
	if err != nil {
		subscriptionRepository.logger.Error("Failed to query subscriptions",
			zap.String("query", query),
			zap.String("userID", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			subscriptionRepository.logger.Error("Failed to close rows",
				zap.String("userID", userID.String()),
				zap.Error(closeErr))
		}
	}()

	for rows.Next() {
		var sub sql_models.Subscription
		var endDate sql.NullTime

		if err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
			&sub.CreatedAt,
		); err != nil {
			subscriptionRepository.logger.Error("Failed to scan subscription row",
				zap.String("userID", userID.String()),
				zap.Error(err))
			return nil, fmt.Errorf("error with scanning: %w", err)
		}

		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		subscriptionRepository.logger.Error("Row iteration error",
			zap.String("userID", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("iteration error: %w", err)
	}

	subscriptionRepository.logger.Debug("Retrieved subscriptions count",
		zap.String("userID", userID.String()),
		zap.Int("count", len(subscriptions)))
	return subscriptions, nil
}

func (subscriptionRepository SubscriptionRepository) UpdateSubscription(ctx context.Context, subscriptionID string, data json_models.SubscriptionUpdate) error {
	subscriptionRepository.logger.Debug("Updating subscription",
		zap.String("SubscriptionID", subscriptionID),
		zap.Any("updateData", data))

	query := `
		UPDATE subscriptions
		SET 
			service_name = COALESCE($1, service_name),
			price = COALESCE($2, price),
			start_date = COALESCE($3, start_date),
			end_date = COALESCE($4, end_date)
		WHERE id = $5
	`

	_, err := subscriptionRepository.db.ExecContext(ctx, query,
		data.ServiceName,
		data.Price,
		data.StartDate,
		data.EndDate,
		subscriptionID,
	)

	if err != nil {
		subscriptionRepository.logger.Error("Failed to update subscription",
			zap.String("query", query),
			zap.String("subscriptionID", subscriptionID),
			zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	subscriptionRepository.logger.Info("Subscription updated successfully",
		zap.String("subscriptionID", subscriptionID))
	return nil
}

func (subscriptionRepository SubscriptionRepository) DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	subscriptionRepository.logger.Debug("Attempting to delete subscription",
		zap.String("userID", subscriptionUUID.String()))

	query := `DELETE FROM subscriptions 
        WHERE id = $1`

	result, err := subscriptionRepository.db.ExecContext(ctx, query, subscriptionUUID)
	if err != nil {
		subscriptionRepository.logger.Error("Database error when deleting subscription",
			zap.String("query", query),
			zap.String("userID", subscriptionUUID.String()),
			zap.Error(err))
		return fmt.Errorf("database error when deleting subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		subscriptionRepository.logger.Error("Failed to get rows affected count",
			zap.String("userID", subscriptionUUID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to verify deletion: %w", err)
	}

	if rowsAffected == 0 {
		subscriptionRepository.logger.Warn("Subscription not found for deletion",
			zap.String("userID", subscriptionUUID.String()))
		return fmt.Errorf("subscription does not exist")
	}

	subscriptionRepository.logger.Info("Subscription deleted successfully",
		zap.String("userID", subscriptionUUID.String()),
		zap.Int64("rowsAffected", rowsAffected))
	return nil
}

func (subscriptionRepository SubscriptionRepository) GetSubscriptionsCost(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	startDate time.Time,
	endDate *time.Time,
) (int, error) {
	subscriptionRepository.logger.Debug("Calculating subscriptions cost",
		zap.Any("userID", userID),
		zap.Any("serviceName", serviceName),
		zap.Time("startDate", startDate),
		zap.Any("endDate", endDate))

	query := `
        SELECT COALESCE(SUM(price), 0) 
        FROM subscriptions 
        WHERE start_date >= $1
        AND (end_date IS NULL OR end_date <= $2)
    `
	args := []interface{}{startDate}

	checkDate := time.Now()
	if endDate != nil {
		checkDate = *endDate
	}
	args = append(args, checkDate)

	argPos := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argPos)
		args = append(args, *serviceName)
		argPos++
	}

	var totalCost int
	err := subscriptionRepository.db.QueryRowContext(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		subscriptionRepository.logger.Error("Failed to calculate subscriptions cost",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Error(err))
		return 0, fmt.Errorf("failed to calculate subscriptions cost: %w", err)
	}

	subscriptionRepository.logger.Debug("Subscriptions cost calculated",
		zap.Int("totalCost", totalCost))
	return totalCost, nil
}
