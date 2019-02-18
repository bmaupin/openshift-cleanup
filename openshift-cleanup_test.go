package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestListCleanup1(t *testing.T) {
	testContents, err := ioutil.ReadFile("testdata/openshift-list-cleaned1.yaml")
	if err != nil {
		t.Errorf("Unexpected error reading test data file: %s", err)
	}

	cleanedContents := cleanOpenshiftConfigFile("testdata/openshift-list-original1.yaml")
	if bytes.Compare(cleanedContents, testContents) != 0 {
		t.Error("Cleaned list contents don't match")
	}
}

func TestListCleanupNodejsEx(t *testing.T) {
	testContents, err := ioutil.ReadFile("testdata/openshift-list-nodejs-ex-cleaned.yaml")
	if err != nil {
		t.Errorf("Unexpected error reading test data file: %s", err)
	}

	cleanedContents := cleanOpenshiftConfigFile("testdata/openshift-list-nodejs-ex-original.yaml")
	if bytes.Compare(cleanedContents, testContents) != 0 {
		t.Error("Cleaned list contents don't match")
	}
}

func TestListCleanupNodejsExJson(t *testing.T) {
	testContents, err := ioutil.ReadFile("testdata/openshift-list-nodejs-ex-cleaned.yaml")
	if err != nil {
		t.Errorf("Unexpected error reading test data file: %s", err)
	}

	cleanedContents := cleanOpenshiftConfigFile("testdata/openshift-list-nodejs-ex-original.json")
	if bytes.Compare(cleanedContents, testContents) != 0 {
		t.Error("Cleaned list contents don't match")
	}
}

func TestTemplateCleanup1(t *testing.T) {
	testContents, err := ioutil.ReadFile("testdata/openshift-template-cleaned1.yaml")
	if err != nil {
		t.Errorf("Unexpected error reading test data file: %s", err)
	}

	cleanedContents := cleanOpenshiftConfigFile("testdata/openshift-template-original1.yaml")
	if bytes.Compare(cleanedContents, testContents) != 0 {
		t.Error("Cleaned template contents don't match")
	}
}

func TestTemplateCleanup2(t *testing.T) {
	testContents, err := ioutil.ReadFile("testdata/openshift-template-cleaned2.yaml")
	if err != nil {
		t.Errorf("Unexpected error reading test data file: %s", err)
	}

	cleanedContents := cleanOpenshiftConfigFile("testdata/openshift-template-original2.yaml")
	if bytes.Compare(cleanedContents, testContents) != 0 {
		t.Error("Cleaned template contents don't match")
	}
}
