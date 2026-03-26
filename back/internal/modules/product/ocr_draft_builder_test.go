package product

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseIngredients_PreservesMultiWordItems(t *testing.T) {
	items := parseIngredients("SUGAR, COCOA BUTTER, WHOLE MILK POWDER, NATURAL VANILLA FLAVOURING")

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}

	require.Equal(t, []string{
		"sugar",
		"cocoa butter",
		"whole milk powder",
		"natural vanilla flavouring",
	}, names)
}

func TestParseIngredients_DoesNotSplitInsideParentheses(t *testing.T) {
	items := parseIngredients("VEGETABLE OILS (PALM, RAPESEED, SUNFLOWER), SALT")

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}

	require.Equal(t, []string{
		"vegetable oils palm rapeseed sunflower",
		"salt",
	}, names)
}

func TestParseIngredients_JoinsWrappedLinesWhenSeparatorContinuesOnNextLine(t *testing.T) {
	items := parseIngredients("WATER,\nWHOLE MILK\nPOWDER,\nSUGAR,\nSTABILIZER (PECTIN)")

	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}

	require.Equal(t, []string{
		"water",
		"whole milk powder",
		"sugar",
		"stabilizer pectin",
	}, names)
}

func TestNormalizeIngredientName_KeepsUsefulShortParentheses(t *testing.T) {
	require.Equal(t, "emulsifier soy lecithin", normalizeIngredientName("EMULSIFIER (SOY LECITHIN)"))
	require.Equal(t, "vitamin c ascorbic acid", normalizeIngredientName("Vitamin C (ascorbic acid)"))
}
