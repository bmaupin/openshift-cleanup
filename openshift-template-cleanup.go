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

	template = cleanMetadata(template)
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
			object = cleanBuildConfig(object)
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

func cleanBuildConfig(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfig = cleanTemplateObject(buildConfig)
	buildConfig = cleanBuildConfigSpec(buildConfig)

	// TODO: Verify that this is optional :)
	delete(buildConfig, "status")

	return buildConfig
}

func cleanBuildConfigSpec(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfigSpec := buildConfig["spec"].(map[interface{}]interface{})

	// TODO: Verify this defaults to 5
	deleteKeyIfValueMatches(buildConfigSpec, "failedBuildsHistoryLimit", 5)
	deleteKeyIfValueMatches(buildConfigSpec, "nodeSelector", nil)

	deleteKeyIfEmpty(buildConfigSpec, "postCommit")
	deleteKeyIfEmpty(buildConfigSpec, "resources")

	deleteKeyIfValueMatches(buildConfigSpec, "runPolicy", "Serial")
	// TODO: Verify this defaults to 5
	deleteKeyIfValueMatches(buildConfigSpec, "successfulBuildsHistoryLimit", 5)

	if val, ok := buildConfigSpec["triggers"]; ok {
		buildConfigSpecTriggers := val.([]interface{})

		for _, trigger := range buildConfigSpecTriggers {
			trigger := trigger.(map[interface{}]interface{})

			if val, ok := trigger["generic"]; ok {
				delete(val.(map[interface{}]interface{}), "secret")
			} else if val, ok := trigger["github"]; ok {
				delete(val.(map[interface{}]interface{}), "secret")
			}
		}
	}

	return buildConfig
}

func cleanTemplateObject(templateObject map[interface{}]interface{}) map[interface{}]interface{} {
	templateObject = cleanMetadata(templateObject)

	return templateObject
}

func cleanMetadata(openshiftObject map[interface{}]interface{}) map[interface{}]interface{} {
	metadata := openshiftObject["metadata"].(map[interface{}]interface{})

	if val, ok := metadata["annotations"]; ok {
		annotations := val.(map[interface{}]interface{})

		for annotation, _ := range annotations {
			if annotation == "openshift.io/generated-by" {
				delete(annotations, "openshift.io/generated-by")
			}
		}

		deleteKeyIfEmpty(metadata, "annotations")
	}

	delete(metadata, "creationTimestamp")

	return openshiftObject
}

func deleteKeyIfEmpty(mapObject map[interface{}]interface{}, keyToMatch string) map[interface{}]interface{} {
	if val, ok := mapObject[keyToMatch]; ok {
		if len(val.(map[interface{}]interface{})) == 0 {
			delete(mapObject, keyToMatch)
		}
	}

	return mapObject
}

func deleteKeyIfValueMatches(mapObject map[interface{}]interface{}, keyToMatch string, valueToMatch interface{}) map[interface{}]interface{} {
	if val, ok := mapObject[keyToMatch]; ok {
		if val == valueToMatch {
			delete(mapObject, keyToMatch)
		}
	}

	return mapObject
}
