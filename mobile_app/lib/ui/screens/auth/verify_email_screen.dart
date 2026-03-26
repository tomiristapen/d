import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../models/auth_models.dart';
import '../../../services/api_client.dart';
import '../../../state/auth_controller.dart';
import '../../../theme/app_theme.dart';
import '../../widgets/app_buttons.dart';
import '../../widgets/pin_input.dart';
import 'login_screen.dart';

class VerifyEmailScreen extends StatefulWidget {
  static const route = '/verify';

  const VerifyEmailScreen({super.key});

  @override
  State<VerifyEmailScreen> createState() => _VerifyEmailScreenState();
}

class _VerifyEmailScreenState extends State<VerifyEmailScreen> {
  String _code = '';

  Future<void> _verify(String email) async {
    final auth = context.read<AuthController>();
    try {
      await auth.verifyEmail(VerifyEmailRequest(email: email, code: _code));
      if (!mounted) return;
      Navigator.pushReplacementNamed(context, LoginScreen.route, arguments: email);
    } on ApiException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.toString())));
    }
  }

  Future<void> _resend(String email) async {
    final auth = context.read<AuthController>();
    try {
      await auth.resendCode(email);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Verification code sent')));
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
    final email = (ModalRoute.of(context)?.settings.arguments as String?) ?? '';

    return Scaffold(
      appBar: AppBar(leading: IconButton(onPressed: () => Navigator.pop(context), icon: const Icon(Icons.arrow_back))),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            children: [
              const SizedBox(height: 8),
              const Text('Verify your email', style: TextStyle(fontSize: 26, fontWeight: FontWeight.w800)),
              const SizedBox(height: 6),
              const Text(
                'We sent a 6-digit verification code to\nyour email address.',
                textAlign: TextAlign.center,
                style: TextStyle(color: AppTheme.muted),
              ),
              const SizedBox(height: 22),
              Container(
                width: 120,
                height: 120,
                decoration: const BoxDecoration(color: Color(0xFFEAF4F0), shape: BoxShape.circle),
                child: const Icon(Icons.alternate_email, color: AppTheme.primary, size: 50),
              ),
              const SizedBox(height: 18),
              PinInput(
                length: 6,
                onChanged: (v) => setState(() => _code = v),
              ),
              const SizedBox(height: 12),
              const Text('Check your spam folder if you didn\'t receive\nthe email.', textAlign: TextAlign.center, style: TextStyle(color: AppTheme.muted)),
              const Spacer(),
              PrimaryButton(
                text: 'Verify',
                busy: auth.busy,
                onPressed: _code.length == 6 && email.isNotEmpty ? () => _verify(email) : null,
              ),
              const SizedBox(height: 10),
              TextButton(
                onPressed: auth.busy ? null : () => _resend(email),
                child: const Text('Resend Code', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.w700)),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
