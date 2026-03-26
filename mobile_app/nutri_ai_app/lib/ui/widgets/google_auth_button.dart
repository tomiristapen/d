import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../services/api_client.dart';
import '../../services/google_sign_in_helper.dart';
import '../../state/auth_controller.dart';
import '../screens/home/home_screen.dart';
import '../screens/auth/email_code_login_screen.dart';
import '../screens/onboarding/profile_setup_screen.dart';
import 'app_buttons.dart';
import 'google_gis_button.dart';

class GoogleAuthButton extends StatefulWidget {
  final String text;

  const GoogleAuthButton({super.key, this.text = 'Continue with Google'});

  @override
  State<GoogleAuthButton> createState() => _GoogleAuthButtonState();
}

class _GoogleAuthButtonState extends State<GoogleAuthButton> {
  bool _handlingWeb = false;

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

  Future<void> _mobileGoogle() async {
    final auth = context.read<AuthController>();
    try {
      final idToken = await googleSignInGetIdToken(forceAccountPicker: true);
      await auth.googleLogin(idToken);
      if (!mounted) return;
      _goNext();
    } on ApiException catch (e) {
      if (!mounted) return;
      if (_isCancellation(e.message)) {
        return;
      }
      await _showFallbackSheet(e.message);
    } catch (e) {
      if (!mounted) return;
      await _showFallbackSheet(e.toString());
    }
  }

  Future<void> _handleWebAfterGis() async {
    if (_handlingWeb) return;
    _handlingWeb = true;
    final auth = context.read<AuthController>();
    try {
      final idToken = await googleWebGetIdTokenFromCurrentUser();
      await auth.googleLogin(idToken);
      if (!mounted) return;
      _goNext();
    } on ApiException catch (e) {
      if (!mounted) return;
      if (_isCancellation(e.message)) {
        return;
      }
      await _showFallbackSheet(e.message);
    } catch (e) {
      if (!mounted) return;
      await _showFallbackSheet(e.toString());
    } finally {
      _handlingWeb = false;
    }
  }

  bool _isCancellation(String message) {
    return message.toLowerCase().contains('cancelled');
  }

  Future<void> _showFallbackSheet(String message) async {
    final action = await showModalBottomSheet<_GoogleFallbackAction>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      builder: (sheetContext) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(18, 8, 18, 18),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Google sign-in did not work',
                  style: TextStyle(fontSize: 22, fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 8),
                Text(
                  message,
                  style: const TextStyle(color: Colors.black54),
                ),
                const SizedBox(height: 10),
                const Text(
                  'You can still continue with a 6-digit code by email and optionally create a password after that.',
                  style: TextStyle(color: Colors.black54),
                ),
                const SizedBox(height: 18),
                PrimaryButton(
                  text: 'Use email code',
                  onPressed: () => Navigator.pop(
                    sheetContext,
                    _GoogleFallbackAction.emailCode,
                  ),
                ),
                const SizedBox(height: 10),
                OutlineActionButton(
                  text: kIsWeb ? 'Close' : 'Try Google again',
                  icon: Icons.refresh,
                  onPressed: () => Navigator.pop(
                    sheetContext,
                    _GoogleFallbackAction.retry,
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );

    if (!mounted || action == null) {
      return;
    }

    if (action == _GoogleFallbackAction.emailCode) {
      Navigator.pushNamed(context, EmailCodeLoginScreen.route);
      return;
    }

    if (action == _GoogleFallbackAction.retry && !kIsWeb) {
      await _mobileGoogle();
    }
  }

  @override
  void initState() {
    super.initState();
    if (kIsWeb) {
      // Ensure the platform plugin is initialized once.
      // This should not show UI.
      googleSignInInstance().isSignedIn();
      // Whenever GIS completes, the plugin sets currentUser; try to finish login.
      googleSignInInstance().onCurrentUserChanged.listen((user) {
        if (user != null) {
          _handleWebAfterGis();
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthController>();

    if (kIsWeb) {
      return SizedBox(
        height: 50,
        width: double.infinity,
        child: Center(child: googleGisButton()),
      );
    }

    return OutlineActionButton(
      text: widget.text,
      icon: Icons.g_mobiledata,
      onPressed: auth.busy ? null : _mobileGoogle,
    );
  }
}

enum _GoogleFallbackAction {
  emailCode,
  retry,
}
