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
	e, ok := event.(events.SendUserActivationEvent)
	if !ok {
		return nil
	}

	log.Printf("Handling SendUserActivationEvent for %s", e.Email)
	err := l.emailService.SendUserActivation(ctx, e.Email, e.FullName, e.Username, e.ActivationURL)
	if err != nil {
		log.Printf("Error sending user credentials email to %s: %v", e.Email, err)
		return err
	}
	log.Printf("Successfully sent user activation email to %s", e.Email)
	return nil
}
