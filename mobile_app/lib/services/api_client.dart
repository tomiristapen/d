import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final int statusCode;
  final String message;

  ApiException(this.statusCode, this.message);

  @override
  String toString() => 'ApiException($statusCode): $message';
}

class ApiClient {
  final String baseUrl;
  final http.Client _client;

  ApiClient({required this.baseUrl, http.Client? client}) : _client = client ?? http.Client();

  Uri _uri(String path, [Map<String, String>? query]) {
    final base = baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length - 1) : baseUrl;
    return Uri.parse('$base$path').replace(queryParameters: query);
  }

  Future<Map<String, dynamic>> postJson(
    String path, {
    Map<String, dynamic>? body,
    String? bearerToken,
  }) async {
    final resp = await _client.post(
      _uri(path),
      headers: {
        'Content-Type': 'application/json',
        if (bearerToken != null) 'Authorization': 'Bearer $bearerToken',
      },
      body: jsonEncode(body ?? const {}),
    );
    return _decode(resp);
  }

  Future<Map<String, dynamic>> getJson(
    String path, {
    Map<String, String>? query,
    String? bearerToken,
  }) async {
    final resp = await _client.get(
      _uri(path, query),
      headers: {
        'Content-Type': 'application/json',
        if (bearerToken != null) 'Authorization': 'Bearer $bearerToken',
      },
    );
    return _decode(resp);
  }

  Map<String, dynamic> _decode(http.Response resp) {
    Map<String, dynamic> json;
    try {
      json = jsonDecode(resp.body) as Map<String, dynamic>;
    } catch (_) {
      json = {'error': resp.body};
    }

    if (resp.statusCode < 200 || resp.statusCode >= 300) {
      throw ApiException(resp.statusCode, (json['error'] ?? 'Request failed').toString());
    }
    return json;
  }
}

