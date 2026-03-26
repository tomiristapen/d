class OcrDraft {
  final String ocrMode;
  final double ocrQuality;
  final double overallConfidence;
  final List<OcrIngredient> ingredients;
  final OcrNutrition nutrition;
  final List<String> missingFields;
  final List<OcrConflict> conflicts;

  OcrDraft({
    required this.ocrMode,
    required this.ocrQuality,
    required this.overallConfidence,
    required this.ingredients,
    required this.nutrition,
    required this.missingFields,
    required this.conflicts,
  });

  factory OcrDraft.fromJson(Map<String, dynamic> json) {
    return OcrDraft(
      ocrMode: (json['ocrMode'] ?? '').toString(),
      ocrQuality: _toDouble(json['ocrQuality']),
      overallConfidence: _toDouble(json['overallConfidence']),
      ingredients: ((json['ingredients'] ?? const []) as List)
          .map((e) => OcrIngredient.fromJson((e ?? const {}) as Map<String, dynamic>))
          .toList(),
      nutrition: OcrNutrition.fromJson((json['nutrition'] ?? const {}) as Map<String, dynamic>),
      missingFields: ((json['missingFields'] ?? const []) as List).map((e) => e.toString()).toList(),
      conflicts: ((json['conflicts'] ?? const []) as List)
          .map((e) => OcrConflict.fromJson((e ?? const {}) as Map<String, dynamic>))
          .toList(),
    );
  }
}

class OcrIngredient {
  final String clientId;
  final String raw;
  String name;
  final int? matchedProductId;
  final String matchName;
  final double matchScore;
  double confidence;
  bool isVerified;

  OcrIngredient({
    required this.clientId,
    required this.raw,
    required this.name,
    required this.matchedProductId,
    required this.matchName,
    required this.matchScore,
    required this.confidence,
    required this.isVerified,
  });

  factory OcrIngredient.fromJson(Map<String, dynamic> json) {
    return OcrIngredient(
      clientId: (json['clientId'] ?? '').toString(),
      raw: (json['raw'] ?? '').toString(),
      name: (json['name'] ?? '').toString(),
      matchedProductId: _toIntNullable(json['matchedProductId']),
      matchName: (json['matchName'] ?? '').toString(),
      matchScore: _toDouble(json['matchScore']),
      confidence: _toDouble(json['confidence']),
      isVerified: (json['isVerified'] ?? false) == true,
    );
  }
}

class OcrNutrition {
  final String energyUnit; // kcal | kJ | ''
  final String massUnit; // g | ''
  final OcrNutritionField calories;
  final OcrNutritionField protein;
  final OcrNutritionField fat;
  final OcrNutritionField carbs;

  OcrNutrition({
    required this.energyUnit,
    required this.massUnit,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  factory OcrNutrition.fromJson(Map<String, dynamic> json) {
    return OcrNutrition(
      energyUnit: (json['energyUnit'] ?? '').toString(),
      massUnit: (json['massUnit'] ?? '').toString(),
      calories: OcrNutritionField.fromJson((json['calories'] ?? const {}) as Map<String, dynamic>),
      protein: OcrNutritionField.fromJson((json['protein'] ?? const {}) as Map<String, dynamic>),
      fat: OcrNutritionField.fromJson((json['fat'] ?? const {}) as Map<String, dynamic>),
      carbs: OcrNutritionField.fromJson((json['carbs'] ?? const {}) as Map<String, dynamic>),
    );
  }
}

class OcrNutritionField {
  double? value;
  double confidence;
  bool isEstimated;
  bool isVerified;

  OcrNutritionField({
    required this.value,
    required this.confidence,
    required this.isEstimated,
    required this.isVerified,
  });

  factory OcrNutritionField.fromJson(Map<String, dynamic> json) {
    final v = json['value'];
    return OcrNutritionField(
      value: v == null ? null : _toDouble(v),
      confidence: _toDouble(json['confidence']),
      isEstimated: (json['isEstimated'] ?? false) == true,
      isVerified: (json['isVerified'] ?? false) == true,
    );
  }
}

class OcrConflict {
  final String field;
  final String note;

  OcrConflict({required this.field, required this.note});

  factory OcrConflict.fromJson(Map<String, dynamic> json) {
    return OcrConflict(
      field: (json['field'] ?? '').toString(),
      note: (json['note'] ?? '').toString(),
    );
  }
}

double _toDouble(dynamic v) {
  if (v is num) return v.toDouble();
  return double.tryParse(v?.toString() ?? '') ?? 0;
}

int? _toIntNullable(dynamic v) {
  if (v == null) return null;
  if (v is int) return v;
  return int.tryParse(v.toString());
}

