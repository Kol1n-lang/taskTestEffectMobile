package json_models

import "time"

// json_models.CreateSubscription model
// @Description Subscription information
type CreateSubscription struct {
	ServiceName string  `json:"service_name" validate:"required"`
	Price       int     `json:"price" validate:"required,gt=0"`
	UserID      string  `json:"user_id" validate:"required,uuid4"`
	StartDate   string  `json:"start_date" validate:"required,datetime=01-2006"`
	EndDate     *string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006"`
}

// json_models.PutSubscription model
// @Description Subscription information
type PutSubscription struct {
	ServiceName    string  `json:"service_name" validate:"required"`
	Price          int     `json:"price" validate:"gt=0"`
	SubscriptionID string  `json:"subscription_id" validate:"uuid4"`
	StartDate      string  `json:"start_date" validate:"datetime=01-2006"`
	EndDate        *string `json:"end_date,omitempty" validate:"omitempty,datetime=01-2006"`
}

// json_models.SubscriptionUpdate model
// @Description Subscription information
type SubscriptionUpdate struct {
	ServiceName string
	Price       int
	StartDate   *time.Time
	EndDate     *time.Time
}

// json_models.CostRequest model
// @Description Subscription information
type CostRequest struct {
	UserID      *string `schema:"user-id"`
	ServiceName *string `schema:"service-name"`
	StartDate   string  `schema:"start-date" validate:"required,datetime=01-2006"`
	EndDate     *string `schema:"end-date" validate:"omitempty,datetime=01-2006"`
}
