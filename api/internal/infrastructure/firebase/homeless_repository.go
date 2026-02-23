package firebase

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/shared"
)

const homelessCollection = "homeless"

type HomelessRepository struct {
	client *firestore.Client
}

func NewHomelessRepository(client *firestore.Client) *HomelessRepository {
	return &HomelessRepository{client: client}
}

type homelessDoc struct {
	ID        string    `firestore:"id"`
	Name      string    `firestore:"name"`
	Nickname  string    `firestore:"nickname,omitempty"`
	BirthDate time.Time `firestore:"birth_date"`
	Gender    string    `firestore:"gender"`
	Eyes      string    `firestore:"eyes"`
	Hair      string    `firestore:"hair"`
	Skin      string    `firestore:"skin"`
	PhotoURL  string    `firestore:"photo_url,omitempty"`
	Lat       float64   `firestore:"lat"`
	Lng       float64   `firestore:"lng"`
	Slug      string    `firestore:"slug"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}

func toHomelessDoc(h *homeless.Homeless) homelessDoc {
	return homelessDoc{
		ID:        h.ID,
		Name:      h.Name,
		Nickname:  h.Nickname,
		BirthDate: h.BirthDate,
		Gender:    string(h.Gender),
		Eyes:      string(h.Eyes),
		Hair:      string(h.Hair),
		Skin:      string(h.Skin),
		PhotoURL:  h.PhotoURL,
		Lat:       h.Location.Lat,
		Lng:       h.Location.Lng,
		Slug:      h.Slug,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
}

func toHomelessEntity(d homelessDoc) *homeless.Homeless {
	return &homeless.Homeless{
		ID:        d.ID,
		Name:      d.Name,
		Nickname:  d.Nickname,
		BirthDate: d.BirthDate,
		Gender:    shared.Gender(d.Gender),
		Eyes:      shared.EyeColor(d.Eyes),
		Hair:      shared.HairColor(d.Hair),
		Skin:      shared.SkinColor(d.Skin),
		PhotoURL:  d.PhotoURL,
		Location:  shared.GeoPoint{Lat: d.Lat, Lng: d.Lng},
		Slug:      d.Slug,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func (r *HomelessRepository) Create(ctx context.Context, h *homeless.Homeless) error {
	_, err := r.client.Collection(homelessCollection).Doc(h.ID).Set(ctx, toHomelessDoc(h))
	if err != nil {
		return fmt.Errorf("firestore: creating homeless: %w", err)
	}
	return nil
}

func (r *HomelessRepository) FindByID(ctx context.Context, id string) (*homeless.Homeless, error) {
	doc, err := r.client.Collection(homelessCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, homeless.ErrHomelessNotFound
		}
		return nil, fmt.Errorf("firestore: finding homeless: %w", err)
	}

	var d homelessDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding homeless: %w", err)
	}

	return toHomelessEntity(d), nil
}

func (r *HomelessRepository) FindAll(ctx context.Context) ([]*homeless.Homeless, error) {
	docs, err := r.client.Collection(homelessCollection).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: listing homeless: %w", err)
	}

	result := make([]*homeless.Homeless, 0, len(docs))
	for _, doc := range docs {
		var d homelessDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toHomelessEntity(d))
	}

	return result, nil
}

func (r *HomelessRepository) Count(ctx context.Context) (int64, error) {
	docs, err := r.client.Collection(homelessCollection).Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("firestore: counting homeless: %w", err)
	}
	return int64(len(docs)), nil
}

func (r *HomelessRepository) CountByGender(ctx context.Context) ([]homeless.GenderStat, error) {
	docs, err := r.client.Collection(homelessCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: counting homeless by gender: %w", err)
	}

	counts := map[string]int64{}
	for _, doc := range docs {
		var d homelessDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		counts[d.Gender]++
	}

	result := make([]homeless.GenderStat, 0, len(counts))
	for g, c := range counts {
		result = append(result, homeless.GenderStat{Gender: shared.Gender(g), Count: c})
	}
	return result, nil
}
