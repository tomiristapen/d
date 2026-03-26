package profile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateTable(t *testing.T) {
	tests := []struct {
		name      string
		repo      *memoryRepo
		req       UpdateRequest
		wantErr   string
		assertion func(t *testing.T, repo *memoryRepo)
	}{
		{
			name: "create valid profile",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(28),
				Gender:               ptrString("female"),
				HeightCM:             ptrFloat(168),
				WeightKG:             ptrFloat(60),
				ActivityLevel:        ptrString("moderate"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{"milk"}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{"lactose"}),
				DietType:             ptrString("vegetarian"),
				ReligiousRestriction: ptrString("none"),
			},
			assertion: func(t *testing.T, repo *memoryRepo) {
				require.True(t, repo.called)
				require.Equal(t, 28, repo.profile.Age)
			},
		},
		{
			name:    "empty create body fails",
			repo:    &memoryRepo{},
			req:     UpdateRequest{},
			wantErr: "missing required fields",
		},
		{
			name: "age boundary zero fails",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(0),
				Gender:               ptrString("female"),
				HeightCM:             ptrFloat(168),
				WeightKG:             ptrFloat(60),
				ActivityLevel:        ptrString("moderate"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{"lactose"}),
				DietType:             ptrString("none"),
				ReligiousRestriction: ptrString("none"),
			},
			wantErr: "age must be between",
		},
		{
			name: "age boundary one passes",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(1),
				Gender:               ptrString("female"),
				HeightCM:             ptrFloat(50),
				WeightKG:             ptrFloat(3),
				ActivityLevel:        ptrString("sedentary"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{}),
				DietType:             ptrString("none"),
				ReligiousRestriction: ptrString("none"),
			},
			assertion: func(t *testing.T, repo *memoryRepo) {
				require.Equal(t, 1, repo.profile.Age)
			},
		},
		{
			name: "invalid gender fails",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(28),
				Gender:               ptrString("abc"),
				HeightCM:             ptrFloat(168),
				WeightKG:             ptrFloat(60),
				ActivityLevel:        ptrString("moderate"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{"lactose"}),
				DietType:             ptrString("none"),
				ReligiousRestriction: ptrString("none"),
			},
			wantErr: "invalid gender",
		},
		{
			name: "partial update existing profile",
			repo: &memoryRepo{
				called: true,
				profile: NewProfile(
					"user-1",
					28,
					"female",
					168,
					60,
					"moderate",
					"maintain",
					[]string{"milk"},
					[]string{},
					[]string{"lactose"},
					"none",
					"none",
				),
			},
			req: UpdateRequest{
				WeightKG:        ptrFloat(58),
				CustomAllergies: ptrStrings([]string{"honey"}),
			},
			assertion: func(t *testing.T, repo *memoryRepo) {
				require.Equal(t, 58.0, repo.profile.WeightKG)
				require.Equal(t, []string{"milk"}, repo.profile.Allergies)
				require.Equal(t, []string{"honey"}, repo.profile.CustomAllergies)
			},
		},
		{
			name: "very large height fails",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(28),
				Gender:               ptrString("female"),
				HeightCM:             ptrFloat(999),
				WeightKG:             ptrFloat(60),
				ActivityLevel:        ptrString("moderate"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{}),
				DietType:             ptrString("none"),
				ReligiousRestriction: ptrString("none"),
			},
			wantErr: "height_cm must be between",
		},
		{
			name: "very large weight fails",
			repo: &memoryRepo{},
			req: UpdateRequest{
				Age:                  ptrInt(28),
				Gender:               ptrString("female"),
				HeightCM:             ptrFloat(168),
				WeightKG:             ptrFloat(501),
				ActivityLevel:        ptrString("moderate"),
				Goal:                 ptrString("maintain"),
				Allergies:            ptrStrings([]string{}),
				CustomAllergies:      ptrStrings([]string{}),
				Intolerances:         ptrStrings([]string{}),
				DietType:             ptrString("none"),
				ReligiousRestriction: ptrString("none"),
			},
			wantErr: "weight_kg must be between",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(tc.repo)
			_, err := service.Update(context.Background(), "user-1", tc.req)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.assertion != nil {
				tc.assertion(t, tc.repo)
			}
		})
	}
}
