package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type Client struct {
	Auth      *auth.Client
	Firestore *firestore.Client
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	app, err := fb.NewApp(ctx, &fb.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("firebase: initializing app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: initializing auth: %w", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: initializing firestore: %w", err)
	}

	return &Client{
		Auth:      authClient,
		Firestore: firestoreClient,
	}, nil
}

func (c *Client) Close() error {
	if c.Firestore != nil {
		return c.Firestore.Close()
	}
	return nil
}
