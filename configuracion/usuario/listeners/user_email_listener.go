package listeners

import (
	"PlantillaInstitucional/configuracion/usuario/events"
	"PlantillaInstitucional/shared/email"
	sharedEvents "PlantillaInstitucional/shared/events"
	"context"
	"log"
)

type UserEmailListener struct {
	emailService *email.EmailService
}

func NewUserEmailListener(emailService *email.EmailService) *UserEmailListener {
	return &UserEmailListener{
		emailService: emailService,
	}
}

func (l *UserEmailListener) Handle(ctx context.Context, event sharedEvents.Event) error {
	e, ok := event.(events.SendUserCredentialsEvent)
	if !ok {
		return nil
	}

	log.Printf("Handling SendUserCredentialsEvent for %s", e.Email)
	err := l.emailService.SendUserCredentials(e.Email, e.FullName, e.Username, e.Password, e.CredentialsURL)
	if err != nil {
		log.Printf("Error sending user credentials email to %s: %v", e.Email, err)
		return err
	}
	log.Printf("Successfully sent user credentials email to %s", e.Email)
	return nil
}
