// Package sms provides implementations of domain.OTPSender.
package sms

import (
	"context"

	"go.uber.org/zap"
)

// ConsoleSender implements domain.OTPSender by logging the code instead of
// sending an SMS. For beta only: cmd/main.go refuses to use it in prod.
type ConsoleSender struct {
	log *zap.Logger
}

// NewConsoleSender constructs a ConsoleSender that writes to log.
func NewConsoleSender(log *zap.Logger) *ConsoleSender {
	return &ConsoleSender{log: log}
}

// Send logs the code at info level so it can be read from the service output.
func (s *ConsoleSender) Send(_ context.Context, phone, code string) error {
	s.log.Info("OTP code (console sender, not sent by SMS)", zap.String("phone", phone), zap.String("code", code))
	return nil
}
