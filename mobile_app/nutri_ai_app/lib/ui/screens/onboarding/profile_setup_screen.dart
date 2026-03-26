import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../state/profile_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import 'dietary_preferences_screen.dart';

class ProfileSetupScreen extends StatefulWidget {
  static const route = '/profile-setup';

  const ProfileSetupScreen({super.key});

  @override
  State<ProfileSetupScreen> createState() => _ProfileSetupScreenState();
}

class _ProfileSetupScreenState extends State<ProfileSetupScreen> {
  late final TextEditingController _age;
  late final TextEditingController _height;
  late final TextEditingController _weight;

  @override
  void initState() {
    super.initState();
    final p = context.read<ProfileController>();
    _age = TextEditingController(text: p.age.toString());
    _height = TextEditingController(text: _formatNumber(p.heightCm));
    _weight = TextEditingController(text: _formatNumber(p.weightKg));
  }

  @override
  void dispose() {
    _age.dispose();
    _height.dispose();
    _weight.dispose();
    super.dispose();
  }

  void _sync(ProfileController p) {
    p.age = int.tryParse(_age.text.trim()) ?? p.age;
    p.heightCm = _parseDouble(_height.text) ?? p.heightCm;
    p.weightKg = _parseDouble(_weight.text) ?? p.weightKg;
  }

  double? _parseDouble(String value) {
    return double.tryParse(value.trim().replaceAll(',', '.'));
  }

  String _formatNumber(double value) {
    if (value.truncateToDouble() == value) {
      return value.toStringAsFixed(0);
    }
    return value.toStringAsFixed(1);
  }

  void _continue(ProfileController p) {
    _sync(p);
    if (p.gender.isEmpty) {
      _showError('Select gender');
      return;
    }
    if (p.age <= 0) {
      _showError('Enter a valid age');
      return;
    }
    if (p.heightCm <= 0 || p.weightKg <= 0) {
      _showError('Enter valid height and weight');
      return;
    }
    Navigator.pushNamed(context, DietaryPreferencesScreen.route);
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final p = context.watch<ProfileController>();

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          onPressed: () => Navigator.pop(context),
          icon: const Icon(Icons.arrow_back),
        ),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Set up your\ntargets',
                style: TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w800,
                  height: 1.1,
                ),
              ),
              const SizedBox(height: 6),
              const Text(
                'We use these inputs to calculate your daily nutrition goals.',
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 18),
              const Text('Age', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _age,
                keyboardType: TextInputType.number,
                onChanged: (_) => _sync(p),
              ),
              const SizedBox(height: 14),
              const Text('Gender',
                  style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              DropdownButtonFormField<String>(
                initialValue: p.gender.isEmpty ? null : p.gender,
                decoration: const InputDecoration(hintText: 'Select gender'),
                items: const [
                  DropdownMenuItem(value: 'male', child: Text('Male')),
                  DropdownMenuItem(value: 'female', child: Text('Female')),
                ],
                onChanged: (v) => setState(() => p.gender = v ?? ''),
              ),
              const SizedBox(height: 14),
              const Text('Height (cm)',
                  style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _height,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                onChanged: (_) => _sync(p),
              ),
              const SizedBox(height: 14),
              const Text('Weight (kg)',
                  style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _weight,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                onChanged: (_) => _sync(p),
              ),
              const SizedBox(height: 18),
              const Text('Activity Level',
                  style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  _Segment(
                    text: 'Sedentary',
                    selected: p.activityLevel == 'sedentary',
                    onTap: () => setState(() => p.activityLevel = 'sedentary'),
                  ),
                  _Segment(
                    text: 'Light',
                    selected: p.activityLevel == 'light',
                    onTap: () => setState(() => p.activityLevel = 'light'),
                  ),
                  _Segment(
                    text: 'Moderate',
                    selected: p.activityLevel == 'moderate',
                    onTap: () => setState(() => p.activityLevel = 'moderate'),
                  ),
                  _Segment(
                    text: 'Active',
                    selected: p.activityLevel == 'active',
                    onTap: () => setState(() => p.activityLevel = 'active'),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              const Text('Goal',
                  style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  _GoalCard(
                    title: 'Lose',
                    subtitle: '-500 kcal',
                    icon: Icons.trending_down,
                    selected: p.goal == 'lose',
                    onTap: () => setState(() => p.goal = 'lose'),
                  ),
                  _GoalCard(
                    title: 'Maintain',
                    subtitle: 'Keep steady',
                    icon: Icons.balance,
                    selected: p.goal == 'maintain',
                    onTap: () => setState(() => p.goal = 'maintain'),
                  ),
                  _GoalCard(
                    title: 'Gain',
                    subtitle: '+500 kcal',
                    icon: Icons.trending_up,
                    selected: p.goal == 'gain',
                    onTap: () => setState(() => p.goal = 'gain'),
                  ),
                ],
              ),
              const SizedBox(height: 22),
              PrimaryButton(
                text: 'Continue',
                onPressed: () => _continue(p),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _GoalCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final IconData icon;
  final bool selected;
  final VoidCallback onTap;

  const _GoalCard({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        width: (MediaQuery.of(context).size.width - 18 * 2 - 12 * 2) / 3,
        padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 12),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFEAF4F0) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border:
              Border.all(color: selected ? AppTheme.primary : Colors.grey.shade300),
        ),
        child: Column(
          children: [
            Icon(icon, color: selected ? AppTheme.primary : AppTheme.muted),
            const SizedBox(height: 10),
            Text(title, style: const TextStyle(fontWeight: FontWeight.w700)),
            const SizedBox(height: 4),
            Text(
              subtitle,
              textAlign: TextAlign.center,
              style: const TextStyle(color: AppTheme.muted, fontSize: 12),
            ),
          ],
        ),
      ),
    );
  }
}

class _Segment extends StatelessWidget {
  final String text;
  final bool selected;
  final VoidCallback onTap;

  const _Segment({
    required this.text,
    required this.selected,
    required this.onTap,
  });

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
          border:
              Border.all(color: selected ? AppTheme.primary : Colors.grey.shade300),
        ),
        child: Text(
          text,
          style: TextStyle(
            fontWeight: FontWeight.w700,
            color: selected ? Colors.white : AppTheme.text,
          ),
        ),
      ),
    );
  }
}
