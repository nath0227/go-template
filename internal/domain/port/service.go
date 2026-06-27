package port

import "context"

// NotificationService is an example external service interface.
// Add interfaces here for any external services the usecase depends on.
type NotificationService interface {
	Notify(ctx context.Context, userID int64, message string) error
}
