import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../../../features/product/data/product_repository.dart';
import '../../../models/product_models.dart';
import '../../../models/ocr_draft_models.dart';
import '../../../state/auth_controller.dart';
import 'product_result_screen.dart';

class OcrDraftReviewScreen extends StatefulWidget {
  final OcrDraft draft;
  final List<String> images;
  final String lang;

  const OcrDraftReviewScreen({
    super.key,
    required this.draft,
    required this.images,
    required this.lang,
  });

  @override
  State<OcrDraftReviewScreen> createState() => _OcrDraftReviewScreenState();
}

class _OcrDraftReviewScreenState extends State<OcrDraftReviewScreen> {
  late OcrDraft _draft;
  late List<String> _images;
  late String _lang;

  bool _loading = false;
  String? _error;

  final _caloriesCtrl = TextEditingController();
  final _proteinCtrl = TextEditingController();
  final _fatCtrl = TextEditingController();
  final _carbsCtrl = TextEditingController();

  final List<TextEditingController> _ingredientCtrls = [];

  @override
  void initState() {
    super.initState();
    _draft = widget.draft;
    _images = List<String>.from(widget.images);
    _lang = widget.lang;
    _syncControllersFromDraft();
  }

  @override
  void dispose() {
    _caloriesCtrl.dispose();
    _proteinCtrl.dispose();
    _fatCtrl.dispose();
    _carbsCtrl.dispose();
    for (final c in _ingredientCtrls) {
      c.dispose();
    }
    super.dispose();
  }

  void _syncControllersFromDraft() {
    _caloriesCtrl.text = _draft.nutrition.calories.value?.toString() ?? '';
    _proteinCtrl.text = _draft.nutrition.protein.value?.toString() ?? '';
    _fatCtrl.text = _draft.nutrition.fat.value?.toString() ?? '';
    _carbsCtrl.text = _draft.nutrition.carbs.value?.toString() ?? '';

    for (final c in _ingredientCtrls) {
      c.dispose();
    }
    _ingredientCtrls
      ..clear()
      ..addAll(_draft.ingredients.map((i) => TextEditingController(text: i.name)));
  }

  bool _isMissing(String field) => _draft.missingFields.contains(field);

  Future<void> _addPhotoAndMerge() async {
    final auth = context.read<AuthController>();
    final repo = context.read<ProductRepository>();
    final token = auth.accessToken;
    if (token == null || token.isEmpty) {
      setState(() => _error = 'Not authenticated');
      return;
    }

    setState(() {
      _error = null;
      _loading = true;
    });

    try {
      final picker = ImagePicker();
      final file = await picker.pickImage(
        source: ImageSource.camera,
        preferredCameraDevice: CameraDevice.rear,
        maxWidth: 2600,
        maxHeight: 2600,
        imageQuality: 95,
      );
      if (!mounted) return;
      if (file == null) {
        setState(() => _loading = false);
        return;
      }

      final base64Image = await compute(_base64FromFilePath, file.path);
      final nextImages = List<String>.from(_images)..add(base64Image);

      final merged = await repo.buildOcrDraft(images: nextImages, lang: _lang, accessToken: token);
      if (!mounted) return;

      setState(() {
        _images = nextImages;
        _draft = merged;
        _loading = false;
      });
      _syncControllersFromDraft();
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('OCR merge failed: $e');
        debugPrintStack(stackTrace: st);
      }
      if (!mounted) return;
      setState(() {
        _error = kDebugMode ? 'Failed to merge photos: $e' : 'Failed to merge photos. Please try again.';
        _loading = false;
      });
    }
  }

  void _removeIngredientAt(int index) {
    final next = List<OcrIngredient>.from(_draft.ingredients)..removeAt(index);
    setState(() {
      _draft = OcrDraft(
        ocrMode: _draft.ocrMode,
        ocrQuality: _draft.ocrQuality,
        overallConfidence: _draft.overallConfidence,
        ingredients: next,
        nutrition: _draft.nutrition,
        missingFields: _draft.missingFields,
        conflicts: _draft.conflicts,
      );
    });
    _syncControllersFromDraft();
  }

  void _addIngredientRow() {
    final next = List<OcrIngredient>.from(_draft.ingredients)
      ..add(OcrIngredient(
        clientId: 'manual_${DateTime.now().millisecondsSinceEpoch}',
        raw: '',
        name: '',
        matchedProductId: null,
        matchName: '',
        matchScore: 0,
        confidence: 0,
        isVerified: true,
      ));
    setState(() {
      _draft = OcrDraft(
        ocrMode: _draft.ocrMode,
        ocrQuality: _draft.ocrQuality,
        overallConfidence: _draft.overallConfidence,
        ingredients: next,
        nutrition: _draft.nutrition,
        missingFields: _draft.missingFields,
        conflicts: _draft.conflicts,
      );
    });
    _syncControllersFromDraft();
  }

  Product _toProduct() {
    final ingredients = <String>[];
    for (final c in _ingredientCtrls) {
      final v = c.text.trim();
      if (v.isNotEmpty) ingredients.add(v);
    }

    double parseNum(String s) => double.tryParse(s.trim().replaceAll(',', '.')) ?? 0;

    return Product(
      id: 0,
      barcode: '',
      name: 'OCR draft',
      brand: '',
      ingredients: ingredients,
      calories: parseNum(_caloriesCtrl.text),
      protein: parseNum(_proteinCtrl.text),
      fat: parseNum(_fatCtrl.text),
      carbohydrates: parseNum(_carbsCtrl.text),
      confidenceScore: _draft.overallConfidence,
      source: 'ocr_draft',
    );
  }

  @override
  Widget build(BuildContext context) {
    final percent = (_draft.overallConfidence * 100).clamp(0, 100).toStringAsFixed(0);
    final energyUnit = _draft.nutrition.energyUnit.isNotEmpty ? _draft.nutrition.energyUnit : 'kcal?';
    final massUnit = _draft.nutrition.massUnit.isNotEmpty ? _draft.nutrition.massUnit : 'g?';

    return Scaffold(
      appBar: AppBar(
        title: const Text('Review OCR'),
        actions: [
          IconButton(
            onPressed: _loading ? null : _addPhotoAndMerge,
            icon: const Icon(Icons.add_a_photo_outlined),
            tooltip: 'Add photo',
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Confidence: $percent%', style: const TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 6),
              Text('OCR mode: ${_draft.ocrMode} - photos: ${_images.length}'),
              if (_draft.conflicts.isNotEmpty) ...[
                const SizedBox(height: 10),
                Text(
                  'Conflicts detected (please verify):',
                  style: TextStyle(color: Colors.orange.shade900, fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 6),
                ..._draft.conflicts.map((c) => Text('- ${c.field}: ${c.note}')),
              ],
              if (_error != null) ...[
                const SizedBox(height: 10),
                Text(_error!, style: const TextStyle(color: Colors.red)),
              ],
              const SizedBox(height: 14),
              Expanded(
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('Ingredients', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w800)),
                      const SizedBox(height: 10),
                      if (_draft.ingredients.isEmpty)
                        Text(
                          'No ingredients found. Add them manually or take another photo.',
                          style: TextStyle(color: Colors.orange.shade900),
                        )
                      else
                        ...List.generate(_draft.ingredients.length, (i) {
                          final ing = _draft.ingredients[i];
                          final matchHint = ing.matchName.isNotEmpty
                              ? 'match: ${ing.matchName} (${(ing.matchScore * 100).toStringAsFixed(0)}%)'
                              : 'no DB match';
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 10),
                            child: Row(
                              children: [
                                Expanded(
                                  child: TextFormField(
                                    controller: _ingredientCtrls[i],
                                    decoration: InputDecoration(
                                      labelText: matchHint,
                                      helperText: ing.raw.isNotEmpty ? 'raw: ${ing.raw}' : null,
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Checkbox(
                                  value: ing.isVerified,
                                  onChanged: (v) {
                                    setState(() => ing.isVerified = v == true);
                                  },
                                ),
                                IconButton(
                                  onPressed: _loading ? null : () => _removeIngredientAt(i),
                                  icon: const Icon(Icons.close),
                                  tooltip: 'Remove',
                                ),
                              ],
                            ),
                          );
                        }),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: TextButton.icon(
                          onPressed: _loading ? null : _addIngredientRow,
                          icon: const Icon(Icons.add),
                          label: const Text('Add ingredient'),
                        ),
                      ),
                      const SizedBox(height: 16),
                      const Text('Nutrition', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w800)),
                      const SizedBox(height: 10),
                      _numField(
                        controller: _caloriesCtrl,
                        label: 'Calories ($energyUnit)',
                        missing: _isMissing('nutrition.calories') || _isMissing('nutrition.energyUnit'),
                        estimated: _draft.nutrition.calories.isEstimated,
                      ),
                      _numField(
                        controller: _proteinCtrl,
                        label: 'Protein ($massUnit)',
                        missing: _isMissing('nutrition.protein') || _isMissing('nutrition.massUnit'),
                        estimated: _draft.nutrition.protein.isEstimated,
                      ),
                      _numField(
                        controller: _fatCtrl,
                        label: 'Fat ($massUnit)',
                        missing: _isMissing('nutrition.fat') || _isMissing('nutrition.massUnit'),
                        estimated: _draft.nutrition.fat.isEstimated,
                      ),
                      _numField(
                        controller: _carbsCtrl,
                        label: 'Carbs ($massUnit)',
                        missing: _isMissing('nutrition.carbs') || _isMissing('nutrition.massUnit'),
                        estimated: _draft.nutrition.carbs.isEstimated,
                      ),
                      const SizedBox(height: 22),
                      SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: _loading
                              ? null
                              : () {
                                  final p = _toProduct();
                                  Navigator.pushReplacement(
                                    context,
                                    MaterialPageRoute(builder: (_) => ProductResultScreen(product: p)),
                                  );
                                },
                          child: _loading
                              ? const SizedBox(
                                  width: 18,
                                  height: 18,
                                  child: CircularProgressIndicator(strokeWidth: 2),
                                )
                              : const Text('Continue'),
                        ),
                      ),
                      const SizedBox(height: 10),
                      const Text(
                        'Tip: verify highlighted fields. OCR is unreliable by design.',
                        style: TextStyle(color: Colors.black54),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _numField({
    required TextEditingController controller,
    required String label,
    required bool missing,
    required bool estimated,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: TextFormField(
        controller: controller,
        keyboardType: const TextInputType.numberWithOptions(decimal: true),
        decoration: InputDecoration(
          labelText: label,
          helperText: estimated ? 'estimated' : null,
          errorText: missing ? 'missing / needs verification' : null,
        ),
      ),
    );
  }
}

String _base64FromFilePath(String path) {
  final bytes = File(path).readAsBytesSync();
  return base64Encode(bytes);
}
