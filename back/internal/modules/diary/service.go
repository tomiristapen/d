package diary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"back/internal/modules/nutrition"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) AddEntry(ctx context.Context, e Entry) (Entry, error) {
	if strings.TrimSpace(e.UserID) == "" {
		return Entry{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(e.Source) == "" {
		return Entry{}, fmt.Errorf("source is required")
	}
	if strings.TrimSpace(e.Name) == "" {
		return Entry{}, fmt.Errorf("name is required")
	}
	if e.AmountG <= 0 {
		return Entry{}, fmt.Errorf("amount_g must be > 0")
	}
	if strings.TrimSpace(e.EntryDate) == "" {
		e.EntryDate = s.localDateString(0)
	}
	return s.repo.Create(ctx, e)
}

func (s *Service) AddToDiary(ctx context.Context, input AddToDiaryInput) (Entry, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return Entry{}, fmt.Errorf("user_id is required")
	}
	if !isValidSource(input.Source) {
		return Entry{}, fmt.Errorf("invalid source")
	}
	if strings.TrimSpace(input.Name) == "" {
		return Entry{}, fmt.Errorf("name is required")
	}
	if input.AmountG <= 0 {
		return Entry{}, fmt.Errorf("amount_g must be > 0")
	}
	if err := validateTimezoneOffset(input.TimezoneOffsetMinutes); err != nil {
		return Entry{}, err
	}

	scaled, err := nutrition.ScalePer100g(input.Per100G, input.AmountG)
	if err != nil {
		return Entry{}, err
	}

	return s.repo.Create(ctx, Entry{
		UserID:         input.UserID,
		Source:         input.Source,
		Name:           strings.TrimSpace(input.Name),
		AmountG:        input.AmountG,
		Calories:       scaled.Calories,
		Protein:        scaled.Protein,
		Fat:            scaled.Fat,
		Carbs:          scaled.Carbs,
		Ingredients:    sanitizeIngredients(input.Ingredients),
		EntryDate:      s.localDateString(input.TimezoneOffsetMinutes),
		IdempotencyKey: strings.TrimSpace(input.IdempotencyKey),
	})
}

func (s *Service) GetDailyTotals(ctx context.Context, userID string, offsetMinutes int) (DailyTotals, error) {
	if strings.TrimSpace(userID) == "" {
		return DailyTotals{}, fmt.Errorf("user_id is required")
	}
	if err := validateTimezoneOffset(offsetMinutes); err != nil {
		return DailyTotals{}, err
	}
	return s.repo.GetDailyTotals(ctx, userID, s.localDate(offsetMinutes))
}

func (s *Service) GetToday(ctx context.Context, userID string, offsetMinutes int) (TodayResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return TodayResponse{}, fmt.Errorf("user_id is required")
	}
	if err := validateTimezoneOffset(offsetMinutes); err != nil {
		return TodayResponse{}, err
	}

	targets, err := s.repo.GetTargets(ctx, userID)
	if err != nil {
		return TodayResponse{}, err
	}
	totals, err := s.repo.GetDailyTotals(ctx, userID, s.localDate(offsetMinutes))
	if err != nil {
		return TodayResponse{}, err
	}

	return TodayResponse{
		Date:     totals.Date,
		Calories: progressFrom(targets.Calories, totals.Calories),
		Protein:  progressFrom(targets.Protein, totals.Protein),
		Fat:      progressFrom(targets.Fat, totals.Fat),
		Carbs:    progressFrom(targets.Carbs, totals.Carbs),
	}, nil
}

func (s *Service) DeleteDiaryEntry(ctx context.Context, userID string, entryID int64) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user_id is required")
	}
	if entryID <= 0 {
		return fmt.Errorf("entry_id must be > 0")
	}
	return s.repo.DeleteEntry(ctx, userID, entryID)
}

func (s *Service) UpdateDiaryEntry(ctx context.Context, userID string, input UpdateEntryInput) (Entry, error) {
	if strings.TrimSpace(userID) == "" {
		return Entry{}, fmt.Errorf("user_id is required")
	}
	if input.EntryID <= 0 {
		return Entry{}, fmt.Errorf("entry_id must be > 0")
	}
	if !isValidSource(input.Source) {
		return Entry{}, fmt.Errorf("invalid source")
	}
	if strings.TrimSpace(input.Name) == "" {
		return Entry{}, fmt.Errorf("name is required")
	}
	if input.AmountG <= 0 {
		return Entry{}, fmt.Errorf("amount_g must be > 0")
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Ingredients = sanitizeIngredients(input.Ingredients)
	return s.repo.UpdateEntry(ctx, userID, input)
}

func (s *Service) localDate(offsetMinutes int) time.Time {
	local := s.now().Add(time.Duration(offsetMinutes) * time.Minute)
	year, month, day := local.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func (s *Service) localDateString(offsetMinutes int) string {
	return s.localDate(offsetMinutes).Format("2006-01-02")
}

func isValidSource(source string) bool {
	switch strings.TrimSpace(source) {
	case "barcode", "ocr", "manual", "recipe":
		return true
	default:
		return false
	}
}

func validateTimezoneOffset(offsetMinutes int) error {
	if offsetMinutes < -14*60 || offsetMinutes > 14*60 {
		return fmt.Errorf("invalid timezone offset")
	}
	return nil
}

func sanitizeIngredients(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func progressFrom(target float64, consumed float64) Progress {
	progress := 0.0
	if target > 0 {
		progress = consumed / target
	}
	return Progress{
		Target:    target,
		Consumed:  consumed,
		Remaining: target - consumed,
		Progress:  progress,
	}
}
