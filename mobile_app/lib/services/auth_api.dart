import '../models/auth_models.dart';
import 'api_client.dart';

class AuthApi {
  final ApiClient _api;

  AuthApi(this._api);

  Future<void> register(RegisterRequest req) async {
    await _api.postJson('/auth/register', body: req.toJson());
  }

  Future<void> sendVerificationCode(String email) async {
    await _api.postJson('/auth/send-verification-code', body: {'email': email});
  }

  Future<void> verifyEmail(VerifyEmailRequest req) async {
    await _api.postJson('/auth/verify-email', body: req.toJson());
  }

  Future<AuthTokens> login(LoginRequest req) async {
    final json = await _api.postJson('/auth/login', body: req.toJson());
    return AuthTokens.fromJson(json);
  }

  Future<AuthTokens> googleLogin(String idToken) async {
    final json = await _api.postJson('/auth/google', body: {'id_token': idToken});
    return AuthTokens.fromJson(json);
  }

  Future<AuthTokens> refresh(String refreshToken) async {
    final json = await _api.postJson('/auth/refresh', body: {'refresh_token': refreshToken});
    return AuthTokens.fromJson(json);
  }
}

