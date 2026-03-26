import 'package:flutter/foundation.dart';

import '../models/profile_models.dart';
import '../services/onboarding_api.dart';

class ProfileController extends ChangeNotifier {
  final OnboardingApi _api;

  ProfileController(this._api);

  int age = 25;
  String gender = '';
  double heightCm = 170;
  double weightKg = 70;
  String activityLevel = 'moderate';
  String goal = 'maintain';

  final Set<String> allergies = {
    'peanut',
    'milk',
    'egg',
    'fish',
    'soy',
    'wheat',
    'sesame',
  };
  final Set<String> intolerances = {'lactose', 'gluten'};

  final List<String> selectedAllergies = [];
  final List<String> customAllergies = [];
  final List<String> selectedIntolerances = [];

  String dietType = 'none';
  String religiousRestriction = 'none';

  Future<void> submit({required String accessToken}) async {
    final req = CompleteProfileRequest(
      age: age,
      gender: gender,
      heightCm: heightCm,
      weightKg: weightKg,
      activityLevel: activityLevel,
      goal: goal,
      allergies: List.of(selectedAllergies),
      customAllergies: List.of(customAllergies),
      intolerances: List.of(selectedIntolerances),
      dietType: dietType,
      religiousRestriction: religiousRestriction,
    );
    await _api.completeProfile(req, accessToken: accessToken);
  }
}
