import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../services/ingredients_api.dart';
import '../../state/auth_controller.dart';
import '../../theme/app_theme.dart';

class IngredientAutocompleteField extends StatefulWidget {
  final TextEditingController controller;
  final String label;
  final String? hintText;
  final bool enabled;
  final ValueChanged<String>? onSuggestionSelected;

  const IngredientAutocompleteField({
    super.key,
    required this.controller,
    required this.label,
    this.hintText,
    this.enabled = true,
    this.onSuggestionSelected,
  });

  @override
  State<IngredientAutocompleteField> createState() =>
      _IngredientAutocompleteFieldState();
}

class _IngredientAutocompleteFieldState
    extends State<IngredientAutocompleteField> {
  Timer? _debounce;
  bool _ignoreNextChange = false;
  bool _loading = false;
  List<String> _suggestions = const [];

  @override
  void initState() {
    super.initState();
    widget.controller.addListener(_onChanged);
  }

  @override
  void didUpdateWidget(covariant IngredientAutocompleteField oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller == widget.controller) return;
    oldWidget.controller.removeListener(_onChanged);
    widget.controller.addListener(_onChanged);
  }

  @override
  void dispose() {
    _debounce?.cancel();
    widget.controller.removeListener(_onChanged);
    super.dispose();
  }

  void _onChanged() {
    if (_ignoreNextChange) {
      _ignoreNextChange = false;
      return;
    }
    _debounce?.cancel();
    final query = widget.controller.text.trim();
    if (!widget.enabled || query.isEmpty) {
      if (_suggestions.isNotEmpty || _loading) {
        setState(() {
          _loading = false;
          _suggestions = const [];
        });
      }
      return;
    }

    _debounce = Timer(
        const Duration(milliseconds: 250), () => _fetchSuggestions(query));
  }

  Future<void> _fetchSuggestions(String query) async {
    if (!mounted) return;
    setState(() => _loading = true);

    try {
      final auth = context.read<AuthController>();
      final api = context.read<IngredientsApi>();
      final items = await auth.withAuthRetry(
          (token) => api.autocomplete(query, accessToken: token));
      if (!mounted || widget.controller.text.trim() != query) return;

      final current = query.toLowerCase();
      setState(() {
        _suggestions = items
            .where((item) =>
                item.trim().isNotEmpty && item.trim().toLowerCase() != current)
            .take(6)
            .toList();
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _suggestions = const []);
    } finally {
      if (mounted && widget.controller.text.trim() == query) {
        setState(() => _loading = false);
      }
    }
  }

  void _select(String value) {
    _debounce?.cancel();
    _ignoreNextChange = true;
    widget.controller
      ..text = value
      ..selection = TextSelection.collapsed(offset: value.length);
    setState(() => _suggestions = const []);
    widget.onSuggestionSelected?.call(value);
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(
          controller: widget.controller,
          enabled: widget.enabled,
          decoration: InputDecoration(
            labelText: widget.label,
            hintText: widget.hintText,
          ),
        ),
        if (_loading) ...[
          const SizedBox(height: 6),
          const LinearProgressIndicator(minHeight: 2),
        ],
        if (_suggestions.isNotEmpty) ...[
          const SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Column(
              children: _suggestions
                  .map(
                    (item) => ListTile(
                      dense: true,
                      title: Text(item),
                      trailing: const Icon(Icons.north_west,
                          size: 18, color: AppTheme.muted),
                      onTap: () => _select(item),
                    ),
                  )
                  .toList(),
            ),
          ),
        ],
      ],
    );
  }
}
