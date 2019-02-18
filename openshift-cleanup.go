package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var usage = fmt.Sprintf("Usage: %s OPENSHIFT_YAML_FILE", filepath.Base(os.Args[0]))

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ERROR: OpenShift YAML file is required")
		fmt.Println(usage)
		os.Exit(1)
	}

	fmt.Print(string(cleanOpenshiftConfigFile(os.Args[1])))
}

func cleanOpenshiftConfigFile(openshiftConfigFilePath string) []byte {
	contents, err := ioutil.ReadFile(openshiftConfigFilePath)
	if err != nil {
		panic(err)
	}

	unmarshaledConfig := make(map[interface{}]interface{})
	err = yaml.Unmarshal(contents, &unmarshaledConfig)
	if err != nil {
		panic(err)
	}

	unmarshaledConfig = cleanOpenshiftConfig(unmarshaledConfig)

	marshaledConfig, err := yaml.Marshal(&unmarshaledConfig)
	if err != nil {
		panic(err)
	}
	return marshaledConfig
}

func cleanOpenshiftConfig(openshiftConfig map[interface{}]interface{}) map[interface{}]interface{} {
	openshiftConfig = cleanMetadata(openshiftConfig)

	var listKey string
	var newChildObjects []interface{}

	switch openshiftConfig["kind"] {
	// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-listmeta
	case "List":
		listKey = "items"
	// https://github.com/openshift/origin/blob/master/pkg/template/apis/template/types.go
	case "Template":
		listKey = "objects"
	}

	for _, object := range openshiftConfig[listKey].([]interface{}) {
		object := object.(map[interface{}]interface{})

		switch object["kind"] {
		case "BuildConfig":
			object = cleanBuildConfig(object)
			newChildObjects = append(newChildObjects, object)
		case "DeploymentConfig":
			object = cleanDeploymentConfig(object)
			newChildObjects = append(newChildObjects, object)
		case "ImageStream":
			object = cleanImageStream(object)
			newChildObjects = append(newChildObjects, object)
		case "Route":
			object = cleanRoute(object)
			newChildObjects = append(newChildObjects, object)
		case "Service":
			object = cleanService(object)
			newChildObjects = append(newChildObjects, object)

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
			newChildObjects = append(newChildObjects, object)
		}
	}

	openshiftConfig[listKey] = newChildObjects

	return openshiftConfig
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-buildconfig
func cleanBuildConfig(buildConfig map[interface{}]interface{}) map[interface{}]interface{} {
	buildConfig = cleanOpenshiftObject(buildConfig)
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
	deploymentConfig = cleanOpenshiftObject(deploymentConfig)
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
	// ports is an optional parameter
	if val, ok := container["ports"]; ok {
		ports := val.([]interface{})
		for _, port := range ports {
			deleteKeyIfValueMatches(port.(map[interface{}]interface{}), "protocol", "TCP")
		}
	}

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
	imageStream = cleanOpenshiftObject(imageStream)

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

			if tagReference["annotations"] != nil {
				cleanAnnotations(tagReference["annotations"].(map[interface{}]interface{}))
				deleteKeyIfEmpty(tagReference, "annotations")
			}
			deleteKeyIfValueMatches(tagReference, "annotations", nil)
			delete(tagReference, "generation")
			deleteKeyIfEmpty(tagReference, "importPolicy")

			if val, ok := tagReference["referencePolicy"]; ok {
				tagReferencePolicy := val.(map[interface{}]interface{})
				deleteKeyIfValueMatches(tagReferencePolicy, "type", "")
				deleteKeyIfValueMatches(tagReferencePolicy, "type", "Source")
				deleteKeyIfEmpty(tagReference, "referencePolicy")
			}
		}
	}

	return imageStream
}

// https://docs.openshift.com/container-platform/3.6/rest_api/openshift_v1.html#v1-route
func cleanRoute(route map[interface{}]interface{}) map[interface{}]interface{} {
	route = cleanOpenshiftObject(route)

	routeSpec := route["spec"].(map[interface{}]interface{})
	routeSpecTo := routeSpec["to"].(map[interface{}]interface{})
	deleteKeyIfValueMatches(routeSpecTo, "weight", 100)
	deleteKeyIfValueMatches(routeSpec, "wildcardPolicy", "None")

	return route
}

// https://kubernetes.io/docs/reference/federation/v1/definitions/#_v1_service
func cleanService(service map[interface{}]interface{}) map[interface{}]interface{} {
	service = cleanOpenshiftObject(service)

	serviceSpec := service["spec"].(map[interface{}]interface{})

	// ports is an optional parameter
	if val, ok := serviceSpec["ports"]; ok {
		ports := val.([]interface{})
		for _, port := range ports {
			deleteKeyIfValueMatches(port.(map[interface{}]interface{}), "protocol", "TCP")
		}
	}

	deleteKeyIfValueMatches(serviceSpec, "sessionAffinity", "None")
	deleteKeyIfValueMatches(serviceSpec, "type", "ClusterIP")

	return service
}

func cleanOpenshiftObject(openshiftObject map[interface{}]interface{}) map[interface{}]interface{} {
	openshiftObject = cleanMetadata(openshiftObject)

	// Status properties across different objects are populated by the server
	delete(openshiftObject, "status")

	return openshiftObject
}

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/types.go
func cleanMetadata(openshiftObject map[interface{}]interface{}) map[interface{}]interface{} {
	metadata := openshiftObject["metadata"].(map[interface{}]interface{})

	if val, ok := metadata["annotations"]; ok {
		cleanAnnotations(val.(map[interface{}]interface{}))
		deleteKeyIfEmpty(metadata, "annotations")
	}

	delete(metadata, "creationTimestamp")
	// "Populated by the system. Read-only."
	delete(metadata, "generation")
	// "Populated by the system. Read-only."
	delete(metadata, "resourceVersion")
	// "Populated by the system. Read-only."
	delete(metadata, "selfLink")

	deleteKeyIfEmpty(openshiftObject, "metadata")

	return openshiftObject
}

func cleanAnnotations(annotations map[interface{}]interface{}) map[interface{}]interface{} {
	annotationsToDelete := []string{
		"openshift.io/generated-by",
		"openshift.io/host.generated",
		"openshift.io/image.dockerRepositoryCheck",
		"openshift.io/imported-from",
	}

	for _, annotationToDelete := range annotationsToDelete {
		delete(annotations, annotationToDelete)
	}

	return annotations
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
