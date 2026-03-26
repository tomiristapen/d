import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'services/api_config.dart';
import 'services/api_client.dart';
import 'services/auth_api.dart';
import 'services/diary_api.dart';
import 'services/ingredients_api.dart';
import 'services/onboarding_api.dart';
import 'services/products_api.dart';
import 'features/product/data/product_api.dart';
import 'features/product/data/product_repository.dart';
import 'state/auth_controller.dart';
import 'state/profile_controller.dart';
import 'theme/app_theme.dart';
import 'ui/screens/auth/auth_choice_screen.dart';
import 'ui/screens/auth/email_code_login_screen.dart';
import 'ui/screens/auth/login_screen.dart';
import 'ui/screens/auth/register_screen.dart';
import 'ui/screens/auth/set_password_screen.dart';
import 'ui/screens/auth/verify_email_screen.dart';
import 'ui/screens/home/home_screen.dart';
import 'ui/screens/onboarding/dietary_preferences_screen.dart';
import 'ui/screens/onboarding/profile_setup_screen.dart';
import 'ui/screens/products/barcode_scan_screen.dart';
import 'ui/screens/products/manual_product_screen.dart';
import 'ui/screens/products/product_details_screen.dart';
import 'ui/screens/products/ocr_scan_screen.dart';
import 'ui/screens/products/recipe_create_screen.dart';
import 'ui/screens/welcome/welcome_screen.dart';

class NutriApp extends StatelessWidget {
  final ApiConfig config;

  const NutriApp({super.key, required this.config});

  @override
  Widget build(BuildContext context) {
    final apiClient = ApiClient(baseUrl: config.baseUrl);

    return MultiProvider(
      providers: [
        Provider.value(value: apiClient),
        Provider(create: (_) => AuthApi(apiClient)),
        Provider(create: (_) => DiaryApi(apiClient)),
        Provider(create: (_) => OnboardingApi(apiClient)),
        Provider(create: (_) => IngredientsApi(apiClient)),
        Provider(create: (ctx) => ProductApi(ctx.read<ApiClient>())),
        Provider(create: (ctx) => ProductRepository(ctx.read<ProductApi>())),
        Provider(create: (ctx) => ProductsApi(ctx.read<ProductApi>())),
        ChangeNotifierProvider(
          create: (ctx) => AuthController(ctx.read<AuthApi>())..init(),
        ),
        ChangeNotifierProvider(
          create: (ctx) => ProfileController(ctx.read<OnboardingApi>()),
        ),
      ],
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        title: 'Nutri AI',
        theme: AppTheme.light(),
        routes: {
          WelcomeScreen.route: (_) => const WelcomeScreen(),
          AuthChoiceScreen.route: (_) => const AuthChoiceScreen(),
          RegisterScreen.route: (_) => const RegisterScreen(),
          VerifyEmailScreen.route: (_) => const VerifyEmailScreen(),
          LoginScreen.route: (_) => const LoginScreen(),
          EmailCodeLoginScreen.route: (_) => const EmailCodeLoginScreen(),
          SetPasswordScreen.route: (_) => const SetPasswordScreen(),
          ProfileSetupScreen.route: (_) => const ProfileSetupScreen(),
          DietaryPreferencesScreen.route: (_) =>
              const DietaryPreferencesScreen(),
          HomeScreen.route: (_) => const HomeScreen(),
          BarcodeScanScreen.route: (_) => const BarcodeScanScreen(),
          OcrScanScreen.route: (_) => const OcrScanScreen(),
          ManualProductScreen.route: (_) => const ManualProductScreen(),
          RecipeCreateScreen.route: (_) => const RecipeCreateScreen(),
        },
        initialRoute: WelcomeScreen.route,
        onGenerateRoute: (settings) {
          if (settings.name == ProductDetailsScreen.route) {
            final barcode = settings.arguments as String?;
            if (barcode == null || barcode.isEmpty) return null;
            return MaterialPageRoute(
                builder: (_) => ProductDetailsScreen(barcode: barcode));
          }
          return null;
        },
      ),
    );
  }
}
