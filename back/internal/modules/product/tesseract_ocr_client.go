package product

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"strings"
)

// TesseractOCRClient extracts text from base64-encoded image bytes.
// For backwards compatibility with earlier stub behavior, it also supports
// base64-encoded UTF-8 text payloads.
type TesseractOCRClient struct {
	bin string
}

func NewTesseractOCRClient() (*TesseractOCRClient, error) {
	bin, err := exec.LookPath("tesseract")
	if err != nil {
		if st, statErr := os.Stat("/usr/bin/tesseract"); statErr == nil && !st.IsDir() {
			return &TesseractOCRClient{bin: "/usr/bin/tesseract"}, nil
		}
		return nil, fmt.Errorf("tesseract not found in PATH")
	}
	return &TesseractOCRClient{bin: bin}, nil
}

func (c *TesseractOCRClient) ExtractText(ctx context.Context, imageBase64, lang, _ string) (string, error) {
	imageBase64 = strings.TrimSpace(imageBase64)
	if imageBase64 == "" {
		return "", fmt.Errorf("%w: ocr image required", ErrInvalidOCRRequest)
	}

	decoded, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 image", ErrInvalidOCRRequest)
	}

	// Backwards-compatible behavior: allow passing base64(utf8(text)).
	if txt, ok := tryDecodeUTF8Text(decoded); ok {
		return strings.TrimSpace(txt), nil
	}

	lang = normalizeTesseractLang(lang)

	variants := make([]ocrImageVariant, 0, 2)
	variants = append(variants, ocrImageVariant{ext: detectImageExt(decoded), bytes: decoded})
	if bin, ok := tryMakeBinaryPNG(decoded); ok {
		variants = append(variants, ocrImageVariant{ext: ".png", bytes: bin})
	}

	psmModes := []string{"6", "4", "11"} // block, column, sparse text
	bestText := ""
	bestScore := 0.0
	var lastErr error
	for _, v := range variants {
		tmpPath, cleanup, err := writeTempOCRImage(v.ext, v.bytes)
		if err != nil {
			lastErr = err
			continue
		}
		for _, psm := range psmModes {
			text, runErr := runTesseract(ctx, c.bin, tmpPath, lang, psm)
			if runErr != nil {
				lastErr = runErr
				continue
			}
			score := scoreOCRText(text, lang)
			if score > bestScore || bestText == "" {
				bestText = text
				bestScore = score
			}
		}
		cleanup()
	}

	if strings.TrimSpace(bestText) == "" {
		if lastErr == nil {
			lastErr = fmt.Errorf("%w: no text found", ErrInvalidOCRRequest)
		}
		return "", lastErr
	}
	return strings.TrimSpace(bestText), nil
}

func normalizeTesseractLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "rus+eng"
	}
	lang = strings.ReplaceAll(lang, " ", "")
	switch lang {
	case "ru", "rus", "ru-ru", "russian":
		return "rus+eng"
	case "en", "eng", "en-us", "english":
		return "eng"
	}
	parts := strings.Split(lang, "+")
	hasRus := false
	hasEng := false
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		if p == "" {
			continue
		}
		switch p {
		case "ru":
			p = "rus"
		case "en":
			p = "eng"
		}
		if p == "rus" {
			hasRus = true
		}
		if p == "eng" {
			hasEng = true
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if hasRus && hasEng {
		return "rus+eng"
	}
	if len(out) == 0 {
		return "rus+eng"
	}
	return strings.Join(out, "+")
}

func detectImageExt(b []byte) string {
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return ".jpg"
	}
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}) {
		return ".png"
	}
	return ".img"
}

type ocrImageVariant struct {
	ext   string
	bytes []byte
}

func writeTempOCRImage(ext string, content []byte) (string, func(), error) {
	tmp, err := os.CreateTemp("", "ocr-*"+ext)
	if err != nil {
		return "", func() {}, fmt.Errorf("ocr temp file: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	if err := os.WriteFile(tmpPath, content, 0o600); err != nil {
		_ = os.Remove(tmpPath)
		return "", func() {}, fmt.Errorf("ocr write temp file: %w", err)
	}
	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}

func runTesseract(ctx context.Context, bin string, imgPath string, lang string, psm string) (string, error) {
	cmd := exec.CommandContext(ctx, bin,
		imgPath,
		"stdout",
		"-l", lang,
		"--oem", "1",
		"--psm", psm,
		"-c", "preserve_interword_spaces=1",
		"-c", "user_defined_dpi=300",
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%w: ocr failed: %s", ErrInvalidOCRRequest, msg)
	}
	return strings.TrimSpace(out.String()), nil
}

func scoreOCRText(text string, lang string) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return 0
	}

	letters := 0
	other := 0
	cyr := 0
	lat := 0
	for _, r := range runes {
		switch {
		case r == '\uFFFD':
			other++
		case r == '\n' || r == '\t' || r == ' ':
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			letters++
			lat++
		case r >= 0x0400 && r <= 0x052F:
			letters++
			cyr++
		default:
			if (r >= '0' && r <= '9') || strings.ContainsRune(".,:;/%()[]{}+-—–•'\"!?_", r) {
				continue
			}
			if r >= 0x20 {
				other++
			}
		}
	}

	base := estimateOCRQuality(text)
	letterQ := float64(letters) / float64(len(runes))
	otherPenalty := float64(other) / float64(len(runes))

	lang = strings.ToLower(lang)
	scriptBoost := 0.0
	if letters > 0 {
		switch {
		case strings.Contains(lang, "rus") && !strings.Contains(lang, "eng"):
			scriptBoost = float64(cyr) / float64(letters)
		case strings.Contains(lang, "eng") && !strings.Contains(lang, "rus"):
			scriptBoost = float64(lat) / float64(letters)
		case strings.Contains(lang, "rus") && strings.Contains(lang, "eng"):
			if cyr > lat {
				scriptBoost = float64(cyr) / float64(letters)
			} else {
				scriptBoost = float64(lat) / float64(letters)
			}
		}
	}

	return clamp01(0.55*base + 0.25*letterQ + 0.25*scriptBoost - 0.35*otherPenalty)
}

func tryMakeBinaryPNG(imgBytes []byte) ([]byte, bool) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil || img == nil {
		return nil, false
	}
	gray := toGray(img)
	thr := otsuThreshold(gray)
	bin := applyThreshold(gray, uint8(thr))
	if shouldInvertBinary(bin) {
		invertGray(bin)
	}
	if bin.Bounds().Dx() < 1800 && bin.Bounds().Dy() < 1800 {
		bin = upscaleGrayNearest(bin, 2)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, bin); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

func toGray(img image.Image) *image.Gray {
	b := img.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			g := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			out.SetGray(x, y, g)
		}
	}
	return out
}

func otsuThreshold(img *image.Gray) int {
	if img == nil {
		return 128
	}
	var hist [256]uint64
	b := img.Bounds()
	total := uint64(b.Dx() * b.Dy())
	if total == 0 {
		return 128
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := img.PixOffset(b.Min.X, y)
		for x := b.Min.X; x < b.Max.X; x++ {
			hist[img.Pix[i]]++
			i++
		}
	}
	var sum uint64
	for t := 0; t < 256; t++ {
		sum += uint64(t) * hist[t]
	}
	var sumB uint64
	var wB uint64
	var maxVar float64
	threshold := 128
	for t := 0; t < 256; t++ {
		wB += hist[t]
		if wB == 0 {
			continue
		}
		wF := total - wB
		if wF == 0 {
			break
		}
		sumB += uint64(t) * hist[t]
		mB := float64(sumB) / float64(wB)
		mF := float64(sum-sumB) / float64(wF)
		varBetween := float64(wB) * float64(wF) * (mB - mF) * (mB - mF)
		if varBetween > maxVar {
			maxVar = varBetween
			threshold = t
		}
	}
	return threshold
}

func applyThreshold(img *image.Gray, threshold uint8) *image.Gray {
	if img == nil {
		return image.NewGray(image.Rect(0, 0, 0, 0))
	}
	b := img.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := img.PixOffset(b.Min.X, y)
		j := out.PixOffset(b.Min.X, y)
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[i] > threshold {
				out.Pix[j] = 255
			} else {
				out.Pix[j] = 0
			}
			i++
			j++
		}
	}
	return out
}

func shouldInvertBinary(img *image.Gray) bool {
	if img == nil {
		return false
	}
	b := img.Bounds()
	total := uint64(b.Dx() * b.Dy())
	if total == 0 {
		return false
	}
	black := uint64(0)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := img.PixOffset(b.Min.X, y)
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[i] == 0 {
				black++
			}
			i++
		}
	}
	return float64(black)/float64(total) > 0.55
}

func invertGray(img *image.Gray) {
	if img == nil {
		return
	}
	for i := range img.Pix {
		img.Pix[i] = 255 - img.Pix[i]
	}
}

func upscaleGrayNearest(src *image.Gray, factor int) *image.Gray {
	if src == nil || factor <= 1 {
		return src
	}
	sb := src.Bounds()
	w := sb.Dx() * factor
	h := sb.Dy() * factor
	dst := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y/factor
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x/factor
			dst.SetGray(x, y, src.GrayAt(sx, sy))
		}
	}
	return dst
}
