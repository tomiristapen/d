String normalizeBarcode(String raw) {
  final digits = raw.replaceAll(RegExp(r'[^0-9]'), '');
  return digits.trim();
}

bool isValidBarcode(String raw) {
  final barcode = normalizeBarcode(raw);
  return RegExp(r'^\d{8,14}$').hasMatch(barcode);
}
