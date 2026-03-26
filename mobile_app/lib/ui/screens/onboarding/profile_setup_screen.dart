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
    _height = TextEditingController(text: p.heightCm.toString());
    _weight = TextEditingController(text: p.weightKg.toString());
  }

  @override
  void dispose() {
    _age.dispose();
    _height.dispose();
    _weight.dispose();
    super.dispose();
  }

  void _sync(ProfileController p) {
    p.age = int.tryParse(_age.text) ?? p.age;
    p.heightCm = int.tryParse(_height.text) ?? p.heightCm;
    p.weightKg = int.tryParse(_weight.text) ?? p.weightKg;
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
              const Text('Tell us about\nyourself', style: TextStyle(fontSize: 28, fontWeight: FontWeight.w800, height: 1.1)),
              const SizedBox(height: 6),
              const Text('Help us personalize your experience', style: TextStyle(color: AppTheme.muted)),
              const SizedBox(height: 18),
              const Text('Age', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(controller: _age, keyboardType: TextInputType.number, onChanged: (_) => _sync(p)),
              const SizedBox(height: 14),
              const Text('Gender', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              DropdownButtonFormField<String>(
                initialValue: p.gender.isEmpty ? null : p.gender,
                decoration: const InputDecoration(hintText: 'Select gender'),
                items: const [
                  DropdownMenuItem(value: 'male', child: Text('Male')),
                  DropdownMenuItem(value: 'female', child: Text('Female')),
                  DropdownMenuItem(value: 'other', child: Text('Other')),
                ],
                onChanged: (v) => setState(() => p.gender = v ?? ''),
              ),
              const SizedBox(height: 14),
              const Text('Height (cm)', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(controller: _height, keyboardType: TextInputType.number, onChanged: (_) => _sync(p)),
              const SizedBox(height: 14),
              const Text('Weight (kg)', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(controller: _weight, keyboardType: TextInputType.number, onChanged: (_) => _sync(p)),
              const SizedBox(height: 18),
              const Text('Nutrition Goal', style: TextStyle(fontWeight: FontWeight.w800)),
              const SizedBox(height: 10),
              Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  _GoalCard(
                    title: 'Lose weight',
                    icon: Icons.trending_down,
                    selected: p.nutritionGoal == 'lose_weight',
                    onTap: () => setState(() => p.nutritionGoal = 'lose_weight'),
                  ),
                  _GoalCard(
                    title: 'Maintain weight',
                    icon: Icons.balance,
                    selected: p.nutritionGoal == 'maintain_weight',
                    onTap: () => setState(() => p.nutritionGoal = 'maintain_weight'),
                  ),
                  _GoalCard(
                    title: 'Gain weight',
                    icon: Icons.trending_up,
                    selected: p.nutritionGoal == 'gain_weight',
                    onTap: () => setState(() => p.nutritionGoal = 'gain_weight'),
                  ),
                  _GoalCard(
                    title: 'Healthy eating',
                    icon: Icons.eco,
                    selected: p.nutritionGoal == 'healthy_eating',
                    onTap: () => setState(() => p.nutritionGoal = 'healthy_eating'),
                  ),
                ],
              ),
              const SizedBox(height: 22),
              PrimaryButton(
                text: 'Continue',
                onPressed: () => Navigator.pushNamed(context, DietaryPreferencesScreen.route),
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
  final IconData icon;
  final bool selected;
  final VoidCallback onTap;

  const _GoalCard({required this.title, required this.icon, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        width: (MediaQuery.of(context).size.width - 18 * 2 - 12) / 2,
        padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 14),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFEAF4F0) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: selected ? AppTheme.primary : Colors.grey.shade300),
        ),
        child: Column(
          children: [
            Icon(icon, color: selected ? AppTheme.primary : AppTheme.muted),
            const SizedBox(height: 10),
            Text(title, style: const TextStyle(fontWeight: FontWeight.w700)),
          ],
        ),
      ),
    );
  }
}
