package meta

import (
	cassdcapi "github.com/k8ssandra/cass-operator/apis/cassandra/v1beta1"
)

// Tags is a generic struct for objects that have labels and annotations. It implements
// annotations.Annotated and labels.Labeled.
// +kubebuilder:object:generate=true
type Tags struct {

	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ResourceMeta holds labels and annotations for objects created and managed by the K8ssandra,
// Reaper and Stargate controllers.
// +kubebuilder:object:generate=true
type ResourceMeta struct {

	// Specific labels and annotations for the custom resource itself (Reaper or Stargate).
	// +optional
	Tags `json:",inline"`

	// Specific labels and annotations for Service resources.
	// +optional
	Services Tags `json:"services,omitempty"`

	// Specific labels and annotations for Pod resources.
	// +optional
	Pods Tags `json:"pods,omitempty"`

	// CommonLabels allows to define additional labels that will be applied to ALL objects created
	// by k8ssandra-operator as part of the reconciliation of a Stargate or Reaper resource.
	// +optional
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
}

// CassandraDatacenterResourceMeta holds labels and annotations for Cassandra objects.
// +kubebuilder:object:generate=true
type CassandraDatacenterResourceMeta struct {

	// Additional labels and annotations for the CassandraDatacenter resource itself.
	// Note: this is inlined to stay backwards-compatible with the existing K8ssandra API.
	Tags `json:",inline"`

	// Meta tags for specific CassandraDatacenter Services.
	Services CassandraDatacenterServicesMeta `json:"services,omitempty"`

	// Meta tags for Pod resources.
	// +optional
	Pods Tags `json:"pods,omitempty"`

	// CommonLabels allows to define additional labels that will be applied to ALL objects
	// created by cass-operator as part of the reconciliation of a CassandraDatacenter resource.
	// These include: StatefulSets, Pods, PVs, PVCs, all Services, and a few others.
	CommonLabels map[string]string `json:"commonLabels,omitempty"`
}

// CassandraDatacenterServicesMeta is very similar to cassdcapi.ServiceConfig and is passed to
// cass-operator in the AdditionalServiceConfig field of the CassandraDatacenter spec.
// +kubebuilder:object:generate=true
type CassandraDatacenterServicesMeta struct {
	DatacenterService     Tags `json:"dcService,omitempty"`
	SeedService           Tags `json:"seedService,omitempty"`
	AllPodsService        Tags `json:"allPodsService,omitempty"`
	AdditionalSeedService Tags `json:"additionalSeedService,omitempty"`
	NodePortService       Tags `json:"nodePortService,omitempty"`
}

func (in *CassandraDatacenterServicesMeta) ToCassAdditionalServiceConfig() cassdcapi.ServiceConfig {
	return cassdcapi.ServiceConfig{
		DatacenterService: cassdcapi.ServiceConfigAdditions{
			Annotations: in.DatacenterService.Annotations,
			Labels:      in.DatacenterService.Labels,
		},
		SeedService: cassdcapi.ServiceConfigAdditions{
			Annotations: in.SeedService.Annotations,
			Labels:      in.SeedService.Labels,
		},
		AdditionalSeedService: cassdcapi.ServiceConfigAdditions{
			Annotations: in.AdditionalSeedService.Annotations,
			Labels:      in.AdditionalSeedService.Labels,
		},
		AllPodsService: cassdcapi.ServiceConfigAdditions{
			Annotations: in.AllPodsService.Annotations,
			Labels:      in.AllPodsService.Labels,
		},
		NodePortService: cassdcapi.ServiceConfigAdditions{
			Annotations: in.NodePortService.Annotations,
			Labels:      in.NodePortService.Labels,
		},
	}
}
