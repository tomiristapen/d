package product

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"
)

// StubOCRClient is a safe fallback when a real OCR engine isn't available.
// It accepts base64-encoded UTF-8 text payloads (dev/legacy) and rejects binary images.
type StubOCRClient struct{}

func NewStubOCRClient() *StubOCRClient { return &StubOCRClient{} }

func (c *StubOCRClient) ExtractText(_ context.Context, imageBase64, _ string, _ string) (string, error) {
	imageBase64 = strings.TrimSpace(imageBase64)
	if imageBase64 == "" {
		return "", fmt.Errorf("%w: ocr image required", ErrInvalidOCRRequest)
	}
	decoded, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 image", ErrInvalidOCRRequest)
	}
	if txt, ok := tryDecodeUTF8Text(decoded); ok {
		return strings.TrimSpace(txt), nil
	}
	return "", fmt.Errorf("%w: ocr engine not configured (install tesseract in backend)", ErrInvalidOCRRequest)
}

func tryDecodeUTF8Text(b []byte) (string, bool) {
	if bytes.IndexByte(b, 0) >= 0 {
		return "", false
	}
	if !utf8.Valid(b) {
		return "", false
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", false
	}
	printable := 0
	runes := []rune(s)
	for _, r := range runes {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			printable++
			continue
		}
		if r >= 0x20 && r != 0x7F {
			printable++
		}
	}
	if len(runes) == 0 {
		return "", false
	}
	if float64(printable)/float64(len(runes)) >= 0.9 {
		return s, true
	}
	return "", false
}
