package networks

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	nmstate "github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstatev1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/types"
)

//NetworkCreator is responsible for finding network attachment definitions
type NetworkCreator struct {
	Client client.Client
}

//Create networks and network attachment definition
func (nm *NetworkCreator) Create() error {
	policy := &nmstatev1beta1.NodeNetworkConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "br-10",
		},
		Spec: nmstate.NodeNetworkConfigurationPolicySpec{
			DesiredState: linuxBrUp("br-10", "enp5s0"),
		},
	}
	err := nm.Client.Create(context.TODO(), policy)
	if err != nil {
		return err
	}
	podList := corev1.PodList{}
	filterHandlers := client.MatchingLabels{"component": "kubernetes-nmstate-handler"}
	err = nm.Client.List(context.TODO(), &podList, filterHandlers)
	if err != nil {
		return err
	}
	var allNodes []string
	for _, pod := range podList.Items {
		allNodes = append(allNodes, pod.Spec.NodeName)
	}
	var nodes []string
	nodeList := corev1.NodeList{}
	filterWorkers := client.MatchingLabels{"node-role.kubernetes.io/worker": ""}
	err = nm.Client.List(context.TODO(), &nodeList, filterWorkers)
	for _, node := range nodeList.Items {
		if containsNode(allNodes, node.Name) {
			nodes = append(nodes, node.Name)
		}
	}
	for _, nodeName := range nodes {
		node := corev1.Node{}
		if policy.Status.Conditions.Find(nmstate.NodeNetworkConfigurationPolicyConditionDegraded) != nil {
			return fmt.Errorf("Error while creating network")
		}

		err = nm.Client.Get(context.TODO(), types.NamespacedName{Namespace: nodeName}, &node)
		if err != nil {
			return err
		}
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				return nil
			}
		}
	}

	netAttachSet := &netv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bridge-br10",
		},
		Spec: netv1.NetworkAttachmentDefinitionSpec{
			Config: createNetowrkAttachmentConfig("test", "bridge", "br-10"),
		},
	}
	err = nm.Client.Create(context.TODO(), netAttachSet)
	if err != nil {
		return err
	}
	return nil
}

func createNetowrkAttachmentConfig(attName, attType, bridgeName string) string {
	return fmt.Sprintf(`{
	"cniVersion": "0.3.1",
	"name": %s,
	"type": %s,
	"bridge": %s,
	"vlan": 100,
	"ipam": {}
}`, attName, attType, bridgeName)
}

func linuxBrUp(bridgeName, firstSecondaryNic string) nmstate.State {
	return nmstate.NewState(fmt.Sprintf(`interfaces:
  - name: %s
    type: linux-bridge
    state: up
    bridge:
      port:
        - name: %s
`, bridgeName, firstSecondaryNic))
}

func containsNode(nodes []string, node string) bool {
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}