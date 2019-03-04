[![CircleCI](https://circleci.com/gh/bmaupin/openshift-cleanup.svg?style=shield)](https://circleci.com/gh/bmaupin/openshift-cleanup)
[![License](https://img.shields.io/badge/license-Apache_2-blue.svg)](https://github.com/bmaupin/openshift-cleanup/blob/master/LICENSE)
---

This removes unnecessary objects and parameters from an exported OpenShift configuration so that it can be more easily
managed and optionally turned into a template.


### Features

- Removes unnecessary objects
    - For example, Builds are removed because they're recreated by the BuildConfig
- Removes read-only properties that are set by the server
    - For example, creation timestamps, unique IDs, etc
- Removes certain properties when they're set to the default value
- Removes empty or blank properties


### Examples

Cleaning up the exported configuration from OpenShift's nodejs-ex example gives us a file that's nearly 5 times smaller:

```
$ cat testdata/openshift-list-nodejs-ex-original.yaml | wc -l
501
$ ./openshift-cleanup testdata/openshift-list-nodejs-ex-original.yaml | wc -l
105
```

Cleaning up the exported configuration from a deployed application with a database gives us a file that's nearly 8 times smaller:

```
$ cat testdata/openshift-list-2-original.yaml | wc -l
2060
$ ./openshift-cleanup testdata/openshift-list-2-original.yaml | wc -l
265
```


### Usage

1. Export the configuration

    (Pick whichever one best suits your needs):
    ```
    oc get all --export -o json > configuration-original.json
    oc get all --export -o yaml > configuration-original.yaml
    oc get all --export -o json -l app=appname > configuration-original.json
    oc get all --export -o yaml -l app=appname > configuration-original.yaml
    ```

1. Export persistent volume claims and add them to the configuration

    ```
    oc get pvc -o yaml ...
    ```

1. Clean up the configuration (with this app)

    ```
    ./openshift-cleanup configuration-original.yaml > configuration.yaml
    ```

    Or:
    ```
    ./openshift-cleanup configuration-original.json > configuration.yaml
    ```

1. (Optional) Parameterize the configs and convert into a template

    [https://docs.openshift.com/container-platform/latest/dev_guide/templates.html](https://docs.openshift.com/container-platform/latest/dev_guide/templates.html)
