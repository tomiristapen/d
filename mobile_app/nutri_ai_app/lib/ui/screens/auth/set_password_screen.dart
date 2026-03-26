import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/auth_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../home/home_screen.dart';
import '../onboarding/profile_setup_screen.dart';

class SetPasswordScreen extends StatefulWidget {
  static const route = '/set-password';

  const SetPasswordScreen({super.key});

  @override
  State<SetPasswordScreen> createState() => _SetPasswordScreenState();
}

class _SetPasswordScreenState extends State<SetPasswordScreen> {
  final _password = TextEditingController();
  final _confirm = TextEditingController();
  bool _hidePassword = true;
  bool _hideConfirm = true;

  @override
  void dispose() {
    _password.dispose();
    _confirm.dispose();
    super.dispose();
  }

  void _goNext() {
    final auth = context.read<AuthController>();
    if (auth.profileCompleted) {
      Navigator.pushNamedAndRemoveUntil(
          context, HomeScreen.route, (_) => false);
    } else {
      Navigator.pushNamedAndRemoveUntil(
          context, ProfileSetupScreen.route, (_) => false);
    }
  }

  Future<void> _submit() async {
    final auth = context.read<AuthController>();
    try {
      await auth.setPassword(
        SetPasswordRequest(
          password: _password.text,
          confirmPassword: _confirm.text,
        ),
      );
      if (!mounted) return;
      _goNext();
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(e.message)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(e.toString())));
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthController>();

    return Scaffold(
      appBar: AppBar(
        automaticallyImplyLeading: false,
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 6),
              const Text(
                'Create a password',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 6),
              const Text(
                'Optional, but helpful. Next time you can sign in even if Google is unavailable.',
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 20),
              const Text('Password',
                  style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _password,
                obscureText: _hidePassword,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.lock_outline),
                  hintText: 'Create a password',
                  suffixIcon: IconButton(
                    onPressed: () =>
                        setState(() => _hidePassword = !_hidePassword),
                    icon: Icon(
                      _hidePassword
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const Text(
                'Confirm Password',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _confirm,
                obscureText: _hideConfirm,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.lock_outline),
                  hintText: 'Repeat your password',
                  suffixIcon: IconButton(
                    onPressed: () =>
                        setState(() => _hideConfirm = !_hideConfirm),
                    icon: Icon(
                      _hideConfirm
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                    ),
                  ),
                ),
              ),
              const Spacer(),
              PrimaryButton(
                text: 'Save password',
                onPressed: _submit,
                busy: auth.busy,
              ),
              const SizedBox(height: 10),
              Center(
                child: TextButton(
                  onPressed: auth.busy ? null : _goNext,
                  child: const Text(
                    'Skip for now',
                    style: TextStyle(
                      color: AppTheme.primary,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
