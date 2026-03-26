import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import 'api_host_discovery_stub.dart'
    if (dart.library.io) 'api_host_discovery_io.dart' as host_discovery;

class ApiConfig {
  static const _defaultBaseUrl = 'http://localhost:8080';
  static const _savedBaseUrlKey = 'api.base_url';
  static const _healthPath = '/api/v1/healthz';

  final String baseUrl;

  const ApiConfig({required this.baseUrl});

  static Future<ApiConfig> resolve() async {
    final prefs = await SharedPreferences.getInstance();
    final defined = const String.fromEnvironment('API_BASE_URL').trim();
    final fromDotenv = _readDotenv('API_BASE_URL');
    final saved = prefs.getString(_savedBaseUrlKey)?.trim();

    final candidates = _unique([
      if (defined.isNotEmpty) defined,
      if (fromDotenv?.isNotEmpty == true) fromDotenv!,
      if (saved?.isNotEmpty == true) saved!,
      ...host_discovery.defaultCandidates(),
      _defaultBaseUrl,
    ]);

    for (final candidate in candidates) {
      if (await _isHealthy(candidate)) {
        await prefs.setString(_savedBaseUrlKey, candidate);
        return ApiConfig(baseUrl: candidate);
      }
    }

    final discovered = await host_discovery.discoverOnLocalNetwork();
    if (discovered != null && discovered.isNotEmpty) {
      await prefs.setString(_savedBaseUrlKey, discovered);
      return ApiConfig(baseUrl: discovered);
    }

    final fallback = candidates.isNotEmpty ? candidates.first : _defaultBaseUrl;
    return ApiConfig(baseUrl: fallback);
  }

  static String? _readDotenv(String key) {
    try {
      final value = dotenv.env[key]?.trim();
      return value == null || value.isEmpty ? null : value;
    } catch (_) {
      return null;
    }
  }

  static Future<bool> _isHealthy(String baseUrl) async {
    final client = http.Client();
    try {
      final normalized =
          baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length - 1) : baseUrl;
      final uri = Uri.parse('$normalized$_healthPath');
      final response = await client
          .get(uri, headers: const {'Accept': 'application/json'})
          .timeout(const Duration(milliseconds: 700));
      return response.statusCode == 200 && response.body.contains('"status"');
    } catch (_) {
      return false;
    } finally {
      client.close();
    }
  }

  static List<String> _unique(List<String> values) {
    final seen = <String>{};
    final result = <String>[];
    for (final raw in values) {
      final value = raw.trim();
      if (value.isEmpty || !seen.add(value)) {
        continue;
      }
      result.add(value);
    }
    return result;
  }
}
