package cmd

import gemara "github.com/gemaraproj/go-gemara"

// knownGroups defines the canonical group definitions shared across all core
// catalog types (controls, threats, capabilities). When a catalog entry
// references one of these group IDs, the full definition is automatically
// injected into the catalog's Groups during generation.
var knownGroups = map[string]gemara.Group{
	"CCC.Core.Encryption": {
		Id:    "CCC.Core.Encryption",
		Title: "Encryption",
		Description: "The Encryption group covers controls, threats, and capabilities related to\n" +
			"protecting data confidentiality and integrity through cryptographic mechanisms.\n" +
			"This includes encryption in transit and at rest, key management, and certificate\n" +
			"lifecycle management.\n",
	},
	"CCC.Core.Access": {
		Id:    "CCC.Core.Access",
		Title: "Access Control",
		Description: "The Access Control group covers controls, threats, and capabilities related to\n" +
			"authentication, authorization, and trust perimeter enforcement. This includes\n" +
			"multi-factor authentication, least privilege access, network access rules, and\n" +
			"prevention of unauthorized access or reconnaissance.\n",
	},
	"CCC.Core.Observability": {
		Id:    "CCC.Core.Observability",
		Title: "Observability",
		Description: "The Observability group covers controls, threats, and capabilities related to\n" +
			"logging, monitoring, metrics, alerting, and event publication. This includes\n" +
			"audit trail integrity, enumeration detection, and protection against tampering\n" +
			"or unauthorized access to operational telemetry.\n",
	},
	"CCC.Core.Data": {
		Id:    "CCC.Core.Data",
		Title: "Data Resilience",
		Description: "The Data Resilience group covers controls, threats, and capabilities related to\n" +
			"ensuring data availability, integrity, and sovereignty across its lifecycle.\n" +
			"This includes replication, backup, recovery, region restrictions, and protection\n" +
			"against data loss or corruption.\n",
	},
	"CCC.Core.Resource": {
		Id:    "CCC.Core.Resource",
		Title: "Resource Management",
		Description: "The Resource Management group covers threats and capabilities related to\n" +
			"the lifecycle, configuration, and operational integrity of cloud resources.\n" +
			"This includes resource exhaustion, tag manipulation, version rollback,\n" +
			"scaling, and cost management.\n",
	},
	"CCC.Core.Compute": {
		Id:    "CCC.Core.Compute",
		Title: "Compute",
		Description: "The Compute group covers capabilities related to processing, execution,\n" +
			"and runtime infrastructure. This includes CPU, memory, storage allocation,\n" +
			"network ports, command-line interfaces, and elastic scaling.\n",
	},
	"CCC.Core.Ingestion": {
		Id:    "CCC.Core.Ingestion",
		Title: "Ingestion",
		Description: "The Ingestion group covers capabilities related to how a service receives\n" +
			"or retrieves data, inputs, or commands for processing. This includes both\n" +
			"active (pull-based) and passive (push-based) ingestion patterns.\n",
	},
}

// injectGroups adds known group definitions to a catalog's group list for any
// group IDs that are referenced by entries but not already present.
func injectGroups(groups *[]gemara.Group, referencedGroupIDs []string) {
	existing := map[string]bool{}
	for _, g := range *groups {
		existing[g.Id] = true
	}
	for _, id := range referencedGroupIDs {
		if id == "" || existing[id] {
			continue
		}
		if g, ok := knownGroups[id]; ok {
			*groups = append(*groups, g)
			existing[id] = true
		}
	}
}
