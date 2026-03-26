import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/diary_models.dart';
import '../../../services/api_client.dart';
import '../../../services/diary_api.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_card.dart';
import '../products/barcode_scan_screen.dart';
import '../products/manual_product_screen.dart';
import '../products/product_details_screen.dart';
import '../products/ocr_scan_screen.dart';
import '../products/recipe_create_screen.dart';
import '../welcome/welcome_screen.dart';

class HomeScreen extends StatefulWidget {
  static const route = '/home';

  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _index = 0;
  bool _loadingDashboard = true;
  String? _dashboardError;
  DiaryTodayResponse? _dashboard;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadDashboard());
  }

  Future<void> _loadDashboard() async {
    final auth = context.read<AuthController>();
    if (!auth.isAuthed) {
      if (!mounted) return;
      setState(() {
        _dashboard = null;
        _dashboardError = 'Not authenticated';
        _loadingDashboard = false;
      });
      return;
    }

    setState(() {
      _loadingDashboard = true;
      _dashboardError = null;
    });

    try {
      final diary = context.read<DiaryApi>();
      final dashboard =
          await auth.withAuthRetry((token) => diary.getToday(accessToken: token));
      if (!mounted) return;
      setState(() {
        _dashboard = dashboard;
        _loadingDashboard = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _dashboardError = e.message;
        _loadingDashboard = false;
      });
    }
  }

  Future<void> _showAddFoodMenu() async {
    final choice = await showModalBottomSheet<_AddFoodChoice>(
      context: context,
      showDragHandle: true,
      builder: (ctx) => const _AddFoodSheet(),
    );
    if (!mounted || choice == null) return;

    switch (choice) {
      case _AddFoodChoice.barcode:
        final barcode =
            await Navigator.pushNamed(context, BarcodeScanScreen.route)
                as String?;
        if (!mounted) return;
        if (barcode == null || barcode.isEmpty) return;
        Navigator.pushNamed(context, ProductDetailsScreen.route,
            arguments: barcode);
        return;
      case _AddFoodChoice.ocr:
        await Navigator.pushNamed(context, OcrScanScreen.route);
        return;
      case _AddFoodChoice.manual:
        await Navigator.pushNamed(context, ManualProductScreen.route);
        return;
      case _AddFoodChoice.recipe:
        await Navigator.pushNamed(context, RecipeCreateScreen.route);
        return;
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.read<AuthController>();
    final dashboard = _dashboard;
    final calories = dashboard?.calories;
    final protein = dashboard?.protein;
    final carbs = dashboard?.carbs;
    final fat = dashboard?.fat;

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Good morning, Alex!',
                  style: TextStyle(fontSize: 24, fontWeight: FontWeight.w800)),
              const SizedBox(height: 4),
              const Text('Keep up the great work',
                  style: TextStyle(color: AppTheme.muted)),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(18),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(18),
                  gradient: const LinearGradient(
                      colors: [Color(0xFF0B7A61), Color(0xFF2D9A80)]),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Daily Goal',
                        style: TextStyle(
                            color: Colors.white70,
                            fontWeight: FontWeight.w600)),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Text(_formatMetric(calories?.consumed ?? 0),
                            style: const TextStyle(
                                color: Colors.white,
                                fontSize: 32,
                                fontWeight: FontWeight.w800)),
                        Text(
                            ' / ${_formatMetric(calories?.target ?? 0)} kcal',
                            style: const TextStyle(
                                color: Colors.white70,
                                fontWeight: FontWeight.w600)),
                      ],
                    ),
                    const SizedBox(height: 10),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(999),
                      child: LinearProgressIndicator(
                        value: _clampProgress(calories?.progress ?? 0),
                        minHeight: 8,
                        backgroundColor: Colors.white24,
                        valueColor:
                            const AlwaysStoppedAnimation(Color(0xFFC8F11A)),
                      ),
                    ),
                    const SizedBox(height: 10),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                            '${_formatMetric(calories?.remaining ?? 0)} kcal remaining',
                            style: const TextStyle(
                                color: Colors.white70,
                                fontWeight: FontWeight.w600)),
                        Text(
                            '${((_clampProgress(calories?.progress ?? 0)) * 100).toStringAsFixed(0)}%',
                            style: const TextStyle(
                                color: Colors.white70,
                                fontWeight: FontWeight.w700)),
                      ],
                    ),
                    if (_loadingDashboard) ...[
                      const SizedBox(height: 10),
                      const LinearProgressIndicator(
                        minHeight: 2,
                        backgroundColor: Colors.white24,
                        valueColor:
                            AlwaysStoppedAnimation<Color>(Colors.white70),
                      ),
                    ] else if (_dashboardError != null) ...[
                      const SizedBox(height: 10),
                      Text(
                        _dashboardError!,
                        style: const TextStyle(color: Colors.white70),
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
                    const Text('Nutrition',
                        style: TextStyle(
                            fontWeight: FontWeight.w800, fontSize: 16)),
                    const SizedBox(height: 12),
                    _MacroRow(
                        name: 'Protein',
                        value: protein?.consumed ?? 0,
                        total: protein?.target ?? 0,
                        color: Color(0xFF2563EB)),
                    const SizedBox(height: 10),
                    _MacroRow(
                        name: 'Carbs',
                        value: carbs?.consumed ?? 0,
                        total: carbs?.target ?? 0,
                        color: Color(0xFFF97316)),
                    const SizedBox(height: 10),
                    _MacroRow(
                        name: 'Fat',
                        value: fat?.consumed ?? 0,
                        total: fat?.target ?? 0,
                        color: Color(0xFFF59E0B)),
                  ],
                ),
              ),
              const SizedBox(height: 14),
              AppCard(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Water Intake',
                            style: TextStyle(
                                fontWeight: FontWeight.w800, fontSize: 16)),
                        SizedBox(height: 12),
                        Text('4 / 8 glasses',
                            style: TextStyle(color: AppTheme.muted)),
                      ],
                    ),
                    FilledButton(
                      style: FilledButton.styleFrom(
                          backgroundColor: const Color(0xFFEAF4F0)),
                      onPressed: () {},
                      child: const Text('Add water',
                          style: TextStyle(
                              color: AppTheme.primary,
                              fontWeight: FontWeight.w700)),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 18),
              OutlinedButton(
                onPressed: () async {
                  await auth.logout();
                  if (!context.mounted) return;
                  Navigator.pushNamedAndRemoveUntil(
                      context, WelcomeScreen.route, (_) => false);
                },
                child: const Text('Log out'),
              ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        backgroundColor: AppTheme.primary,
        onPressed: _showAddFoodMenu,
        child: const Icon(Icons.add, color: Colors.white),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        height: 72,
        color: Colors.white,
        surfaceTintColor: Colors.white,
        shape: const CircularNotchedRectangle(),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _NavItem(
                icon: Icons.home_outlined,
                label: 'Home',
                selected: _index == 0,
                onTap: () => setState(() => _index = 0)),
            _NavItem(
                icon: Icons.history,
                label: 'History',
                selected: _index == 1,
                onTap: () => setState(() => _index = 1)),
            const SizedBox(width: 36),
            _NavItem(
                icon: Icons.book_outlined,
                label: 'Diary',
                selected: _index == 2,
                onTap: () => setState(() => _index = 2)),
            _NavItem(
                icon: Icons.person_outline,
                label: 'Profile',
                selected: _index == 3,
                onTap: () => setState(() => _index = 3)),
          ],
        ),
      ),
    );
  }
}

class _NavItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool selected;
  final VoidCallback onTap;

  const _NavItem(
      {required this.icon,
      required this.label,
      required this.selected,
      required this.onTap});

  @override
  Widget build(BuildContext context) {
    final color = selected ? AppTheme.primary : AppTheme.muted;
    return InkWell(
      onTap: onTap,
      child: SizedBox(
        width: 72,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, color: color, size: 22),
            const SizedBox(height: 4),
            Text(label,
                style: TextStyle(
                    color: color, fontSize: 11, fontWeight: FontWeight.w700)),
          ],
        ),
      ),
    );
  }
}

class _MacroRow extends StatelessWidget {
  final String name;
  final double value;
  final double total;
  final Color color;

  const _MacroRow(
      {required this.name,
      required this.value,
      required this.total,
      required this.color});

  @override
  Widget build(BuildContext context) {
    final pct = total <= 0 ? 0.0 : (value / total).clamp(0.0, 1.0).toDouble();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(name, style: const TextStyle(fontWeight: FontWeight.w700)),
            Text('${_formatMetric(value)} / ${_formatMetric(total)} g',
                style: const TextStyle(color: AppTheme.muted)),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: pct,
            minHeight: 6,
            backgroundColor: Colors.black12,
            valueColor: AlwaysStoppedAnimation(color),
          ),
        ),
      ],
    );
  }
}

double _clampProgress(double value) => value.clamp(0.0, 1.0).toDouble();

String _formatMetric(double value) {
  if (value.truncateToDouble() == value) {
    return value.toStringAsFixed(0);
  }
  return value.toStringAsFixed(1);
}

enum _AddFoodChoice { barcode, ocr, manual, recipe }

class _AddFoodSheet extends StatelessWidget {
  const _AddFoodSheet();

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const ListTile(
              title: Text('Add food',
                  style: TextStyle(fontWeight: FontWeight.w800)),
              subtitle: Text('Choose an input method'),
            ),
            ListTile(
              leading: const Icon(Icons.qr_code_scanner),
              title: const Text('Barcode'),
              subtitle: const Text('Scan product barcode'),
              onTap: () => Navigator.pop(context, _AddFoodChoice.barcode),
            ),
            ListTile(
              leading: const Icon(Icons.document_scanner_outlined),
              title: const Text('Scan label (OCR)'),
              subtitle: const Text('Scan ingredients label'),
              onTap: () => Navigator.pop(context, _AddFoodChoice.ocr),
            ),
            ListTile(
              leading: const Icon(Icons.edit_note),
              title: const Text('Manual product'),
              subtitle: const Text('Enter one product and amount'),
              onTap: () => Navigator.pop(context, _AddFoodChoice.manual),
            ),
            ListTile(
              leading: const Icon(Icons.restaurant_menu_outlined),
              title: const Text('Create recipe'),
              subtitle: const Text('Combine multiple ingredients'),
              onTap: () => Navigator.pop(context, _AddFoodChoice.recipe),
            ),
          ],
        ),
      ),
    );
  }
}
