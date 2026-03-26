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

class ManualProductScreen extends StatefulWidget {
  static const route = '/products/manual';

  const ManualProductScreen({super.key});

  @override
  State<ManualProductScreen> createState() => _ManualProductScreenState();
}

class _ManualProductScreenState extends State<ManualProductScreen> {
  final _nameCtrl = TextEditingController();
  final _amountCtrl = TextEditingController(text: '100');
  final _caloriesCtrl = TextEditingController();
  final _proteinCtrl = TextEditingController();
  final _fatCtrl = TextEditingController();
  final _carbsCtrl = TextEditingController();

  bool _loading = false;
  bool _savingCustom = false;
  bool _showCustom = false;
  String? _error;
  List<String> _suggestions = const [];

  @override
  void dispose() {
    _nameCtrl.dispose();
    _amountCtrl.dispose();
    _caloriesCtrl.dispose();
    _proteinCtrl.dispose();
    _fatCtrl.dispose();
    _carbsCtrl.dispose();
    super.dispose();
  }

  double? _parseNumber(String value) {
    return double.tryParse(value.trim().replaceAll(',', '.'));
  }

  String _formatAmount(double value) {
    if (value.truncateToDouble() == value) {
      return value.toStringAsFixed(0);
    }
    return value.toStringAsFixed(1);
  }

  Future<void> _analyze({String? overrideName}) async {
    final auth = context.read<AuthController>();
    if (!auth.isAuthed) {
      setState(() => _error = 'Not authenticated');
      return;
    }

    final name = (overrideName ?? _nameCtrl.text).trim();
    final amount = _parseNumber(_amountCtrl.text);

    if (name.isEmpty) {
      setState(() => _error = 'Enter a product name');
      return;
    }
    if (amount == null || amount <= 0) {
      setState(() => _error = 'Enter a valid amount in grams');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
      _suggestions = const [];
    });

    try {
      final repo = context.read<ProductRepository>();
      final response = await auth.withAuthRetry(
        (token) => repo.analyzeManual(
          ManualAnalyzeRequest(name: name, amount: amount),
          accessToken: token,
        ),
      );
      if (!mounted) return;

      if (response.product != null) {
        setState(() => _loading = false);
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => AnalysisResultScreen(
              title: response.product!.name,
              subtitle: 'Manual entry • ${_formatAmount(amount)} g',
              product: response.product!,
              confidence: response.confidence ?? 0,
              diaryRequest: response.per100g == null
                  ? null
                  : DiaryAddRequest(
                      source: 'manual',
                      name: response.product!.name,
                      amountG: response.amountG,
                      per100g: DiaryNutrients(
                        calories: response.per100g!.calories,
                        protein: response.per100g!.protein,
                        fat: response.per100g!.fat,
                        carbs: response.per100g!.carbs,
                      ),
                      ingredients: [response.product!.name],
                    ),
            ),
          ),
        );
        return;
      }

      setState(() {
        _loading = false;
        _suggestions = response.suggestions;
        _showCustom = response.suggestions.isEmpty;
        _error = response.suggestions.isEmpty
            ? 'No exact match yet. You can create a custom product below.'
            : null;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _showCustom = e.statusCode == 404;
        _error = e.statusCode == 404
            ? 'No exact match yet. You can create a custom product below.'
            : e.message;
      });
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('Manual analyze failed: $e');
        debugPrintStack(stackTrace: st);
      }
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = 'Failed to analyze product. Please try again.';
      });
    }
  }

  Future<void> _saveCustomAndAnalyze() async {
    final auth = context.read<AuthController>();
    if (!auth.isAuthed) {
      setState(() => _error = 'Not authenticated');
      return;
    }

    final name = _nameCtrl.text.trim();
    final calories = _parseNumber(_caloriesCtrl.text);
    final protein = _parseNumber(_proteinCtrl.text);
    final fat = _parseNumber(_fatCtrl.text);
    final carbs = _parseNumber(_carbsCtrl.text);

    if (name.isEmpty) {
      setState(() => _error = 'Enter a product name first');
      return;
    }
    if ([calories, protein, fat, carbs]
        .any((value) => value == null || value < 0)) {
      setState(() => _error = 'Fill custom nutrition values per 100g');
      return;
    }

    setState(() {
      _savingCustom = true;
      _error = null;
    });

    try {
      final repo = context.read<ProductRepository>();
      await auth.withAuthRetry(
        (token) => repo.createCustomManualProduct(
          ManualCustomRequest(
            name: name,
            calories: calories!,
            protein: protein!,
            fat: fat!,
            carbs: carbs!,
          ),
          accessToken: token,
        ),
      );
      if (!mounted) return;
      setState(() => _savingCustom = false);
      await _analyze();
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _savingCustom = false;
        _error = e.message;
      });
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('Manual custom create failed: $e');
        debugPrintStack(stackTrace: st);
      }
      if (!mounted) return;
      setState(() {
        _savingCustom = false;
        _error = 'Failed to save custom product. Please try again.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Manual product')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Check manual logic end-to-end',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 6),
              const Text(
                'Search a product, confirm suggestions, or create a custom fallback if it is missing.',
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 16),
              AppCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    IngredientAutocompleteField(
                      controller: _nameCtrl,
                      label: 'Product name',
                      hintText: 'Try honey, broccoli, whole milk...',
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _amountCtrl,
                      keyboardType:
                          const TextInputType.numberWithOptions(decimal: true),
                      decoration: const InputDecoration(
                        labelText: 'Amount',
                        suffixText: 'g',
                      ),
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: 12),
                      Text(_error!, style: const TextStyle(color: Colors.red)),
                    ],
                    const SizedBox(height: 14),
                    PrimaryButton(
                      text: 'Analyze product',
                      busy: _loading,
                      onPressed: _loading ? null : _analyze,
                    ),
                  ],
                ),
              ),
              if (_suggestions.isNotEmpty) ...[
                const SizedBox(height: 16),
                AppCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Did you mean one of these?',
                          style: TextStyle(
                              fontSize: 16, fontWeight: FontWeight.w800)),
                      const SizedBox(height: 10),
                      Wrap(
                        spacing: 10,
                        runSpacing: 10,
                        children: _suggestions
                            .map(
                              (item) => ActionChip(
                                label: Text(item),
                                onPressed: _loading
                                    ? null
                                    : () {
                                        _nameCtrl.text = item;
                                        _analyze(overrideName: item);
                                      },
                              ),
                            )
                            .toList(),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 16),
              AppCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Expanded(
                          child: Text(
                            'Custom product fallback',
                            style: TextStyle(
                                fontSize: 16, fontWeight: FontWeight.w800),
                          ),
                        ),
                        TextButton(
                          onPressed: () =>
                              setState(() => _showCustom = !_showCustom),
                          child: Text(_showCustom ? 'Hide' : 'Show'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      'Use this only when the search logic cannot find a match yet. Values are per 100g.',
                      style: TextStyle(color: AppTheme.muted),
                    ),
                    if (_showCustom) ...[
                      const SizedBox(height: 14),
                      Row(
                        children: [
                          Expanded(
                            child: TextField(
                              controller: _caloriesCtrl,
                              keyboardType:
                                  const TextInputType.numberWithOptions(
                                      decimal: true),
                              decoration:
                                  const InputDecoration(labelText: 'Calories'),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: TextField(
                              controller: _proteinCtrl,
                              keyboardType:
                                  const TextInputType.numberWithOptions(
                                      decimal: true),
                              decoration:
                                  const InputDecoration(labelText: 'Protein'),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      Row(
                        children: [
                          Expanded(
                            child: TextField(
                              controller: _fatCtrl,
                              keyboardType:
                                  const TextInputType.numberWithOptions(
                                      decimal: true),
                              decoration:
                                  const InputDecoration(labelText: 'Fat'),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: TextField(
                              controller: _carbsCtrl,
                              keyboardType:
                                  const TextInputType.numberWithOptions(
                                      decimal: true),
                              decoration:
                                  const InputDecoration(labelText: 'Carbs'),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      OutlineActionButton(
                        text: 'Save custom product and analyze',
                        icon: Icons.playlist_add_check_circle_outlined,
                        onPressed: _savingCustom ? null : _saveCustomAndAnalyze,
                      ),
                      if (_savingCustom) ...[
                        const SizedBox(height: 10),
                        const LinearProgressIndicator(minHeight: 2),
                      ],
                    ],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
