import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/auth_models.dart';
import '../services/auth_api.dart';
import '../services/api_client.dart';

class AuthController extends ChangeNotifier {
  static const _kAccessToken = 'access_token';
  static const _kRefreshToken = 'refresh_token';
  static const _kProfileCompleted = 'profile_completed';
  static const _kHasPassword = 'has_password';

  final AuthApi _api;

  bool _initialized = false;
  bool _busy = false;
  String? _accessToken;
  String? _refreshToken;
  bool _profileCompleted = false;
  bool _hasPassword = false;

  AuthController(this._api);

  bool get initialized => _initialized;
  bool get busy => _busy;
  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshToken;
  bool get isAuthed => _accessToken != null && _accessToken!.isNotEmpty;
  bool get profileCompleted => _profileCompleted;
  bool get hasPassword => _hasPassword;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _accessToken = prefs.getString(_kAccessToken);
    _refreshToken = prefs.getString(_kRefreshToken);
    _profileCompleted = prefs.getBool(_kProfileCompleted) ?? false;
    _hasPassword = prefs.getBool(_kHasPassword) ?? false;
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

  Future<void> sendLoginCode(String email) async {
    await _wrap(() => _api.sendLoginCode(email));
  }

  Future<void> loginWithCode(EmailCodeLoginRequest req) async {
    await _wrap(() async {
      final tokens = await _api.loginWithCode(req);
      await _persist(tokens);
    });
  }

  Future<void> googleLogin(String idToken) async {
    await _wrap(() async {
      final tokens = await _api.googleLogin(idToken);
      await _persist(tokens);
    });
  }

  Future<void> setPassword(SetPasswordRequest req) async {
    await _wrap(() async {
      await withAuthRetry<void>(
          (token) => _api.setPassword(req, accessToken: token));
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool(_kHasPassword, true);
      _hasPassword = true;
    });
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_kAccessToken);
    await prefs.remove(_kRefreshToken);
    await prefs.remove(_kProfileCompleted);
    await prefs.remove(_kHasPassword);
    _accessToken = null;
    _refreshToken = null;
    _profileCompleted = false;
    _hasPassword = false;
    notifyListeners();
  }

  Future<void> markProfileCompleted() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_kProfileCompleted, true);
    _profileCompleted = true;
    notifyListeners();
  }

  /// Runs an authenticated API call and retries once on 401 by using the stored refresh token.
  /// This prevents the app from breaking after the access token (15m) expires.
  Future<T> withAuthRetry<T>(Future<T> Function(String accessToken) fn) async {
    final token = _accessToken;
    if (token == null || token.isEmpty) {
      throw ApiException(401, 'unauthorized');
    }

    try {
      return await fn(token);
    } on ApiException catch (e) {
      if (e.statusCode != 401) rethrow;
      final refreshToken = _refreshToken;
      if (refreshToken == null || refreshToken.isEmpty) rethrow;

      final tokens = await _api.refresh(refreshToken);
      await _persist(tokens);
      notifyListeners();
      return await fn(tokens.accessToken);
    }
  }

  Future<void> _persist(AuthTokens tokens) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kAccessToken, tokens.accessToken);
    await prefs.setString(_kRefreshToken, tokens.refreshToken);
    await prefs.setBool(_kProfileCompleted, tokens.profileCompleted);
    await prefs.setBool(_kHasPassword, tokens.hasPassword);
    _accessToken = tokens.accessToken;
    _refreshToken = tokens.refreshToken;
    _profileCompleted = tokens.profileCompleted;
    _hasPassword = tokens.hasPassword;
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
