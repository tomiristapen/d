import '../models/product_models.dart';
import 'api_client.dart';

class ProductsApi {
  final ApiClient _api;

  ProductsApi(this._api);

  Future<Product> getByBarcode(String barcode, {required String accessToken}) async {
    final json = await _api.getJson(
      '/products/by-barcode/$barcode',
      bearerToken: accessToken,
    );
    return Product.fromJson(json);
  }
}

