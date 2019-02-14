package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func main() {
	// TODO
	contents, err := ioutil.ReadFile(filepath.Join("testdata", "nodejs.yml"))
	if err != nil {
		panic(err)
	}

	template := make(map[interface{}]interface{})
	err = yaml.Unmarshal(contents, &template)
	if err != nil {
		panic(err)
	}

	template = cleanTemplate(template)

	marshaledTemplate, err := yaml.Marshal(&template)
	fmt.Printf("--- template dump:\n%s\n", string(marshaledTemplate))
}

// https://github.com/openshift/origin/blob/master/pkg/template/apis/template/types.go
func cleanTemplate(template map[interface{}]interface{}) map[interface{}]interface{} {
	template = cleanMetadata(template)
	template = cleanTemplateObjects(template)

	return template
}

func cleanTemplateObjects(template map[interface{}]interface{}) map[interface{}]interface{} {
	var newTemplateObjects []interface{}

	for _, object := range template["objects"].([]interface{}) {
		object := object.(map[interface{}]interface{})

		switch object["kind"] {
		case "BuildConfig":
			object = cleanBuildConfig(object)
			newTemplateObjects = append(newTemplateObjects, object)
		case "DeploymentConfig":
			object = cleanDeploymentConfig(object)
			newTemplateObjects = append(newTemplateObjects, object)
		case "ImageStream":
			object = cleanImageStream(object)
			newTemplateObjects = append(newTemplateObjects, object)

		// TODO: handle Route, Service, ...

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

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
func cleanBuildConfig(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfig = cleanTemplateObject(buildConfig)
	buildConfig = cleanBuildConfigSpec(buildConfig)

	delete(buildConfig, "status")

	return buildConfig
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
func cleanBuildConfigSpec(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfigSpec := buildConfig["spec"].(map[interface{}]interface{})

	deleteKeyIfValueMatches(buildConfigSpec, "failedBuildsHistoryLimit", 5)
	// This is technically required but the server will create it
	deleteKeyIfValueMatches(buildConfigSpec, "nodeSelector", nil)

	deleteKeyIfEmpty(buildConfigSpec, "postCommit")
	deleteKeyIfEmpty(buildConfigSpec, "resources")

	deleteKeyIfValueMatches(buildConfigSpec, "runPolicy", "Serial")
	deleteKeyIfValueMatches(buildConfigSpec, "successfulBuildsHistoryLimit", 5)

	/*
	 * TODO: Generic/Github triggers cannot be specified without secrets
	 * - Templatize with randomly generated secret?
	 * - Replace with SecretReference?
	 */
	// if val, ok := buildConfigSpec["triggers"]; ok {
	// 	buildConfigSpecTriggers := val.([]interface{})

	// 	for _, trigger := range buildConfigSpecTriggers {
	// 		trigger := trigger.(map[interface{}]interface{})

	// 		if val, ok := trigger["generic"]; ok {
	// 			delete(val.(map[interface{}]interface{}), "secret")
	// 		} else if val, ok := trigger["github"]; ok {
	// 			delete(val.(map[interface{}]interface{}), "secret")
	// 		}
	// 	}
	// }

	return buildConfig
}

func cleanDeploymentConfig(deploymentConfig map[interface{}]interface{}) map[interface{}]interface{} {
	deploymentConfig = cleanTemplateObject(deploymentConfig)
	deploymentConfigSpec := deploymentConfig["spec"].(map[interface{}]interface{})
	deploymentConfigSpec = cleanDeploymentConfigSpec(deploymentConfigSpec)

	// TODO: Could we just delete the status from all template objects?
	delete(deploymentConfig, "status")

	return deploymentConfig
}

func cleanDeploymentConfigSpec(deploymentConfigSpec map[interface{}]interface{}) map[interface{}]interface{} {
	deleteKeyIfValueMatches(deploymentConfigSpec, "revisionHistoryLimit", 10)

	deploymentStrategy := deploymentConfigSpec["strategy"].(map[interface{}]interface{})
	deploymentStrategy = cleanDeploymentStrategy(deploymentStrategy)

	deploymentTemplate := deploymentConfigSpec["template"].(map[interface{}]interface{})
	deploymentTemplate = cleanDeploymentTemplate(deploymentTemplate)

	deleteKeyIfValueMatches(deploymentConfigSpec, "test", false)

	deploymentTriggers := deploymentConfigSpec["triggers"].([]interface{})
	for _, deploymentTrigger := range deploymentTriggers {
		deploymentTrigger := deploymentTrigger.(map[interface{}]interface{})
		deploymentTrigger = cleanDeploymentTrigger(deploymentTrigger)
	}

	return deploymentConfigSpec
}

func cleanDeploymentStrategy(deploymentStrategy map[interface{}]interface{}) map[interface{}]interface{} {
	deleteKeyIfValueMatches(deploymentStrategy, "activeDeadlineSeconds", 21600)
	deleteKeyIfEmpty(deploymentStrategy, "resources")

	if val, ok := deploymentStrategy["rollingParams"]; ok {
		rollingParams := val.(map[interface{}]interface{})
		deleteKeyIfValueMatches(rollingParams, "intervalSeconds", 1)
		deleteKeyIfValueMatches(rollingParams, "maxSurge", "25%")
		deleteKeyIfValueMatches(rollingParams, "maxUnavailable", "25%")
		deleteKeyIfValueMatches(rollingParams, "timeoutSeconds", 600)
		deleteKeyIfValueMatches(rollingParams, "updatePeriodSeconds", 1)
	}
	deleteKeyIfEmpty(deploymentStrategy, "rollingParams")

	return deploymentStrategy
}

func cleanDeploymentTemplate(deploymentTemplate map[interface{}]interface{}) map[interface{}]interface{} {
	deploymentTemplate = cleanMetadata(deploymentTemplate)
	podSpec := deploymentTemplate["spec"].(map[interface{}]interface{})
	podSpec = cleanPodSpec(podSpec)

	return deploymentTemplate
}

func cleanPodSpec(podSpec map[interface{}]interface{}) map[interface{}]interface{} {
	containers := podSpec["containers"].([]interface{})
	for _, container := range containers {
		container := container.(map[interface{}]interface{})
		container = cleanContainer(container)
	}

	deleteKeyIfValueMatches(podSpec, "dnsPolicy", "ClusterFirst")
	deleteKeyIfValueMatches(podSpec, "restartPolicy", "Always")
	deleteKeyIfValueMatches(podSpec, "schedulerName", "default-scheduler")
	deleteKeyIfEmpty(podSpec, "securityContext")
	deleteKeyIfValueMatches(podSpec, "terminationGracePeriodSeconds", 30)

	return podSpec
}

func cleanContainer(container map[interface{}]interface{}) map[interface{}]interface{} {
	deleteKeyIfEmpty(container, "resources")
	deleteKeyIfValueMatches(container, "terminationMessagePath", "/dev/termination-log")
	deleteKeyIfValueMatches(container, "terminationMessagePolicy", "File")

	return container
}

func cleanDeploymentTrigger(deploymentTrigger map[interface{}]interface{}) map[interface{}]interface{} {
	if val, ok := deploymentTrigger["imageChangeParams"]; ok {
		imageChangeParams := val.(map[interface{}]interface{})
		delete(imageChangeParams, "lastTriggeredImage")
	}

	return deploymentTrigger
}

// https://github.com/openshift/origin/blob/master/pkg/image/apis/image/types.go
func cleanImageStream(imageStream map[interface{}]interface{}) map[interface{}]interface{} {
	imageStream = cleanTemplateObject(imageStream)

	imageStreamSpec := imageStream["spec"].(map[interface{}]interface{})

	if val, ok := imageStreamSpec["lookupPolicy"]; ok {
		imageLookupPolicy := val.(map[interface{}]interface{})
		deleteKeyIfValueMatches(imageLookupPolicy, "local", false)
	}
	deleteKeyIfEmpty(imageStreamSpec, "lookupPolicy")

	// Tags seem to be ignored and recreated by the server
	delete(imageStreamSpec, "tags")
	deleteKeyIfEmpty(imageStream, "spec")
	delete(imageStream, "status")

	return imageStream
}

func cleanTemplateObject(templateObject map[interface{}]interface{}) map[interface{}]interface{} {
	templateObject = cleanMetadata(templateObject)

	return templateObject
}

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/types.go
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
	// "Populated by the system. Read-only."
	delete(metadata, "generation")

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
