// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:nutri_ai_app/nutri_app.dart';
import 'package:nutri_ai_app/services/api_config.dart';

void main() {
  testWidgets('App builds smoke test', (WidgetTester tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    SharedPreferences.setMockInitialValues({});

    await tester.binding.setSurfaceSize(const Size(430, 932));
    addTearDown(() async {
      await tester.binding.setSurfaceSize(null);
    });

    await tester.pumpWidget(
      const NutriApp(config: ApiConfig(baseUrl: 'http://localhost:8080')),
    );
    await tester.pumpAndSettle();

    expect(find.text('Get Started'), findsOneWidget);
  });
}
