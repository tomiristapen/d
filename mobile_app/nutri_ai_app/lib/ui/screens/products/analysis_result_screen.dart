import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/diary_models.dart';
import '../../../models/product_models.dart';
import '../../../services/api_client.dart';
import '../../../services/diary_api.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/app_card.dart';
import '../home/home_screen.dart';

class AnalysisResultScreen extends StatefulWidget {
  final String title;
  final String subtitle;
  final ProductResult product;
  final double confidence;
  final List<ResolvedIngredient> ingredients;
  final DiaryAddRequest? diaryRequest;

  const AnalysisResultScreen({
    super.key,
    required this.title,
    required this.subtitle,
    required this.product,
    required this.confidence,
    this.ingredients = const [],
    this.diaryRequest,
  });

  @override
  State<AnalysisResultScreen> createState() => _AnalysisResultScreenState();
}

class _AnalysisResultScreenState extends State<AnalysisResultScreen> {
  bool _saving = false;

  Future<void> _addToDiary() async {
    final request = widget.diaryRequest;
    if (request == null) return;

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
        (token) => diary.addEntry(request, accessToken: token),
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

    return Scaffold(
      appBar: AppBar(title: const Text('Analysis result')),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.title.trim().isEmpty ? product.name : widget.title,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 6),
              Text(
                widget.subtitle,
                style: const TextStyle(
                  color: AppTheme.muted,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 16),
              AppCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Nutrition summary',
                      style:
                          TextStyle(fontSize: 16, fontWeight: FontWeight.w800),
                    ),
                    const SizedBox(height: 14),
                    Wrap(
                      spacing: 10,
                      runSpacing: 10,
                      children: [
                        _MetricTile(
                          label: 'Calories',
                          value: '${product.calories.toStringAsFixed(0)} kcal',
                        ),
                        _MetricTile(
                          label: 'Protein',
                          value: '${product.protein.toStringAsFixed(1)} g',
                        ),
                        _MetricTile(
                          label: 'Fat',
                          value: '${product.fat.toStringAsFixed(1)} g',
                        ),
                        _MetricTile(
                          label: 'Carbs',
                          value: '${product.carbs.toStringAsFixed(1)} g',
                        ),
                      ],
                    ),
                    const SizedBox(height: 14),
                    Row(
                      children: [
                        const Icon(
                          Icons.verified_outlined,
                          size: 18,
                          color: AppTheme.primary,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Confidence ${(widget.confidence * 100).toStringAsFixed(0)}%',
                          style: const TextStyle(
                            fontWeight: FontWeight.w700,
                            color: AppTheme.text,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              if (widget.ingredients.isNotEmpty) ...[
                const SizedBox(height: 16),
                AppCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Resolved ingredients',
                        style:
                            TextStyle(fontSize: 16, fontWeight: FontWeight.w800),
                      ),
                      const SizedBox(height: 12),
                      ...widget.ingredients.map(
                        (item) => Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: AppTheme.bg,
                              borderRadius: BorderRadius.circular(14),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
                                  children: [
                                    Expanded(
                                      child: Text(
                                        item.name,
                                        style: const TextStyle(
                                          fontWeight: FontWeight.w700,
                                        ),
                                      ),
                                    ),
                                    Text(
                                      '${item.amount.toStringAsFixed(item.amount.truncateToDouble() == item.amount ? 0 : 1)} g',
                                      style: const TextStyle(
                                        color: AppTheme.muted,
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  '${item.calories.toStringAsFixed(0)} kcal | P ${item.protein.toStringAsFixed(1)} | F ${item.fat.toStringAsFixed(1)} | C ${item.carbs.toStringAsFixed(1)}',
                                  style:
                                      const TextStyle(color: AppTheme.muted),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              if (widget.diaryRequest != null) ...[
                const SizedBox(height: 18),
                PrimaryButton(
                  text: 'Add to diary',
                  busy: _saving,
                  onPressed: _saving ? null : _addToDiary,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _MetricTile extends StatelessWidget {
  final String label;
  final String value;

  const _MetricTile({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 145,
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: AppTheme.bg,
          borderRadius: BorderRadius.circular(14),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: const TextStyle(
                color: AppTheme.muted,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              value,
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w800),
            ),
          ],
        ),
      ),
    );
  }
}
