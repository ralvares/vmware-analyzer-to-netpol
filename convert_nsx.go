package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// ServiceEntry represents a single service entry
type ServiceEntry struct {
	DisplayName      string   `json:"display_name"`
	L4Protocol       string   `json:"l4_protocol"`
	DestinationPorts []string `json:"destination_ports"`
	SourcePorts      []string `json:"source_ports"`
}

// Service represents a service with its entries
type Service struct {
	DisplayName    string         `json:"display_name"`
	ServiceEntries []ServiceEntry `json:"service_entries"`
}

// Root represents the root of the JSON structure
type Root struct {
	Services []Service `json:"services"`
}

// NetworkPolicy represents a Kubernetes NetworkPolicy
type NetworkPolicy struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		PodSelector struct {
			MatchLabels map[string]string `yaml:"matchLabels"`
		} `yaml:"podSelector"`
		PolicyTypes []string `yaml:"policyTypes"`
		Ingress     []struct {
			Ports []struct {
				Port     int    `yaml:"port"`
				Protocol string `yaml:"protocol"`
			} `yaml:"ports"`
		} `yaml:"ingress"`
		Egress []struct {
			Ports []struct {
				Port     int    `yaml:"port"`
				Protocol string `yaml:"protocol"`
			} `yaml:"ports"`
		} `yaml:"egress,omitempty"`
	} `yaml:"spec"`
}

func main() {
	// Command-line flags for the JSON file path and namespace
	jsonFile := flag.String("f", "", "Path to the JSON file containing service data")
	namespace := flag.String("n", "default", "Kubernetes namespace for the NetworkPolicy")
	flag.Parse()

	if *jsonFile == "" {
		log.Fatal("Usage: go run main.go -f <path_to_json_file> -n <namespace>")
	}

	// Read the JSON file
	data, err := ioutil.ReadFile(*jsonFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Parse the JSON data
	var root Root
	if err := json.Unmarshal(data, &root); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Generate NetworkPolicies
	for _, service := range root.Services {
		policy := NetworkPolicy{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		}

		// Sanitize display name to ensure it is a valid DNS-1123 label
		sanitizedName := sanitizeName(service.DisplayName)
		policy.Metadata.Name = sanitizedName
		policy.Metadata.Namespace = *namespace
		policy.Spec.PodSelector.MatchLabels = map[string]string{"app": sanitizedName}

		// Initialize ingress and egress sections
		var ingressRules []struct {
			Ports []struct {
				Port     int    `yaml:"port"`
				Protocol string `yaml:"protocol"`
			} `yaml:"ports"`
		}
		var egressRules []struct {
			Ports []struct {
				Port     int    `yaml:"port"`
				Protocol string `yaml:"protocol"`
			} `yaml:"ports"`
		}

		// Process service entries
		for _, entry := range service.ServiceEntries {
			// Create ingress rule if destination ports exist
			if len(entry.DestinationPorts) > 0 {
				ingress := struct {
					Ports []struct {
						Port     int    `yaml:"port"`
						Protocol string `yaml:"protocol"`
					} `yaml:"ports"`
				}{}
				for _, port := range entry.DestinationPorts {
					portInt, _ := strconv.Atoi(port)
					ingress.Ports = append(ingress.Ports, struct {
						Port     int    `yaml:"port"`
						Protocol string `yaml:"protocol"`
					}{Port: portInt, Protocol: entry.L4Protocol})
				}
				ingressRules = append(ingressRules, ingress)
			}

			// Create egress rule if source ports exist
			if len(entry.SourcePorts) > 0 {
				egress := struct {
					Ports []struct {
						Port     int    `yaml:"port"`
						Protocol string `yaml:"protocol"`
					} `yaml:"ports"`
				}{}
				for _, port := range entry.SourcePorts {
					portInt, _ := strconv.Atoi(port)
					egress.Ports = append(egress.Ports, struct {
						Port     int    `yaml:"port"`
						Protocol string `yaml:"protocol"`
					}{Port: portInt, Protocol: entry.L4Protocol})
				}
				egressRules = append(egressRules, egress)
			}
		}

		// Add rules to policy spec
		if len(ingressRules) > 0 {
			policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, "Ingress")
			policy.Spec.Ingress = ingressRules
		}
		if len(egressRules) > 0 {
			policy.Spec.PolicyTypes = append(policy.Spec.PolicyTypes, "Egress")
			policy.Spec.Egress = egressRules
		}

		// Convert to YAML and print
		yamlData, err := yaml.Marshal(&policy)
		if err != nil {
			log.Fatalf("Error marshaling to YAML: %v", err)
		}

		fmt.Printf("---\n%s\n", string(yamlData))
	}
}

// sanitizeName ensures a name complies with DNS-1123 naming conventions
func sanitizeName(name string) string {
	// Replace invalid characters with a hyphen
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	name = re.ReplaceAllString(name, "-")
	// Ensure it starts and ends with an alphanumeric character
	name = strings.Trim(name, "-")
	return name
}
