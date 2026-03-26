class CompleteProfileRequest {
  final int age;
  final String gender;
  final int heightCm;
  final int weightKg;
  final String nutritionGoal;
  final List<String> allergies;
  final List<String> customAllergies;
  final List<String> intolerances;
  final String dietType;
  final String religiousRestriction;

  CompleteProfileRequest({
    required this.age,
    required this.gender,
    required this.heightCm,
    required this.weightKg,
    required this.nutritionGoal,
    required this.allergies,
    required this.customAllergies,
    required this.intolerances,
    required this.dietType,
    required this.religiousRestriction,
  });

  Map<String, dynamic> toJson() => {
        'age': age,
        'gender': gender,
        'height_cm': heightCm,
        'weight_kg': weightKg,
        'nutrition_goal': nutritionGoal,
        'allergies': allergies,
        'custom_allergies': customAllergies,
        'intolerances': intolerances,
        'diet_type': dietType,
        'religious_restriction': religiousRestriction,
      };
}

