package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type OpenShiftObject struct {
	TypeMeta `yaml:",inline"`
	ObjectMeta `yaml:",inline"`

	Spec OpenShiftObjectSpec `yaml:",omitempty"`
}

type OpenShiftObjectMetadata struct {
	Labels struct {
		App string `yaml:"app,omitempty"`
	} `yaml:",omitempty"`
	Name string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type OpenShiftObjectSpec struct {
	NodeSelector *struct {
		Type string `yaml:"type,omitempty"`
	} `yaml:"nodeSelector"`
	Output *struct {
		To *struct {
			Kind string `yaml:"kind,omitempty"`
			Name string `yaml:"name,omitempty"`
		}
	}
	RunPolicy string `yaml:"runPolicy,omitempty"`
	Source *struct {
		Git *struct {
			Ref string `yaml:"ref,omitempty"`
			Uri string `yaml:"uri,omitempty"`
		}
		Type string `yaml:"type,omitempty"`
	}
	// Status holds any relevant information about a build config
	Status BuildConfigStatus
	Strategy *struct {
		SourceStrategy *struct {
			From *struct {
				Kind string `yaml:"kind,omitempty"`
				Name string `yaml:"name,omitempty"`
				Namespace string `yaml:"namespace,omitempty"`
			}
		}
		Type string `yaml:"type,omitempty"`
	}
	Triggers []BuildTriggerPolicy
}

// BuildConfigStatus contains current state of the build config object.
type BuildConfigStatus struct {
	// LastVersion is used to inform about number of last triggered build.
	LastVersion int64
}

type BuildTriggerPolicy struct {
	Generic WebHookTrigger `yaml:",omitempty"`
	Github WebHookTrigger `yaml:",omitempty"`
	ImageChange ImageChangeTrigger `yaml:"imageChange,omitempty"`
	Type string `yaml:"type"`
}

type ImageChangeTrigger struct {}

type ObjectMeta struct {
	Metadata OpenShiftObjectMetadata `yaml:"metadata,omitempty"`
}

type Template struct {
	TypeMeta `yaml:",inline"`
	ObjectMeta `yaml:",inline"`

	Objects []OpenShiftObject
}

type TypeMeta struct {
	Kind string `yaml:"kind,omitempty"`
	APIVersion string `yaml:"apiVersion,omitempty"`
}

type WebHookTrigger struct {}

func main() {
	contents, err := ioutil.ReadFile(filepath.Join("testdata", "exported-openshift-template.yml"))
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	template := Template{}
	err = yaml.Unmarshal(contents, &template)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// fmt.Printf("--- template:\n%v\n\n", template)

	fmt.Printf("DEBUG: %d\n", len(template.Objects))

	template = cleanTemplateObjects(template)

	cleanedtemplate, err := yaml.Marshal(&template)
	fmt.Printf("--- template dump:\n%s\n\n", string(cleanedtemplate))

	fmt.Printf("DEBUG: %d\n", len(template.Objects))
}

func cleanTemplateObjects(template Template) Template {
	newTemplateObjects := []OpenShiftObject{}

	for _, templateObject := range template.Objects {
		// Builds will be recreated by the BuildConfig
		if templateObject.Kind == "Build" {
			continue
		// Pods will be recreated by the DeploymentConfig
		} else if templateObject.Kind == "Pod" {
			continue
		} else if templateObject.Kind == "ReplicationController" {
			continue
		} else {
			newTemplateObjects = append(newTemplateObjects, templateObject)
		}
		template.Objects = newTemplateObjects
	}

	return template
}
