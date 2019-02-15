package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var usage = fmt.Sprintf("Usage: %s TEMPLATE_FILE", filepath.Base(os.Args[0]))

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ERROR: Template file is required")
		fmt.Println(usage)
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(os.Args[1])
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
	fmt.Print(string(marshaledTemplate))
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
		case "Route":
			object = cleanRoute(object)
			newTemplateObjects = append(newTemplateObjects, object)
		case "Service":
			object = cleanService(object)
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
			newTemplateObjects = append(newTemplateObjects, object)
		}
	}

	template["objects"] = newTemplateObjects

	return template
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-buildconfig
func cleanBuildConfig(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfig = cleanTemplateObject(buildConfig)
	buildConfig = cleanBuildConfigSpec(buildConfig)

	return buildConfig
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-buildconfigspec
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

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-deploymentconfig
func cleanDeploymentConfig(deploymentConfig map[interface{}]interface{}) map[interface{}]interface{} {
	deploymentConfig = cleanTemplateObject(deploymentConfig)
	deploymentConfigSpec := deploymentConfig["spec"].(map[interface{}]interface{})
	deploymentConfigSpec = cleanDeploymentConfigSpec(deploymentConfigSpec)

	return deploymentConfig
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-deploymentconfigspec
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

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-deploymentstrategy
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

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-podtemplatespec
func cleanDeploymentTemplate(deploymentTemplate map[interface{}]interface{}) map[interface{}]interface{} {
	deploymentTemplate = cleanMetadata(deploymentTemplate)
	podSpec := deploymentTemplate["spec"].(map[interface{}]interface{})
	podSpec = cleanPodSpec(podSpec)

	return deploymentTemplate
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-podspec
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

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-container
func cleanContainer(container map[interface{}]interface{}) map[interface{}]interface{} {
	deleteKeyIfEmpty(container, "resources")
	deleteKeyIfValueMatches(container, "terminationMessagePath", "/dev/termination-log")
	deleteKeyIfValueMatches(container, "terminationMessagePolicy", "File")

	return container
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-deploymenttriggerpolicy
func cleanDeploymentTrigger(deploymentTrigger map[interface{}]interface{}) map[interface{}]interface{} {
	if val, ok := deploymentTrigger["imageChangeParams"]; ok {
		imageChangeParams := val.(map[interface{}]interface{})
		delete(imageChangeParams, "lastTriggeredImage")
	}

	return deploymentTrigger
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-imagestream
func cleanImageStream(imageStream map[interface{}]interface{}) map[interface{}]interface{} {
	imageStream = cleanTemplateObject(imageStream)

	imageStreamSpec := imageStream["spec"].(map[interface{}]interface{})

	if val, ok := imageStreamSpec["lookupPolicy"]; ok {
		imageLookupPolicy := val.(map[interface{}]interface{})
		deleteKeyIfValueMatches(imageLookupPolicy, "local", false)
	}
	deleteKeyIfEmpty(imageStreamSpec, "lookupPolicy")

	if val, ok := imageStreamSpec["tags"]; ok {
		tagReferences := val.([]interface{})

		for _, tagReference := range tagReferences {
			tagReference := tagReference.(map[interface{}]interface{})
			deleteKeyIfValueMatches(tagReference, "annotations", nil)
			deleteKeyIfValueMatches(tagReference, "generation", nil)
			deleteKeyIfEmpty(tagReference, "importPolicy")

			if val, ok := tagReference["referencePolicy"]; ok {
				tagReferencePolicy := val.(map[interface{}]interface{})
				deleteKeyIfValueMatches(tagReferencePolicy, "type", "")
				deleteKeyIfEmpty(tagReference, "referencePolicy")
			}
		}
	}

	return imageStream
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-route
func cleanRoute(route map[interface{}]interface{}) map[interface{}]interface{} {
	route = cleanTemplateObject(route)

	routeSpec := route["spec"].(map[interface{}]interface{})
	routeSpecTo := routeSpec["to"].(map[interface{}]interface{})
	deleteKeyIfValueMatches(routeSpecTo, "weight", 100)
	deleteKeyIfValueMatches(routeSpec, "wildcardPolicy", "None")

	return route
}

// https://kubernetes.io/docs/reference/federation/v1/definitions/#_v1_service
func cleanService(service map[interface{}]interface{}) map[interface{}]interface{} {
	service = cleanTemplateObject(service)

	serviceSpec := service["spec"].(map[interface{}]interface{})
	deleteKeyIfValueMatches(serviceSpec, "sessionAffinity", "None")
	deleteKeyIfValueMatches(serviceSpec, "type", "ClusterIP")

	return service
}

func cleanTemplateObject(templateObject map[interface{}]interface{}) map[interface{}]interface{} {
	templateObject = cleanMetadata(templateObject)

	// Status properties across different objects are populated by the server
	delete(templateObject, "status")

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
			} else if annotation == "openshift.io/host.generated" {
				delete(annotations, "openshift.io/host.generated")
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
