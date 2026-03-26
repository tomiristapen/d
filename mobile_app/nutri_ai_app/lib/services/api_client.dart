import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final int statusCode;
  final String message;
  final Uri? uri;

  ApiException(this.statusCode, this.message, {this.uri});

  @override
  String toString() => uri == null ? 'ApiException($statusCode): $message' : 'ApiException($statusCode): $message ($uri)';
}

class ApiClient {
  final String baseUrl;
  final http.Client _client;
  final Duration _timeout;

  ApiClient({
    required this.baseUrl,
    http.Client? client,
    Duration timeout = const Duration(seconds: 60),
  })  : _client = client ?? http.Client(),
        _timeout = timeout;

  Uri _uri(String path, [Map<String, String>? query]) {
    final base = baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length - 1) : baseUrl;
    return Uri.parse('$base$path').replace(queryParameters: query);
  }

  Future<Map<String, dynamic>> postJson(
    String path, {
    Map<String, dynamic>? body,
    String? bearerToken,
    Map<String, String>? headers,
  }) async {
    final uri = _uri(path);
    final resp = await _send(
      () => _client.post(
        uri,
        headers: _headers(
          bearerToken: bearerToken,
          extra: headers,
        ),
        body: jsonEncode(body ?? const {}),
      ),
      uri,
    );
    return _decode(resp, uri);
  }

  Future<Map<String, dynamic>> putJson(
    String path, {
    Map<String, dynamic>? body,
    String? bearerToken,
    Map<String, String>? headers,
  }) async {
    final uri = _uri(path);
    final resp = await _send(
      () => _client.put(
        uri,
        headers: _headers(
          bearerToken: bearerToken,
          extra: headers,
        ),
        body: jsonEncode(body ?? const {}),
      ),
      uri,
    );
    return _decode(resp, uri);
  }

  Future<void> delete(
    String path, {
    String? bearerToken,
    Map<String, String>? headers,
  }) async {
    final uri = _uri(path);
    final resp = await _send(
      () => _client.delete(
        uri,
        headers: _headers(
          bearerToken: bearerToken,
          extra: headers,
        ),
      ),
      uri,
    );
    _decode(resp, uri);
  }

  Future<Map<String, dynamic>> getJson(
    String path, {
    Map<String, String>? query,
    String? bearerToken,
    Map<String, String>? headers,
  }) async {
    final uri = _uri(path, query);
    final resp = await _send(
      () => _client.get(
        uri,
        headers: _headers(
          bearerToken: bearerToken,
          extra: headers,
        ),
      ),
      uri,
    );
    return _decode(resp, uri);
  }

  Future<http.Response> _send(Future<http.Response> Function() fn, Uri uri) async {
    try {
      return await fn().timeout(_timeout);
    } on TimeoutException {
      throw ApiException(
        0,
        'Timeout. Backend is not responding: $uri',
        uri: uri,
      );
    } on http.ClientException catch (e) {
      final msg = e.message.trim().isNotEmpty ? e.message.trim() : 'Connection failed';
      throw ApiException(
        0,
        'Cannot connect to backend: $uri ($msg)',
        uri: uri,
      );
    } catch (e) {
      throw ApiException(0, 'Network error calling $uri: $e', uri: uri);
    }
  }

  Map<String, dynamic> _decode(http.Response resp, Uri uri) {
    if (resp.body.isEmpty) {
      if (resp.statusCode >= 200 && resp.statusCode < 300) {
        return const <String, dynamic>{};
      }
      return <String, dynamic>{'error': ''};
    }

    Map<String, dynamic> json;
    try {
      json = jsonDecode(resp.body) as Map<String, dynamic>;
    } catch (_) {
      json = {'error': resp.body};
    }

    if (resp.statusCode < 200 || resp.statusCode >= 300) {
      final err = json['error'];
      if (err is String) {
        throw ApiException(resp.statusCode, err, uri: uri);
      }
      if (err is Map) {
        final msg = err['message']?.toString();
        throw ApiException(resp.statusCode, msg?.isNotEmpty == true ? msg! : 'Request failed', uri: uri);
      }
      throw ApiException(resp.statusCode, 'Request failed', uri: uri);
    }
    return json;
  }

  Map<String, String> _headers({
    String? bearerToken,
    Map<String, String>? extra,
  }) {
    return {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
      'X-Timezone-Offset-Minutes':
          DateTime.now().timeZoneOffset.inMinutes.toString(),
      if (bearerToken != null) 'Authorization': 'Bearer $bearerToken',
      ...?extra,
    };
  }
}
