import 'package:nutri_ai_app/features/product/data/product_api.dart';
import 'package:nutri_ai_app/features/product/domain/product_model.dart';

class ProductsApi {
  final ProductApi _api;

  ProductsApi(this._api);

  Future<Product> getByBarcode(String barcode, {required String accessToken}) async {
    return _api.getByBarcode(barcode, accessToken: accessToken);
  }

  Future<Product> analyzeProduct(AnalyzeRequest request, {required String accessToken}) async {
    return _api.analyze(request, accessToken: accessToken);
  }

  Future<OcrDraft> buildOcrDraft({
    required List<String> images,
    String lang = 'eng+rus',
    required String accessToken,
  }) async {
    return _api.buildOcrDraft(images: images, lang: lang, accessToken: accessToken);
  }
}
