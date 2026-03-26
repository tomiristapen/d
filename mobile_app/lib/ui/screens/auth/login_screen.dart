import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/auth_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/google_auth_button.dart';
import '../home/home_screen.dart';
import '../onboarding/profile_setup_screen.dart';
import 'register_screen.dart';

class LoginScreen extends StatefulWidget {
  static const route = '/login';

  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _email = TextEditingController();
  final _pass = TextEditingController();
  bool _hide = true;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final emailArg = ModalRoute.of(context)?.settings.arguments as String?;
    if (emailArg != null && _email.text.isEmpty) {
      _email.text = emailArg;
    }
  }

  @override
  void dispose() {
    _email.dispose();
    _pass.dispose();
    super.dispose();
  }

  void _goNext() {
    final auth = context.read<AuthController>();
    if (auth.profileCompleted) {
      Navigator.pushNamedAndRemoveUntil(context, HomeScreen.route, (_) => false);
    } else {
      Navigator.pushNamedAndRemoveUntil(context, ProfileSetupScreen.route, (_) => false);
    }
  }

  Future<void> _login() async {
    final auth = context.read<AuthController>();
    try {
      await auth.login(LoginRequest(email: _email.text.trim(), password: _pass.text));
      if (!mounted) return;
      _goNext();
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
              const Text('Welcome back', style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800)),
              const SizedBox(height: 6),
              const Text('Sign in to continue your healthy journey', style: TextStyle(color: AppTheme.muted)),
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
                obscureText: _hide,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.lock_outline),
                  hintText: 'Enter your password',
                  suffixIcon: IconButton(
                    onPressed: () => setState(() => _hide = !_hide),
                    icon: Icon(_hide ? Icons.visibility_outlined : Icons.visibility_off_outlined),
                  ),
                ),
              ),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () {},
                  child: const Text('Forgot Password?', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w600)),
                ),
              ),
              const SizedBox(height: 6),
              PrimaryButton(text: 'Sign In', onPressed: _login, busy: auth.busy),
              const SizedBox(height: 14),
              Row(
                children: [
                  Expanded(child: Divider(color: Colors.grey.shade300)),
                  const Padding(padding: EdgeInsets.symmetric(horizontal: 12), child: Text('or', style: TextStyle(color: AppTheme.muted))),
                  Expanded(child: Divider(color: Colors.grey.shade300)),
                ],
              ),
              const SizedBox(height: 14),
              const GoogleAuthButton(),
              const Spacer(),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text('Don\'t have an account? ', style: TextStyle(color: AppTheme.muted)),
                  GestureDetector(
                    onTap: () => Navigator.pushReplacementNamed(context, RegisterScreen.route),
                    child: const Text('Sign Up', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w700)),
                  ),
                ],
              ),
              const SizedBox(height: 10),
            ],
          ),
        ),
      ),
    );
  }
}
