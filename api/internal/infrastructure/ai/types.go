package ai

type FaceComparison struct {
	SimilarityScore   float64  `json:"similarity_score"`
	Analysis          string   `json:"analysis"`
	MatchingFeatures  []string `json:"matching_features"`
	DifferentFeatures []string `json:"different_features"`
	Confidence        string   `json:"confidence"`
}
