package main

import (
	"fmt"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"

	appsv1 "github.com/openshift/api/apps/v1"
	authorizationv1 "github.com/openshift/api/authorization/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	networkv1 "github.com/openshift/api/network/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	projectv1 "github.com/openshift/api/project/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	templatev1 "github.com/openshift/api/template/v1"
	userv1 "github.com/openshift/api/user/v1"

	// corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	// The Kubernetes Go client (nested within the OpenShift Go client)
	// automatically registers its types in scheme.Scheme, however the
	// additional OpenShift types must be registered manually.  AddToScheme
	// registers the API group types (e.g. route.openshift.io/v1, Route) only.
	appsv1.AddToScheme(scheme.Scheme)
	authorizationv1.AddToScheme(scheme.Scheme)
	buildv1.AddToScheme(scheme.Scheme)
	imagev1.AddToScheme(scheme.Scheme)
	networkv1.AddToScheme(scheme.Scheme)
	oauthv1.AddToScheme(scheme.Scheme)
	projectv1.AddToScheme(scheme.Scheme)
	quotav1.AddToScheme(scheme.Scheme)
	routev1.AddToScheme(scheme.Scheme)
	securityv1.AddToScheme(scheme.Scheme)
	templatev1.AddToScheme(scheme.Scheme)
	userv1.AddToScheme(scheme.Scheme)

	// If you need to serialize/deserialize legacy (non-API group) OpenShift
	// types (e.g. v1, Route), these must be additionally registered using
	// AddToSchemeInCoreGroup.
	appsv1.AddToSchemeInCoreGroup(scheme.Scheme)
	authorizationv1.AddToSchemeInCoreGroup(scheme.Scheme)
	buildv1.AddToSchemeInCoreGroup(scheme.Scheme)
	imagev1.AddToSchemeInCoreGroup(scheme.Scheme)
	networkv1.AddToSchemeInCoreGroup(scheme.Scheme)
	oauthv1.AddToSchemeInCoreGroup(scheme.Scheme)
	projectv1.AddToSchemeInCoreGroup(scheme.Scheme)
	quotav1.AddToSchemeInCoreGroup(scheme.Scheme)
	routev1.AddToSchemeInCoreGroup(scheme.Scheme)
	securityv1.AddToSchemeInCoreGroup(scheme.Scheme)
	templatev1.AddToSchemeInCoreGroup(scheme.Scheme)
	userv1.AddToSchemeInCoreGroup(scheme.Scheme)
}

func main() {
	templateYAML, err := ioutil.ReadFile(filepath.ToSlash("testdata/exported-openshift-template.yml"))
	if err != nil {
		panic(err)
	}

	// Create a YAML serializer.  JSON is a subset of YAML, so is supported too.
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
		scheme.Scheme)

	// Decode the YAML to an object.
	var template templatev1.Template
	_, _, err = s.Decode(templateYAML, nil, &template)
	if err != nil {
		panic(err)
	}

	// Some types, e.g. List, contain RawExtensions.  If the appropriate types
	// are registered, these can be decoded in a second pass.
	for i, o := range template.Objects {
		o.Object, _, err = s.Decode(o.Raw, nil, nil)
		if err != nil {
			panic(err)
		}
		o.Raw = nil

		template.Objects[i] = o
	}

	fmt.Printf("%#v\n", template)

	// Encode the object to YAML.
	err = s.Encode(&template, os.Stdout)
	if err != nil {
		panic(err)
	}
}
