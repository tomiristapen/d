import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/auth_models.dart';
import '../services/auth_api.dart';

class AuthController extends ChangeNotifier {
  static const _kAccessToken = 'access_token';
  static const _kRefreshToken = 'refresh_token';
  static const _kProfileCompleted = 'profile_completed';

  final AuthApi _api;

  bool _initialized = false;
  bool _busy = false;
  String? _accessToken;
  String? _refreshToken;
  bool _profileCompleted = false;

  AuthController(this._api);

  bool get initialized => _initialized;
  bool get busy => _busy;
  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshToken;
  bool get isAuthed => _accessToken != null && _accessToken!.isNotEmpty;
  bool get profileCompleted => _profileCompleted;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _accessToken = prefs.getString(_kAccessToken);
    _refreshToken = prefs.getString(_kRefreshToken);
    _profileCompleted = prefs.getBool(_kProfileCompleted) ?? false;
    _initialized = true;
    notifyListeners();
  }

  Future<void> register(RegisterRequest req) async {
    await _wrap(() => _api.register(req));
  }

  Future<void> resendCode(String email) async {
    await _wrap(() => _api.sendVerificationCode(email));
  }

  Future<void> verifyEmail(VerifyEmailRequest req) async {
    await _wrap(() => _api.verifyEmail(req));
  }

  Future<void> login(LoginRequest req) async {
    await _wrap(() async {
      final tokens = await _api.login(req);
      await _persist(tokens);
    });
  }

  Future<void> googleLogin(String idToken) async {
    await _wrap(() async {
      final tokens = await _api.googleLogin(idToken);
      await _persist(tokens);
    });
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_kAccessToken);
    await prefs.remove(_kRefreshToken);
    await prefs.remove(_kProfileCompleted);
    _accessToken = null;
    _refreshToken = null;
    _profileCompleted = false;
    notifyListeners();
  }

  Future<void> _persist(AuthTokens tokens) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kAccessToken, tokens.accessToken);
    await prefs.setString(_kRefreshToken, tokens.refreshToken);
    await prefs.setBool(_kProfileCompleted, tokens.profileCompleted);
    _accessToken = tokens.accessToken;
    _refreshToken = tokens.refreshToken;
    _profileCompleted = tokens.profileCompleted;
  }

  Future<void> _wrap(Future<void> Function() fn) async {
    _busy = true;
    notifyListeners();
    try {
      await fn();
    } finally {
      _busy = false;
      notifyListeners();
    }
  }
}
