package main

import (
	"encoding/binary"
	"log"
	"os"
	"sigs.k8s.io/kube-scheduler-wasm-extension/guest/api"
	protoapi "sigs.k8s.io/kube-scheduler-wasm-extension/kubernetes/proto/api"
	"time"
)

func main() {
	if os.Args[1] == "w" {
		write()
	} else {
		read()
	}
}

func read() {
	start := time.Now()
	szB := make([]byte, 8)
	os.Stdin.Read(szB)
	sz := binary.LittleEndian.Uint64(szB)
	bytes := make([]byte, sz)
	os.Stdin.Read(bytes)
	node := &protoapi.IoK8SApiCoreV1Node{}
	node.UnmarshalVT(bytes)

	szB = make([]byte, 8)
	os.Stdin.Read(szB)
	sz = binary.LittleEndian.Uint64(szB)

	bytes = make([]byte, sz)
	os.Stdin.Read(bytes)
	podSpec := &protoapi.IoK8SApiCoreV1PodSpec{}
	podSpec.UnmarshalVT(bytes)

	elapsed := time.Since(start)
	log.Printf("Elapsed: %s", elapsed)
	log.Println(node.ApiVersion)
	log.Println(podSpec.NodeName)
}

func write() {
	start := time.Now()

	node := makeNode()
	bytes := marshall(node)
	sz := binary.LittleEndian.AppendUint64(nil, uint64(len(bytes)))
	os.Stdout.Write(sz)
	os.Stdout.Write(bytes)

	pod := makePodSpec()
	bytes = marshall(pod)
	sz = binary.LittleEndian.AppendUint64(nil, uint64(len(bytes)))
	os.Stdout.Write(sz)
	os.Stdout.Write(bytes)

	elapsed := time.Since(start)
	log.Printf("Elapsed: %s", elapsed)
}

func makeNode() *protoapi.IoK8SApiCoreV1Node {
	msg := &protoapi.IoK8SApiCoreV1Node{}
	msg.Kind = "Node"
	msg.ApiVersion = "v1"
	msg.Metadata = &protoapi.IoK8SApimachineryPkgApisMetaV1ObjectMeta{
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Annotations: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	msg.Spec = &protoapi.IoK8SApiCoreV1NodeSpec{
		ProviderID:    "wazero",
		Unschedulable: false,
		PodCIDR:       "10.244.0.0/24",
		Taints: []*protoapi.IoK8SApiCoreV1Taint{
			{
				Key:    "key1",
				Value:  "value1",
				Effect: "NoSchedule",
			},
			{
				Key:    "key2",
				Value:  "value2",
				Effect: "PreferNoSchedule",
			},
		},
	}
	return msg
}

func marshall(vt valueType) []byte {
	vLen := vt.SizeVT()
	if vLen == 0 {
		return nil // nothing to write
	}
	// Write directly to the wasm memory.
	wasmMem := make([]byte, vLen)
	if _, err := vt.MarshalToSizedBufferVT(wasmMem); err != nil {
		panic(err)
	}
	return wasmMem
}

func makePodSpec() *protoapi.IoK8SApiCoreV1PodSpec {
	return &protoapi.IoK8SApiCoreV1PodSpec{
		NodeName: "Node1",
		Volumes: []*protoapi.IoK8SApiCoreV1Volume{
			{
				Name: "volume1",
				PersistentVolumeClaim: &protoapi.IoK8SApiCoreV1PersistentVolumeClaimVolumeSource{
					ClaimName: "claim1",
				},
			},
			{
				Name:     "volume2",
				EmptyDir: &protoapi.IoK8SApiCoreV1EmptyDirVolumeSource{},
			},
		},
		Containers: []*protoapi.IoK8SApiCoreV1Container{
			{
				Name:  "container1",
				Image: "nginx:latest",
				Ports: []*protoapi.IoK8SApiCoreV1ContainerPort{
					{
						ContainerPort: 80,
					},
				},
				Env: []*protoapi.IoK8SApiCoreV1EnvVar{
					{
						Name:  "ENV_VAR1",
						Value: "value1",
					},
					{
						Name: "ENV_VAR2",
						ValueFrom: &protoapi.IoK8SApiCoreV1EnvVarSource{
							SecretKeyRef: &protoapi.IoK8SApiCoreV1SecretKeySelector{
								Name: "my-secret",
								Key:  "secret-key",
							},
						},
					},
				},
				Resources: &protoapi.IoK8SApiCoreV1ResourceRequirements{
					Limits: map[string]string{
						"cpu":    "1",
						"memory": "1Gi",
					},
					Requests: map[string]string{
						"cpu":    "500m",
						"memory": "512Mi",
					},
				},
			},
			{
				Name:  "container2",
				Image: "redis:latest",
				Ports: []*protoapi.IoK8SApiCoreV1ContainerPort{
					{
						ContainerPort: 6379,
					},
				},
				Env: []*protoapi.IoK8SApiCoreV1EnvVar{
					{
						Name:  "ENV_VAR3",
						Value: "value3",
					},
				},
				Resources: &protoapi.IoK8SApiCoreV1ResourceRequirements{
					Limits: map[string]string{
						"cpu":    "500m",
						"memory": "512Mi",
					},
					Requests: map[string]string{
						"cpu":    "250m",
						"memory": "256Mi",
					},
				},
			},
		},
		RestartPolicy: "Always",
		DnsPolicy:     "ClusterFirst",
		NodeSelector: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

}

type valueType interface {
	SizeVT() (n int)
	MarshalToSizedBufferVT(dAtA []byte) (int, error)
	UnmarshalVT(dAtA []byte) error
}

// nameEqualsPodSpec schedules this node if its name equals its pod spec.
func nameEqualsPodSpec(nodeInfo api.NodeInfo, pod api.Pod) (api.StatusCode, string) {
	nodeName := nodeInfo.Node().Metadata.Name
	podSpecNodeName := pod.Spec().NodeName

	if len(podSpecNodeName) == 0 || podSpecNodeName == nodeName {
		return api.StatusCodeSuccess, ""
	} else {
		return api.StatusCodeUnschedulable, ""
	}
}
