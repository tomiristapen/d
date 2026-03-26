import '../models/diary_models.dart';
import 'api_client.dart';

class DiaryApi {
  final ApiClient _api;

  DiaryApi(this._api);

  Future<DiaryEntry> addEntry(
    DiaryAddRequest request, {
    required String accessToken,
    String? idempotencyKey,
  }) async {
    final json = await _api.postJson(
      '/api/v1/diary/entries',
      body: request.toJson(),
      bearerToken: accessToken,
      headers: {
        'Idempotency-Key': idempotencyKey ??
            '${request.source}-${DateTime.now().microsecondsSinceEpoch}',
      },
    );
    return DiaryEntry.fromJson(json);
  }

  Future<DiaryTodayResponse> getToday({required String accessToken}) async {
    final json = await _api.getJson(
      '/api/v1/diary/today',
      bearerToken: accessToken,
    );
    return DiaryTodayResponse.fromJson(json);
  }
}
