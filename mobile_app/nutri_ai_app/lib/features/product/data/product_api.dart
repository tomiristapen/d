import 'package:nutri_ai_app/features/product/domain/product_model.dart';
import 'package:nutri_ai_app/services/api_client.dart';

class ProductApi {
  final ApiClient _api;

  ProductApi(this._api);

  Future<Product> getByBarcode(String barcode,
      {required String accessToken}) async {
    final json = await _api.getJson(
      '/api/v1/products/$barcode',
      bearerToken: accessToken,
    );
    return Product.fromJson(json);
  }

  Future<Product> analyze(AnalyzeRequest request,
      {required String accessToken}) async {
    final json = await _api.postJson(
      '/api/v1/products/analyze',
      body: request.toJson(),
      bearerToken: accessToken,
    );
    return Product.fromJson(json);
  }

  Future<ManualAnalyzeResponse> analyzeManual(
    ManualAnalyzeRequest request, {
    required String accessToken,
  }) async {
    final json = await _api.postJson(
      '/api/v1/manual/analyze',
      body: request.toJson(),
      bearerToken: accessToken,
    );
    return ManualAnalyzeResponse.fromJson(json);
  }

  Future<ManualCustomResponse> createCustomManualProduct(
    ManualCustomRequest request, {
    required String accessToken,
  }) async {
    final json = await _api.postJson(
      '/api/v1/manual/custom',
      body: request.toJson(),
      bearerToken: accessToken,
    );
    return ManualCustomResponse.fromJson(json);
  }

  Future<RecipeAnalyzeResponse> analyzeRecipe(
    RecipeAnalyzeRequest request, {
    required String accessToken,
  }) async {
    final json = await _api.postJson(
      '/api/v1/recipe/analyze',
      body: request.toJson(),
      bearerToken: accessToken,
    );
    return RecipeAnalyzeResponse.fromJson(json);
  }

  Future<OcrDraft> buildOcrDraft({
    required List<String> images,
    String lang = 'eng+rus',
    required String accessToken,
  }) async {
    final body = <String, dynamic>{'lang': lang};
    if (images.length == 1) {
      body['image'] = images.first;
    } else {
      body['images'] = images;
    }

    final json = await _api.postJson(
      '/api/v1/products/ocr/draft',
      body: body,
      bearerToken: accessToken,
    );
    return OcrDraft.fromJson(json);
  }
}
