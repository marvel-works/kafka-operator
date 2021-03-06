// Copyright © 2019 Banzai Cloud
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

package util

import (
	"reflect"
	"testing"

	"github.com/banzaicloud/kafka-operator/api/v1beta1"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// we expect the final result will be the two segment's union
func TestGetBrokerConfigAffinityMergeBrokerNodeAffinityWithGroupsAntiAffinity(t *testing.T) {
	expected := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchFields: []corev1.NodeSelectorRequirement{
							{Key: "fruit", Operator: "in", Values: []string{"apple"}},
						},
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_broker"}},
					Namespaces:    nil,
					TopologyKey:   "kubernetes.io/hostname",
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
	}

	cluster := v1beta1.KafkaClusterSpec{
		BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
			"default": {
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: nil,
									MatchFields: []corev1.NodeSelectorRequirement{
										{
											Key:      "fruit",
											Operator: "in",
											Values:   []string{"apple"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	broker := v1beta1.Broker{
		Id:                0,
		BrokerConfigGroup: "default",
		BrokerConfig: &v1beta1.BrokerConfig{
			Affinity: &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_broker"},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		},
	}

	config, err := GetBrokerConfig(broker, cluster)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, expected, config.Affinity)
}

// we expect the final result will be the two segment's union
func TestGetBrokerConfigAffinityMergeEmptyBrokerConfigWithDefaultConfig(t *testing.T) {
	expected := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchFields: []corev1.NodeSelectorRequirement{
							{Key: "fruit", Operator: "in", Values: []string{"apple"}},
						},
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_config_group"}},
					Namespaces:    nil,
					TopologyKey:   "kubernetes.io/hostname",
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
	}

	cluster := v1beta1.KafkaClusterSpec{
		BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
			"default": {
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_config_group"},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: nil,
									MatchFields: []corev1.NodeSelectorRequirement{
										{
											Key:      "fruit",
											Operator: "in",
											Values:   []string{"apple"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	broker := v1beta1.Broker{
		Id:                0,
		BrokerConfigGroup: "default",
	}

	config, err := GetBrokerConfig(broker, cluster)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, expected, config.Affinity)
}

// the equal values got duplicated
func TestGetBrokerConfigAffinityMergeEqualPodAntiAffinity(t *testing.T) {
	expected := &corev1.Affinity{
		NodeAffinity: nil,
		PodAffinity:  nil,
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_broker"},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: nil,
		},
	}

	broker := v1beta1.Broker{
		Id:                0,
		BrokerConfigGroup: "default",
		BrokerConfig: &v1beta1.BrokerConfig{
			Affinity: &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_broker"},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		},
	}

	cluster := v1beta1.KafkaClusterSpec{
		BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
			"default": {
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "kafka", "kafka_cr": "kafka_config_group"},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
		},
	}

	config, err := GetBrokerConfig(broker, cluster)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, expected, config.Affinity)
}

func TestGetBrokerConfig(t *testing.T) {
	expected := &v1beta1.BrokerConfig{
		StorageConfigs: []v1beta1.StorageConfig{
			{
				MountPath: "kafka-test/log",
				PvcSpec: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
					},
				},
			},
			{
				MountPath: "kafka-test1/log",
				PvcSpec: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
					},
				},
			},
		},
	}
	broker := v1beta1.Broker{
		Id:                0,
		BrokerConfigGroup: "default",
		BrokerConfig: &v1beta1.BrokerConfig{
			StorageConfigs: []v1beta1.StorageConfig{
				{
					MountPath: "kafka-test/log",
					PvcSpec: &corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
						},
					},
				},
			},
		},
	}

	cluster := v1beta1.KafkaClusterSpec{
		BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
			"default": {
				StorageConfigs: []v1beta1.StorageConfig{
					{
						MountPath: "kafka-test1/log",
						PvcSpec: &corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
							},
						},
					},
				},
			},
		},
	}

	result, err := GetBrokerConfig(broker, cluster)
	if err != nil {
		t.Error("Error GetBrokerConfig throw an unexpected error")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Error("Expected:", expected, "Got:", result)
	}
}

func TestGetBrokerConfigUniqueStorage(t *testing.T) {
	expected := &v1beta1.BrokerConfig{
		StorageConfigs: []v1beta1.StorageConfig{
			{
				MountPath: "kafka-test/log",
				PvcSpec: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("20Gi")},
					},
				},
			},
			{
				MountPath: "kafka-test1/log",
				PvcSpec: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
					},
				},
			},
		},
	}
	broker := v1beta1.Broker{
		Id:                0,
		BrokerConfigGroup: "default",
		BrokerConfig: &v1beta1.BrokerConfig{
			StorageConfigs: []v1beta1.StorageConfig{
				{
					MountPath: "kafka-test/log",
					PvcSpec: &corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("20Gi")},
						},
					},
				},
			},
		},
	}

	cluster := v1beta1.KafkaClusterSpec{
		BrokerConfigGroups: map[string]v1beta1.BrokerConfig{
			"default": {
				StorageConfigs: []v1beta1.StorageConfig{
					{
						MountPath: "kafka-test1/log",
						PvcSpec: &corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
							},
						},
					},
					{
						MountPath: "kafka-test/log",
						PvcSpec: &corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
							},
						},
					},
				},
			},
		},
	}

	result, err := GetBrokerConfig(broker, cluster)
	if err != nil {
		t.Error("Error GetBrokerConfig throw an unexpected error")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Error("Expected:", expected, "Got:", result)
	}
}

func TestParsePropertiesFormat(t *testing.T) {
	testProp := `
broker.id=1
advertised.listener=broker-1:29092
empty.config=
 trailing.whitespace=foobar
`
	props := ParsePropertiesFormat(testProp)

	if props["broker.id"] != "1" ||
		props["advertised.listener"] != "broker-1:29092" ||
		props["empty.config"] != "" ||
		props["trailing.whitespace"] != "foobar" {
		t.Error("Error properties not loaded correctly")
	}
}

func TestIntstrPointer(t *testing.T) {
	i := int(10)
	ptr := IntstrPointer(i)
	if ptr.String() != "10" {
		t.Error("Expected 10, got:", ptr.String())
	}
}

func TestInt64Pointer(t *testing.T) {
	i := int64(10)
	ptr := Int64Pointer(i)
	if *ptr != int64(10) {
		t.Error("Expected pointer to 10, got:", *ptr)
	}
}

func TestInt32Pointer(t *testing.T) {
	i := int32(10)
	ptr := Int32Pointer(i)
	if *ptr != int32(10) {
		t.Error("Expected pointer to 10, got:", *ptr)
	}
}

func TestBoolPointer(t *testing.T) {
	b := true
	ptr := BoolPointer(b)
	if !*ptr {
		t.Error("Expected ptr to true, got false")
	}

	b = false
	ptr = BoolPointer(b)
	if *ptr {
		t.Error("Expected ptr to false, got true")
	}
}

func TestStringPointer(t *testing.T) {
	str := "test"
	ptr := StringPointer(str)
	if *ptr != "test" {
		t.Error("Expected ptr to 'test', got:", *ptr)
	}
}

func TestMapStringStringPointer(t *testing.T) {
	m := map[string]string{
		"test-key": "test-value",
	}
	ptr := MapStringStringPointer(m)
	for k, v := range ptr {
		if k != "test-key" {
			t.Error("Expected ptr map key 'test-key', got:", k)
		}
		if *v != "test-value" {
			t.Error("Expected ptr map value 'test-value', got:", *v)
		}
	}
}

func TestMergeLabels(t *testing.T) {
	m1 := map[string]string{"key1": "value1"}
	m2 := map[string]string{"key2": "value2"}
	expected := map[string]string{"key1": "value1", "key2": "value2"}
	merged := MergeLabels(m1, m2)
	if !reflect.DeepEqual(merged, expected) {
		t.Error("Expected:", expected, "Got:", merged)
	}

	m1 = nil
	expected = m2
	merged = MergeLabels(m1, m2)
	if !reflect.DeepEqual(merged, expected) {
		t.Error("Expected:", expected, "Got:", merged)
	}
}

func TestConvertStringToInt32(t *testing.T) {
	i := ConvertStringToInt32("10")
	if i != 10 {
		t.Error("Expected 10, got:", i)
	}
	i = ConvertStringToInt32("string")
	if i != -1 {
		t.Error("Expected -1 for non-convertable string, got:", i)
	}
}

func TestIsSSLEnabledForInternalCommunication(t *testing.T) {
	lconfig := []v1beta1.InternalListenerConfig{
		{
			CommonListenerSpec: v1beta1.CommonListenerSpec{Type: "ssl"},
		},
	}
	if !IsSSLEnabledForInternalCommunication(lconfig) {
		t.Error("Expected ssl enabled for internal communication, got disabled")
	}
	lconfig = []v1beta1.InternalListenerConfig{
		{
			CommonListenerSpec: v1beta1.CommonListenerSpec{Type: "plaintext"},
		},
	}
	if IsSSLEnabledForInternalCommunication(lconfig) {
		t.Error("Expected ssl disabled for internal communication, got enabled")
	}
}

func TestStringSliceContains(t *testing.T) {
	slice := []string{"1", "2", "3"}
	if !StringSliceContains(slice, "1") {
		t.Error("Expected slice contains 1, got false")
	}
	if StringSliceContains(slice, "4") {
		t.Error("Expected slice not contains 4, got true")
	}
}

func TestStringSliceRemove(t *testing.T) {
	slice := []string{"1", "2", "3"}
	removed := StringSliceRemove(slice, "3")
	expected := []string{"1", "2"}
	if !reflect.DeepEqual(removed, expected) {
		t.Error("Expected:", expected, "got:", removed)
	}
}

func TestMergeAnnotations(t *testing.T) {
	annotations := map[string]string{"foo": "bar", "bar": "foo"}
	annotations2 := map[string]string{"thing": "1", "other_thing": "2"}

	combined := MergeAnnotations(annotations, annotations2)

	if len(combined) != 4 {
		t.Error("Annotations didn't combine correctly")
	}
}

func TestCreateLogger(t *testing.T) {
	logger := CreateLogger(false, false)
	if logger == nil {
		t.Fatal("created Logger instance should not be nil")
	}
	if logger.V(1).Enabled() {
		t.Error("debug level should not be enabled")
	}
	if !logger.V(0).Enabled() {
		t.Error("info level should be enabled")
	}

	logger = CreateLogger(true, true)
	if !logger.V(1).Enabled() {
		t.Error("debug level should be enabled")
	}
	if !logger.V(0).Enabled() {
		t.Error("info level should be enabled")
	}
}

func TestGetBrokerIdsFromStatusAndSpec(t *testing.T) {
	testCases := []struct {
		states         map[string]v1beta1.BrokerState
		brokers        []v1beta1.Broker
		expectedOutput []int
	}{
		// the states and the spec matches
		{
			map[string]v1beta1.BrokerState{
				"1": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
				"2": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
				"3": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
			},
			[]v1beta1.Broker{
				{
					Id: 1,
				},
				{
					Id: 2,
				},
				{
					Id: 3,
				},
			},
			[]int{1, 2, 3},
		},
		// broker has been deleted from spec
		{
			map[string]v1beta1.BrokerState{
				"1": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
				"2": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
				"3": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
			},
			[]v1beta1.Broker{
				{
					Id: 1,
				},
				{
					Id: 2,
				},
			},
			[]int{1, 2, 3},
		},
		// broker is added to spec
		{
			map[string]v1beta1.BrokerState{
				"1": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
				"2": {
					ConfigurationState: v1beta1.ConfigInSync,
				},
			},
			[]v1beta1.Broker{
				{
					Id: 1,
				},
				{
					Id: 2,
				},
				{
					Id: 3,
				},
			},
			[]int{1, 2, 3},
		},
	}

	logger := zap.New()
	for _, testCase := range testCases {
		brokerIds := GetBrokerIdsFromStatusAndSpec(testCase.states, testCase.brokers, logger)
		if len(brokerIds) != len(testCase.expectedOutput) {
			t.Errorf("size of the merged slice of broker ids mismatch - expected: %d, actual: %d", len(testCase.expectedOutput), len(brokerIds))
		}
		for i, brokerId := range brokerIds {
			if brokerId != testCase.expectedOutput[i] {
				t.Errorf("broker id is not the expected - index: %d, expected: %d, actual: %d", i, testCase.expectedOutput[i], brokerId)
			}
		}
	}
}
