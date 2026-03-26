import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../../features/product/domain/barcode_normalizer.dart';

class BarcodeScanScreen extends StatefulWidget {
  static const route = '/scan';

  const BarcodeScanScreen({super.key});

  @override
  State<BarcodeScanScreen> createState() => _BarcodeScanScreenState();
}

class _BarcodeScanScreenState extends State<BarcodeScanScreen> {
  bool _found = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Scan barcode')),
      body: MobileScanner(
        onDetect: (capture) {
          if (_found) return;
          final barcodes = capture.barcodes;
          if (barcodes.isEmpty) return;
          final raw = barcodes.first.rawValue?.trim();
          if (raw == null || raw.isEmpty) return;
          final normalized = normalizeBarcode(raw);
          if (!isValidBarcode(normalized)) return;
          _found = true;
          Navigator.pop(context, normalized);
        },
      ),
    );
  }
}
