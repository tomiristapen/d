class Product {
  final int id;
  final String barcode;
  final String name;
  final String brand;
  final List<String> ingredients;
  final double calories;
  final double protein;
  final double fat;
  final double carbohydrates;
  final double confidenceScore;
  final String source;

  Product({
    required this.id,
    required this.barcode,
    required this.name,
    required this.brand,
    required this.ingredients,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbohydrates,
    required this.confidenceScore,
    required this.source,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: (json['id'] ?? 0) as int,
      barcode: (json['barcode'] ?? '').toString(),
      name: (json['name'] ?? '').toString(),
      brand: (json['brand'] ?? '').toString(),
      ingredients: ((json['ingredients'] ?? const []) as List)
          .map((e) => e.toString())
          .toList(),
      calories: _toDouble(json['calories']),
      protein: _toDouble(json['protein']),
      fat: _toDouble(json['fat']),
      carbohydrates: _toDouble(json['carbohydrates']),
      confidenceScore: _toDouble(json['confidenceScore']),
      source: (json['source'] ?? '').toString(),
    );
  }
}

double _toDouble(dynamic v) {
  if (v is num) return v.toDouble();
  return double.tryParse(v?.toString() ?? '') ?? 0;
}

class AnalyzeRequest {
  final String type;
  final Map<String, dynamic> data;

  AnalyzeRequest({required this.type, required this.data});

  Map<String, dynamic> toJson() => {
        'type': type,
        'data': data,
      };
}

class ProductResult {
  final String name;
  final double calories;
  final double protein;
  final double fat;
  final double carbs;

  ProductResult({
    required this.name,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  factory ProductResult.fromJson(Map<String, dynamic> json) {
    return ProductResult(
      name: (json['name'] ?? '').toString(),
      calories: _toDouble(json['calories']),
      protein: _toDouble(json['protein']),
      fat: _toDouble(json['fat']),
      carbs: _toDouble(json['carbs']),
    );
  }
}

class ManualAnalyzeRequest {
  final String name;
  final double amount;

  ManualAnalyzeRequest({
    required this.name,
    required this.amount,
  });

  Map<String, dynamic> toJson() => {
        'name': name,
        'amount': amount,
      };
}

class ManualAnalyzeResponse {
  final ProductResult? product;
  final ProductResult? per100g;
  final List<String> suggestions;
  final double? confidence;
  final double amountG;

  ManualAnalyzeResponse({
    required this.product,
    required this.per100g,
    required this.suggestions,
    required this.confidence,
    required this.amountG,
  });

  factory ManualAnalyzeResponse.fromJson(Map<String, dynamic> json) {
    final productJson = json['product'];
    final per100gJson = json['per_100g'];
    return ManualAnalyzeResponse(
      product: productJson is Map<String, dynamic>
          ? ProductResult.fromJson(productJson)
          : null,
      per100g: per100gJson is Map<String, dynamic>
          ? ProductResult.fromJson(per100gJson)
          : null,
      suggestions: ((json['suggestions'] ?? const []) as List)
          .map((e) => e.toString())
          .toList(),
      confidence:
          json['confidence'] == null ? null : _toDouble(json['confidence']),
      amountG: _toDouble(json['amount_g']),
    );
  }
}

class ManualCustomRequest {
  final String name;
  final double calories;
  final double protein;
  final double fat;
  final double carbs;

  ManualCustomRequest({
    required this.name,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  Map<String, dynamic> toJson() => {
        'name': name,
        'calories': calories,
        'protein': protein,
        'fat': fat,
        'carbs': carbs,
      };
}

class ManualCustomResponse {
  final String status;

  ManualCustomResponse({required this.status});

  factory ManualCustomResponse.fromJson(Map<String, dynamic> json) {
    return ManualCustomResponse(status: (json['status'] ?? '').toString());
  }
}

class RecipeIngredientInput {
  final String name;
  final double amount;

  RecipeIngredientInput({
    required this.name,
    required this.amount,
  });

  Map<String, dynamic> toJson() => {
        'name': name,
        'amount': amount,
      };
}

class ResolvedIngredient {
  final String name;
  final double amount;
  final double calories;
  final double protein;
  final double fat;
  final double carbs;

  ResolvedIngredient({
    required this.name,
    required this.amount,
    required this.calories,
    required this.protein,
    required this.fat,
    required this.carbs,
  });

  factory ResolvedIngredient.fromJson(Map<String, dynamic> json) {
    return ResolvedIngredient(
      name: (json['name'] ?? '').toString(),
      amount: _toDouble(json['amount']),
      calories: _toDouble(json['calories']),
      protein: _toDouble(json['protein']),
      fat: _toDouble(json['fat']),
      carbs: _toDouble(json['carbs']),
    );
  }
}

class RecipeAnalyzeRequest {
  final String name;
  final List<RecipeIngredientInput> ingredients;

  RecipeAnalyzeRequest({
    required this.name,
    required this.ingredients,
  });

  Map<String, dynamic> toJson() => {
        if (name.trim().isNotEmpty) 'name': name.trim(),
        'ingredients': ingredients.map((item) => item.toJson()).toList(),
      };
}

class RecipeAnalyzeResponse {
  final ProductResult product;
  final ProductResult per100g;
  final List<ResolvedIngredient> ingredients;
  final double confidence;
  final double amountG;

  RecipeAnalyzeResponse({
    required this.product,
    required this.per100g,
    required this.ingredients,
    required this.confidence,
    required this.amountG,
  });

  factory RecipeAnalyzeResponse.fromJson(Map<String, dynamic> json) {
    return RecipeAnalyzeResponse(
      product: ProductResult.fromJson((json['product'] ??
          const <String, dynamic>{}) as Map<String, dynamic>),
      per100g: ProductResult.fromJson((json['per_100g'] ??
          const <String, dynamic>{}) as Map<String, dynamic>),
      ingredients: ((json['ingredients'] ?? const []) as List)
          .whereType<Map<String, dynamic>>()
          .map(ResolvedIngredient.fromJson)
          .toList(),
      confidence: _toDouble(json['confidence']),
      amountG: _toDouble(json['amount_g']),
    );
  }
}
