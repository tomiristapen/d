import 'package:flutter/material.dart';

import '../../../models/product_models.dart';
import 'product_view.dart';

class ProductResultScreen extends StatelessWidget {
  static const route = '/product_result';

  final Product product;

  const ProductResultScreen({super.key, required this.product});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Product')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: ProductView(
            product: product,
            diarySource: product.source == 'ocr_draft' ? 'ocr' : 'manual',
          ),
        ),
      ),
    );
  }
}
