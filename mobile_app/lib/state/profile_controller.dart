import 'package:flutter/foundation.dart';

import '../models/profile_models.dart';
import '../services/onboarding_api.dart';

class ProfileController extends ChangeNotifier {
  final OnboardingApi _api;

  ProfileController(this._api);

  int age = 25;
  String gender = '';
  int heightCm = 170;
  int weightKg = 70;
  String nutritionGoal = 'maintain_weight';

  final Set<String> allergies = {'Peanuts', 'Milk', 'Eggs', 'Fish', 'Soy', 'Wheat', 'Sesame'};
  final Set<String> intolerances = {'Lactose', 'Gluten'};

  final List<String> selectedAllergies = [];
  final List<String> customAllergies = [];
  final List<String> selectedIntolerances = [];

  String dietType = 'none';
  String religiousRestriction = 'none';

  Future<void> submit({required String accessToken}) async {
    final req = CompleteProfileRequest(
      age: age,
      gender: gender.isEmpty ? 'other' : gender,
      heightCm: heightCm,
      weightKg: weightKg,
      nutritionGoal: nutritionGoal,
      allergies: List.of(selectedAllergies),
      customAllergies: List.of(customAllergies),
      intolerances: List.of(selectedIntolerances),
      dietType: dietType,
      religiousRestriction: religiousRestriction,
    );
    await _api.completeProfile(req, accessToken: accessToken);
  }
}

