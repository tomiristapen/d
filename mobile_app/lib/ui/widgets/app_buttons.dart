import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class PrimaryButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final bool busy;

  const PrimaryButton({super.key, required this.text, required this.onPressed, this.busy = false});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 50,
      width: double.infinity,
      child: FilledButton(
        style: FilledButton.styleFrom(
          backgroundColor: AppTheme.primary,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
        ),
        onPressed: busy ? null : onPressed,
        child: busy
            ? const SizedBox(height: 18, width: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
            : Text(text, style: const TextStyle(fontWeight: FontWeight.w600)),
      ),
    );
  }
}

class OutlineActionButton extends StatelessWidget {
  final String text;
  final IconData icon;
  final VoidCallback? onPressed;

  const OutlineActionButton({super.key, required this.text, required this.icon, required this.onPressed});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 50,
      width: double.infinity,
      child: OutlinedButton.icon(
        style: OutlinedButton.styleFrom(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          side: BorderSide(color: Colors.grey.shade300),
        ),
        onPressed: onPressed,
        icon: Icon(icon, size: 18, color: AppTheme.primary),
        label: Text(text, style: const TextStyle(fontWeight: FontWeight.w600)),
      ),
    );
  }
}

