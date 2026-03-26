import 'package:flutter/material.dart';

import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/google_auth_button.dart';
import 'login_screen.dart';
import 'register_screen.dart';

class AuthChoiceScreen extends StatelessWidget {
  static const route = '/auth-choice';

  const AuthChoiceScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            children: [
              const SizedBox(height: 10),
              const SizedBox(height: 56),
              Container(
                width: 120,
                height: 120,
                decoration: const BoxDecoration(color: Color(0xFFEAF4F0), shape: BoxShape.circle),
                child: const Icon(Icons.lock_outline, color: AppTheme.primary, size: 52),
              ),
              const SizedBox(height: 22),
              const Text('Create your account', style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800)),
              const SizedBox(height: 6),
              const Text('Choose how you want to continue.', style: TextStyle(color: AppTheme.muted)),
              const Spacer(),
              OutlineActionButton(
                text: 'Continue with Email',
                icon: Icons.mail_outline,
                onPressed: () => Navigator.pushNamed(context, RegisterScreen.route),
              ),
              const SizedBox(height: 12),
              const GoogleAuthButton(text: 'Continue with Google'),
              const SizedBox(height: 18),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text('Already have an account? ', style: TextStyle(color: AppTheme.muted)),
                  GestureDetector(
                    onTap: () => Navigator.pushNamed(context, LoginScreen.route),
                    child: const Text('Sign In', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w700)),
                  ),
                ],
              ),
              const SizedBox(height: 18),
            ],
          ),
        ),
      ),
    );
  }
}
