import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../../../features/product/data/product_repository.dart';
import '../../../state/auth_controller.dart';
import 'ocr_draft_review_screen.dart';

class OcrScanScreen extends StatefulWidget {
  static const route = '/products/ocr';

  const OcrScanScreen({super.key});

  @override
  State<OcrScanScreen> createState() => _OcrScanScreenState();
}

class _OcrScanScreenState extends State<OcrScanScreen> {
  bool _loading = false;
  String? _error;

  Future<void> _captureAndAnalyze() async {
    final auth = context.read<AuthController>();
    final repo = context.read<ProductRepository>();
    if (!auth.isAuthed) {
      setState(() => _error = 'Not authenticated');
      return;
    }

    setState(() {
      _error = null;
      _loading = true;
    });

    try {
      final picker = ImagePicker();
      final file = await picker.pickImage(
        source: ImageSource.camera,
        preferredCameraDevice: CameraDevice.rear,
        maxWidth: 2600,
        maxHeight: 2600,
        imageQuality: 95,
      );
      if (!mounted) return;
      if (file == null) {
        setState(() => _loading = false);
        return;
      }

      // Base64 encoding a full-res photo can freeze the UI on some devices (ANR).
      // Do it in a background isolate.
      final base64Image = await compute(_base64FromFilePath, file.path);

      final draft = await auth.withAuthRetry(
        (token) => repo.buildOcrDraft(images: [base64Image], lang: 'rus+eng', accessToken: token),
      );
      if (!mounted) return;

      setState(() => _loading = false);
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (_) => OcrDraftReviewScreen(draft: draft, images: [base64Image], lang: 'rus+eng')),
      );
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('OCR draft failed: $e');
        debugPrintStack(stackTrace: st);
      }
      if (!mounted) return;
      setState(() {
        _error = kDebugMode ? 'Failed to analyze label: $e' : 'Failed to analyze label. Please try again.';
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Scan label (OCR)')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Take a photo of the label. We will extract ingredients + nutrition and ask you to verify it.',
                style: TextStyle(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 14),
              if (_error != null) ...[
                Text(_error!, style: const TextStyle(color: Colors.red)),
                const SizedBox(height: 10),
              ],
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  onPressed: _loading ? null : _captureAndAnalyze,
                  icon: _loading
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.camera_alt_outlined),
                  label: Text(_loading ? 'Analyzing...' : 'Take photo'),
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                'Tip: make sure the text is sharp and well-lit.',
                style: TextStyle(color: Colors.black54),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

String _base64FromFilePath(String path) {
  final bytes = File(path).readAsBytesSync();
  return base64Encode(bytes);
}
