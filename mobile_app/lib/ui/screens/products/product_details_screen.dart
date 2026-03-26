import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/product_models.dart';
import '../../../services/products_api.dart';
import '../../../state/auth_controller.dart';
import '../../widgets/app_card.dart';

class ProductDetailsScreen extends StatefulWidget {
  static const route = '/product';

  final String barcode;

  const ProductDetailsScreen({super.key, required this.barcode});

  @override
  State<ProductDetailsScreen> createState() => _ProductDetailsScreenState();
}

class _ProductDetailsScreenState extends State<ProductDetailsScreen> {
  Product? _product;
  String? _error;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = context.read<ProductsApi>();
    final auth = context.read<AuthController>();
    final token = auth.accessToken;

    if (token == null || token.isEmpty) {
      setState(() {
        _error = 'Not authenticated';
        _loading = false;
      });
      return;
    }

    try {
      final product = await api.getByBarcode(widget.barcode, accessToken: token);
      if (!mounted) return;
      setState(() {
        _product = product;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Product')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: _loading
              ? const Center(child: CircularProgressIndicator())
              : _error != null
                  ? Center(child: Text(_error!))
                  : _product == null
                      ? const Center(child: Text('No product'))
                      : _ProductView(product: _product!),
        ),
      ),
    );
  }
}

class _ProductView extends StatelessWidget {
  final Product product;

  const _ProductView({required this.product});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(product.name, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800)),
          const SizedBox(height: 6),
          Text('Barcode: ${product.barcode}'),
          if (product.brand.isNotEmpty) ...[
            const SizedBox(height: 6),
            Text('Brand: ${product.brand}'),
          ],
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Nutrition (per 100g)', style: TextStyle(fontWeight: FontWeight.w800)),
                const SizedBox(height: 10),
                _Row(label: 'Calories', value: '${product.calories.toStringAsFixed(0)} kcal'),
                _Row(label: 'Protein', value: '${product.protein.toStringAsFixed(1)} g'),
                _Row(label: 'Fat', value: '${product.fat.toStringAsFixed(1)} g'),
                _Row(label: 'Carbs', value: '${product.carbohydrates.toStringAsFixed(1)} g'),
              ],
            ),
          ),
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Ingredients', style: TextStyle(fontWeight: FontWeight.w800)),
                const SizedBox(height: 10),
                if (product.ingredients.isEmpty)
                  const Text('No ingredients found')
                else
                  Text(product.ingredients.join(', ')),
              ],
            ),
          ),
          const SizedBox(height: 14),
          Text('Source: ${product.source} • Confidence: ${(product.confidenceScore * 100).toStringAsFixed(0)}%'),
        ],
      ),
    );
  }
}

class _Row extends StatelessWidget {
  final String label;
  final String value;

  const _Row({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontWeight: FontWeight.w700)),
          Text(value),
        ],
      ),
    );
  }
}

