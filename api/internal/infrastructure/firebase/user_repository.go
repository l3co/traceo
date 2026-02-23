package firebase

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/l3co/traceo-api/internal/domain/user"
)

const usersCollection = "users"

type UserRepository struct {
	client *firestore.Client
}

func NewUserRepository(client *firestore.Client) *UserRepository {
	return &UserRepository{client: client}
}

type userDoc struct {
	ID            string    `firestore:"id"`
	Name          string    `firestore:"name"`
	Email         string    `firestore:"email"`
	Phone         string    `firestore:"phone,omitempty"`
	CellPhone     string    `firestore:"cell_phone,omitempty"`
	AvatarURL     string    `firestore:"avatar_url,omitempty"`
	AcceptedTerms bool      `firestore:"accepted_terms"`
	CreatedAt     time.Time `firestore:"created_at"`
	UpdatedAt     time.Time `firestore:"updated_at"`
}

func toDoc(u *user.User) userDoc {
	return userDoc{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Phone:         u.Phone,
		CellPhone:     u.CellPhone,
		AvatarURL:     u.AvatarURL,
		AcceptedTerms: u.AcceptedTerms,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func toEntity(d userDoc) *user.User {
	return &user.User{
		ID:            d.ID,
		Name:          d.Name,
		Email:         d.Email,
		Phone:         d.Phone,
		CellPhone:     d.CellPhone,
		AvatarURL:     d.AvatarURL,
		AcceptedTerms: d.AcceptedTerms,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.client.Collection(usersCollection).Doc(u.ID).Set(ctx, toDoc(u))
	if err != nil {
		return fmt.Errorf("firestore: creating user %s: %w", u.ID, err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	doc, err := r.client.Collection(usersCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("firestore: finding user %s: %w", id, err)
	}

	var d userDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding user %s: %w", id, err)
	}

	return toEntity(d), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	iter := r.client.Collection(usersCollection).Where("email", "==", email).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	var d userDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding user by email %s: %w", email, err)
	}

	return toEntity(d), nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.client.Collection(usersCollection).Doc(u.ID).Set(ctx, toDoc(u))
	if err != nil {
		return fmt.Errorf("firestore: updating user %s: %w", u.ID, err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(usersCollection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestore: deleting user %s: %w", id, err)
	}
	return nil
}
