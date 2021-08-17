package validators

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//NetworkAttachmentDefinitions is responsible for finding network attachment definitions
type NetworkAttachmentDefinitions struct {
	Client client.Client
}

//Find retrieves network attachment definition with provided name and namespace
func (finder *NetworkAttachmentDefinitions) Find(name string, namespace string) (*netv1.NetworkAttachmentDefinition, error) {
	netAttachDef := &netv1.NetworkAttachmentDefinition{}
	err := finder.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, netAttachDef)
	return netAttachDef, err
}

func (finder *NetworkAttachmentDefinitions) Create() (*netv1.NetworkAttachmentDefinition, error) {
	netAttach := &netv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bridge-br10",
		},
		Spec: netv1.NetworkAttachmentDefinitionSpec{
			Config: createNetowrkAttachmentConfig("test", "bridge", "br-10"),
		},
	}
	err := finder.Client.Create(context.TODO(), netAttach)
	return netAttach, err
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