import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

import 'nutri_app.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  // Loads `mobile_app/.env` (added as an asset) when present.
  // `--dart-define` values can still override it.
  try {
    await dotenv.load(fileName: '.env');
  } catch (_) {
    // Allow running without .env (e.g. first launch), using defaults or --dart-define.
  }
  runApp(const NutriApp());
}
