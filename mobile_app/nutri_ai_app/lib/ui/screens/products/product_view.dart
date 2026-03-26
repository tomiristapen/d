import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/diary_models.dart';
import '../../../models/product_models.dart';
import '../../../services/api_client.dart';
import '../../../services/diary_api.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_card.dart';
import '../../widgets/app_buttons.dart';
import '../home/home_screen.dart';

class ProductView extends StatefulWidget {
  final Product product;
  final String diarySource;

  const ProductView({
    super.key,
    required this.product,
    required this.diarySource,
  });

  @override
  State<ProductView> createState() => _ProductViewState();
}

class _ProductViewState extends State<ProductView> {
  final _amountCtrl = TextEditingController(text: '100');
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _amountCtrl.addListener(_handleAmountChanged);
  }

  @override
  void dispose() {
    _amountCtrl
      ..removeListener(_handleAmountChanged)
      ..dispose();
    super.dispose();
  }

  void _handleAmountChanged() {
    setState(() {});
  }

  double? _parseAmount() {
    return double.tryParse(_amountCtrl.text.trim().replaceAll(',', '.'));
  }

  double _scale(double per100g, double amountG) {
    return per100g * amountG / 100;
  }

  String _formatAmount(double value) {
    if (value.truncateToDouble() == value) {
      return value.toStringAsFixed(0);
    }
    return value.toStringAsFixed(1);
  }

  Future<void> _addToDiary() async {
    final amount = _parseAmount();
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enter a valid amount in grams')),
      );
      return;
    }

    final auth = context.read<AuthController>();
    if (!auth.isAuthed) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Not authenticated')),
      );
      return;
    }

    setState(() => _saving = true);
    try {
      final diary = context.read<DiaryApi>();
      await auth.withAuthRetry(
        (token) => diary.addEntry(
          DiaryAddRequest(
            source: widget.diarySource,
            name: widget.product.name,
            amountG: amount,
            per100g: DiaryNutrients(
              calories: widget.product.calories,
              protein: widget.product.protein,
              fat: widget.product.fat,
              carbs: widget.product.carbohydrates,
            ),
            ingredients: widget.product.ingredients,
          ),
          accessToken: token,
        ),
      );
      if (!mounted) return;
      Navigator.pushNamedAndRemoveUntil(
        context,
        HomeScreen.route,
        (_) => false,
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.message)),
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final product = widget.product;
    final amount = _parseAmount();
    final hasValidAmount = amount != null && amount > 0;
    final currentAmount = amount ?? 0.0;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            product.name,
            style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 6),
          const Text(
            'Nutrition is stored per 100 g. Enter how much you ate to see recalculated totals.',
            style: TextStyle(color: AppTheme.muted),
          ),
          if (product.barcode.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text('Barcode: ${product.barcode}'),
          ],
          if (product.brand.isNotEmpty) ...[
            const SizedBox(height: 6),
            Text('Brand: ${product.brand}'),
          ],
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Nutrition (per 100 g)',
                  style: TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 10),
                _Row(
                  label: 'Calories',
                  value: '${product.calories.toStringAsFixed(0)} kcal',
                ),
                _Row(
                  label: 'Protein',
                  value: '${product.protein.toStringAsFixed(1)} g',
                ),
                _Row(
                  label: 'Fat',
                  value: '${product.fat.toStringAsFixed(1)} g',
                ),
                _Row(
                  label: 'Carbs',
                  value: '${product.carbohydrates.toStringAsFixed(1)} g',
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'How much did you eat?',
                  style: TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: _amountCtrl,
                  keyboardType:
                      const TextInputType.numberWithOptions(decimal: true),
                  decoration: const InputDecoration(
                    labelText: 'Amount eaten',
                    suffixText: 'g',
                  ),
                ),
                const SizedBox(height: 10),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    for (final quickAmount in const [30.0, 100.0, 250.0])
                      ActionChip(
                        label: Text('${_formatAmount(quickAmount)} g'),
                        onPressed: () {
                          _amountCtrl.text = _formatAmount(quickAmount);
                        },
                      ),
                  ],
                ),
                if (!hasValidAmount) ...[
                  const SizedBox(height: 10),
                  const Text(
                    'Enter a valid amount in grams to recalculate nutrients.',
                    style: TextStyle(color: Colors.red),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  hasValidAmount
                      ? 'Nutrition for ${_formatAmount(currentAmount)} g'
                      : 'Nutrition for this amount',
                  style: const TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 10),
                if (hasValidAmount) ...[
                  _Row(
                    label: 'Calories',
                    value:
                        '${_scale(product.calories, currentAmount).toStringAsFixed(0)} kcal',
                  ),
                  _Row(
                    label: 'Protein',
                    value:
                        '${_scale(product.protein, currentAmount).toStringAsFixed(1)} g',
                  ),
                  _Row(
                    label: 'Fat',
                    value:
                        '${_scale(product.fat, currentAmount).toStringAsFixed(1)} g',
                  ),
                  _Row(
                    label: 'Carbs',
                    value:
                        '${_scale(product.carbohydrates, currentAmount).toStringAsFixed(1)} g',
                  ),
                ] else
                  const Text(
                    'We will show recalculated totals here after you enter grams.',
                    style: TextStyle(color: AppTheme.muted),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 14),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Ingredients',
                  style: TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 10),
                if (product.ingredients.isEmpty)
                  const Text('No ingredients found')
                else
                  Text(product.ingredients.join(', ')),
              ],
            ),
          ),
          const SizedBox(height: 14),
          Text(
            'Source: ${product.source} | Confidence: ${(product.confidenceScore * 100).toStringAsFixed(0)}%',
            style: const TextStyle(color: AppTheme.muted),
          ),
          const SizedBox(height: 18),
          PrimaryButton(
            text: 'Add to diary',
            busy: _saving,
            onPressed: _saving ? null : _addToDiary,
          ),
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
