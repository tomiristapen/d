class RegisterRequest {
  final String email;
  final String password;
  final String confirmPassword;

  RegisterRequest({required this.email, required this.password, required this.confirmPassword});

  Map<String, dynamic> toJson() => {
        'email': email,
        'password': password,
        'confirm_password': confirmPassword,
      };
}

class VerifyEmailRequest {
  final String email;
  final String code;

  VerifyEmailRequest({required this.email, required this.code});

  Map<String, dynamic> toJson() => {
        'email': email,
        'code': code,
      };
}

class LoginRequest {
  final String email;
  final String password;

  LoginRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() => {
        'email': email,
        'password': password,
      };
}

class AuthTokens {
  final String accessToken;
  final String refreshToken;
  final bool profileCompleted;

  AuthTokens({required this.accessToken, required this.refreshToken, required this.profileCompleted});

  factory AuthTokens.fromJson(Map<String, dynamic> json) {
    return AuthTokens(
      accessToken: (json['access_token'] ?? '').toString(),
      refreshToken: (json['refresh_token'] ?? '').toString(),
      profileCompleted: (json['profile_completed'] ?? false) == true,
    );
  }
}

