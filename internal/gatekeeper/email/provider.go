package email

// Provider defines the interface for sending email messages.
type Provider interface {
	Send(to, subject, body string) error
}
