package oauth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleProfile struct {
	GoogleID string
	Email    string
	Name     string
}

// VerifyIDToken verifikasi ID token yang dikirim Flutter (hasil google_sign_in)
// langsung ke server Google, bukan decode manual — supaya gak bisa dipalsukan.
func VerifyIDToken(ctx context.Context, rawToken, audience string) (*GoogleProfile, error) {
	payload, err := idtoken.Validate(ctx, rawToken, audience)
	if err != nil {
		return nil, fmt.Errorf("invalid google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)

	return &GoogleProfile{
		GoogleID: payload.Subject,
		Email:    email,
		Name:     name,
	}, nil
}
