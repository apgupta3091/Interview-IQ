package service

// Billing/Stripe logic is preserved here but disabled.
// Payments are removed for now — all users get full access.
// Re-enable by uncommenting this file and wiring it back into main.go.

/*
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v82"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/webhook"

	"github.com/apgupta3091/interview-iq/internal/repository"
)

type BillingService interface {
	// CreateCheckoutSession returns a Stripe-hosted checkout URL for the given price.
	// Redirect the user to this URL to complete payment.
	CreateCheckoutSession(ctx context.Context, userID int, priceID, successURL, cancelURL string) (string, error)
	// CreatePortalSession returns a Stripe Billing Portal URL so the user can manage
	// their subscription (cancel, update payment method, etc.).
	CreatePortalSession(ctx context.Context, userID int, returnURL string) (string, error)
	// HandleWebhook verifies the Stripe signature and updates the user's tier based
	// on the event type. Unknown events are silently acknowledged.
	HandleWebhook(ctx context.Context, payload []byte, sigHeader, secret string) error
}

type billingService struct {
	users repository.UserRepository
}

// NewBillingService creates a billing service and initialises the Stripe SDK with secretKey.
func NewBillingService(users repository.UserRepository, secretKey string) BillingService {
	stripe.Key = secretKey
	return &billingService{users: users}
}

func (s *billingService) CreateCheckoutSession(ctx context.Context, userID int, priceID, successURL, cancelURL string) (string, error) {
	customerID, _, err := s.users.GetBilling(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("CreateCheckoutSession: get billing: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// Store our internal user_id so the webhook can resolve it without a
		// customer lookup (handles the case where no customer exists yet).
		Metadata: map[string]string{
			"user_id": strconv.Itoa(userID),
		},
	}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("CreateCheckoutSession: stripe: %w", err)
	}
	return sess.URL, nil
}

func (s *billingService) CreatePortalSession(ctx context.Context, userID int, returnURL string) (string, error) {
	customerID, _, err := s.users.GetBilling(ctx, userID)
	if err != nil || customerID == "" {
		return "", fmt.Errorf("CreatePortalSession: no Stripe customer for user %d", userID)
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	sess, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("CreatePortalSession: stripe: %w", err)
	}
	return sess.URL, nil
}

// HandleWebhook verifies the Stripe signature then dispatches on event type.
func (s *billingService) HandleWebhook(ctx context.Context, payload []byte, sigHeader, secret string) error {
	event, err := webhook.ConstructEvent(payload, sigHeader, secret)
	if err != nil {
		return fmt.Errorf("HandleWebhook: verify signature: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		// User completed checkout — activate Pro and store their Stripe customer ID.
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return fmt.Errorf("HandleWebhook: unmarshal session: %w", err)
		}
		userID, err := strconv.Atoi(sess.Metadata["user_id"])
		if err != nil {
			return fmt.Errorf("HandleWebhook: invalid user_id in metadata: %w", err)
		}
		customerID := ""
		if sess.Customer != nil {
			customerID = sess.Customer.ID
		}
		if err := s.users.UpdateBilling(ctx, userID, customerID, "pro"); err != nil {
			return fmt.Errorf("HandleWebhook: update billing (activate): %w", err)
		}

	case "customer.subscription.deleted":
		// Subscription cancelled — revert to free tier.
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("HandleWebhook: unmarshal subscription: %w", err)
		}
		if sub.Customer == nil {
			return nil
		}
		userID, err := s.users.GetUserIDByStripeCustomerID(ctx, sub.Customer.ID)
		if err != nil {
			return fmt.Errorf("HandleWebhook: find user by customer: %w", err)
		}
		if err := s.users.UpdateBilling(ctx, userID, sub.Customer.ID, "free"); err != nil {
			return fmt.Errorf("HandleWebhook: update billing (cancel): %w", err)
		}

	case "customer.subscription.updated":
		// Status may have changed (e.g. payment failed → past_due, or reactivated).
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return fmt.Errorf("HandleWebhook: unmarshal subscription update: %w", err)
		}
		if sub.Customer == nil {
			return nil
		}
		userID, err := s.users.GetUserIDByStripeCustomerID(ctx, sub.Customer.ID)
		if err != nil {
			return fmt.Errorf("HandleWebhook: find user by customer (update): %w", err)
		}
		tier := "free"
		if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing {
			tier = "pro"
		}
		if err := s.users.UpdateBilling(ctx, userID, sub.Customer.ID, tier); err != nil {
			return fmt.Errorf("HandleWebhook: update billing (update): %w", err)
		}
	}

	return nil
}
*/
