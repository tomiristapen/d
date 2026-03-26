package nutrition

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScalePer100g(t *testing.T) {
	per := Nutrients{
		Calories: 100,
		Protein:  10,
		Fat:      5,
		Carbs:    20,
	}

	got, err := ScalePer100g(per, 200)
	require.NoError(t, err)
	require.Equal(t, Nutrients{
		Calories: 200,
		Protein:  20,
		Fat:      10,
		Carbs:    40,
	}, got)
}

func TestScalePer100g_InvalidAmount(t *testing.T) {
	_, err := ScalePer100g(Nutrients{Calories: 100}, 0)
	require.Error(t, err)

	_, err = ScalePer100g(Nutrients{Calories: 100}, -5)
	require.Error(t, err)
}

func TestAdd(t *testing.T) {
	a := Nutrients{Calories: 10, Protein: 1, Fat: 2, Carbs: 3}
	b := Nutrients{Calories: 5, Protein: 0.5, Fat: 1, Carbs: 0}
	require.Equal(t, Nutrients{Calories: 15, Protein: 1.5, Fat: 3, Carbs: 3}, Add(a, b))
}

func TestSubtract(t *testing.T) {
	a := Nutrients{Calories: 10, Protein: 4, Fat: 3, Carbs: 9}
	b := Nutrients{Calories: 5, Protein: 1.5, Fat: 1, Carbs: 2}
	require.Equal(t, Nutrients{Calories: 5, Protein: 2.5, Fat: 2, Carbs: 7}, Subtract(a, b))
}

func TestPer100gFromTotal(t *testing.T) {
	got, err := Per100gFromTotal(Nutrients{
		Calories: 210,
		Protein:  7,
		Fat:      3.5,
		Carbs:    28,
	}, 350)
	require.NoError(t, err)
	require.Equal(t, Nutrients{
		Calories: 60,
		Protein:  2,
		Fat:      1,
		Carbs:    8,
	}, got)
}

func TestCalculateDailyTargets(t *testing.T) {
	got, err := CalculateDailyTargets(TargetProfile{
		WeightKG:      70,
		HeightCM:      175,
		Age:           30,
		Gender:        "male",
		ActivityLevel: "moderate",
		Goal:          "maintain",
	})
	require.NoError(t, err)
	require.Equal(t, Nutrients{
		Calories: 2556,
		Protein:  126,
		Fat:      63,
		Carbs:    371.14,
	}, got)
}

func TestCalculateDailyTargets_ClampsNegativeCarbs(t *testing.T) {
	got, err := CalculateDailyTargets(TargetProfile{
		WeightKG:      200,
		HeightCM:      120,
		Age:           70,
		Gender:        "female",
		ActivityLevel: "sedentary",
		Goal:          "lose",
	})
	require.NoError(t, err)
	require.Zero(t, got.Carbs)
}
