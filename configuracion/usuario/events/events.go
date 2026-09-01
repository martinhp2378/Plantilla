package events

// SendUserActivationEvent is triggered when a newly-created user must choose a password.
type SendUserActivationEvent struct {
	Email         string
	FullName      string
	Username      string
	ActivationURL string
}

func (e SendUserActivationEvent) Name() string {
	return "SendUserActivationEvent"
}
