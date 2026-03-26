import 'dart:math' as math;
import 'dart:io';

import 'package:http/http.dart' as http;

const _backendPort = 8080;
const _healthPath = '/api/v1/healthz';

List<String> defaultCandidates() {
  if (Platform.isAndroid) {
    return const [
      'http://10.0.2.2:8080',
      'http://127.0.0.1:8080',
      'http://localhost:8080',
    ];
  }

  return const [
    'http://127.0.0.1:8080',
    'http://localhost:8080',
  ];
}

Future<String?> discoverOnLocalNetwork() async {
  try {
    final subnets = await _localSubnets();
    for (final subnet in subnets) {
      final found = await _scanSubnet(subnet);
      if (found != null) {
        return found;
      }
    }
  } catch (_) {
    // Best-effort discovery only.
  }
  return null;
}

Future<String?> _scanSubnet(_Subnet subnet) async {
  const batchSize = 24;
  final hosts = _prioritizedHosts(subnet.localHostOctet);
  for (var index = 0; index < hosts.length; index += batchSize) {
    final batch = hosts.sublist(index, math.min(index + batchSize, hosts.length));
    final results = await Future.wait(
      batch.map((host) => _probe('http://${subnet.prefix}.$host:$_backendPort')),
    );
    for (final result in results) {
      if (result != null) {
        return result;
      }
    }
  }
  return null;
}

Future<String?> _probe(String baseUrl) async {
  final client = http.Client();
  try {
    final uri = Uri.parse('$baseUrl$_healthPath');
    final response = await client
        .get(uri, headers: const {'Accept': 'application/json'})
        .timeout(const Duration(milliseconds: 450));
    if (response.statusCode == 200 && response.body.contains('"status"')) {
      return baseUrl;
    }
  } catch (_) {
    // Ignore network misses while scanning the subnet.
  } finally {
    client.close();
  }
  return null;
}

Future<List<_Subnet>> _localSubnets() async {
  final interfaces = await NetworkInterface.list(
    includeLoopback: false,
    includeLinkLocal: false,
    type: InternetAddressType.IPv4,
  );

  final seen = <String>{};
  final subnets = <_Subnet>[];

  for (final iface in interfaces) {
    for (final address in iface.addresses) {
      final ip = address.address.trim();
      if (!_isPrivateIpv4(ip)) {
        continue;
      }

      final octets = ip.split('.');
      final prefix = '${octets[0]}.${octets[1]}.${octets[2]}';
      if (!seen.add(prefix)) {
        continue;
      }

      subnets.add(
        _Subnet(
          prefix: prefix,
          localHostOctet: int.tryParse(octets[3]) ?? 0,
        ),
      );
    }
  }

  return subnets;
}

List<int> _prioritizedHosts(int localHostOctet) {
  final seen = <int>{};
  final ordered = <int>[];

  void add(int value) {
    if (value < 1 || value > 254 || !seen.add(value)) {
      return;
    }
    ordered.add(value);
  }

  for (var delta = 1; delta <= 16; delta++) {
    add(localHostOctet - delta);
    add(localHostOctet + delta);
  }

  add(1);
  add(2);
  add(10);

  for (var host = 1; host <= 254; host++) {
    add(host);
  }

  return ordered;
}

bool _isPrivateIpv4(String value) {
  final parts = value.split('.');
  if (parts.length != 4) {
    return false;
  }

  final first = int.tryParse(parts[0]);
  final second = int.tryParse(parts[1]);
  if (first == null || second == null) {
    return false;
  }

  if (first == 10) {
    return true;
  }
  if (first == 172 && second >= 16 && second <= 31) {
    return true;
  }
  if (first == 192 && second == 168) {
    return true;
  }
  return false;
}

class _Subnet {
  final String prefix;
  final int localHostOctet;

  const _Subnet({
    required this.prefix,
    required this.localHostOctet,
  });
}
