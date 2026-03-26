import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/auth_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import 'login_screen.dart';
import 'verify_email_screen.dart';

class RegisterScreen extends StatefulWidget {
  static const route = '/register';

  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _email = TextEditingController();
  final _pass = TextEditingController();
  final _confirm = TextEditingController();
  bool _hide1 = true;
  bool _hide2 = true;

  @override
  void dispose() {
    _email.dispose();
    _pass.dispose();
    _confirm.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final auth = context.read<AuthController>();
    final req = RegisterRequest(
      email: _email.text.trim(),
      password: _pass.text,
      confirmPassword: _confirm.text,
    );
    try {
      await auth.register(req);
      if (!mounted) return;
      Navigator.pushReplacementNamed(
        context,
        VerifyEmailScreen.route,
        arguments: _email.text.trim(),
      );
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.toString())));
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthController>();

    return Scaffold(
      appBar: AppBar(leading: IconButton(onPressed: () => Navigator.pop(context), icon: const Icon(Icons.arrow_back))),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 6),
              const Text('Create an account', style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800)),
              const SizedBox(height: 6),
              const Text('Enter your details to get started', style: TextStyle(color: AppTheme.muted)),
              const SizedBox(height: 18),
              const Text('Email', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _email,
                keyboardType: TextInputType.emailAddress,
                decoration: const InputDecoration(prefixIcon: Icon(Icons.mail_outline), hintText: 'your.email@example.com'),
              ),
              const SizedBox(height: 14),
              const Text('Password', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _pass,
                obscureText: _hide1,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.lock_outline),
                  hintText: 'Enter your password',
                  suffixIcon: IconButton(
                    onPressed: () => setState(() => _hide1 = !_hide1),
                    icon: Icon(_hide1 ? Icons.visibility_outlined : Icons.visibility_off_outlined),
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const Text('Confirm Password', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _confirm,
                obscureText: _hide2,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.lock_outline),
                  hintText: 'Confirm your password',
                  suffixIcon: IconButton(
                    onPressed: () => setState(() => _hide2 = !_hide2),
                    icon: Icon(_hide2 ? Icons.visibility_outlined : Icons.visibility_off_outlined),
                  ),
                ),
              ),
              const Spacer(),
              PrimaryButton(text: 'Create Account', onPressed: _submit, busy: auth.busy),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text('Already have an account? ', style: TextStyle(color: AppTheme.muted)),
                  GestureDetector(
                    onTap: () => Navigator.pushReplacementNamed(context, LoginScreen.route),
                    child: const Text('Sign In', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w700)),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
