# Specification

This repo is a Go application based on the ConfigHub SDK. Specifically it is a variant of the bridge example in https://github.com/confighub/sdk/tree/main/examples/hello-world-bridge.

This Go app implements a worker in the same way as the example. This worker should include all existing standard bridges and all standard functions. The example in https://github.com/confighub/sdk/tree/main/examples/hello-world-function shows how to register all standard functions.

The one difference is that instead of the standard Kubernetes bridge, this app registers its own bridge which is a wrapper around the standard Kubernetes bridge. It should pass through all operations and then it should additionally do the following:

* On Apply, it should save the config unit data and metadata in a directory structure. The parent directory can be configured for the app. All subdirs and files are in the format `space-slug/unit-slug`. For each unit it saves two files it saves a yaml file with content of the "Data" field on config unit (base64 decoded and saved as readable yaml) and a json file with the rest of the config unit serialized as a json object, without the Data field.

* On Destroy, it should delete these files.
