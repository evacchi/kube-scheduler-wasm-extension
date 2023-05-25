//go:build _tinygo.wasm

package imports

//go:wasm-module k8s.io/scheduler
//go:export status_reason
func _statusReason(ptr, size uint32)

//go:wasm-module k8s.io/api
//go:export nodeInfo/node
func _nodeInfoNode(ptr uint32, limit bufLimit) (len uint32)

//go:wasm-module k8s.io/api
//go:export pod/spec
func _podSpec(ptr uint32, limit bufLimit) (len uint32)
