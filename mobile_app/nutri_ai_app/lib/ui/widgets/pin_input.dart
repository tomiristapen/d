import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class PinInput extends StatefulWidget {
  final int length;
  final ValueChanged<String> onChanged;

  const PinInput({super.key, this.length = 6, required this.onChanged});

  @override
  State<PinInput> createState() => _PinInputState();
}

class _PinInputState extends State<PinInput> {
  late final List<TextEditingController> _controllers;
  late final List<FocusNode> _focusNodes;

  @override
  void initState() {
    super.initState();
    _controllers = List.generate(widget.length, (_) => TextEditingController());
    _focusNodes = List.generate(widget.length, (_) => FocusNode());
  }

  @override
  void dispose() {
    for (final c in _controllers) {
      c.dispose();
    }
    for (final f in _focusNodes) {
      f.dispose();
    }
    super.dispose();
  }

  void _emit() {
    final code = _controllers.map((c) => c.text).join();
    widget.onChanged(code);
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: List.generate(widget.length, (i) {
        return SizedBox(
          width: 44,
          child: TextField(
            controller: _controllers[i],
            focusNode: _focusNodes[i],
            keyboardType: TextInputType.number,
            textAlign: TextAlign.center,
            maxLength: 1,
            decoration: InputDecoration(
              counterText: '',
              contentPadding: const EdgeInsets.symmetric(vertical: 14),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(color: AppTheme.primary, width: 1.2),
              ),
            ),
            onChanged: (v) {
              if (v.isNotEmpty && i < widget.length - 1) {
                _focusNodes[i + 1].requestFocus();
              }
              if (v.isEmpty && i > 0) {
                _focusNodes[i - 1].requestFocus();
              }
              _emit();
            },
          ),
        );
      }),
    );
  }
}

