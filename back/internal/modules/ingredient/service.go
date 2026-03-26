package ingredient

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Autocomplete(ctx context.Context, query string) ([]string, error) {
	if strings.TrimSpace(query) == "" {
		return []string{}, nil
	}
	items, err := s.repo.Search(ctx, strings.ToLower(strings.TrimSpace(query)))
	if err != nil {
		return nil, err
	}
	return dedupeItems(items), nil
}

func (s *Service) LookupByName(ctx context.Context, name string) (Ingredient, error) {
	if strings.TrimSpace(name) == "" {
		return Ingredient{}, ErrNotFound
	}
	return s.repo.FindBestMatch(ctx, strings.ToLower(strings.TrimSpace(name)))
}

func dedupeItems(items []string) []string {
	if len(items) < 2 {
		return items
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.Join(strings.Fields(item), " "))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
