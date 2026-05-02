package smc

import (
	"context"
	"fmt"
)

// SMSProvider is an interface for SMS sending.
// Production would integrate Alibaba Cloud / Tencent Cloud SMS.
type SMSProvider interface {
	SendCode(ctx context.Context, mobile, code string) error
}

type MockSMSProvider struct{}

func (m *MockSMSProvider) SendCode(ctx context.Context, mobile, code string) error {
	// In production, integrate with Alibaba Cloud / Tencent Cloud SMS API.
	// This mock logs the code for testing.
	fmt.Printf("[MockSMS] Sending code %s to mobile %s\n", code, mobile)
	return nil
}
