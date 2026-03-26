import 'package:flutter/material.dart';

class AppTheme {
  static const Color primary = Color(0xFF0B7A61);
  static const Color primaryDark = Color(0xFF0A6A55);
  static const Color bg = Color(0xFFF6F8F7);
  static const Color text = Color(0xFF0F172A);
  static const Color muted = Color(0xFF64748B);

  static ThemeData light() {
    final scheme = ColorScheme.fromSeed(seedColor: primary, primary: primary);
    return ThemeData(
      colorScheme: scheme,
      scaffoldBackgroundColor: Colors.white,
      useMaterial3: true,
      textTheme: const TextTheme(
        headlineSmall: TextStyle(fontSize: 26, fontWeight: FontWeight.w700, color: text),
        titleLarge: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: text),
        bodyMedium: TextStyle(fontSize: 14, color: text),
        bodySmall: TextStyle(fontSize: 12, color: muted),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: Colors.white,
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(14), borderSide: BorderSide(color: Colors.grey.shade300)),
        enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(14), borderSide: BorderSide(color: Colors.grey.shade300)),
        focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(14), borderSide: const BorderSide(color: primary, width: 1.2)),
      ),
    );
  }
}

