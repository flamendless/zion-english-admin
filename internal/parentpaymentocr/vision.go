package parentpaymentocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const visionAnnotateURL = "https://vision.googleapis.com/v1/images:annotate"

type visionRequest struct {
	Requests []visionImageRequest `json:"requests"`
}

type visionImageRequest struct {
	Image    visionImage     `json:"image"`
	Features []visionFeature `json:"features"`
}

type visionImage struct {
	Content string `json:"content"`
}

type visionFeature struct {
	Type string `json:"type"`
}

type visionResponse struct {
	Responses []visionAnnotateResponse `json:"responses"`
}

type visionAnnotateResponse struct {
	FullTextAnnotation *visionFullText `json:"fullTextAnnotation"`
	Error              *visionError    `json:"error"`
}

type visionFullText struct {
	Text string `json:"text"`
}

type visionError struct {
	Message string `json:"message"`
}

// DetectDocumentText runs Google Cloud Vision DOCUMENT_TEXT_DETECTION on image bytes.
func DetectDocumentText(ctx context.Context, apiKey string, image []byte) (string, error) {
	if strings.TrimSpace(apiKey) == "" {
		return "", ErrVisionAPIKeyMissing
	}
	payload, err := json.Marshal(visionRequest{
		Requests: []visionImageRequest{{
			Image: visionImage{Content: base64.StdEncoding.EncodeToString(image)},
			Features: []visionFeature{{
				Type: "DOCUMENT_TEXT_DETECTION",
			}},
		}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, visionAnnotateURL+"?key="+apiKey, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrVisionRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d", ErrVisionRequestFailed, resp.StatusCode)
	}

	var parsed visionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Responses) == 0 {
		return "", ErrVisionNoText
	}
	first := parsed.Responses[0]
	if first.Error != nil && first.Error.Message != "" {
		return "", fmt.Errorf("%w: %s", ErrVisionRequestFailed, first.Error.Message)
	}
	if first.FullTextAnnotation == nil || strings.TrimSpace(first.FullTextAnnotation.Text) == "" {
		return "", ErrVisionNoText
	}
	return first.FullTextAnnotation.Text, nil
}
