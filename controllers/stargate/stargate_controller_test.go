package stargate

import (
	"context"
	telemetryapi "github.com/k8ssandra/k8ssandra-operator/apis/telemetry/v1alpha1"
	"github.com/k8ssandra/k8ssandra-operator/pkg/stargate"
	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"testing"
	"time"

	cassdcapi "github.com/k8ssandra/cass-operator/apis/cassandra/v1beta1"
	api "github.com/k8ssandra/k8ssandra-operator/apis/stargate/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/k8ssandra/k8ssandra-operator/pkg/config"
	testutils "github.com/k8ssandra/k8ssandra-operator/pkg/test"
)

const (
	timeout  = time.Second * 5
	interval = time.Millisecond * 500
)

func TestStargate(t *testing.T) {
	ctx := testutils.TestSetup(t)
	ctx, cancel := context.WithCancel(ctx)
	testEnv := &testutils.TestEnv{}
	err := testEnv.Start(ctx, t, func(mgr manager.Manager) error {
		err := (&StargateReconciler{
			ReconcilerConfig: config.InitConfig(),
			Client:           mgr.GetClient(),
			Scheme:           scheme.Scheme,
		}).SetupWithManager(mgr)
		return err
	})
	if err != nil {
		t.Fatalf("failed to start test environment: %s", err)
	}

	defer testEnv.Stop(t)
	defer cancel()

	t.Run("CreateStargateSingleRack", func(t *testing.T) {
		testCreateStargateSingleRack(t, testEnv.TestClient)
	})
	t.Run("CreateStargateMultiRack", func(t *testing.T) {
		testCreateStargateMultiRack(t, testEnv.TestClient)
	})
}

func testCreateStargateSingleRack(t *testing.T, testClient client.Client) {

	namespace := "default"
	ctx := context.Background()

	dc := &cassdcapi.CassandraDatacenter{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dc1",
		},
		Spec: cassdcapi.CassandraDatacenterSpec{
			Size:          1,
			ServerVersion: "3.11.10",
			ServerType:    "cassandra",
			ClusterName:   "test",
			StorageConfig: cassdcapi.StorageConfig{
				CassandraDataVolumeClaimSpec: &v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				},
			},
		},
	}

	err := testClient.Create(ctx, dc)
	require.NoError(t, err, "failed to create CassandraDatacenter")

	t.Log("check that the datacenter was created")
	dcKey := types.NamespacedName{Namespace: namespace, Name: "dc1"}

	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, dcKey, dc)
		return err == nil
	}, timeout, interval)

	sg := &api.Stargate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dc1-stargate",
		},
		Spec: api.StargateSpec{
			StargateDatacenterTemplate: api.StargateDatacenterTemplate{
				StargateClusterTemplate: api.StargateClusterTemplate{
					Size: 1,
					StargateTemplate: api.StargateTemplate{
						Telemetry: &telemetryapi.TelemetrySpec{
							Prometheus: &telemetryapi.PrometheusTelemetrySpec{
								Enabled: true,
							},
						},
					},
				},
			},
			DatacenterRef: corev1.LocalObjectReference{Name: "dc1"},
		},
	}

	// artificially put the DC in a ready state
	dc.SetCondition(cassdcapi.DatacenterCondition{
		Type:               cassdcapi.DatacenterReady,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	})
	dc.Status.CassandraOperatorProgress = cassdcapi.ProgressReady
	dc.Status.LastServerNodeStarted = metav1.Now()
	dc.Status.NodeStatuses = cassdcapi.CassandraStatusMap{"node1": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"}}
	dc.Status.NodeReplacements = []string{}
	dc.Status.LastRollingRestart = metav1.Now()
	dc.Status.QuietPeriod = metav1.Now()
	//goland:noinspection GoDeprecation
	dc.Status.SuperUserUpserted = metav1.Now()
	dc.Status.UsersUpserted = metav1.Now()

	err = testClient.Status().Update(ctx, dc)
	require.NoError(t, err, "failed to update dc")

	err = testClient.Create(ctx, sg)
	require.NoError(t, err, "failed to create Stargate")

	t.Log("check that the Stargate resource was created")
	stargateKey := types.NamespacedName{Namespace: namespace, Name: "dc1-stargate"}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, stargateKey, sg)
		return err == nil && sg.Status.Progress == api.StargateProgressDeploying
	}, timeout, interval)

	deploymentKey := types.NamespacedName{Namespace: namespace, Name: "test-dc1-default-stargate"}
	deployment := &appsv1.Deployment{}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, deploymentKey, deployment)
		return err == nil
	}, timeout, interval)

	t.Log("check that the owner reference is set on the Stargate deployment")
	assert.Len(t, deployment.OwnerReferences, 1, "expected to find 1 owner reference for Stargate deployment")
	assert.Equal(t, sg.UID, deployment.OwnerReferences[0].UID)

	assert.Equal(t, "docker.io/stargateio/stargate-3_11:v"+stargate.DefaultVersion, deployment.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, corev1.PullIfNotPresent, deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy)

	t.Log("check that authentication is enabled on the Stargate deployment")
	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	assert.Equal(t, "ENABLE_AUTH", envVars[len(envVars)-1].Name)
	assert.Equal(t, "true", envVars[len(envVars)-1].Value)

	deployment.Status.Replicas = 1
	deployment.Status.ReadyReplicas = 1
	deployment.Status.AvailableReplicas = 1
	deployment.Status.UpdatedReplicas = 1
	err = testClient.Status().Update(ctx, deployment)
	require.NoError(t, err, "failed to update deployment")

	serviceKey := types.NamespacedName{Namespace: namespace, Name: "test-dc1-stargate-service"}
	service := &corev1.Service{}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, serviceKey, service)
		return err == nil
	}, timeout, interval)

	t.Log("check that the Stargate resource is fully reconciled")
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, stargateKey, sg)
		return err == nil && sg.Status.Progress == api.StargateProgressRunning
	}, timeout, interval)

	t.Log("check Stargate status")
	assert.EqualValues(t, 1, sg.Status.Replicas, "expected to find 1 replica for Stargate")
	assert.EqualValues(t, 1, sg.Status.ReadyReplicas, "expected to find 1 ready replica for Stargate")
	assert.EqualValues(t, 1, sg.Status.AvailableReplicas, "expected to find 1 available replica for Stargate")
	assert.EqualValues(t, 1, sg.Status.UpdatedReplicas, "expected to find 1 updated replica for Stargate")
	assert.Equal(t, "1/1", *sg.Status.ReadyReplicasRatio)
	assert.Len(t, sg.Status.DeploymentRefs, 1)
	assert.NotNil(t, sg.Status.ServiceRef)

	t.Log("check Stargate condition")
	assert.Len(t, sg.Status.Conditions, 1, "expected to find 1 condition for Stargate")
	assert.Equal(t, api.StargateReady, sg.Status.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, sg.Status.Conditions[0].Status)

	smKey := types.NamespacedName{Name: sg.Name + "-" + "stargate-servicemonitor", Namespace: "default"}
	//	Check for presence of expected ServiceMonitor
	sm := &promapi.ServiceMonitor{}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, smKey, sm)
		return err == nil
	}, timeout, interval)
	assert.NotNil(t, sm.Spec.Endpoints)

	// Ensure that removing the telemetry spec does delete the ServiceMonitor
	sgPatch := client.MergeFrom(sg.DeepCopy())
	sg.Spec.Telemetry = nil
	if err := testClient.Patch(ctx, sg, sgPatch); err != nil {
		assert.Fail(t, "failed to patch stargate", "error", err)
	}
	assert.Eventually(t, func() bool {
		err := testClient.Get(ctx, smKey, sm)
		if err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, timeout, interval)
}

func testCreateStargateMultiRack(t *testing.T, testClient client.Client) {

	namespace := "default"
	ctx := context.Background()

	dc := &cassdcapi.CassandraDatacenter{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dc2",
		},
		Spec: cassdcapi.CassandraDatacenterSpec{
			Size: 9, // 3 nodes per rack
			Racks: []cassdcapi.Rack{
				{
					Name:               "rack1",
					NodeAffinityLabels: map[string]string{"topology.kubernetes.io/zone": "us-east-1a"},
				},
				{
					Name:               "rack2",
					NodeAffinityLabels: map[string]string{"topology.kubernetes.io/zone": "us-east-1b"},
				},
				{
					Name:               "rack3",
					NodeAffinityLabels: map[string]string{"topology.kubernetes.io/zone": "us-east-1c"},
				},
			},
			ServerVersion: "3.11.10",
			ServerType:    "cassandra",
			ClusterName:   "cluster1",
			StorageConfig: cassdcapi.StorageConfig{
				CassandraDataVolumeClaimSpec: &v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				},
			},
		},
	}

	err := testClient.Create(ctx, dc)
	require.NoError(t, err, "failed to create CassandraDatacenter")

	t.Log("check that the datacenter was created")
	dcKey := types.NamespacedName{Namespace: namespace, Name: "dc2"}

	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, dcKey, dc)
		return err == nil
	}, timeout, interval)

	// artificially put the DC in a ready state
	dc.SetCondition(cassdcapi.DatacenterCondition{
		Type:               cassdcapi.DatacenterReady,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
	})
	dc.Status.CassandraOperatorProgress = cassdcapi.ProgressReady
	dc.Status.LastServerNodeStarted = metav1.Now()
	dc.Status.NodeStatuses = cassdcapi.CassandraStatusMap{
		"node1": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node2": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node3": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node4": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node5": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node6": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node7": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node8": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
		"node9": cassdcapi.CassandraNodeStatus{HostID: "irrelevant"},
	}
	dc.Status.NodeReplacements = []string{}
	dc.Status.LastRollingRestart = metav1.Now()
	dc.Status.QuietPeriod = metav1.Now()
	//goland:noinspection GoDeprecation
	dc.Status.SuperUserUpserted = metav1.Now()
	dc.Status.UsersUpserted = metav1.Now()

	err = testClient.Status().Update(ctx, dc)
	require.NoError(t, err, "failed to update dc")

	stargate := &api.Stargate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dc2-stargate",
		},
		Spec: api.StargateSpec{
			StargateDatacenterTemplate: api.StargateDatacenterTemplate{
				StargateClusterTemplate: api.StargateClusterTemplate{
					Size: 3, // 1 node per rack
				},
			},
			DatacenterRef: corev1.LocalObjectReference{Name: "dc2"},
		},
	}

	err = testClient.Create(ctx, stargate)
	require.NoError(t, err, "failed to create Stargate")

	t.Log("check that the Stargate resource was created")
	stargateKey := types.NamespacedName{Namespace: namespace, Name: "dc2-stargate"}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, stargateKey, stargate)
		return err == nil && stargate.Status.Progress == api.StargateProgressDeploying
	}, timeout, interval)

	t.Log("check that the first Stargate deployment was created")
	deploymentList := &appsv1.DeploymentList{}
	require.Eventually(t, func() bool {
		err := testClient.List(
			ctx,
			deploymentList,
			client.InNamespace(namespace),
			client.MatchingLabels{api.StargateLabel: stargate.Name},
		)
		return err == nil && len(deploymentList.Items) == 1
	}, timeout, interval)

	deployment1 := deploymentList.Items[0]
	assert.Equal(t, "cluster1-dc2-rack1-stargate", deployment1.Name)
	assert.EqualValues(t, 1, *deployment1.Spec.Replicas)
	requirement1 := deployment1.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
	assert.Equal(t, "topology.kubernetes.io/zone", requirement1.Key)
	assert.Contains(t, requirement1.Values[0], "us-east-1a")

	t.Log("update the first Stargate deployment to ready")
	deployment1.Status.Replicas = 1
	deployment1.Status.ReadyReplicas = 1
	deployment1.Status.AvailableReplicas = 1
	deployment1.Status.UpdatedReplicas = 1
	err = testClient.Status().Update(ctx, &deployment1)
	require.NoError(t, err, "failed to update deployment1")

	t.Log("check that the other Stargate deployments were created")
	require.Eventually(t, func() bool {
		err := testClient.List(
			ctx,
			deploymentList,
			client.InNamespace(namespace),
			client.MatchingLabels{api.StargateLabel: stargate.Name},
		)
		return err == nil && len(deploymentList.Items) == 3
	}, timeout, interval)

	deployment1 = deploymentList.Items[0]
	assert.Equal(t, "cluster1-dc2-rack1-stargate", deployment1.Name)
	assert.EqualValues(t, 1, *deployment1.Spec.Replicas)
	requirement1 = deployment1.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
	assert.Equal(t, "topology.kubernetes.io/zone", requirement1.Key)
	assert.Contains(t, requirement1.Values[0], "us-east-1a")

	deployment2 := deploymentList.Items[1]
	assert.Equal(t, "cluster1-dc2-rack2-stargate", deployment2.Name)
	assert.EqualValues(t, 1, *deployment2.Spec.Replicas)
	requirement2 := deployment2.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
	assert.Equal(t, "topology.kubernetes.io/zone", requirement2.Key)
	assert.Contains(t, requirement2.Values[0], "us-east-1b")

	deployment3 := deploymentList.Items[2]
	assert.Equal(t, "cluster1-dc2-rack3-stargate", deployment3.Name)
	assert.EqualValues(t, 1, *deployment3.Spec.Replicas)
	requirement3 := deployment3.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
	assert.Equal(t, "topology.kubernetes.io/zone", requirement3.Key)
	assert.Contains(t, requirement3.Values[0], "us-east-1c")

	t.Log("update the other Stargate deployments to ready")
	deployment2.Status.Replicas = 1
	deployment2.Status.ReadyReplicas = 1
	deployment2.Status.AvailableReplicas = 1
	deployment2.Status.UpdatedReplicas = 1
	err = testClient.Status().Update(ctx, &deployment2)
	require.NoError(t, err, "failed to update deployment2")

	deployment3.Status.Replicas = 1
	deployment3.Status.ReadyReplicas = 1
	deployment3.Status.AvailableReplicas = 1
	deployment3.Status.UpdatedReplicas = 1
	err = testClient.Status().Update(ctx, &deployment3)
	require.NoError(t, err, "failed to update deployment3")

	t.Log("check that the Stargate service was created")
	serviceKey := types.NamespacedName{Namespace: namespace, Name: "cluster1-dc2-stargate-service"}
	service := &corev1.Service{}
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, serviceKey, service)
		return err == nil
	}, timeout, interval)

	t.Log("check that the Stargate resource is fully reconciled")
	require.Eventually(t, func() bool {
		err := testClient.Get(ctx, stargateKey, stargate)
		return err == nil && stargate.Status.Progress == api.StargateProgressRunning && stargate.Status.IsReady()
	}, timeout, interval)

	t.Log("check Stargate status")
	assert.EqualValues(t, 3, stargate.Status.Replicas, "expected to find 3 replicas for Stargate")
	assert.EqualValues(t, 3, stargate.Status.ReadyReplicas, "expected to find 3 ready replicas for Stargate")
	assert.EqualValues(t, 3, stargate.Status.AvailableReplicas, "expected to find 3 available replicas for Stargate")
	assert.EqualValues(t, 3, stargate.Status.UpdatedReplicas, "expected to find 3 updated replicas for Stargate")
	assert.Equal(t, "3/3", *stargate.Status.ReadyReplicasRatio)
	assert.Len(t, stargate.Status.DeploymentRefs, 3)
	assert.NotNil(t, stargate.Status.ServiceRef)

	t.Log("check Stargate condition")
	assert.Len(t, stargate.Status.Conditions, 1, "expected to find 1 condition for Stargate")
	assert.Equal(t, api.StargateReady, stargate.Status.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, stargate.Status.Conditions[0].Status)

}
