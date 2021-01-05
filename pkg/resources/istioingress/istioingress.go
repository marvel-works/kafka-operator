// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istioingress

import (
	"github.com/banzaicloud/kafka-operator/api/v1beta1"
	"github.com/banzaicloud/kafka-operator/pkg/k8sutil"
	"github.com/banzaicloud/kafka-operator/pkg/resources"
	"github.com/banzaicloud/kafka-operator/pkg/util/istioingress"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	componentName          = "istioingress"
	gatewayNameTemplate    = "%s-%s-gateway"
	virtualServiceTemplate = "%s-%s-virtualservice"
)

// labelsForIstioIngress returns the labels for selecting the resources
// belonging to the given kafka CR name.
func labelsForIstioIngress(crName, eLName string) map[string]string {
	return map[string]string{"app": "istioingress", "eListenerName": eLName, "kafka_cr": crName}
}

// Reconciler implements the Component Reconciler
type Reconciler struct {
	resources.Reconciler
}

// New creates a new reconciler for IstioIngress
func New(client client.Client, cluster *v1beta1.KafkaCluster) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client:       client,
			KafkaCluster: cluster,
		},
	}
}

// Reconcile implements the reconcile logic for IstioIngress
func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.V(1).Info("Reconciling")
	if r.KafkaCluster.Spec.ListenersConfig.ExternalListeners != nil && r.KafkaCluster.Spec.GetIngressController() == istioingress.IngressControllerName {

		for _, eListener := range r.KafkaCluster.Spec.ListenersConfig.ExternalListeners {
			if eListener.GetAccessMethod() == corev1.ServiceTypeLoadBalancer {

				// Prefer specific external listener configuration but fall back to global one if none specified
				var istioIngressConfig v1beta1.IstioIngressConfig
				if eListener.Config != nil && eListener.Config.IstioIngressConfig != nil {
					istioIngressConfig = *eListener.Config.IstioIngressConfig
				} else {
					istioIngressConfig = r.KafkaCluster.Spec.IstioIngressConfig
				}

				for _, res := range []resources.ResourceWithLogAndExternalListenerConfigAndIstioIngressConfig{
					r.meshgateway,
					r.gateway,
					r.virtualService,
				} {
					o := res(log, eListener, istioIngressConfig)
					err := k8sutil.Reconcile(log, r.Client, o, r.KafkaCluster)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	log.V(1).Info("Reconciled")

	return nil
}
