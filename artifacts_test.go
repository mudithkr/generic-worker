package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"

	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/slugid-go/slugid"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

var (
	// all tests can share taskGroupId so we can view all test tasks in same
	// graph later for troubleshooting
	taskGroupID string = slugid.Nice()
)

func validateArtifacts(
	t *testing.T,
	payloadArtifacts []struct {
		Expires tcclient.Time `json:"expires"`
		Path    string        `json:"path"`
		Type    string        `json:"type"`
	},
	expected []Artifact) {

	// to test, create a dummy task run with given artifacts
	// and then call PayloadArtifacts() method to see what
	// artifacts would get uploaded...
	tr := &TaskRun{
		Payload: GenericWorkerPayload{
			Artifacts: payloadArtifacts,
		},
	}
	artifacts := tr.PayloadArtifacts()

	// compare expected vs actual artifacts by converting artifacts to strings...
	if fmt.Sprintf("%q", artifacts) != fmt.Sprintf("%q", expected) {
		t.Fatalf("Expected different artifacts to be generated...\nExpected:\n%q\nActual:\n%q", expected, artifacts)
	}
}

// See the testdata/SampleArtifacts subdirectory of this project. This
// simulates adding it as a directory artifact in a task payload, and checks
// that all files underneath this directory are discovered and created as s3
// artifacts.
func TestDirectoryArtifacts(t *testing.T) {

	setup(t)
	validateArtifacts(t,

		// what appears in task payload
		[]struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{{
			Expires: inAnHour,
			Path:    "SampleArtifacts",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]Artifact{
			S3Artifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "SampleArtifacts/%%%/v/X",
					Expires:       inAnHour,
				},
				MimeType: "application/octet-stream",
			},
			S3Artifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "SampleArtifacts/_/X.txt",
					Expires:       inAnHour,
				},
				MimeType: "text/plain; charset=utf-8",
			},
			S3Artifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "SampleArtifacts/b/c/d.jpg",
					Expires:       inAnHour,
				},
				MimeType: "image/jpeg",
			},
		})
}

// Task payload specifies a file artifact which doesn't exist on worker
func TestMissingFileArtifact(t *testing.T) {

	setup(t)
	validateArtifacts(t,

		// what appears in task payload
		[]struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{{
			Expires: inAnHour,
			Path:    "TestMissingFileArtifact/no_such_file",
			Type:    "file",
		}},

		// what we expect to discover on file system
		[]Artifact{
			ErrorArtifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "TestMissingFileArtifact/no_such_file",
					Expires:       inAnHour,
				},
				Message: "Could not read file '" + filepath.Join(TaskUser.HomeDir, "TestMissingFileArtifact", "no_such_file") + "'",
				Reason:  "file-missing-on-worker",
			},
		})
}

// Task payload specifies a directory artifact which doesn't exist on worker
func TestMissingDirectoryArtifact(t *testing.T) {

	setup(t)
	validateArtifacts(t,

		// what appears in task payload
		[]struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{{
			Expires: inAnHour,
			Path:    "TestMissingDirectoryArtifact/no_such_dir",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]Artifact{
			ErrorArtifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "TestMissingDirectoryArtifact/no_such_dir",
					Expires:       inAnHour,
				},
				Message: "Could not read directory '" + filepath.Join(TaskUser.HomeDir, "TestMissingDirectoryArtifact", "no_such_dir") + "'",
				Reason:  "file-missing-on-worker",
			},
		})
}

// Task payload specifies a file artifact which is actually a directory on worker
func TestFileArtifactIsDirectory(t *testing.T) {

	setup(t)
	validateArtifacts(t,

		// what appears in task payload
		[]struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{{
			Expires: inAnHour,
			Path:    "SampleArtifacts/b/c",
			Type:    "file",
		}},

		// what we expect to discover on file system
		[]Artifact{
			ErrorArtifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "SampleArtifacts/b/c",
					Expires:       inAnHour,
				},
				Message: "File artifact '" + filepath.Join(TaskUser.HomeDir, "SampleArtifacts", "b", "c") + "' exists as a directory, not a file, on the worker",
				Reason:  "invalid-resource-on-worker",
			},
		})
}

// Task payload specifies a directory artifact which is a regular file on worker
func TestDirectoryArtifactIsFile(t *testing.T) {

	setup(t)
	validateArtifacts(t,

		// what appears in task payload
		[]struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{{
			Expires: inAnHour,
			Path:    "SampleArtifacts/b/c/d.jpg",
			Type:    "directory",
		}},

		// what we expect to discover on file system
		[]Artifact{
			ErrorArtifact{
				BaseArtifact: BaseArtifact{
					CanonicalPath: "SampleArtifacts/b/c/d.jpg",
					Expires:       inAnHour,
				},
				Message: "Directory artifact '" + filepath.Join(TaskUser.HomeDir, "SampleArtifacts", "b", "c", "d.jpg") + "' exists as a file, not a directory, on the worker",
				Reason:  "invalid-resource-on-worker",
			},
		})
}

func TestUpload(t *testing.T) {

	setup(t)

	expires := tcclient.Time(time.Now().Add(time.Minute * 30))

	payload := GenericWorkerPayload{
		Command:    helloGoodbye(),
		MaxRunTime: 7200,
		Artifacts: []struct {
			Expires tcclient.Time `json:"expires"`
			Path    string        `json:"path"`
			Type    string        `json:"type"`
		}{
			{
				Path:    "SampleArtifacts/_/X.txt",
				Expires: expires,
				Type:    "file",
			},
		},
		Features: struct {
			ChainOfTrust bool `json:"chainOfTrust,omitempty"`
		}{
			ChainOfTrust: true,
		},
	}

	td := testTask()

	taskID, myQueue := submitTask(t, td, payload)
	runWorker()

	// some required substrings - not all, just a selection
	expectedArtifacts := map[string]struct {
		extracts        []string
		contentEncoding string
		expires         tcclient.Time
	}{
		"public/logs/live_backing.log": {
			extracts: []string{
				"hello world!",
				"goodbye world!",
				`"instance-type": "p3.enormous"`,
			},
			contentEncoding: "gzip",
			expires:         td.Expires,
		},
		"public/logs/live.log": {
			extracts: []string{
				"hello world!",
				"goodbye world!",
				"=== Task Finished ===",
				"Exit Code: 0",
			},
			contentEncoding: "gzip",
			expires:         td.Expires,
		},
		"public/logs/certified.log": {
			extracts: []string{
				"hello world!",
				"goodbye world!",
				"=== Task Finished ===",
				"Exit Code: 0",
			},
			contentEncoding: "gzip",
			expires:         td.Expires,
		},
		"public/logs/chainOfTrust.json.asc": {
			// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  ./%%%/v/X
			// 8308d593eb56527137532595a60255a3fcfbe4b6b068e29b22d99742bad80f6f  ./_/X.txt
			// a0ed21ab50992121f08da55365da0336062205fd6e7953dbff781a7de0d625b7  ./b/c/d.jpg
			extracts: []string{
				"8308d593eb56527137532595a60255a3fcfbe4b6b068e29b22d99742bad80f6f",
			},
			contentEncoding: "gzip",
			expires:         td.Expires,
		},
		"SampleArtifacts/_/X.txt": {
			extracts: []string{
				"test artifact",
			},
			contentEncoding: "",
			expires:         payload.Artifacts[0].Expires,
		},
	}

	artifacts, err := myQueue.ListArtifacts(taskID, "0", "", "")

	if err != nil {
		t.Fatalf("Error listing artifacts: %v", err)
	}

	actualArtifacts := make(map[string]struct {
		ContentType string        `json:"contentType"`
		Expires     tcclient.Time `json:"expires"`
		Name        string        `json:"name"`
		StorageType string        `json:"storageType"`
	}, len(artifacts.Artifacts))

	for _, actualArtifact := range artifacts.Artifacts {
		actualArtifacts[actualArtifact.Name] = actualArtifact
	}

	for artifact := range expectedArtifacts {
		if a, ok := actualArtifacts[artifact]; ok {
			if a.ContentType != "text/plain; charset=utf-8" {
				t.Errorf("Artifact %s should have mime type 'text/plain; charset=utf-8' but has '%s'", artifact, a.ContentType)
			}
			if a.Expires.String() != expectedArtifacts[artifact].expires.String() {
				t.Errorf("Artifact %s should have expiry '%s' but has '%s'", artifact, expires, a.Expires)
			}
		} else {
			t.Errorf("Artifact '%s' not created", artifact)
		}
	}

	// now check content was uploaded to Amazon, and is correct

	// signer of public/logs/chainOfTrust.json.asc
	signer := &openpgp.Entity{}
	cotCert := &ChainOfTrustData{}

	for artifact, content := range expectedArtifacts {
		url, err := myQueue.GetLatestArtifact_SignedURL(taskID, artifact, 10*time.Minute)
		if err != nil {
			t.Fatalf("Error trying to fetch artifacts from Amazon...\n%s", err)
		}
		// need to do this so Content-Encoding header isn't swallowed by Go for test later on
		tr := &http.Transport{
			DisableCompression: true,
		}
		client := &http.Client{Transport: tr}
		rawResp, _, err := httpbackoff.ClientGet(client, url.String())
		if err != nil {
			t.Fatalf("Error trying to fetch decompressed artifact from signed URL %s ...\n%s", url.String(), err)
		}
		resp, _, err := httpbackoff.Get(url.String())
		if err != nil {
			t.Fatalf("Error trying to fetch artifact from signed URL %s ...\n%s", url.String(), err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Error trying to read response body of artifact from signed URL %s ...\n%s", url.String(), err)
		}
		for _, requiredSubstring := range content.extracts {
			if strings.Index(string(b), requiredSubstring) < 0 {
				t.Errorf("Artifact '%s': Could not find substring %q in '%s'", artifact, requiredSubstring, string(b))
			}
		}
		if actualContentEncoding := rawResp.Header.Get("Content-Encoding"); actualContentEncoding != content.contentEncoding {
			t.Fatalf("Expected Content-Encoding %q but got Content-Encoding %q for artifact %q from url %v", content.contentEncoding, actualContentEncoding, artifact, url)
		}
		if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "text/plain; charset=utf-8" {
			t.Fatalf("Content-Type in Signed URL response does not match Content-Type of artifact")
		}
		// check openpgp signature is valid
		if artifact == "public/logs/chainOfTrust.json.asc" {
			pubKey, err := os.Open(filepath.Join("testdata", "public-openpgp-key"))
			if err != nil {
				t.Fatalf("Error opening public key file")
			}
			defer pubKey.Close()
			entityList, err := openpgp.ReadArmoredKeyRing(pubKey)
			if err != nil {
				t.Fatalf("Error decoding public key file")
			}
			block, _ := clearsign.Decode(b)
			signer, err = openpgp.CheckDetachedSignature(entityList, bytes.NewBuffer(block.Bytes), block.ArmoredSignature.Body)
			if err != nil {
				t.Fatalf("Not able to validate openpgp signature of public/logs/chainOfTrust.json.asc")
			}
			err = json.Unmarshal(block.Plaintext, cotCert)
			if err != nil {
				t.Fatalf("Could not interpret public/logs/chainOfTrust.json as json")
			}
		}
	}
	if signer == nil {
		t.Fatalf("Signer of public/logs/chainOfTrust.json.asc could not be established (is nil)")
	}
	if signer.Identities["Generic-Worker <taskcluster-accounts+gpgsigning@mozilla.com>"] == nil {
		t.Fatalf("Did not get correct signer identity in public/logs/chainOfTrust.json.asc - %#v", signer.Identities)
	}

	// This trickery is to convert a TaskDefinitionResponse into a
	// TaskDefinitionRequest in order that we can compare. We cannot cast, so
	// need to transform to json as an intermediary step.
	b, err := json.Marshal(cotCert.Task)
	if err != nil {
		t.Fatalf("Cannot marshal task into json - %#v\n%v", cotCert.Task, err)
	}
	cotCertTaskRequest := &queue.TaskDefinitionRequest{}
	err = json.Unmarshal(b, cotCertTaskRequest)
	if err != nil {
		t.Fatalf("Cannot unmarshal json into task request - %#v\n%v", string(b), err)
	}

	// The Payload, Tags and Extra fields are raw bytes, so differences may not
	// be valid. Since we are comparing the rest, let's skip these two fields,
	// as the rest should give us good enough coverage already
	cotCertTaskRequest.Payload = nil
	cotCertTaskRequest.Tags = nil
	cotCertTaskRequest.Extra = nil
	td.Payload = nil
	td.Tags = nil
	td.Extra = nil
	if !reflect.DeepEqual(cotCertTaskRequest, td) {
		t.Fatalf("Did not get back expected task definition in chain of trust certificate:\n%#v\n ** vs **\n%#v", cotCertTaskRequest, td)
	}
	if len(cotCert.Artifacts) != 2 {
		t.Fatalf("Expected 2 artifact hashes to be listed")
	}
	if cotCert.TaskID != taskID {
		t.Fatalf("Expected taskId to be %q but was %q", taskID, cotCert.TaskID)
	}
	if cotCert.RunID != 0 {
		t.Fatalf("Expected runId to be 0 but was %v", cotCert.RunID)
	}
	if cotCert.WorkerGroup != "test-worker-group" {
		t.Fatalf("Expected workerGroup to be \"test-worker-group\" but was %q", cotCert.WorkerGroup)
	}
	if cotCert.WorkerID != "test-worker-id" {
		t.Fatalf("Expected workerGroup to be \"test-worker-id\" but was %q", cotCert.WorkerID)
	}
	if cotCert.Environment.PublicIPAddress != "12.34.56.78" {
		t.Fatalf("Expected publicIpAddress to be 12.34.56.78 but was %v", cotCert.Environment.PublicIPAddress)
	}
	if cotCert.Environment.PrivateIPAddress != "87.65.43.21" {
		t.Fatalf("Expected privateIpAddress to be 87.65.43.21 but was %v", cotCert.Environment.PrivateIPAddress)
	}
	if cotCert.Environment.InstanceID != "test-instance-id" {
		t.Fatalf("Expected instanceId to be \"test-instance-id\" but was %v", cotCert.Environment.InstanceID)
	}
	if cotCert.Environment.InstanceType != "p3.enormous" {
		t.Fatalf("Expected instanceType to be \"p3.enormous\" but was %v", cotCert.Environment.InstanceType)
	}
	if cotCert.Environment.Region != "outer-space" {
		t.Fatalf("Expected region to be \"outer-space\" but was %v", cotCert.Environment.Region)
	}
}
