package webapp

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/opencontainers/go-digest"

	hub "git.act3-ace.com/ace/hub/api/v6/pkg/apis/hub.act3-ace.io/v1beta1"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

func Test_getViewerURL(t *testing.T) {
	const bottleSha string = "sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9"
	bottleDigest, err := digest.Parse(bottleSha)
	if err != nil {
		t.Fatal("could not parse bottle sha", "sha", bottleSha)
	}
	type args struct {
		spec          v1alpha2.ViewerSpec
		artifact      *db.PublicArtifact
		hubInstance   v1alpha2.ACEHubInstance
		bottle        digest.Digest
		partSelectors []string
	}
	defaultHub := hub.HubEnvTemplateSpec{
		ServiceAccountName: "test-account",
		EnvSecretPrefix:    "pfx",
		QueueName:          "myQueue",
		GPU: &hub.GPU{
			Type:  "myGPU",
			Count: 2,
		},
		Resources: corev1.ResourceRequirements{},
		Image:     "reg.act3.git/myImage:v1",
		Env: map[string]string{
			"ENV1": "TEST1",
			"ENV2": "TEST2",
		},
		Script:       `#!/bin/bash; echo "Hello World!";`,
		SharedMemory: &resource.Quantity{},
		Bottles: []hub.BottleSpec{
			{
				Name:      "myBottle",
				BottleRef: bottleSha,
				Selector:  []string{"testing=true"},
				IPS:       "myImagePullSecret",
			},
		},
		Ports: []hub.Port{{Name: "myPort", Number: 8080, ProxyType: "normal"}},
	}

	defaultArgs := args{
		spec:          v1alpha2.ViewerSpec{},
		hubInstance:   v1alpha2.ACEHubInstance{},
		bottle:        bottleDigest,
		partSelectors: []string{},
		artifact:      &db.PublicArtifact{},
	}
	tests := []struct {
		want    string
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty hub instance",
			args:    defaultArgs,
			want:    "bottles%5B0%5D%5BbottleRef%5D=sha256%3A05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9&bottles%5B0%5D%5Bname%5D=dataset&env%5BaCE_OPEN_PATH%5D=%2Face%2Fbottle%2Fdataset",
			wantErr: false,
		},
		{
			name: "example hub instance",
			args: args{
				spec: v1alpha2.ViewerSpec{
					Name:   "hub-viewer",
					Accept: "text/html",
					ACEHub: defaultHub,
				},
				artifact: &db.PublicArtifact{},
				hubInstance: v1alpha2.ACEHubInstance{
					Name: "myHub",
					URL:  "https://myhub.io",
				},
				bottle:        bottleDigest,
				partSelectors: []string{},
			},
			want:    "bottles%5B0%5D%5BbottleRef%5D=sha256%3A05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9&bottles%5B0%5D%5Bips%5D=myImagePullSecret&bottles%5B0%5D%5Bname%5D=myBottle&bottles%5B0%5D%5Bselector%5D%5B%5D=testing%3Dtrue&bottles%5B1%5D%5BbottleRef%5D=sha256%3A05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9&bottles%5B1%5D%5Bname%5D=dataset&envSecretPrefix=pfx&env%5BaCE_OPEN_PATH%5D=%2Face%2Fbottle%2Fdataset&env%5BeNV1%5D=TEST1&env%5BeNV2%5D=TEST2&gPU%5Bcount%5D=2&gPU%5Btype%5D=myGPU&image=reg.act3.git%2FmyImage%3Av1&ports%5B0%5D%5Bname%5D=myPort&ports%5B0%5D%5Bnumber%5D=8080&ports%5B0%5D%5BproxyType%5D=normal&queueName=myQueue&script=%23%21%2Fbin%2Fbash%3B+echo+%22Hello+World%21%22%3B&serviceAccountName=test-account",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := getViewerURL(tt.args.spec, tt.args.hubInstance.URL, tt.args.bottle, tt.args.partSelectors, tt.args.artifact)
			if (err != nil) != tt.wantErr {
				t.Errorf("getViewerURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotURL.RawQuery, tt.want) {
				t.Errorf("getViewerURL() = %v, want %v", gotURL.RawQuery, tt.want)
			}
		})
	}
}
