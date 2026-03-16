package application

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type firebaseClaims struct {
	Subject string
	Name    string
	Email   string
}

func verifyFirebaseIDToken(ctx context.Context, credentialsPath, idToken string) (*firebaseClaims, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("firebase credentials not configured")
	}
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("firebase credentials file not found: %s", credentialsPath)
	}

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app init: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth client: %w", err)
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("verify id token: %w", err)
	}

	claims := token.Claims
	sub, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)
	email, _ := claims["email"].(string)

	return &firebaseClaims{
		Subject: sub,
		Name:    name,
		Email:   email,
	}, nil
}
