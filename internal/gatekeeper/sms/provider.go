package sms

// Provider defines the interface for sending SMS messages.
type Provider interface {
	Send(phoneNumber, message string) error
}
