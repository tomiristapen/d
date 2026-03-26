import 'api_client.dart';

class IngredientsApi {
  final ApiClient _api;

  IngredientsApi(this._api);

  Future<List<String>> autocomplete(String query, {required String accessToken}) async {
    final json = await _api.getJson(
      '/api/v1/ingredients/autocomplete',
      query: {'q': query},
      bearerToken: accessToken,
    );
    final items = json['items'];
    if (items is List) {
      return items.map((e) => e.toString()).toList();
    }
    return const [];
  }
}

