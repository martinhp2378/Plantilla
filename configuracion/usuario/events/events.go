package events

// SendUserCredentialsEvent is the event triggered when a new user is created and needs their credentials.
type SendUserCredentialsEvent struct {
	Email          string
	FullName       string
	Username       string
	Password       string
	CredentialsURL string
}

func (e SendUserCredentialsEvent) Name() string {
	return "SendUserCredentialsEvent"
}
