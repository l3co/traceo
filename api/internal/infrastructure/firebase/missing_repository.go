package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/l3co/traceo-api/internal/domain/missing"
)

const missingCollection = "missing"

type MissingRepository struct {
	client *firestore.Client
}

func NewMissingRepository(client *firestore.Client) *MissingRepository {
	return &MissingRepository{client: client}
}

type missingDoc struct {
	ID                  string    `firestore:"id"`
	UserID              string    `firestore:"user_id"`
	Name                string    `firestore:"name"`
	Nickname            string    `firestore:"nickname,omitempty"`
	BirthDate           time.Time `firestore:"birth_date"`
	DateOfDisappearance time.Time `firestore:"date_of_disappearance"`
	Height              string    `firestore:"height,omitempty"`
	Clothes             string    `firestore:"clothes,omitempty"`
	Gender              string    `firestore:"gender"`
	Eyes                string    `firestore:"eyes"`
	Hair                string    `firestore:"hair"`
	Skin                string    `firestore:"skin"`
	PhotoURL            string    `firestore:"photo_url,omitempty"`
	Lat                 float64   `firestore:"lat"`
	Lng                 float64   `firestore:"lng"`
	Status              string    `firestore:"status"`
	EventReport         string    `firestore:"event_report,omitempty"`
	TattooDescription   string    `firestore:"tattoo_description,omitempty"`
	ScarDescription     string    `firestore:"scar_description,omitempty"`
	WasChild            bool      `firestore:"was_child"`
	AgeProgressionURLs  []string  `firestore:"age_progression_urls,omitempty"`
	Slug                string    `firestore:"slug"`
	NameLowercase       string    `firestore:"name_lowercase"`
	CreatedAt           time.Time `firestore:"created_at"`
	UpdatedAt           time.Time `firestore:"updated_at"`
}

func toMissingDoc(m *missing.Missing) missingDoc {
	return missingDoc{
		ID:                  m.ID,
		UserID:              m.UserID,
		Name:                m.Name,
		Nickname:            m.Nickname,
		BirthDate:           m.BirthDate,
		DateOfDisappearance: m.DateOfDisappearance,
		Height:              m.Height,
		Clothes:             m.Clothes,
		Gender:              string(m.Gender),
		Eyes:                string(m.Eyes),
		Hair:                string(m.Hair),
		Skin:                string(m.Skin),
		PhotoURL:            m.PhotoURL,
		Lat:                 m.Location.Lat,
		Lng:                 m.Location.Lng,
		Status:              string(m.Status),
		EventReport:         m.EventReport,
		TattooDescription:   m.TattooDescription,
		ScarDescription:     m.ScarDescription,
		WasChild:            m.WasChild,
		AgeProgressionURLs:  m.AgeProgressionURLs,
		Slug:                m.Slug,
		NameLowercase:       m.NameLowercase,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func toMissingEntity(d missingDoc) *missing.Missing {
	return &missing.Missing{
		ID:                  d.ID,
		UserID:              d.UserID,
		Name:                d.Name,
		Nickname:            d.Nickname,
		BirthDate:           d.BirthDate,
		DateOfDisappearance: d.DateOfDisappearance,
		Height:              d.Height,
		Clothes:             d.Clothes,
		Gender:              missing.Gender(d.Gender),
		Eyes:                missing.EyeColor(d.Eyes),
		Hair:                missing.HairColor(d.Hair),
		Skin:                missing.SkinColor(d.Skin),
		PhotoURL:            d.PhotoURL,
		Location:            missing.GeoPoint{Lat: d.Lat, Lng: d.Lng},
		Status:              missing.Status(d.Status),
		EventReport:         d.EventReport,
		TattooDescription:   d.TattooDescription,
		ScarDescription:     d.ScarDescription,
		WasChild:            d.WasChild,
		AgeProgressionURLs:  d.AgeProgressionURLs,
		Slug:                d.Slug,
		NameLowercase:       d.NameLowercase,
		Timestamps: missing.Timestamps{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		},
	}
}

func (r *MissingRepository) Create(ctx context.Context, m *missing.Missing) error {
	_, err := r.client.Collection(missingCollection).Doc(m.ID).Set(ctx, toMissingDoc(m))
	if err != nil {
		return fmt.Errorf("firestore: creating missing %s: %w", m.ID, err)
	}
	return nil
}

func (r *MissingRepository) FindByID(ctx context.Context, id string) (*missing.Missing, error) {
	doc, err := r.client.Collection(missingCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, missing.ErrMissingNotFound
		}
		return nil, fmt.Errorf("firestore: finding missing %s: %w", id, err)
	}

	var d missingDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding missing %s: %w", id, err)
	}

	return toMissingEntity(d), nil
}

func (r *MissingRepository) Update(ctx context.Context, m *missing.Missing) error {
	_, err := r.client.Collection(missingCollection).Doc(m.ID).Set(ctx, toMissingDoc(m))
	if err != nil {
		return fmt.Errorf("firestore: updating missing %s: %w", m.ID, err)
	}
	return nil
}

func (r *MissingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(missingCollection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestore: deleting missing %s: %w", id, err)
	}
	return nil
}

func (r *MissingRepository) FindByUserID(ctx context.Context, userID string) ([]*missing.Missing, error) {
	iter := r.client.Collection(missingCollection).
		Where("user_id", "==", userID).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx)

	docs, err := iter.GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding missing by user %s: %w", userID, err)
	}

	result := make([]*missing.Missing, 0, len(docs))
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toMissingEntity(d))
	}

	return result, nil
}

func (r *MissingRepository) FindAll(ctx context.Context, opts missing.ListOptions) ([]*missing.Missing, string, error) {
	query := r.client.Collection(missingCollection).
		OrderBy("created_at", firestore.Desc).
		Limit(opts.PageSize)

	if opts.After != "" {
		cursorDoc, err := r.client.Collection(missingCollection).Doc(opts.After).Get(ctx)
		if err == nil {
			query = query.StartAfter(cursorDoc)
		}
	}

	if opts.UserID != "" {
		query = query.Where("user_id", "==", opts.UserID)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, "", fmt.Errorf("firestore: listing missing: %w", err)
	}

	result := make([]*missing.Missing, 0, len(docs))
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toMissingEntity(d))
	}

	var nextCursor string
	if len(docs) == opts.PageSize {
		nextCursor = docs[len(docs)-1].Ref.ID
	}

	return result, nextCursor, nil
}

func (r *MissingRepository) Count(ctx context.Context) (int64, error) {
	docs, err := r.client.Collection(missingCollection).Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("firestore: counting missing: %w", err)
	}
	return int64(len(docs)), nil
}

func (r *MissingRepository) Search(ctx context.Context, query string, limit int) ([]*missing.Missing, error) {
	q := strings.ToLower(query)
	docs, err := r.client.Collection(missingCollection).
		Where("name_lowercase", ">=", q).
		Where("name_lowercase", "<=", q+"\uf8ff").
		Limit(limit).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: searching missing: %w", err)
	}

	result := make([]*missing.Missing, 0, len(docs))
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toMissingEntity(d))
	}
	return result, nil
}

func (r *MissingRepository) CountByGender(ctx context.Context) ([]missing.GenderStat, error) {
	docs, err := r.client.Collection(missingCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: counting by gender: %w", err)
	}

	counts := map[string]int64{}
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		counts[d.Gender]++
	}

	result := make([]missing.GenderStat, 0, len(counts))
	for g, c := range counts {
		result = append(result, missing.GenderStat{Gender: g, Count: c})
	}
	return result, nil
}

func (r *MissingRepository) CountByYear(ctx context.Context) ([]missing.YearStat, error) {
	docs, err := r.client.Collection(missingCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: counting by year: %w", err)
	}

	counts := map[int]int64{}
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		if !d.DateOfDisappearance.IsZero() {
			counts[d.DateOfDisappearance.Year()]++
		}
	}

	result := make([]missing.YearStat, 0, len(counts))
	for y, c := range counts {
		result = append(result, missing.YearStat{Year: y, Count: c})
	}
	return result, nil
}

func (r *MissingRepository) CountChildren(ctx context.Context) (int64, error) {
	docs, err := r.client.Collection(missingCollection).
		Where("was_child", "==", true).
		Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("firestore: counting children: %w", err)
	}
	return int64(len(docs)), nil
}

func (r *MissingRepository) FindLocations(ctx context.Context, limit int) ([]missing.LocationPoint, error) {
	docs, err := r.client.Collection(missingCollection).
		Limit(limit).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding locations: %w", err)
	}

	result := make([]missing.LocationPoint, 0, len(docs))
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		if d.Lat == 0 && d.Lng == 0 {
			continue
		}
		result = append(result, missing.LocationPoint{
			ID:     d.ID,
			Name:   d.Name,
			Lat:    d.Lat,
			Lng:    d.Lng,
			Status: missing.Status(d.Status),
		})
	}
	return result, nil
}

func (r *MissingRepository) FindCandidates(ctx context.Context, filter missing.CandidateFilter) ([]*missing.Missing, error) {
	query := r.client.Collection(missingCollection).
		Where("status", "==", string(filter.Status)).
		Where("gender", "==", string(filter.Gender)).
		Where("skin", "==", string(filter.Skin))

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding candidates: %w", err)
	}

	result := make([]*missing.Missing, 0, len(docs))
	for _, doc := range docs {
		var d missingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		entity := toMissingEntity(d)
		age := entity.Age()
		if filter.MinAge > 0 && age < filter.MinAge {
			continue
		}
		if filter.MaxAge > 0 && age > filter.MaxAge {
			continue
		}
		result = append(result, entity)
	}

	return result, nil
}

func (r *MissingRepository) UpdateAgeProgressionURLs(ctx context.Context, id string, urls []string) error {
	_, err := r.client.Collection(missingCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "age_progression_urls", Value: urls},
		{Path: "updated_at", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("firestore: updating age progression urls for %s: %w", id, err)
	}
	return nil
}
