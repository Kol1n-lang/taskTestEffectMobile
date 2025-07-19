package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"taskTestEffectMobile/internal/models/json_models"
	"taskTestEffectMobile/internal/models/sql_models"
	"taskTestEffectMobile/internal/repository"
	"time"
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *zap.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger.With(zap.String("layer", "service")),
	}
}

func (subscriptionService SubscriptionService) CreateSubscription(ctx context.Context, sub json_models.CreateSubscription) (string, error) {
	subscriptionService.logger.Info("Creating subscription",
		zap.String("userID", sub.UserID),
		zap.String("service", sub.ServiceName))

	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		subscriptionService.logger.Error("Invalid start date format",
			zap.String("date", sub.StartDate),
			zap.Error(err))
		return "", fmt.Errorf("invalid start date format: %w", err)
	}

	if sub.EndDate == nil {
		subscriptionService.logger.Debug("Creating subscription without end date")
		return subscriptionService.repo.InsertSubscription(ctx, sub.ServiceName, sub.Price, sub.UserID, startDate, nil)
	}

	endDate, err := time.Parse("01-2006", *sub.EndDate)
	if err != nil {
		subscriptionService.logger.Error("Invalid end date format",
			zap.String("date", *sub.EndDate),
			zap.Error(err))
		return "", fmt.Errorf("invalid end date format: %w", err)
	}

	subscriptionService.logger.Debug("Creating subscription with end date")
	return subscriptionService.repo.InsertSubscription(ctx, sub.ServiceName, sub.Price, sub.UserID, startDate, &endDate)
}

func (subscriptionService SubscriptionService) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]sql_models.Subscription, error) {
	subscriptionService.logger.Info("Getting user subscriptions",
		zap.String("userID", userID.String()))

	subscriptions, err := subscriptionService.repo.GetSubscriptions(ctx, userID)
	if err != nil {
		subscriptionService.logger.Error("Failed to get subscriptions",
			zap.String("userID", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	subscriptionService.logger.Info("Successfully retrieved subscriptions",
		zap.String("userID", userID.String()),
		zap.Int("count", len(subscriptions)))
	return subscriptions, nil
}

func (subscriptionService SubscriptionService) UpdateSubscription(ctx context.Context, req json_models.PutSubscription) error {
	subscriptionService.logger.Info("Updating subscription",
		zap.String("subscriptionID", req.SubscriptionID),
		zap.String("service", req.ServiceName))

	var startDate, endDate *time.Time

	if req.StartDate != "" {
		sd, err := time.Parse("01-2006", req.StartDate)
		if err != nil {
			subscriptionService.logger.Error("Invalid start date format",
				zap.String("date", req.StartDate),
				zap.Error(err))
			return fmt.Errorf("invalid start date format: %w", err)
		}
		startDate = &sd
	}

	if req.EndDate != nil {
		ed, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			subscriptionService.logger.Error("Invalid end date format",
				zap.String("date", *req.EndDate),
				zap.Error(err))
			return fmt.Errorf("invalid end date format: %w", err)
		}
		endDate = &ed
	}

	updateData := json_models.SubscriptionUpdate{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := subscriptionService.repo.UpdateSubscription(ctx, req.SubscriptionID, updateData); err != nil {
		subscriptionService.logger.Error("Failed to update subscription",
			zap.String("subscriptionID", req.SubscriptionID),
			zap.String("service", req.ServiceName),
			zap.Error(err))
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	subscriptionService.logger.Info("Subscription updated successfully",
		zap.String("subscriptionID", req.SubscriptionID),
		zap.String("service", req.ServiceName))
	return nil
}

func (subscriptionService SubscriptionService) DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error {
	subscriptionService.logger.Info("Deleting subscription",
		zap.String("userID", subscriptionUUID.String()))

	if err := subscriptionService.repo.DeleteSubscription(ctx, subscriptionUUID); err != nil {
		subscriptionService.logger.Error("Failed to delete subscription",
			zap.String("userID", subscriptionUUID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	subscriptionService.logger.Info("Subscription deleted successfully",
		zap.String("userID", subscriptionUUID.String()))
	return nil
}

func (subscriptionService SubscriptionService) CalculateSubscriptionsCost(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	startDateStr string,
	endDateStr *string,
) (int, error) {
	subscriptionService.logger.Info("Calculating subscriptions cost",
		zap.Any("userID", userID),
		zap.Any("serviceName", serviceName),
		zap.String("startDate", startDateStr),
		zap.Any("endDate", endDateStr))

	startDate, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		subscriptionService.logger.Error("Invalid start date format",
			zap.String("date", startDateStr),
			zap.Error(err))
		return 0, fmt.Errorf("invalid start date format: %w", err)
	}

	var endDate *time.Time
	if endDateStr != nil {
		parsedEndDate, err := time.Parse("01-2006", *endDateStr)
		if err != nil {
			subscriptionService.logger.Error("Invalid end date format",
				zap.String("date", *endDateStr),
				zap.Error(err))
			return 0, fmt.Errorf("invalid end date format: %w", err)
		}
		endDate = &parsedEndDate
	}

	return subscriptionService.repo.GetSubscriptionsCost(ctx, userID, serviceName, startDate, endDate)
}
