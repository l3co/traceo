package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client     *genai.Client
	visionModel *genai.GenerativeModel
	httpClient *http.Client
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("creating gemini client: %w", err)
	}

	visionModel := client.GenerativeModel("gemini-2.0-flash")
	visionModel.SetTemperature(0.4)
	visionModel.ResponseMIMEType = "application/json"

	return &GeminiClient{
		client:      client,
		visionModel: visionModel,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (g *GeminiClient) Close() error {
	return g.client.Close()
}

func (g *GeminiClient) CompareFaces(ctx context.Context, photo1URL, photo2URL string) (*FaceComparison, error) {
	img1, mime1, err := g.downloadImage(ctx, photo1URL)
	if err != nil {
		return nil, fmt.Errorf("downloading photo1: %w", err)
	}
	img2, mime2, err := g.downloadImage(ctx, photo2URL)
	if err != nil {
		return nil, fmt.Errorf("downloading photo2: %w", err)
	}

	prompt := `Analyze these two facial photos and compare them.
Consider: facial structure, eye shape and color, nose shape, lip shape,
skin tone, face shape, distinguishing features.

Respond in JSON format:
{
    "similarity_score": 0.0-1.0,
    "analysis": "detailed explanation in Portuguese",
    "matching_features": ["feature1", "feature2"],
    "different_features": ["feature1", "feature2"],
    "confidence": "high" | "medium" | "low"
}

Be conservative with scores. Only score above 0.7 if there is strong facial resemblance.
Consider that one photo may be older (age difference is expected).`

	resp, err := g.visionModel.GenerateContent(ctx,
		genai.ImageData(mime1, img1),
		genai.ImageData(mime2, img2),
		genai.Text(prompt),
	)
	if err != nil {
		return nil, fmt.Errorf("gemini compare faces: %w", err)
	}

	return parseJSONFromGemini[FaceComparison](resp)
}

func (g *GeminiClient) DescribeFace(ctx context.Context, photoURL string, currentAge int, gender string) (string, error) {
	img, mime, err := g.downloadImage(ctx, photoURL)
	if err != nil {
		return "", fmt.Errorf("downloading photo: %w", err)
	}

	prompt := fmt.Sprintf(`Describe this person's facial features in detail for age progression.
Current age: %d years old. Gender: %s.
Focus on: bone structure, eye shape, nose shape, lip shape, skin characteristics,
hair pattern, distinguishing marks.
Be specific and detailed. Respond in English.`, currentAge, gender)

	resp, err := g.visionModel.GenerateContent(ctx,
		genai.ImageData(mime, img),
		genai.Text(prompt),
	)
	if err != nil {
		return "", fmt.Errorf("gemini describe face: %w", err)
	}

	return extractText(resp), nil
}

func (g *GeminiClient) downloadImage(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("image download status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading image body: %w", err)
	}

	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/jpeg"
	}

	return data, mime, nil
}
