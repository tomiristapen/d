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
      ingredients: ((json['ingredients'] ?? const []) as List).map((e) => e.toString()).toList(),
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

