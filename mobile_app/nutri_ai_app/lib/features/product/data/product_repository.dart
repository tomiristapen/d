import 'package:nutri_ai_app/features/product/data/product_api.dart';
import 'package:nutri_ai_app/features/product/domain/product_model.dart';

class ProductRepository {
  final ProductApi _api;

  ProductRepository(this._api);

  Future<Product> getByBarcode(String barcode, {required String accessToken}) {
    return _api.getByBarcode(barcode, accessToken: accessToken);
  }

  Future<Product> analyzeProduct(AnalyzeRequest request,
      {required String accessToken}) {
    return _api.analyze(request, accessToken: accessToken);
  }

  Future<ManualAnalyzeResponse> analyzeManual(
    ManualAnalyzeRequest request, {
    required String accessToken,
  }) {
    return _api.analyzeManual(request, accessToken: accessToken);
  }

  Future<ManualCustomResponse> createCustomManualProduct(
    ManualCustomRequest request, {
    required String accessToken,
  }) {
    return _api.createCustomManualProduct(request, accessToken: accessToken);
  }

  Future<RecipeAnalyzeResponse> analyzeRecipe(
    RecipeAnalyzeRequest request, {
    required String accessToken,
  }) {
    return _api.analyzeRecipe(request, accessToken: accessToken);
  }

  Future<OcrDraft> buildOcrDraft({
    required List<String> images,
    String lang = 'eng+rus',
    required String accessToken,
  }) {
    return _api.buildOcrDraft(
        images: images, lang: lang, accessToken: accessToken);
  }
}
