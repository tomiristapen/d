import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../features/product/data/product_repository.dart';
import '../../../models/diary_models.dart';
import '../../../models/product_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/app_card.dart';
import '../../widgets/ingredient_autocomplete_field.dart';
import 'analysis_result_screen.dart';

class RecipeCreateScreen extends StatefulWidget {
  static const route = '/products/recipe';

  const RecipeCreateScreen({super.key});

  @override
  State<RecipeCreateScreen> createState() => _RecipeCreateScreenState();
}

class _RecipeCreateScreenState extends State<RecipeCreateScreen> {
  final _nameCtrl = TextEditingController();
  final List<_RecipeItemCtrls> _items = [_RecipeItemCtrls()];

  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _nameCtrl.dispose();
    for (final item in _items) {
      item.dispose();
    }
    super.dispose();
  }

  void _addItem() {
    setState(() => _items.add(_RecipeItemCtrls()));
  }

  void _removeItem(int index) {
    if (_items.length <= 1) return;
    setState(() {
      final removed = _items.removeAt(index);
      removed.dispose();
    });
  }

  Future<void> _submit() async {
    final auth = context.read<AuthController>();
    if (!auth.isAuthed) {
      setState(() => _error = 'Not authenticated');
      return;
    }

    final ingredients = <RecipeIngredientInput>[];
    for (final item in _items) {
      final name = item.name.text.trim();
      if (name.isEmpty) continue;

      final amount =
          double.tryParse(item.amount.text.trim().replaceAll(',', '.'));
      if (amount == null || amount <= 0) {
        setState(
            () => _error = 'Each ingredient needs a valid amount in grams');
        return;
      }

      ingredients.add(RecipeIngredientInput(name: name, amount: amount));
    }

    if (ingredients.isEmpty) {
      setState(() => _error = 'Add at least one ingredient');
      return;
    }

    setState(() {
      _error = null;
      _loading = true;
    });

    try {
      final repo = context.read<ProductRepository>();
      final req = RecipeAnalyzeRequest(
        name: _nameCtrl.text.trim(),
        ingredients: ingredients,
      );
      final result = await auth.withAuthRetry(
          (token) => repo.analyzeRecipe(req, accessToken: token));
      if (!mounted) return;
      setState(() => _loading = false);
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => AnalysisResultScreen(
            title: _nameCtrl.text.trim().isEmpty
                ? result.product.name
                : _nameCtrl.text.trim(),
            subtitle: 'Recipe total • ${result.ingredients.length} ingredients',
            product: result.product,
            confidence: result.confidence,
            ingredients: result.ingredients,
            diaryRequest: DiaryAddRequest(
              source: 'recipe',
              name: _nameCtrl.text.trim().isEmpty
                  ? result.product.name
                  : _nameCtrl.text.trim(),
              amountG: result.amountG,
              per100g: DiaryNutrients(
                calories: result.per100g.calories,
                protein: result.per100g.protein,
                fat: result.per100g.fat,
                carbs: result.per100g.carbs,
              ),
              ingredients:
                  result.ingredients.map((item) => item.name).toList(),
            ),
          ),
        ),
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.statusCode == 404
            ? 'One or more ingredients were not found.'
            : e.message;
        _loading = false;
      });
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('Recipe analyze failed: $e');
        debugPrintStack(stackTrace: st);
      }
      if (!mounted) return;
      setState(() {
        _error = 'Failed to analyze recipe. Please try again.';
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Create recipe')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Build a recipe and test the merge logic',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 6),
              const Text(
                'Enter ingredients in grams. We will resolve them through the same backend logic as manual input.',
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 16),
              if (_error != null) ...[
                Text(_error!, style: const TextStyle(color: Colors.red)),
                const SizedBox(height: 12),
              ],
              AppCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    TextField(
                      controller: _nameCtrl,
                      decoration: const InputDecoration(
                          labelText: 'Recipe name (optional)'),
                    ),
                    const SizedBox(height: 16),
                    const Text('Ingredients',
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w800)),
                    const SizedBox(height: 6),
                    const Text('Use grams so the backend math matches exactly.',
                        style: TextStyle(color: AppTheme.muted)),
                    const SizedBox(height: 12),
                    ...List.generate(_items.length, (index) {
                      final item = _items[index];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 14),
                        child: Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: AppTheme.bg,
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Column(
                            children: [
                              Row(
                                children: [
                                  Text(
                                    'Ingredient ${index + 1}',
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w700),
                                  ),
                                  const Spacer(),
                                  IconButton(
                                    onPressed: _loading
                                        ? null
                                        : () => _removeItem(index),
                                    icon: const Icon(Icons.close),
                                  ),
                                ],
                              ),
                              IngredientAutocompleteField(
                                controller: item.name,
                                label: 'Ingredient name',
                                hintText: 'Start typing to get suggestions',
                                enabled: !_loading,
                              ),
                              const SizedBox(height: 10),
                              TextField(
                                controller: item.amount,
                                enabled: !_loading,
                                keyboardType:
                                    const TextInputType.numberWithOptions(
                                        decimal: true),
                                decoration: const InputDecoration(
                                  labelText: 'Amount',
                                  suffixText: 'g',
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    }),
                    const SizedBox(height: 4),
                    OutlineActionButton(
                      text: 'Add ingredient',
                      icon: Icons.add,
                      onPressed: _loading ? null : _addItem,
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              PrimaryButton(
                text: 'Analyze recipe',
                busy: _loading,
                onPressed: _loading ? null : _submit,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _RecipeItemCtrls {
  final name = TextEditingController();
  final amount = TextEditingController();

  void dispose() {
    name.dispose();
    amount.dispose();
  }
}
