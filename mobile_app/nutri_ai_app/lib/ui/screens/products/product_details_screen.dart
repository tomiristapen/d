import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../features/product/domain/barcode_normalizer.dart';
import '../../../features/product/data/product_repository.dart';
import '../../../models/product_models.dart';
import '../../../state/auth_controller.dart';
import 'product_view.dart';

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
    final repo = context.read<ProductRepository>();
    final auth = context.read<AuthController>();
    final barcode = normalizeBarcode(widget.barcode);

    if (!auth.isAuthed) {
      setState(() {
        _error = 'Not authenticated';
        _loading = false;
      });
      return;
    }
    if (!isValidBarcode(barcode)) {
      setState(() {
        _error = 'Invalid barcode';
        _loading = false;
      });
      return;
    }

    try {
      final product = await auth.withAuthRetry(
        (token) => repo.getByBarcode(barcode, accessToken: token),
      );
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
                      : ProductView(
                          product: _product!,
                          diarySource: 'barcode',
                        ),
        ),
      ),
    );
  }
}
