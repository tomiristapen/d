package errcode

import (
	"errors"
	"net/http"
)

type Code string

func (c Code) Error() string { return string(c) }

const (
	InvalidName        Code = "INVALID_NAME"
	InvalidAmount      Code = "INVALID_AMOUNT"
	NotFound           Code = "NOT_FOUND"
	IngredientNotFound Code = "INGREDIENT_NOT_FOUND"
	EmptyRecipe        Code = "EMPTY_RECIPE"
	InvalidNutrients   Code = "INVALID_NUTRIENTS"
)

func FromError(err error) string {
	if err == nil {
		return ""
	}

	var code Code
	if errors.As(err, &code) {
		return string(code)
	}
	return ""
}

func Message(code string) string {
	switch Code(code) {
	case InvalidName:
		return "Invalid name"
	case InvalidAmount:
		return "Invalid amount"
	case NotFound:
		return "Product not found"
	case IngredientNotFound:
		return "Ingredient not found"
	case EmptyRecipe:
		return "Recipe is empty"
	case InvalidNutrients:
		return "Invalid nutrients"
	default:
		return code
	}
}

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, NotFound), errors.Is(err, IngredientNotFound):
		return http.StatusNotFound
	case errors.Is(err, InvalidName), errors.Is(err, InvalidAmount), errors.Is(err, EmptyRecipe), errors.Is(err, InvalidNutrients):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
