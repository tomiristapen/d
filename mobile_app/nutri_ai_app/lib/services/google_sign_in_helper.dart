import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:google_sign_in/google_sign_in.dart';

import 'api_client.dart';

String _googleClientId() {
  const definedClientId = String.fromEnvironment('GOOGLE_CLIENT_ID');
  final dotenvClientId = dotenv.env['GOOGLE_CLIENT_ID']?.trim();
  final clientId = definedClientId.isNotEmpty ? definedClientId : (dotenvClientId ?? '');
  if (clientId.isEmpty) {
    throw ApiException(0, 'Set GOOGLE_CLIENT_ID in `.env` (or pass --dart-define=GOOGLE_CLIENT_ID=...)');
  }
  return clientId;
}

GoogleSignIn? _cached;

GoogleSignIn googleSignInInstance() {
  if (_cached != null) return _cached!;
  final clientId = _googleClientId();
  _cached = GoogleSignIn(
    scopes: const ['email'],
    clientId: kIsWeb ? clientId : null,
    // For Android/iOS you typically need a "Web application" OAuth client ID here,
    // so Google returns an ID token that your backend can verify.
    serverClientId: kIsWeb ? null : clientId,
  );
  return _cached!;
}

Future<String> googleSignInGetIdToken({bool forceAccountPicker = true}) async {
  if (kIsWeb) {
    // On web `signIn()` is deprecated and may not return an idToken.
    // Use the GIS button (`renderButton`) and then read tokens from current user.
    throw ApiException(0, 'On Web use the Google button (GIS) instead of popup sign-in');
  }

  final signIn = googleSignInInstance();
  if (forceAccountPicker) {
    // Google Sign-In tends to reuse the last account without prompting. Best-effort reset.
    try {
      await signIn.signOut();
    } catch (_) {}
    try {
      await signIn.disconnect();
    } catch (_) {}
  }
  final account = await signIn.signIn();
  if (account == null) {
    throw ApiException(0, 'Google sign-in was cancelled');
  }

  final authData = await account.authentication;
  final idToken = authData.idToken;
  if (idToken == null || idToken.isEmpty) {
    throw ApiException(0, 'Google did not return an id_token (check Google Sign-In setup / client ID)');
  }
  return idToken;
}

Future<String> googleWebGetIdTokenFromCurrentUser() async {
  if (!kIsWeb) {
    throw ApiException(0, 'Web-only method');
  }
  final signIn = googleSignInInstance();
  final user = signIn.currentUser;
  if (user == null) {
    throw ApiException(0, 'Google user is not signed in yet');
  }
  final authData = await user.authentication;
  final idToken = authData.idToken;
  if (idToken == null || idToken.isEmpty) {
    throw ApiException(0, 'Google did not return an id_token (ensure you used the GIS button)');
  }
  return idToken;
}
