package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type BuildConfig struct {
	TypeMeta `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`

	Spec BuildConfigSpec `yaml:",omitempty"`
	Status BuildConfigStatus
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type BuildConfigSpec struct {
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
	Strategy BuildStrategy
	Triggers []BuildTriggerPolicy
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type BuildConfigStatus struct {
	LastVersion int64
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type BuildStrategy struct {
	SourceStrategy *SourceBuildStrategy `yaml:"sourceStrategy"`
	Type string `yaml:"type"`
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type BuildTriggerPolicy struct {
	Generic WebHookTrigger `yaml:",omitempty"`
	Github WebHookTrigger `yaml:",omitempty"`
	ImageChange ImageChangeTrigger `yaml:"imageChange,omitempty"`
	Type string `yaml:"type"`
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
type ImageChangeTrigger struct {}

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/runtime/interfaces.go
type Object interface {}

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/types.go
type ObjectMeta struct {
	Labels struct {
		App string `yaml:"app,omitempty"`
	} `yaml:",omitempty"`
	Name string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type SourceBuildStrategy struct {
	From *struct {
		Kind string `yaml:"kind,omitempty"`
		Name string `yaml:"name,omitempty"`
		Namespace string `yaml:"namespace,omitempty"`
	}
}

// https://github.com/openshift/origin/blob/master/pkg/template/apis/template/types.go
type Template struct {
	TypeMeta `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`

	Objects []Object
}

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/types.go
type TypeMeta struct {
	Kind string `yaml:"kind,omitempty"`
	APIVersion string `yaml:"apiVersion,omitempty"`
}

// https://github.com/openshift/origin/blob/master/pkg/build/apis/build/types.go
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

	template = cleanTemplateObjects(template)

	marshaledTemplate, err := yaml.Marshal(&template)
	fmt.Printf("--- template dump:\n%s\n", string(marshaledTemplate))
}

func cleanTemplateObjects(template Template) Template {
	newTemplateObjects := []Object{}

	for _, object := range template.Objects {
		objectKind := object.(map[interface {}]interface {})["kind"]

		// Is there a better way to do this? Seems hacky: marshaling, unmarshaling, and then replacing each object ...
		switch objectKind {
			case "BuildConfig":
				marshaledBuildConfig, err := yaml.Marshal(&object)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				buildConfig := BuildConfig{}
				if err := yaml.Unmarshal(marshaledBuildConfig, &buildConfig); err != nil {
					log.Fatalf("error: %v", err)
				}
				newTemplateObjects = append(newTemplateObjects, buildConfig)

			// Builds will be recreated by the BuildConfig
			case "Build":
				fallthrough
			// Pods will be recreated by the DeploymentConfig
			case "Pod":
				fallthrough
			case "ReplicationController":
				// noop

			default:
				log.Println(fmt.Sprintf("WARNING: Unhandled object kind: %s", objectKind))
		}
	}

	template.Objects = newTemplateObjects

	return template
}
