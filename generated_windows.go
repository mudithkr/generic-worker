// This source code file is AUTO-GENERATED by github.com/taskcluster/jsonschema2go

package main

import (
	"encoding/json"
	"errors"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

type (
	ArtifactContent struct {

		// Max length: 1024
		Artifact string `json:"artifact"`

		// Syntax:     ^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$
		TaskID string `json:"taskId"`
	}

	Content json.RawMessage

	DirectoryMount json.RawMessage

	FileMount struct {

		// Content of the file to be mounted
		Content Content `json:"content"`

		// The filesystem location to mount the file
		File string `json:"file"`
	}

	// This schema defines the structure of the `payload` property referred to in a
	// TaskCluster Task definition.
	GenericWorkerPayload struct {

		// Artifacts to be published. For example:
		// `{ "type": "file", "path": "builds\\firefox.exe", "expires": "2015-08-19T17:30:00.000Z" }`
		Artifacts []struct {

			// Date when artifact should expire must be in the future
			Expires tcclient.Time `json:"expires"`

			// Filesystem path of artifact
			Path string `json:"path"`

			// Artifacts can be either an individual `file` or a `directory` containing
			// potentially multiple files with recursively included subdirectories.
			//
			// Possible values:
			//   * "file"
			//   * "directory"
			Type string `json:"type"`
		} `json:"artifacts,omitempty"`

		// One entry per command (consider each entry to be interpreted as a full line of
		// a Windows™ .bat file). For example:
		// `["set", "echo hello world > hello_world.txt", "set GOPATH=C:\\Go"]`.
		Command []string `json:"command"`

		// Example: ```{ "PATH": "C:\\Windows\\system32;C:\\Windows", "GOOS": "darwin" }```
		Env json.RawMessage `json:"env,omitempty"`

		// Feature flags enable additional functionality.
		Features struct {

			// An artifact named chainOfTrust.json.asc should be generated
			// which will include information for downstream tasks to build
			// a level of trust for the artifacts produced by the task and
			// the environment it ran in.
			ChainOfTrust bool `json:"chainOfTrust,omitempty"`
		} `json:"features,omitempty"`

		// Maximum time the task container can run in seconds
		//
		// Mininum:    1
		// Maximum:    86400
		MaxRunTime int `json:"maxRunTime"`

		// Directories and/or files to be mounted
		Mounts []json.RawMessage `json:"mounts,omitempty"`
	}

	ReadOnlyDirectory struct {

		// Contents of read only directory.
		Content Content `json:"content"`

		// The filesystem location to mount the directory volume
		Directory string `json:"directory"`

		// Archive format of content for read only directory
		//
		// Possible values:
		//   * "tar.gz"
		//   * "zip"
		Format string `json:"format"`
	}

	// URL to download content from
	URLContent struct {

		// URL to download content from
		URL string `json:"url"`
	}

	Var FileMount

	Var1 DirectoryMount

	WritableDirectoryCache struct {

		// Implies a read/write cache directory volume. A unique name for the cache volume. Note if this cache is loaded from an artifact, you will require scope `queue:get-artifact:<artifact-name>` to use this cache.
		CacheName string `json:"cacheName"`

		// Optional content to be loaded when initially creating the cache.
		Content Content `json:"content,omitempty"`

		// The filesystem location to mount the directory volume
		Directory string `json:"directory"`

		// Archive format of content for writable directory cache
		//
		// Possible values:
		//   * "rar"
		//   * "tar.bz2"
		//   * "tar.gz"
		//   * "zip"
		Format string `json:"format,omitempty"`
	}
)

// MarshalJSON calls json.RawMessage method of the same name. Required since
// Content is of type json.RawMessage...
func (this *Content) MarshalJSON() ([]byte, error) {
	x := json.RawMessage(*this)
	return (&x).MarshalJSON()
}

// UnmarshalJSON is a copy of the json.RawMessage implementation.
func (this *Content) UnmarshalJSON(data []byte) error {
	if this == nil {
		return errors.New("Content: UnmarshalJSON on nil pointer")
	}
	*this = append((*this)[0:0], data...)
	return nil
}

// MarshalJSON calls json.RawMessage method of the same name. Required since
// DirectoryMount is of type json.RawMessage...
func (this *DirectoryMount) MarshalJSON() ([]byte, error) {
	x := json.RawMessage(*this)
	return (&x).MarshalJSON()
}

// UnmarshalJSON is a copy of the json.RawMessage implementation.
func (this *DirectoryMount) UnmarshalJSON(data []byte) error {
	if this == nil {
		return errors.New("DirectoryMount: UnmarshalJSON on nil pointer")
	}
	*this = append((*this)[0:0], data...)
	return nil
}

// Returns json schema for the payload part of the task definition. Please
// note we use a go string and do not load an external file, since we want this
// to be *part of the compiled executable*. If this sat in another file that
// was loaded at runtime, it would not be burned into the build, which would be
// bad for the following two reasons:
//  1) we could no longer distribute a single binary file that didn't require
//     installation/extraction
//  2) the payload schema is specific to the version of the code, therefore
//     should be versioned directly with the code and *frozen on build*.
//
// Run `generic-worker show-payload-schema` to output this schema to standard
// out.
func taskPayloadSchema() string {
	return `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "additionalProperties": false,
  "definitions": {
    "content": {
      "oneOf": [
        {
          "additionalProperties": false,
          "properties": {
            "artifact": {
              "maxLength": 1024,
              "type": "string"
            },
            "taskId": {
              "pattern": "^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$",
              "type": "string"
            }
          },
          "required": [
            "taskId",
            "artifact"
          ],
          "title": "Artifact Content",
          "type": "object"
        },
        {
          "additionalProperties": false,
          "description": "URL to download content from",
          "properties": {
            "url": {
              "description": "URL to download content from",
              "format": "uri",
              "title": "URL",
              "type": "string"
            }
          },
          "required": [
            "url"
          ],
          "title": "URL Content",
          "type": "object"
        }
      ]
    },
    "directoryMount": {
      "oneOf": [
        {
          "additionalProperties": false,
          "dependencies": {
            "content": [
              "format"
            ],
            "format": [
              "content"
            ]
          },
          "properties": {
            "cacheName": {
              "description": "Implies a read/write cache directory volume. A unique name for the cache volume. Note if this cache is loaded from an artifact, you will require scope ` + "`" + `queue:get-artifact:\u003cartifact-name\u003e` + "`" + ` to use this cache.",
              "title": "Cache Name",
              "type": "string"
            },
            "content": {
              "$ref": "#/definitions/content",
              "description": "Optional content to be loaded when initially creating the cache.",
              "title": "Content",
              "type": "object"
            },
            "directory": {
              "description": "The filesystem location to mount the directory volume",
              "title": "Directory Volume",
              "type": "string"
            },
            "format": {
              "description": "Archive format of content for writable directory cache",
              "enum": [
                "rar",
                "tar.bz2",
                "tar.gz",
                "zip"
              ],
              "title": "Format",
              "type": "string"
            }
          },
          "required": [
            "directory",
            "cacheName"
          ],
          "title": "Writable Directory Cache",
          "type": "object"
        },
        {
          "additionalProperties": false,
          "properties": {
            "content": {
              "$ref": "#/definitions/content",
              "description": "Contents of read only directory.",
              "title": "Content",
              "type": "object"
            },
            "directory": {
              "description": "The filesystem location to mount the directory volume",
              "title": "Directory",
              "type": "string"
            },
            "format": {
              "description": "Archive format of content for read only directory",
              "enum": [
                "tar.gz",
                "zip"
              ],
              "title": "Format",
              "type": "string"
            }
          },
          "required": [
            "directory",
            "content",
            "format"
          ],
          "title": "Read Only Directory",
          "type": "object"
        }
      ],
      "type": "object"
    },
    "fileMount": {
      "additionalProperties": false,
      "properties": {
        "content": {
          "$ref": "#/definitions/content",
          "description": "Content of the file to be mounted",
          "title": "Content",
          "type": "object"
        },
        "file": {
          "description": "The filesystem location to mount the file",
          "title": "File",
          "type": "string"
        }
      },
      "required": [
        "file",
        "content"
      ],
      "type": "object"
    }
  },
  "description": "This schema defines the structure of the ` + "`" + `payload` + "`" + ` property referred to in a\nTaskCluster Task definition.",
  "id": "http://schemas.taskcluster.net/generic-worker/v1/payload.json#",
  "properties": {
    "artifacts": {
      "description": "Artifacts to be published. For example:\n` + "`" + `{ \"type\": \"file\", \"path\": \"builds\\\\firefox.exe\", \"expires\": \"2015-08-19T17:30:00.000Z\" }` + "`" + `",
      "items": {
        "additionalProperties": false,
        "properties": {
          "expires": {
            "description": "Date when artifact should expire must be in the future",
            "format": "date-time",
            "title": "Expiry date and time",
            "type": "string"
          },
          "path": {
            "description": "Filesystem path of artifact",
            "title": "Artifact location",
            "type": "string"
          },
          "type": {
            "description": "Artifacts can be either an individual ` + "`" + `file` + "`" + ` or a ` + "`" + `directory` + "`" + ` containing\npotentially multiple files with recursively included subdirectories.",
            "enum": [
              "file",
              "directory"
            ],
            "title": "Artifact upload type.",
            "type": "string"
          }
        },
        "required": [
          "type",
          "path",
          "expires"
        ],
        "type": "object"
      },
      "title": "Artifacts to be published",
      "type": "array"
    },
    "command": {
      "description": "One entry per command (consider each entry to be interpreted as a full line of\na Windows™ .bat file). For example:\n` + "`" + `[\"set\", \"echo hello world \u003e hello_world.txt\", \"set GOPATH=C:\\\\Go\"]` + "`" + `.",
      "items": {
        "type": "string"
      },
      "minItems": 1,
      "title": "Commands to run",
      "type": "array"
    },
    "env": {
      "description": "Example: ` + "`" + `` + "`" + `` + "`" + `{ \"PATH\": \"C:\\\\Windows\\\\system32;C:\\\\Windows\", \"GOOS\": \"darwin\" }` + "`" + `` + "`" + `` + "`" + `",
      "title": "Environment variable mappings.",
      "type": "object"
    },
    "features": {
      "additionalProperties": false,
      "description": "Feature flags enable additional functionality.",
      "properties": {
        "chainOfTrust": {
          "description": "An artifact named chainOfTrust.json.asc should be generated\nwhich will include information for downstream tasks to build\na level of trust for the artifacts produced by the task and\nthe environment it ran in.",
          "title": "Enable generation of a openpgp signed Chain of Trust artifact",
          "type": "boolean"
        }
      },
      "title": "Feature flags",
      "type": "object"
    },
    "maxRunTime": {
      "description": "Maximum time the task container can run in seconds",
      "maximum": 86400,
      "minimum": 1,
      "multipleOf": 1,
      "title": "Maximum run time in seconds",
      "type": "integer"
    },
    "mounts": {
      "description": "Directories and/or files to be mounted",
      "items": {
        "oneOf": [
          {
            "$ref": "#/definitions/fileMount"
          },
          {
            "$ref": "#/definitions/directoryMount"
          }
        ],
        "type": "object"
      },
      "title": "Mounts",
      "type": "array"
    }
  },
  "required": [
    "command",
    "maxRunTime"
  ],
  "title": "Generic worker payload",
  "type": "object"
}`
}
