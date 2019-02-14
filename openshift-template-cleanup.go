package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// See here for more information on OpenShift/Kubernetes objects:
// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/types.go
// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/runtime/interfaces.go
// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
// https://github.com/openshift/origin/blob/master/pkg/template/apis/template/types.go

func main() {
	contents, err := ioutil.ReadFile(filepath.Join("testdata", "exported-openshift-template1.yml"))
	if err != nil {
		panic(err)
	}

	template := make(map[interface{}]interface{})
	err = yaml.Unmarshal(contents, &template)
	if err != nil {
		panic(err)
	}

	template = cleanTemplateObjects(template)

	marshaledTemplate, err := yaml.Marshal(&template)
	fmt.Printf("--- template dump:\n%s\n", string(marshaledTemplate))
}

func cleanTemplateObjects(template map[interface{}]interface{}) map[interface{}]interface{} {
	var newTemplateObjects []interface{}

	for _, object := range template["objects"].([]interface{}) {
		object := object.(map[interface{}]interface{})

		switch object["kind"] {
		// TODO: handle DeploymentConfig, ImageStream, Route, Service, ...

		case "BuildConfig":
			newTemplateObjects = append(newTemplateObjects, object)

		// Builds will be recreated by the BuildConfig
		case "Build":
			continue
		// Pods will be recreated by the DeploymentConfig
		case "Pod":
			continue
		case "ReplicationController":
			continue

		default:
			log.Println(fmt.Sprintf("WARNING: Unhandled object kind: %s", object["kind"]))
			continue
		}
	}

	template["objects"] = newTemplateObjects

	return template
}
