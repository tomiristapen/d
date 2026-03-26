import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../state/profile_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../home/home_screen.dart';

class DietaryPreferencesScreen extends StatefulWidget {
  static const route = '/dietary-preferences';

  const DietaryPreferencesScreen({super.key});

  @override
  State<DietaryPreferencesScreen> createState() => _DietaryPreferencesScreenState();
}

class _DietaryPreferencesScreenState extends State<DietaryPreferencesScreen> {
  final _custom = TextEditingController();
  bool _busy = false;

  @override
  void dispose() {
    _custom.dispose();
    super.dispose();
  }

  Future<void> _finish() async {
    final auth = context.read<AuthController>();
    final profile = context.read<ProfileController>();
    final token = auth.accessToken;
    if (token == null || token.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Not authenticated')));
      return;
    }

    setState(() => _busy = true);
    try {
      await profile.submit(accessToken: token);
      if (!mounted) return;
      Navigator.pushNamedAndRemoveUntil(context, HomeScreen.route, (_) => false);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final p = context.watch<ProfileController>();

    return Scaffold(
      appBar: AppBar(leading: IconButton(onPressed: () => Navigator.pop(context), icon: const Icon(Icons.arrow_back))),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Your dietary\npreferences', style: TextStyle(fontSize: 28, fontWeight: FontWeight.w800, height: 1.1)),
              const SizedBox(height: 6),
              const Text('Help us keep you safe and healthy', style: TextStyle(color: AppTheme.muted)),
              const SizedBox(height: 18),
              const Text('Allergies', style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: p.allergies.map((a) {
                  final selected = p.selectedAllergies.contains(a);
                  return _Chip(
                    text: a,
                    selected: selected,
                    onTap: () => setState(() => selected ? p.selectedAllergies.remove(a) : p.selectedAllergies.add(a)),
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(child: TextField(controller: _custom, decoration: const InputDecoration(hintText: 'Add custom allergy'))),
                  const SizedBox(width: 10),
                  SizedBox(
                    width: 46,
                    height: 46,
                    child: FilledButton(
                      style: FilledButton.styleFrom(
                        backgroundColor: AppTheme.primary,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                      ),
                      onPressed: () {
                        final v = _custom.text.trim();
                        if (v.isEmpty) return;
                        setState(() {
                          p.customAllergies.add(v);
                          _custom.clear();
                        });
                      },
                      child: const Icon(Icons.add, color: Colors.white),
                    ),
                  ),
                ],
              ),
              if (p.customAllergies.isNotEmpty) ...[
                const SizedBox(height: 10),
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: p.customAllergies
                      .map((a) => _Chip(
                            text: a,
                            selected: true,
                            onTap: () => setState(() => p.customAllergies.remove(a)),
                          ))
                      .toList(),
                ),
              ],
              const SizedBox(height: 18),
              const Text('Intolerances', style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: p.intolerances.map((a) {
                  final selected = p.selectedIntolerances.contains(a);
                  return _Chip(
                    text: a,
                    selected: selected,
                    onTap: () => setState(() => selected ? p.selectedIntolerances.remove(a) : p.selectedIntolerances.add(a)),
                  );
                }).toList(),
              ),
              const SizedBox(height: 18),
              const Text('Diet Type', style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  _Segment(text: 'None', selected: p.dietType == 'none', onTap: () => setState(() => p.dietType = 'none')),
                  _Segment(text: 'Vegetarian', selected: p.dietType == 'vegetarian', onTap: () => setState(() => p.dietType = 'vegetarian')),
                  _Segment(text: 'Vegan', selected: p.dietType == 'vegan', onTap: () => setState(() => p.dietType = 'vegan')),
                  _Segment(text: 'Pescatarian', selected: p.dietType == 'pescatarian', onTap: () => setState(() => p.dietType = 'pescatarian')),
                ],
              ),
              const SizedBox(height: 18),
              const Text('Religious Restriction', style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  _Segment(text: 'None', selected: p.religiousRestriction == 'none', onTap: () => setState(() => p.religiousRestriction = 'none')),
                  _Segment(text: 'Halal', selected: p.religiousRestriction == 'halal', onTap: () => setState(() => p.religiousRestriction = 'halal')),
                  _Segment(text: 'Kosher', selected: p.religiousRestriction == 'kosher', onTap: () => setState(() => p.religiousRestriction = 'kosher')),
                ],
              ),
              const SizedBox(height: 22),
              PrimaryButton(text: 'Finish Setup', busy: _busy, onPressed: _finish),
              const SizedBox(height: 8),
            ],
          ),
        ),
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  final String text;
  final bool selected;
  final VoidCallback onTap;

  const _Chip({required this.text, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(999),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFEAF4F0) : Colors.white,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(color: selected ? AppTheme.primary : Colors.grey.shade300),
        ),
        child: Text(text, style: const TextStyle(fontWeight: FontWeight.w600)),
      ),
    );
  }
}

class _Segment extends StatelessWidget {
  final String text;
  final bool selected;
  final VoidCallback onTap;

  const _Segment({required this.text, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(14),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
        decoration: BoxDecoration(
          color: selected ? AppTheme.primary : Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: selected ? AppTheme.primary : Colors.grey.shade300),
        ),
        child: Text(
          text,
          style: TextStyle(fontWeight: FontWeight.w700, color: selected ? Colors.white : AppTheme.text),
        ),
      ),
    );
  }
}

