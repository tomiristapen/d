import 'package:flutter_dotenv/flutter_dotenv.dart';

class ApiConfig {
  final String baseUrl;

  ApiConfig({required this.baseUrl});

  factory ApiConfig.fromEnvironment() {
    const defined = String.fromEnvironment('API_BASE_URL');
    String? fromDotenv;
    try {
      fromDotenv = dotenv.env['API_BASE_URL'];
    } catch (_) {
      fromDotenv = null;
    }
    final baseUrl = defined.isNotEmpty ? defined : (fromDotenv?.trim().isNotEmpty == true ? fromDotenv!.trim() : 'http://localhost:8080');
    return ApiConfig(baseUrl: baseUrl);
  }
}
