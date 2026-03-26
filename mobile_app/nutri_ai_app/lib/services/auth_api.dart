import '../models/auth_models.dart';
import 'api_client.dart';

class AuthApi {
  final ApiClient _api;

  AuthApi(this._api);

  Future<void> register(RegisterRequest req) async {
    await _api.postJson('/api/v1/auth/register', body: req.toJson());
  }

  Future<void> sendVerificationCode(String email) async {
    await _api.postJson('/api/v1/auth/send-verification-code',
        body: {'email': email});
  }

  Future<void> verifyEmail(VerifyEmailRequest req) async {
    await _api.postJson('/api/v1/auth/verify-email', body: req.toJson());
  }

  Future<AuthTokens> login(LoginRequest req) async {
    final json = await _api.postJson('/api/v1/auth/login', body: req.toJson());
    return AuthTokens.fromJson(json);
  }

  Future<void> sendLoginCode(String email) async {
    await _api.postJson('/api/v1/auth/send-login-code', body: {'email': email});
  }

  Future<AuthTokens> loginWithCode(EmailCodeLoginRequest req) async {
    final json =
        await _api.postJson('/api/v1/auth/login-with-code', body: req.toJson());
    return AuthTokens.fromJson(json);
  }

  Future<AuthTokens> googleLogin(String idToken) async {
    final json =
        await _api.postJson('/api/v1/auth/google', body: {'id_token': idToken});
    return AuthTokens.fromJson(json);
  }

  Future<void> setPassword(SetPasswordRequest req,
      {required String accessToken}) async {
    await _api.postJson(
      '/api/v1/auth/set-password',
      body: req.toJson(),
      bearerToken: accessToken,
    );
  }

  Future<AuthTokens> refresh(String refreshToken) async {
    final json = await _api.postJson('/api/v1/auth/refresh',
        body: {'refresh_token': refreshToken});
    return AuthTokens.fromJson(json);
  }
}
