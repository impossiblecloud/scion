// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hubclient

import (
	"context"
	"net/url"
	"time"

	"github.com/GoogleCloudPlatform/scion/pkg/apiclient"
)

// NotificationService handles notification operations.
type NotificationService interface {
	// List returns notifications for the current user.
	List(ctx context.Context, opts *ListNotificationsOptions) ([]Notification, error)

	// Acknowledge marks a single notification as acknowledged.
	Acknowledge(ctx context.Context, id string) error

	// AcknowledgeAll marks all unacknowledged notifications as acknowledged.
	AcknowledgeAll(ctx context.Context) error
}

// notificationService is the implementation of NotificationService.
type notificationService struct {
	c *client
}

// ListNotificationsOptions configures notification listing.
type ListNotificationsOptions struct {
	OnlyUnacknowledged bool
}

// Notification represents a notification from the Hub API.
type Notification struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscriptionId"`
	AgentID        string    `json:"agentId"`
	GroveID        string    `json:"groveId"`
	SubscriberType string    `json:"subscriberType"`
	SubscriberID   string    `json:"subscriberId"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	Dispatched     bool      `json:"dispatched"`
	Acknowledged   bool      `json:"acknowledged"`
	CreatedAt      time.Time `json:"createdAt"`
}

// List returns notifications for the current user.
func (s *notificationService) List(ctx context.Context, opts *ListNotificationsOptions) ([]Notification, error) {
	query := url.Values{}
	if opts != nil && !opts.OnlyUnacknowledged {
		query.Set("acknowledged", "true")
	}

	resp, err := s.c.transport.GetWithQuery(ctx, "/api/v1/notifications", query, nil)
	if err != nil {
		return nil, err
	}
	result, err := apiclient.DecodeResponse[[]Notification](resp)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []Notification{}, nil
	}
	return *result, nil
}

// Acknowledge marks a single notification as acknowledged.
func (s *notificationService) Acknowledge(ctx context.Context, id string) error {
	resp, err := s.c.transport.Post(ctx, "/api/v1/notifications/"+url.PathEscape(id)+"/ack", nil, nil)
	if err != nil {
		return err
	}
	return apiclient.CheckResponse(resp)
}

// AcknowledgeAll marks all unacknowledged notifications as acknowledged.
func (s *notificationService) AcknowledgeAll(ctx context.Context) error {
	resp, err := s.c.transport.Post(ctx, "/api/v1/notifications/ack-all", nil, nil)
	if err != nil {
		return err
	}
	return apiclient.CheckResponse(resp)
}

// SubscriptionService handles notification subscription operations.
type SubscriptionService interface {
	// Create creates a new notification subscription.
	Create(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error)

	// List returns subscriptions for the current user.
	List(ctx context.Context, opts *ListSubscriptionsOptions) ([]Subscription, error)

	// Delete removes a subscription by ID.
	Delete(ctx context.Context, id string) error
}

// subscriptionService is the implementation of SubscriptionService.
type subscriptionService struct {
	c *client
}

// CreateSubscriptionRequest is the request body for creating a subscription.
type CreateSubscriptionRequest struct {
	Scope             string   `json:"scope"`
	AgentID           string   `json:"agentId,omitempty"`
	GroveID           string   `json:"groveId"`
	TriggerActivities []string `json:"triggerActivities"`
}

// ListSubscriptionsOptions configures subscription listing.
type ListSubscriptionsOptions struct {
	GroveID string
	AgentID string
	Scope   string
}

// Subscription represents a notification subscription from the Hub API.
type Subscription struct {
	ID                string    `json:"id"`
	Scope             string    `json:"scope"`
	AgentID           string    `json:"agentId,omitempty"`
	SubscriberType    string    `json:"subscriberType"`
	SubscriberID      string    `json:"subscriberId"`
	GroveID           string    `json:"groveId"`
	TriggerActivities []string  `json:"triggerActivities"`
	CreatedAt         time.Time `json:"createdAt"`
	CreatedBy         string    `json:"createdBy"`
}

// Create creates a new notification subscription.
func (s *subscriptionService) Create(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error) {
	resp, err := s.c.transport.Post(ctx, "/api/v1/notifications/subscriptions", req, nil)
	if err != nil {
		return nil, err
	}
	return apiclient.DecodeResponse[Subscription](resp)
}

// List returns subscriptions for the current user.
func (s *subscriptionService) List(ctx context.Context, opts *ListSubscriptionsOptions) ([]Subscription, error) {
	query := url.Values{}
	if opts != nil {
		if opts.GroveID != "" {
			query.Set("groveId", opts.GroveID)
		}
		if opts.AgentID != "" {
			query.Set("agentId", opts.AgentID)
		}
		if opts.Scope != "" {
			query.Set("scope", opts.Scope)
		}
	}

	resp, err := s.c.transport.GetWithQuery(ctx, "/api/v1/notifications/subscriptions", query, nil)
	if err != nil {
		return nil, err
	}
	result, err := apiclient.DecodeResponse[[]Subscription](resp)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []Subscription{}, nil
	}
	return *result, nil
}

// Delete removes a subscription by ID.
func (s *subscriptionService) Delete(ctx context.Context, id string) error {
	resp, err := s.c.transport.Delete(ctx, "/api/v1/notifications/subscriptions/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	return apiclient.CheckResponse(resp)
}
