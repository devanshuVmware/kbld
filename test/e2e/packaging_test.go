//go:build e2e

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestPkgUnpkgSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	// redis:5.0.4
	input := `
kind: Object
spec:
# references image index
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
# references plain image
- image: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`

	path := "/tmp/kbld-test-pkg-unpkg-successful"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{"unpackage", "-f", "-", "--input", path, "--repository", env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg")}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut = env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`)

	if out != expectedOut {
		t.Fatalf("Expected unpackage output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestPkgUnpkgLockSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
overrides:
# ignored because it's not preresolved
- image: gcs-fetcher
  newImage: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
- image: redis
  newImage: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`

	path := "/tmp/kbld-test-pkg-unpkg-with-lock-successful"
	defer os.RemoveAll(path)

	relocatedLockPath := path + "-relocated"
	defer os.RemoveAll(relocatedLockPath)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{
		"unpackage", "-f", "-", "--input", path, "--repository",
		env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg"),
		"--lock-output", relocatedLockPath,
	}, RunOpts{StdinReader: strings.NewReader(input)})

	lockOutBs, err := ioutil.ReadFile(relocatedLockPath)
	if err != nil {
		t.Fatalf("Expected to find relocated lock file")
	}

	kbldVersionOutput, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})
	kbldVersion := strings.SplitAfter(kbldVersionOutput, " ")[2]

	expectedLockOut := strings.ReplaceAll(env.WithRegistries(`apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: __ver__
overrides:
- image: redis
  newImage: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  newImage: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`), "__ver__", strings.TrimSuffix(kbldVersion, "\n"))

	if string(lockOutBs) != expectedLockOut {
		t.Fatalf("Expected unpackage lock output >>>%s<<< to match >>>%s<<<", lockOutBs, expectedLockOut)
	}
}

func TestPkgUnpkgSuccessfulWithForeignLayers(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	// Mongo has 2 foreign layers
	input := `
kind: Object
spec:
- image: index.docker.io/library/mongo@sha256:633ec3ae6db954a65a1abadb482bae73375d0098005cb36a3851b32cd891b22e
`

	path := "/tmp/kbld-test-pkg-unpkg-successful-foreign-layers"
	defer os.RemoveAll(path)

	out, _ := kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	out, _ = kbld.RunWithOpts([]string{"unpackage", "-f", "-", "--input", path, "--repository", env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg")}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut = env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-pkg-unpkg@sha256:633ec3ae6db954a65a1abadb482bae73375d0098005cb36a3851b32cd891b22e
`)

	if out != expectedOut {
		t.Fatalf("Expected unpackage output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestPkgUnpkgSuccessfulWithManyImages(t *testing.T) {
	env := BuildEnv(t)

	if env.SkipStressTests {
		fmt.Printf("This is a stress test; skipping.")
		return
	}

	kbld := Kbld{t, env.KbldBinaryPath, Logger{}}

	input := `
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: 0.20.0
overrides:
- image: cfidentity/uaa@sha256:9f1e7e399c96309935145624d1824b2c2bf93656fd9c4dcf1c593b55f98aa6a8
  newImage: index.docker.io/cfidentity/uaa@sha256:9f1e7e399c96309935145624d1824b2c2bf93656fd9c4dcf1c593b55f98aa6a8
  preresolved: true
- image: cloudfoundry/capi:nginx@sha256:51e4e48c457d5cb922cf0f569e145054e557e214afa78fb2b312a39bb2f938b6
  newImage: index.docker.io/cloudfoundry/capi@sha256:51e4e48c457d5cb922cf0f569e145054e557e214afa78fb2b312a39bb2f938b6
  preresolved: true
- image: cloudfoundry/cloud-controller-ng:1ebab1cbb5270a3d51c0a098a37cd9e8ca0f721d@sha256:374f967edd7db4d7efc2f38cb849988aa36a8248dd240d56f49484b8159fd800
  newImage: index.docker.io/cloudfoundry/cloud-controller-ng@sha256:374f967edd7db4d7efc2f38cb849988aa36a8248dd240d56f49484b8159fd800
  preresolved: true
- image: cloudfoundry/cnb:0.0.94-bionic@sha256:5b03a853e636b78c44e475bbc514e2b7b140cc41cca8ab907e9753431ae8c0b0
  newImage: index.docker.io/cloudfoundry/cnb@sha256:5b03a853e636b78c44e475bbc514e2b7b140cc41cca8ab907e9753431ae8c0b0
  preresolved: true
- image: docker.io/istio/citadel:1.4.5
  newImage: index.docker.io/istio/citadel@sha256:420a331a528886aca47bed5b8c549c78d594e52d771f876f3137d3557207712f
  preresolved: true
- image: docker.io/istio/galley:1.4.5
  newImage: index.docker.io/istio/galley@sha256:26e744bdfd3db289d4cfc9be63e38e7c7a424a9f76d1224cbdbbe58374229b68
  preresolved: true
- image: docker.io/istio/mixer:1.4.5
  newImage: index.docker.io/istio/mixer@sha256:ff6f39732c31999911790b00b484a471b6fe87192223d3266b0b6e752a374287
  preresolved: true
- image: docker.io/istio/node-agent-k8s:1.4.5
  newImage: index.docker.io/istio/node-agent-k8s@sha256:7e17ab509777a54f3c0dfb4518692a9ca179d1e8c41df87dc81a734339b37152
  preresolved: true
- image: docker.io/istio/pilot:1.4.5
  newImage: index.docker.io/istio/pilot@sha256:2bca5900d6bf20d5f0bf2b6673bc4d2885bab8cca2a9060336a0024930665b59
  preresolved: true
- image: docker.io/istio/proxyv2:1.4.5
  newImage: index.docker.io/istio/proxyv2@sha256:fc09ea0f969147a4843a564c5b677fbf3a6f94b56627d00b313b4c30d5fef094
  preresolved: true
- image: docker.io/istio/sidecar_injector:1.4.5
  newImage: index.docker.io/istio/sidecar_injector@sha256:ba446f8cf98bafdad4514fd492432dd180243cbc55a0b9c6bebfe31cb169033d
  preresolved: true
- image: eirini/opi@sha256:2e0b84c5fcb1e6e5cdb07a70210f2e462aa52119f7a330660a7444a938deefbb
  newImage: index.docker.io/eirini/opi@sha256:2e0b84c5fcb1e6e5cdb07a70210f2e462aa52119f7a330660a7444a938deefbb
  preresolved: true
- image: gcr.io/cf-build-service-public/kpack/controller@sha256:1d7d80257e2019a474417ba0c7dcfff5612aeec55e24d91ef7b2e4bd0a521a40
  newImage: gcr.io/cf-build-service-public/kpack/controller@sha256:1d7d80257e2019a474417ba0c7dcfff5612aeec55e24d91ef7b2e4bd0a521a40
  preresolved: true
- image: gcr.io/cf-build-service-public/kpack/webhook@sha256:c2461ef9634c771f2a06bc0371040b43c9a78dd0e4ac1c9fde3f4525e0ae21f2
  newImage: gcr.io/cf-build-service-public/kpack/webhook@sha256:c2461ef9634c771f2a06bc0371040b43c9a78dd0e4ac1c9fde3f4525e0ae21f2
  preresolved: true
- image: index.docker.io/bitnami/postgresql@sha256:9762d9a80b90a5efe299d4848057ac5c45fb384570b36f60aad38fe2b1704bd6
  newImage: index.docker.io/bitnami/postgresql@sha256:9762d9a80b90a5efe299d4848057ac5c45fb384570b36f60aad38fe2b1704bd6
  preresolved: true
- image: index.docker.io/metacontroller/metacontroller@sha256:ad85cb5f5ad9a61a3f38277fed371df43ea0fc55d9073dfa8f4fc2e27c127603
  newImage: index.docker.io/metacontroller/metacontroller@sha256:ad85cb5f5ad9a61a3f38277fed371df43ea0fc55d9073dfa8f4fc2e27c127603
  preresolved: true
- image: index.docker.io/minio/minio@sha256:5e96d539583afd9a7da14e0d9bf2360d316e4e8219659d82b8ef106a9d75b16c
  newImage: index.docker.io/minio/minio@sha256:5e96d539583afd9a7da14e0d9bf2360d316e4e8219659d82b8ef106a9d75b16c
  preresolved: true
- image: index.docker.io/cloudfoundry/cf-k8s-logging@sha256:d8c73e6c87b2a71c8b6798205761bb7870fb2080a4329c4eefb0b4620656eeaa
  newImage: index.docker.io/cloudfoundry/cf-k8s-logging@sha256:d8c73e6c87b2a71c8b6798205761bb7870fb2080a4329c4eefb0b4620656eeaa
  preresolved: true
- image: index.docker.io/cloudfoundry/log-cache-cf-auth-proxy@sha256:f75fc1585dbf869a1a98ee094d334525bb944a07aa0ebe522019c3a479b843de
  newImage: index.docker.io/cloudfoundry/log-cache-cf-auth-proxy@sha256:f75fc1585dbf869a1a98ee094d334525bb944a07aa0ebe522019c3a479b843de
  preresolved: true
- image: index.docker.io/cloudfoundry/log-cache-gateway@sha256:51bc54b4a901228537f43efe3b687c3d8c07f49c015775141a0dffa66cd83ec6
  newImage: index.docker.io/cloudfoundry/log-cache-gateway@sha256:51bc54b4a901228537f43efe3b687c3d8c07f49c015775141a0dffa66cd83ec6
  preresolved: true
- image: index.docker.io/cloudfoundry/log-cache@sha256:616ce5401945100a84df5380e913d1136f2ed714aa90c78f592839de88928dfb
  newImage: index.docker.io/cloudfoundry/log-cache@sha256:616ce5401945100a84df5380e913d1136f2ed714aa90c78f592839de88928dfb
  preresolved: true
- image: index.docker.io/cloudfoundry/syslog-server@sha256:39a386521f94c70071eab4a7fb67cc7e28adba2e2dd8113d6df155c17b19f5a5
  newImage: index.docker.io/cloudfoundry/syslog-server@sha256:39a386521f94c70071eab4a7fb67cc7e28adba2e2dd8113d6df155c17b19f5a5
  preresolved: true
- image: index.docker.io/cloudfoundry/metric-proxy@sha256:0f10d093fbb4437e071dcf31411ef344b4059bf26494e4b23e2625a7d9d4e77e
  newImage: index.docker.io/cloudfoundry/metric-proxy@sha256:0f10d093fbb4437e071dcf31411ef344b4059bf26494e4b23e2625a7d9d4e77e
  preresolved: true
- image: prom/statsd-exporter:v0.15.0@sha256:e3174186628b401e4a441b78513ba06e957644267332436be0c77dd7af9bdddc
  newImage: index.docker.io/prom/statsd-exporter@sha256:e3174186628b401e4a441b78513ba06e957644267332436be0c77dd7af9bdddc
  preresolved: true
`

	expectedPackagedSHA := "e2c66f42fdac6993741440ba157fe4fb3f505eb3"

	path := "/tmp/kbld-test-pkg-unpkg-successful-with-many-images"
	defer os.RemoveAll(path)

	kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path, "--concurrency=1"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	actualSHA := sha1File(t, path)

	// Assert that concurrently writing to tar doesn't affect sha
	if actualSHA != expectedPackagedSHA {
		t.Fatalf("Expected package sha to be same >>>%s<<< to match >>>%s<<<", actualSHA, expectedPackagedSHA)
	}

	kbld.RunWithOpts([]string{"package", "-f", "-", "--output", path, "--concurrency=5"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	actualSHA = sha1File(t, path)

	if actualSHA != expectedPackagedSHA {
		t.Fatalf("Expected package sha to be same >>>%s<<< to match >>>%s<<<", actualSHA, expectedPackagedSHA)
	}

	kbld.RunWithOpts([]string{
		"unpackage", "-f", "-", "--input", path,
		"--repository", env.WithRegistries("docker.io/*username*/kbld-test-pkg-unpkg-successful-with-many-images"),
	}, RunOpts{StdinReader: strings.NewReader(input)})
}

func sha1File(t *testing.T, path string) string {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	hs := sha1.New()
	if _, err := io.Copy(hs, f); err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf("%x", hs.Sum(nil))
}
