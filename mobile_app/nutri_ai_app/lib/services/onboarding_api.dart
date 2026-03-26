import '../models/profile_models.dart';
import 'api_client.dart';

class OnboardingApi {
  final ApiClient _api;

  OnboardingApi(this._api);

  Future<void> completeProfile(CompleteProfileRequest req, {required String accessToken}) async {
    await _api.putJson(
      '/api/v1/profile',
      body: req.toJson(),
      bearerToken: accessToken,
    );
  }
}
