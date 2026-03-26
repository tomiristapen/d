import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'dart:async';

import '../../../services/api_client.dart';
import '../../../services/ingredients_api.dart';
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
  Timer? _debounce;
  bool _loadingSuggestions = false;
  List<String> _suggestions = const [];

  @override
  void dispose() {
    _debounce?.cancel();
    _custom.dispose();
    super.dispose();
  }

  void _onCustomChanged(String value) {
    _debounce?.cancel();
    final q = value.trim();
    if (q.isEmpty) {
      setState(() => _suggestions = const []);
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 250), () async {
      final auth = context.read<AuthController>();
      final token = auth.accessToken;
      if (token == null || token.isEmpty) return;

      setState(() => _loadingSuggestions = true);
      try {
        final api = context.read<IngredientsApi>();
        final items = await api.autocomplete(q, accessToken: token);
        if (!mounted) return;
        setState(() {
          _suggestions = items.where((e) => e.isNotEmpty).take(8).toList();
        });
      } catch (_) {
        if (!mounted) return;
        setState(() => _suggestions = const []);
      } finally {
        if (mounted) setState(() => _loadingSuggestions = false);
      }
    });
  }

  void _addCustom(ProfileController p, String value) {
    final v = value.trim();
    if (v.isEmpty) return;
    setState(() {
      p.customAllergies.add(v);
      _custom.clear();
      _suggestions = const [];
    });
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
      await auth.markProfileCompleted();
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
                    text: _prettyKey(a),
                    selected: selected,
                    onTap: () => setState(() => selected ? p.selectedAllergies.remove(a) : p.selectedAllergies.add(a)),
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _custom,
                      decoration: const InputDecoration(hintText: 'Add custom allergy'),
                      onChanged: _onCustomChanged,
                      onSubmitted: (v) => _addCustom(p, v),
                    ),
                  ),
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
                        _addCustom(p, _custom.text);
                      },
                      child: const Icon(Icons.add, color: Colors.white),
                    ),
                  ),
                ],
              ),
              if (_loadingSuggestions) ...[
                const SizedBox(height: 8),
                const LinearProgressIndicator(minHeight: 2),
              ],
              if (_suggestions.isNotEmpty) ...[
                const SizedBox(height: 8),
                Container(
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(color: Colors.grey.shade200),
                  ),
                  child: Column(
                    children: _suggestions
                        .map((s) => ListTile(
                              dense: true,
                              title: Text(s),
                              onTap: () => _addCustom(p, s),
                            ))
                        .toList(),
                  ),
                ),
              ],
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
                    text: _prettyKey(a),
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

String _prettyKey(String value) {
  final words = value.split('_').where((part) => part.isNotEmpty).map((part) {
    if (part.length == 1) return part.toUpperCase();
    return '${part[0].toUpperCase()}${part.substring(1)}';
  });
  return words.join(' ');
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
