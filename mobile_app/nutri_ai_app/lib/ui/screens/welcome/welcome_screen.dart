import 'package:flutter/material.dart';

import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../auth/auth_choice_screen.dart';
import '../auth/login_screen.dart';

class WelcomeScreen extends StatelessWidget {
  static const route = '/';

  const WelcomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            children: [
              const SizedBox(height: 12),
              Container(
                height: 250,
                width: double.infinity,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(28),
                  gradient: const LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [Color(0xFF1B8A68), Color(0xFF9EDB45)],
                  ),
                ),
                child: const Center(
                  child: Icon(Icons.auto_awesome, color: Colors.white, size: 78),
                ),
              ),
              const SizedBox(height: 18),
              const Text(
                'Understand Your Food\nwith AI',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800, color: AppTheme.text, height: 1.15),
              ),
              const SizedBox(height: 10),
              const Text(
                'Scan products, analyze ingredients, and make\nhealthier decisions instantly.',
                textAlign: TextAlign.center,
                style: TextStyle(color: AppTheme.muted, height: 1.3),
              ),
              const SizedBox(height: 18),
              const _FeatureRow(icon: Icons.auto_fix_high, text: 'AI ingredient analysis'),
              const SizedBox(height: 10),
              const _FeatureRow(icon: Icons.shield_outlined, text: 'Allergy detection'),
              const SizedBox(height: 10),
              const _FeatureRow(icon: Icons.trending_up, text: 'Health score system'),
              const Spacer(),
              PrimaryButton(
                text: 'Get Started',
                onPressed: () => Navigator.pushNamed(context, AuthChoiceScreen.route),
              ),
              const SizedBox(height: 12),
              TextButton(
                onPressed: () => Navigator.pushNamed(context, LoginScreen.route),
                child: const Text('Sign In', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w600)),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _FeatureRow extends StatelessWidget {
  final IconData icon;
  final String text;

  const _FeatureRow({required this.icon, required this.text});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 34,
          height: 34,
          decoration: const BoxDecoration(color: Color(0xFFEAF4F0), shape: BoxShape.circle),
          child: Icon(icon, size: 18, color: AppTheme.primary),
        ),
        const SizedBox(width: 10),
        Text(text, style: const TextStyle(fontWeight: FontWeight.w600)),
      ],
    );
  }
}
