package cmd

import gemara "github.com/gemaraproj/go-gemara"

// knownGroups defines the canonical group definitions shared across all core
// catalog types (controls, threats, capabilities). When a catalog entry
// references one of these group IDs, the full definition is automatically
// injected into the catalog's Groups during generation.
var knownGroups = map[string]gemara.Group{
	"Encryption": {
		Id:    "Encryption",
		Title: "Encryption",
		Description: "The Encryption group covers entries related to\n" +
			"protecting data confidentiality and integrity through cryptographic mechanisms.\n" +
			"This includes encryption in transit and at rest, key management, and certificate\n" +
			"lifecycle management.\n",
	},
	"Access": {
		Id:    "Access",
		Title: "Access Control",
		Description: "The Access Control group covers entries related to\n" +
			"authentication, authorization, and trust perimeter enforcement. This includes\n" +
			"multi-factor authentication, least privilege access, network access rules, and\n" +
			"prevention of unauthorized access or reconnaissance.\n",
	},
	"Observability": {
		Id:    "Observability",
		Title: "Observability",
		Description: "The Observability group covers entries related to\n" +
			"logging, monitoring, metrics, alerting, and event publication. This includes\n" +
			"audit trail integrity, enumeration detection, and protection against tampering\n" +
			"or unauthorized access to operational telemetry.\n",
	},
	"Data": {
		Id:    "Data",
		Title: "Data Resilience",
		Description: "The Data Resilience group covers entries related to\n" +
			"ensuring data availability, integrity, and sovereignty across its lifecycle.\n" +
			"This includes replication, backup, recovery, region restrictions, and protection\n" +
			"against data loss or corruption.\n",
	},
	"Resource": {
		Id:    "Resource",
		Title: "Resource Management",
		Description: "The Resource Management group covers entries related to\n" +
			"the lifecycle, configuration, and operational integrity of cloud resources.\n" +
			"This includes resource exhaustion, tag manipulation, version rollback,\n" +
			"scaling, and cost management.\n",
	},
	"Compute": {
		Id:    "Compute",
		Title: "Compute",
		Description: "The Compute group covers entries related to processing, execution,\n" +
			"and runtime infrastructure. This includes CPU, memory, storage allocation,\n" +
			"network ports, command-line interfaces, and elastic scaling.\n",
	},
	"Ingestion": {
		Id:    "Ingestion",
		Title: "Ingestion",
		Description: "The Ingestion group covers entries related to how a service receives\n" +
			"or retrieves data, inputs, or commands for processing. This includes both\n" +
			"active (pull-based) and passive (push-based) ingestion patterns.\n",
	},
	"Networking": {
		Id:    "Networking",
		Title: "Networking",
		Description: "The Networking group covers entries related to network infrastructure,\n" +
			"connectivity, and traffic management. This includes virtual networks, subnets,\n" +
			"load balancing, DNS, routing, peering, and network-level access controls.\n",
	},
	"Orchestration": {
		Id:    "Orchestration",
		Title: "Orchestration",
		Description: "The Orchestration group covers entries related to coordinating and\n" +
			"managing workloads across distributed systems. This includes container\n" +
			"orchestration, job scheduling, CI/CD pipelines, build automation, and\n" +
			"service mesh management.\n",
	},
	"Processing": {
		Id:    "Processing",
		Title: "Data Processing",
		Description: "The Data Processing group covers entries related to transforming,\n" +
			"enriching, and moving data through pipelines. This includes ETL/ELT,\n" +
			"stream and batch processing, data lineage, schema evolution, and\n" +
			"workflow orchestration for data workloads.\n",
	},
	"Messaging": {
		Id:    "Messaging",
		Title: "Messaging",
		Description: "The Messaging group covers entries related to asynchronous\n" +
			"communication between services. This includes publish-subscribe,\n" +
			"message queuing, topic management, delivery guarantees, ordering,\n" +
			"filtering, and dead-letter handling.\n",
	},
	"MachineLearning": {
		Id:    "MachineLearning",
		Title: "Machine Learning",
		Description: "The Machine Learning group covers entries related to building,\n" +
			"training, deploying, and managing ML models and AI systems. This includes\n" +
			"development environments, experiment tracking, model registries, inference,\n" +
			"generative AI, prompt engineering, and model governance.\n",
	},
}

// tlpApplicabilityGroups defines the Traffic Light Protocol (TLP) v2.0
// applicability groups injected into every catalog's metadata. These describe
// the sharing boundaries applied to assessment requirements.
// Reference: https://www.first.org/tlp/
var tlpApplicabilityGroups = []gemara.Group{
	{
		Id:    "tlp_red",
		Title: "TLP:RED",
		Description: "For the eyes and ears of individual recipients only, no further disclosure. " +
			"Sources may use TLP:RED when information cannot be effectively acted upon without significant " +
			"risk for the privacy, reputation, or operations of the organizations involved.",
	},
	{
		Id:    "tlp_amber_strict",
		Title: "TLP:AMBER+STRICT",
		Description: "Limited disclosure, recipients can only spread this on a need-to-know basis " +
			"within their organization only. Sources may use TLP:AMBER+STRICT when information requires " +
			"support to be effectively acted upon, yet carries risk to privacy, reputation, or operations " +
			"if shared outside of the organization.",
	},
	{
		Id:    "tlp_amber",
		Title: "TLP:AMBER",
		Description: "Limited disclosure, recipients can spread this on a need-to-know basis within " +
			"their organization and to its clients. Sources may use TLP:AMBER when information requires " +
			"support to be effectively acted upon, yet carries risk to privacy, reputation, or operations " +
			"if shared outside of the organizations involved.",
	},
	{
		Id:    "tlp_green",
		Title: "TLP:GREEN",
		Description: "Limited disclosure, recipients can spread this within their community. " +
			"Sources may use TLP:GREEN when information is useful to increase awareness within their " +
			"wider community.",
	},
	{
		Id:    "tlp_clear",
		Title: "TLP:CLEAR",
		Description: "Recipients can spread this to the world, there is no limit on disclosure. " +
			"Sources may use TLP:CLEAR when information carries minimal or no foreseeable risk of misuse, " +
			"in accordance with applicable rules and procedures for public release.",
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
