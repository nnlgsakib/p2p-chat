package chat

// Notifier is an interface for sending notifications.
type Notifier interface {
	NotifyNewMessage(senderID, message, messageType string)
}
