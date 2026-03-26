package diary

import (
	"errors"
	"time"

	"back/internal/modules/nutrition"
)

var (
	ErrEntryNotFound   = errors.New("diary entry not found")
	ErrTargetsNotFound = errors.New("user targets not found")
)

type Entry struct {
	ID             int64     `json:"id"`
	UserID         string    `json:"user_id"`
	Source         string    `json:"source"`
	Name           string    `json:"name"`
	AmountG        float64   `json:"amount_g"`
	Calories       float64   `json:"calories"`
	Protein        float64   `json:"protein"`
	Fat            float64   `json:"fat"`
	Carbs          float64   `json:"carbs"`
	Ingredients    []string  `json:"ingredients"`
	EntryDate      string    `json:"entry_date"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"-"`
}

type AddToDiaryInput struct {
	UserID                string
	Source                string              `json:"source"`
	Name                  string              `json:"name"`
	AmountG               float64             `json:"amount_g"`
	Per100G               nutrition.Nutrients `json:"per_100g"`
	Ingredients           []string            `json:"ingredients"`
	TimezoneOffsetMinutes int
	IdempotencyKey        string
}

type UpdateEntryInput struct {
	EntryID     int64
	AmountG     float64
	Name        string
	Source      string
	Per100G     nutrition.Nutrients
	Ingredients []string
}

type DailyTotals struct {
	UserID    string    `json:"user_id"`
	Date      string    `json:"date"`
	Calories  float64   `json:"calories"`
	Protein   float64   `json:"protein"`
	Fat       float64   `json:"fat"`
	Carbs     float64   `json:"carbs"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Progress struct {
	Target    float64 `json:"target"`
	Consumed  float64 `json:"consumed"`
	Remaining float64 `json:"remaining"`
	Progress  float64 `json:"progress"`
}

type TodayResponse struct {
	Date     string   `json:"date"`
	Calories Progress `json:"calories"`
	Protein  Progress `json:"protein"`
	Fat      Progress `json:"fat"`
	Carbs    Progress `json:"carbs"`
}
