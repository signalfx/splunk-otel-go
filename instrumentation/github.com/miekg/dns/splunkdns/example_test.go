// Copyright Splunk Inc.
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

package splunkdns_test

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns"
)

func Example_client() {
	m := new(dns.Msg)
	m.SetQuestion("miek.nl.", dns.TypeMX)
	// Calling splunkdns.Exchange calls dns.Exchange and trace the request.
	reply, err := splunkdns.Exchange(m, "127.0.0.1:53")
	fmt.Println(reply, err)
}

func Example_server() {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		_ = w.WriteMsg(m)
	})
	// Calling splunkdns.ListenAndServe calls dns.ListenAndServe and traces
	// all requests to handled by mux.
	if err := splunkdns.ListenAndServe(":dns", "udp", mux); err != nil {
		fmt.Println(err)
	}
}
