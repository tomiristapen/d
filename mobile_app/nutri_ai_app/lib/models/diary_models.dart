class DiaryNutrients {
  final double calories;
  final double protein;
  final double fat;
  final double carbs;

  const DiaryNutrients({
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  Map<String, dynamic> toJson() => {
        'calories': calories,
        'protein': protein,
        'fat': fat,
        'carbs': carbs,
      };
}

class DiaryAddRequest {
  final String source;
  final String name;
  final double amountG;
  final DiaryNutrients per100g;
  final List<String> ingredients;

  const DiaryAddRequest({
    required this.source,
    required this.name,
    required this.amountG,
    required this.per100g,
    this.ingredients = const [],
  });

  Map<String, dynamic> toJson() => {
        'source': source,
        'name': name,
        'amount_g': amountG,
        'per_100g': per100g.toJson(),
        'ingredients': ingredients,
      };
}

class DiaryEntry {
  final int id;
  final String userId;
  final String source;
  final String name;
  final double amountG;
  final double calories;
  final double protein;
  final double fat;
  final double carbs;
  final List<String> ingredients;
  final String entryDate;
  final DateTime? createdAt;

  const DiaryEntry({
    required this.id,
    required this.userId,
    required this.source,
    required this.name,
    required this.amountG,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
    required this.ingredients,
    required this.entryDate,
    required this.createdAt,
  });

  factory DiaryEntry.fromJson(Map<String, dynamic> json) {
    return DiaryEntry(
      id: (json['id'] ?? 0) as int,
      userId: (json['user_id'] ?? '').toString(),
      source: (json['source'] ?? '').toString(),
      name: (json['name'] ?? '').toString(),
      amountG: _toDouble(json['amount_g']),
      calories: _toDouble(json['calories']),
      protein: _toDouble(json['protein']),
      fat: _toDouble(json['fat']),
      carbs: _toDouble(json['carbs']),
      ingredients: ((json['ingredients'] ?? const []) as List)
          .map((e) => e.toString())
          .toList(),
      entryDate: (json['entry_date'] ?? '').toString(),
      createdAt: json['created_at'] == null
          ? null
          : DateTime.tryParse(json['created_at'].toString()),
    );
  }
}

class DiaryProgressMetric {
  final double target;
  final double consumed;
  final double remaining;
  final double progress;

  const DiaryProgressMetric({
    required this.target,
    required this.consumed,
    required this.remaining,
    required this.progress,
  });

  factory DiaryProgressMetric.fromJson(Map<String, dynamic> json) {
    return DiaryProgressMetric(
      target: _toDouble(json['target']),
      consumed: _toDouble(json['consumed']),
      remaining: _toDouble(json['remaining']),
      progress: _toDouble(json['progress']),
    );
  }
}

class DiaryTodayResponse {
  final String date;
  final DiaryProgressMetric calories;
  final DiaryProgressMetric protein;
  final DiaryProgressMetric fat;
  final DiaryProgressMetric carbs;

  const DiaryTodayResponse({
    required this.date,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  factory DiaryTodayResponse.fromJson(Map<String, dynamic> json) {
    return DiaryTodayResponse(
      date: (json['date'] ?? '').toString(),
      calories: DiaryProgressMetric.fromJson(
          (json['calories'] ?? const {}) as Map<String, dynamic>),
      protein: DiaryProgressMetric.fromJson(
          (json['protein'] ?? const {}) as Map<String, dynamic>),
      fat: DiaryProgressMetric.fromJson(
          (json['fat'] ?? const {}) as Map<String, dynamic>),
      carbs: DiaryProgressMetric.fromJson(
          (json['carbs'] ?? const {}) as Map<String, dynamic>),
    );
  }
}

double _toDouble(dynamic v) {
  if (v is num) return v.toDouble();
  return double.tryParse(v?.toString() ?? '') ?? 0;
}
