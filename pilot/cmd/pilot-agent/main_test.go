// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	"github.com/onsi/gomega"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/api/networking/v1alpha3"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/proxy/envoy"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pkg/config/constants"
)

func TestPilotDefaultDomainKubernetes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	role = &model.Proxy{}
	role.DNSDomain = ""
	registryID = serviceregistry.Kubernetes

	domain := getDNSDomain("default", role.DNSDomain)

	g.Expect(domain).To(gomega.Equal("default.svc.cluster.local"))
}

func TestPilotDefaultDomainConsul(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	role := &model.Proxy{}
	role.DNSDomain = ""
	registryID = serviceregistry.Consul

	domain := getDNSDomain("", role.DNSDomain)

	g.Expect(domain).To(gomega.Equal("service.consul"))
}

func TestPilotDefaultDomainOthers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	role = &model.Proxy{}
	role.DNSDomain = ""
	registryID = serviceregistry.Mock

	domain := getDNSDomain("", role.DNSDomain)

	g.Expect(domain).To(gomega.Equal(""))
}

func TestPilotDomain(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	role.DNSDomain = "my.domain"
	registryID = serviceregistry.Mock

	domain := getDNSDomain("", role.DNSDomain)

	g.Expect(domain).To(gomega.Equal("my.domain"))
}

func TestCustomMixerSanIfAuthenticationMutualDomainKubernetes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	role = &model.Proxy{}
	role.DNSDomain = ""
	trustDomain = "mesh.com"
	mixerIdentity = "mixer-identity"
	registryID = serviceregistry.Kubernetes

	setSpiffeTrustDomain("", role.DNSDomain)
	mixerSAN := envoy.GetSAN("", mixerIdentity)

	g.Expect(mixerSAN).To(gomega.Equal("spiffe://mesh.com/mixer-identity"))
}

func TestDedupeStrings(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	in := []string{
		constants.DefaultCertChain, constants.DefaultKey, constants.DefaultRootCert,
		constants.DefaultCertChain, constants.DefaultKey, constants.DefaultRootCert,
	}
	expected := []string{constants.DefaultCertChain, constants.DefaultKey, constants.DefaultRootCert}

	actual := dedupeStrings(in)

	g.Expect(actual).To(gomega.ConsistOf(expected))
}

func TestIsIPv6Proxy(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []string
		expected bool
	}{
		{
			name:     "ipv4 only",
			addrs:    []string{"1.1.1.1", "127.0.0.1", "2.2.2.2"},
			expected: false,
		},
		{
			name:     "ipv6 only",
			addrs:    []string{"1111:2222::1", "::1", "2222:3333::1"},
			expected: true,
		},
		{
			name:     "mixed ipv4 and ipv6",
			addrs:    []string{"1111:2222::1", "::1", "127.0.0.1", "2.2.2.2", "2222:3333::1"},
			expected: false,
		},
	}
	for _, tt := range tests {
		result := isIPv6Proxy(tt.addrs)
		if result != tt.expected {
			t.Errorf("Test %s failed, expected: %t got: %t", tt.name, tt.expected, result)
		}
	}
}

func Test_fromJSON(t *testing.T) {
	val := `
{"address":"oap.istio-system:11800",
"tlsSettings":{"mode": "DISABLE", "subjectAltNames": [], "caCertificates": null},
"tcpKeepalive":{"interval":"10s","probes":3,"time":"10s"}
}
`

	type args struct {
		j string
	}
	tests := []struct {
		name string
		args args
		want *meshconfig.RemoteService
	}{
		{
			name: "foo",
			args: struct{ j string }{j: val},
			want: &meshconfig.RemoteService{
				Address:     "oap.istio-system:11800",
				TlsSettings: &v1alpha3.ClientTLSSettings{},
				TcpKeepalive: &v1alpha3.ConnectionPoolSettings_TCPSettings_TcpKeepalive{
					Probes:   3,
					Time:     &types.Duration{Seconds: 10},
					Interval: &types.Duration{Seconds: 10},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromJSON(tt.args.j); !proto.Equal(got, tt.want) {
				t.Errorf("got = %v \n want %v", got, tt.want)
			}
		})
	}
}
