package nutrition

import (
	"fmt"
	"math"
)

type Nutrients struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type TargetProfile struct {
	WeightKG      float64
	HeightCM      float64
	Age           int
	Gender        string
	ActivityLevel string
	Goal          string
}

func Add(a, b Nutrients) Nutrients {
	return Nutrients{
		Calories: a.Calories + b.Calories,
		Protein:  a.Protein + b.Protein,
		Fat:      a.Fat + b.Fat,
		Carbs:    a.Carbs + b.Carbs,
	}
}

func Subtract(a, b Nutrients) Nutrients {
	return Nutrients{
		Calories: a.Calories - b.Calories,
		Protein:  a.Protein - b.Protein,
		Fat:      a.Fat - b.Fat,
		Carbs:    a.Carbs - b.Carbs,
	}
}

func ScalePer100g(per100g Nutrients, amountG float64) (Nutrients, error) {
	if amountG <= 0 {
		return Nutrients{}, fmt.Errorf("amount_g must be > 0")
	}
	f := amountG / 100.0
	return Nutrients{
		Calories: per100g.Calories * f,
		Protein:  per100g.Protein * f,
		Fat:      per100g.Fat * f,
		Carbs:    per100g.Carbs * f,
	}, nil
}

func Per100gFromTotal(total Nutrients, amountG float64) (Nutrients, error) {
	if amountG <= 0 {
		return Nutrients{}, fmt.Errorf("amount_g must be > 0")
	}
	f := amountG / 100.0
	if f == 0 {
		return Nutrients{}, fmt.Errorf("amount_g must be > 0")
	}
	return Nutrients{
		Calories: total.Calories / f,
		Protein:  total.Protein / f,
		Fat:      total.Fat / f,
		Carbs:    total.Carbs / f,
	}, nil
}

func CalculateDailyTargets(p TargetProfile) (Nutrients, error) {
	if p.WeightKG <= 0 {
		return Nutrients{}, fmt.Errorf("weight_kg must be > 0")
	}
	if p.HeightCM <= 0 {
		return Nutrients{}, fmt.Errorf("height_cm must be > 0")
	}
	if p.Age <= 0 {
		return Nutrients{}, fmt.Errorf("age must be > 0")
	}

	bmr, err := mifflinStJeor(p)
	if err != nil {
		return Nutrients{}, err
	}

	multiplier, err := activityMultiplier(p.ActivityLevel)
	if err != nil {
		return Nutrients{}, err
	}

	calories := bmr * multiplier
	switch p.Goal {
	case "lose":
		calories -= 500
	case "maintain":
	case "gain":
		calories += 500
	default:
		return Nutrients{}, fmt.Errorf("invalid goal")
	}

	protein := p.WeightKG * 1.8
	fat := p.WeightKG * 0.9
	carbs := (calories - protein*4 - fat*9) / 4
	if carbs < 0 {
		carbs = 0
	}

	return Nutrients{
		Calories: round(calories, 0),
		Protein:  round(protein, 2),
		Fat:      round(fat, 2),
		Carbs:    round(carbs, 2),
	}, nil
}

func mifflinStJeor(p TargetProfile) (float64, error) {
	base := 10*p.WeightKG + 6.25*p.HeightCM - 5*float64(p.Age)
	switch p.Gender {
	case "male":
		return base + 5, nil
	case "female":
		return base - 161, nil
	default:
		return 0, fmt.Errorf("invalid gender")
	}
}

func activityMultiplier(level string) (float64, error) {
	switch level {
	case "sedentary":
		return 1.2, nil
	case "light":
		return 1.375, nil
	case "moderate":
		return 1.55, nil
	case "active":
		return 1.725, nil
	default:
		return 0, fmt.Errorf("invalid activity_level")
	}
}

func round(value float64, digits int) float64 {
	factor := math.Pow(10, float64(digits))
	return math.Round(value*factor) / factor
}
