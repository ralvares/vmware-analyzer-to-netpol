# Network Policy Generator

This project provides a tool to generate Kubernetes NetworkPolicy YAMLs from a JSON file containing service connectivity data. It reads service definitions and translates them into NetworkPolicies suitable for Kubernetes environments.

## Features
- Generates NetworkPolicies based on JSON input.
- Supports specifying Kubernetes namespaces.
- Outputs NetworkPolicies in YAML format.

## Prerequisites
- Go programming language installed (1.16 or later).

### Build the Project
1. Clone this repository or create it locally.
2. Build the Go program:
   ```bash
   go build -o vmware-analyzer-to-netpol convert_nsx.go
   ```

### Run the Program
Run the program with a JSON input file and an optional namespace flag:
```bash
./vmware-analyzer-to-netpol -f json/Example2.json -n custom-namespace
```

- `-f`: Path to the JSON file containing service data.
- `-n`: (Optional) Namespace for the generated NetworkPolicies. Default is `default`.

## Example

### Input JSON File (Example2.json):
```json
{
  "services": [
    {
      "display_name": "Web Service",
      "service_entries": [
        {
          "display_name": "HTTP",
          "l4_protocol": "TCP",
          "destination_ports": ["80"]
        },
        {
          "display_name": "HTTPS",
          "l4_protocol": "TCP",
          "destination_ports": ["443"]
        }
      ]
    }
  ]
}
```

### Output YAML:
```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-service
  namespace: custom-namespace
spec:
  podSelector:
    matchLabels:
      app: web-service
  policyTypes:
  - Ingress
  ingress:
  - ports:
    - port: "80"
      protocol: TCP
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-service
  namespace: custom-namespace
spec:
  podSelector:
    matchLabels:
      app: web-service
  policyTypes:
  - Ingress
  ingress:
  - ports:
    - port: "443"
      protocol: TCP
```
