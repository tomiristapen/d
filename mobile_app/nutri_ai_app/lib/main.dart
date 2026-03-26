import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

import 'services/api_config.dart';
import 'nutri_app.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await dotenv.load(fileName: '.env');
  } catch (_) {
    // Allow running without .env, using auto-discovery or --dart-define.
  }
  runApp(const _AppBootstrap());
}

class _AppBootstrap extends StatefulWidget {
  const _AppBootstrap();

  @override
  State<_AppBootstrap> createState() => _AppBootstrapState();
}

class _AppBootstrapState extends State<_AppBootstrap> {
  late Future<ApiConfig> _configFuture;

  @override
  void initState() {
    super.initState();
    _configFuture = ApiConfig.resolve();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<ApiConfig>(
      future: _configFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: _BootstrapScreen(
              title: 'Connecting to backend',
              subtitle: 'Trying saved and local network addresses...',
              busy: true,
            ),
          );
        }

        if (snapshot.hasError || !snapshot.hasData) {
          return MaterialApp(
            debugShowCheckedModeBanner: false,
            home: _BootstrapScreen(
              title: 'Could not prepare the app',
              subtitle: 'Check that the backend is running, then try again.',
              onRetry: () => setState(() => _configFuture = ApiConfig.resolve()),
            ),
          );
        }

        return NutriApp(config: snapshot.data!);
      },
    );
  }
}

class _BootstrapScreen extends StatelessWidget {
  final String title;
  final String subtitle;
  final bool busy;
  final VoidCallback? onRetry;

  const _BootstrapScreen({
    required this.title,
    required this.subtitle,
    this.busy = false,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 320),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (busy) ...[
                  const CircularProgressIndicator(),
                  const SizedBox(height: 20),
                ],
                Text(
                  title,
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 10),
                Text(
                  subtitle,
                  textAlign: TextAlign.center,
                  style: const TextStyle(color: Colors.black54),
                ),
                if (!busy && onRetry != null) ...[
                  const SizedBox(height: 20),
                  ElevatedButton(
                    onPressed: onRetry,
                    child: const Text('Retry'),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
