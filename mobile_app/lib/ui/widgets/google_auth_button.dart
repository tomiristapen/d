import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../services/api_client.dart';
import '../../services/google_sign_in_helper.dart';
import '../../state/auth_controller.dart';
import '../screens/home/home_screen.dart';
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
      Navigator.pushNamedAndRemoveUntil(context, HomeScreen.route, (_) => false);
    } else {
      Navigator.pushNamedAndRemoveUntil(context, ProfileSetupScreen.route, (_) => false);
    }
  }

  Future<void> _mobileGoogle() async {
    final auth = context.read<AuthController>();
    try {
      final idToken = await googleSignInGetIdToken();
      await auth.googleLogin(idToken);
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
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.toString())));
    } finally {
      _handlingWeb = false;
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
