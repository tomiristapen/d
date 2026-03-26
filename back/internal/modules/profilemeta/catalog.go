package profilemeta

type Option struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type Catalog struct {
	Genders               []Option `json:"genders"`
	ActivityLevels        []Option `json:"activity_levels"`
	Goals                 []Option `json:"goals"`
	NutritionGoals        []Option `json:"nutrition_goals"`
	Allergies             []Option `json:"allergies"`
	Intolerances          []Option `json:"intolerances"`
	DietTypes             []Option `json:"diet_types"`
	ReligiousRestrictions []Option `json:"religious_restrictions"`
}

var genderOptions = []Option{
	{Key: "male", Label: "Male"},
	{Key: "female", Label: "Female"},
}

var activityLevelOptions = []Option{
	{Key: "sedentary", Label: "Sedentary"},
	{Key: "light", Label: "Light"},
	{Key: "moderate", Label: "Moderate"},
	{Key: "active", Label: "Active"},
}

var goalOptions = []Option{
	{Key: "lose", Label: "Lose"},
	{Key: "maintain", Label: "Maintain"},
	{Key: "gain", Label: "Gain"},
}

var allergyOptions = []Option{
	{Key: "milk", Label: "Milk"},
	{Key: "egg", Label: "Egg"},
	{Key: "peanut", Label: "Peanut"},
	{Key: "tree_nuts", Label: "Tree nuts"},
	{Key: "soy", Label: "Soy"},
	{Key: "wheat", Label: "Wheat"},
	{Key: "fish", Label: "Fish"},
	{Key: "shellfish", Label: "Shellfish"},
	{Key: "sesame", Label: "Sesame"},
	{Key: "mustard", Label: "Mustard"},
	{Key: "celery", Label: "Celery"},
	{Key: "lupin", Label: "Lupin"},
	{Key: "mollusks", Label: "Mollusks"},
	{Key: "sulfites", Label: "Sulfites"},
	{Key: "corn", Label: "Corn"},
	{Key: "coconut", Label: "Coconut"},
	{Key: "oat", Label: "Oat"},
	{Key: "rice", Label: "Rice"},
	{Key: "buckwheat", Label: "Buckwheat"},
	{Key: "sunflower_seed", Label: "Sunflower seed"},
	{Key: "poppy_seed", Label: "Poppy seed"},
	{Key: "chickpea", Label: "Chickpea"},
	{Key: "lentil", Label: "Lentil"},
	{Key: "pea", Label: "Pea"},
	{Key: "banana", Label: "Banana"},
	{Key: "avocado", Label: "Avocado"},
	{Key: "kiwi", Label: "Kiwi"},
	{Key: "peach", Label: "Peach"},
	{Key: "apple", Label: "Apple"},
	{Key: "citrus", Label: "Citrus"},
	{Key: "strawberry", Label: "Strawberry"},
	{Key: "tomato", Label: "Tomato"},
	{Key: "garlic", Label: "Garlic"},
	{Key: "onion", Label: "Onion"},
	{Key: "cocoa", Label: "Cocoa"},
	{Key: "coffee", Label: "Coffee"},
	{Key: "cinnamon", Label: "Cinnamon"},
	{Key: "vanilla", Label: "Vanilla"},
}

var intoleranceOptions = []Option{
	{Key: "lactose", Label: "Lactose"},
	{Key: "gluten", Label: "Gluten"},
}

var dietTypeOptions = []Option{
	{Key: "none", Label: "No specific diet"},
	{Key: "vegetarian", Label: "Vegetarian"},
	{Key: "vegan", Label: "Vegan"},
	{Key: "pescatarian", Label: "Pescatarian"},
}

var religiousRestrictionOptions = []Option{
	{Key: "none", Label: "None"},
	{Key: "halal", Label: "Halal"},
	{Key: "kosher", Label: "Kosher"},
}

var allowedGenders = optionKeySet(genderOptions)
var allowedActivityLevels = optionKeySet(activityLevelOptions)
var allowedGoals = optionKeySet(goalOptions)
var allowedAllergies = optionKeySet(allergyOptions)
var allowedIntolerances = optionKeySet(intoleranceOptions)
var allowedDietTypes = optionKeySet(dietTypeOptions)
var allowedReligiousRestrictions = optionKeySet(religiousRestrictionOptions)

func Options() Catalog {
	return Catalog{
		Genders:               cloneOptions(genderOptions),
		ActivityLevels:        cloneOptions(activityLevelOptions),
		Goals:                 cloneOptions(goalOptions),
		NutritionGoals:        cloneOptions(goalOptions),
		Allergies:             cloneOptions(allergyOptions),
		Intolerances:          cloneOptions(intoleranceOptions),
		DietTypes:             cloneOptions(dietTypeOptions),
		ReligiousRestrictions: cloneOptions(religiousRestrictionOptions),
	}
}

func IsValidGender(value string) bool {
	_, ok := allowedGenders[value]
	return ok
}

func IsValidActivityLevel(value string) bool {
	_, ok := allowedActivityLevels[value]
	return ok
}

func IsValidGoal(value string) bool {
	_, ok := allowedGoals[value]
	return ok
}

func IsValidNutritionGoal(value string) bool { return IsValidGoal(value) }

func IsValidAllergy(value string) bool {
	_, ok := allowedAllergies[value]
	return ok
}

func IsValidIntolerance(value string) bool {
	_, ok := allowedIntolerances[value]
	return ok
}

func IsValidDietType(value string) bool {
	_, ok := allowedDietTypes[value]
	return ok
}

func IsValidReligiousRestriction(value string) bool {
	_, ok := allowedReligiousRestrictions[value]
	return ok
}

func cloneOptions(values []Option) []Option {
	result := make([]Option, len(values))
	copy(result, values)
	return result
}

func optionKeySet(values []Option) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value.Key] = struct{}{}
	}
	return result
}
