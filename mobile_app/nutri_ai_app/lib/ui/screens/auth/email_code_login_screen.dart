import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/auth_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/pin_input.dart';
import '../home/home_screen.dart';
import '../onboarding/profile_setup_screen.dart';
import 'set_password_screen.dart';

class EmailCodeLoginScreen extends StatefulWidget {
  static const route = '/email-code-login';

  const EmailCodeLoginScreen({super.key});

  @override
  State<EmailCodeLoginScreen> createState() => _EmailCodeLoginScreenState();
}

class _EmailCodeLoginScreenState extends State<EmailCodeLoginScreen> {
  final _email = TextEditingController();
  bool _initializedEmail = false;
  bool _codeSent = false;
  String _code = '';

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_initializedEmail) {
      return;
    }
    final emailArg = ModalRoute.of(context)?.settings.arguments as String?;
    if (emailArg != null && emailArg.trim().isNotEmpty) {
      _email.text = emailArg.trim();
    }
    _initializedEmail = true;
  }

  @override
  void dispose() {
    _email.dispose();
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

  Future<void> _sendCode() async {
    final auth = context.read<AuthController>();
    final email = _email.text.trim();
    try {
      await auth.sendLoginCode(email);
      if (!mounted) return;
      setState(() {
        _codeSent = true;
        _code = '';
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Code sent to $email')),
      );
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

  Future<void> _verifyCode() async {
    final auth = context.read<AuthController>();
    final email = _email.text.trim();
    try {
      await auth
          .loginWithCode(EmailCodeLoginRequest(email: email, code: _code));
      if (!mounted) return;
      if (!auth.hasPassword) {
        Navigator.pushReplacementNamed(context, SetPasswordScreen.route);
        return;
      }
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
    final email = _email.text.trim();

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          onPressed: () => Navigator.pop(context),
          icon: const Icon(Icons.arrow_back),
        ),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 6),
              const Text(
                'Sign in with email code',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800),
              ),
              const SizedBox(height: 6),
              const Text(
                'Use this when Google is unavailable. We will send a 6-digit code to your email.',
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 22),
              const Text('Email',
                  style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              TextField(
                controller: _email,
                keyboardType: TextInputType.emailAddress,
                onChanged: (_) => setState(() {}),
                readOnly: _codeSent,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.mail_outline),
                  hintText: 'your.email@example.com',
                  suffixIcon: _codeSent
                      ? IconButton(
                          onPressed: auth.busy
                              ? null
                              : () => setState(() {
                                    _codeSent = false;
                                    _code = '';
                                  }),
                          icon: const Icon(Icons.edit_outlined),
                        )
                      : null,
                ),
              ),
              const SizedBox(height: 18),
              if (!_codeSent) ...[
                PrimaryButton(
                  text: 'Send code',
                  onPressed: email.isNotEmpty ? _sendCode : null,
                  busy: auth.busy,
                ),
              ] else ...[
                const Text(
                  'Enter code',
                  style: TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 8),
                PinInput(
                  length: 6,
                  onChanged: (value) => setState(() => _code = value),
                ),
                const SizedBox(height: 12),
                Text(
                  'We sent the code to $email',
                  style: const TextStyle(color: AppTheme.muted),
                ),
                const Spacer(),
                PrimaryButton(
                  text: 'Sign In',
                  onPressed: _code.length == 6 ? _verifyCode : null,
                  busy: auth.busy,
                ),
                const SizedBox(height: 10),
                TextButton(
                  onPressed: auth.busy ? null : _sendCode,
                  child: const Text(
                    'Resend code',
                    style: TextStyle(
                      color: AppTheme.primary,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
