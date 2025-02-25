// SPDX-License-Identifier: Apache-2.0
// Copyright(c) 2018-2020 Saaras Inc.

// Copyright © 2018 Heptio
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

package contour

import (
	"sort"
	"sync"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/proto"
	"github.com/saarasio/enroute/enroute-dp/internal/dag"
	"github.com/saarasio/enroute/enroute-dp/internal/envoy"
	"github.com/saarasio/enroute/enroute-dp/internal/logger"
)

const (
	ENVOY_HTTP_LISTENER            = "ingress_http"
	ENVOY_HTTPS_LISTENER           = "ingress_https"
	DEFAULT_HTTP_ACCESS_LOG        = "/dev/stdout"
	DEFAULT_HTTP_LISTENER_ADDRESS  = "0.0.0.0"
	DEFAULT_HTTP_LISTENER_PORT     = 8080
	DEFAULT_HTTPS_ACCESS_LOG       = "/dev/stdout"
	DEFAULT_HTTPS_LISTENER_ADDRESS = DEFAULT_HTTP_LISTENER_ADDRESS
	DEFAULT_HTTPS_LISTENER_PORT    = 8443
)

// ListenerVisitorConfig holds configuration parameters for visitListeners.
type ListenerVisitorConfig struct {
	// Envoy's HTTP (non TLS) listener address.
	// If not set, defaults to DEFAULT_HTTP_LISTENER_ADDRESS.
	HTTPAddress string

	// Envoy's HTTP (non TLS) listener port.
	// If not set, defaults to DEFAULT_HTTP_LISTENER_PORT.
	HTTPPort int

	// Envoy's HTTP (non TLS) access log path.
	// If not set, defaults to DEFAULT_HTTP_ACCESS_LOG.
	HTTPAccessLog string

	// Envoy's HTTPS (TLS) listener address.
	// If not set, defaults to DEFAULT_HTTPS_LISTENER_ADDRESS.
	HTTPSAddress string

	// Envoy's HTTPS (TLS) listener port.
	// If not set, defaults to DEFAULT_HTTPS_LISTENER_PORT.
	HTTPSPort int

	// Envoy's HTTPS (TLS) access log path.
	// If not set, defaults to DEFAULT_HTTPS_ACCESS_LOG.
	HTTPSAccessLog string

	// UseProxyProto configurs all listeners to expect a PROXY
	// V1 or V2 preamble.
	// If not set, defaults to false.
	UseProxyProto bool
}

// httpAddress returns the port for the HTTP (non TLS)
// listener or DEFAULT_HTTP_LISTENER_ADDRESS if not configured.
func (lvc *ListenerVisitorConfig) httpAddress() string {
	if lvc.HTTPAddress != "" {
		return lvc.HTTPAddress
	}
	return DEFAULT_HTTP_LISTENER_ADDRESS
}

// httpPort returns the port for the HTTP (non TLS)
// listener or DEFAULT_HTTP_LISTENER_PORT if not configured.
func (lvc *ListenerVisitorConfig) httpPort() int {
	if lvc.HTTPPort != 0 {
		return lvc.HTTPPort
	}
	return DEFAULT_HTTP_LISTENER_PORT
}

// httpAccessLog returns the access log for the HTTP (non TLS)
// listener or DEFAULT_HTTP_ACCESS_LOG if not configured.
func (lvc *ListenerVisitorConfig) httpAccessLog() string {
	if lvc.HTTPAccessLog != "" {
		return lvc.HTTPAccessLog
	}
	return DEFAULT_HTTP_ACCESS_LOG
}

// httpsAddress returns the port for the HTTPS (TLS)
// listener or DEFAULT_HTTPS_LISTENER_ADDRESS if not configured.
func (lvc *ListenerVisitorConfig) httpsAddress() string {
	if lvc.HTTPSAddress != "" {
		return lvc.HTTPSAddress
	}
	return DEFAULT_HTTPS_LISTENER_ADDRESS
}

// httpsPort returns the port for the HTTPS (TLS) listener
// or DEFAULT_HTTPS_LISTENER_PORT if not configured.
func (lvc *ListenerVisitorConfig) httpsPort() int {
	if lvc.HTTPSPort != 0 {
		return lvc.HTTPSPort
	}
	return DEFAULT_HTTPS_LISTENER_PORT
}

// httpsAccessLog returns the access log for the HTTPS (TLS)
// listener or DEFAULT_HTTPS_ACCESS_LOG if not configured.
func (lvc *ListenerVisitorConfig) httpsAccessLog() string {
	if lvc.HTTPSAccessLog != "" {
		return lvc.HTTPSAccessLog
	}
	return DEFAULT_HTTPS_ACCESS_LOG
}

// ListenerCache manages the contents of the gRPC LDS cache.
type ListenerCache struct {
	mu           sync.Mutex
	values       map[string]*envoy_config_listener_v3.Listener
	staticValues map[string]*envoy_config_listener_v3.Listener
	waiters      []chan int
	last         int
}

// NewListenerCache returns an instance of a ListenerCache
func NewListenerCache(address string, port int) ListenerCache {
	stats := envoy.StatsListener(address, port)
	return ListenerCache{
		staticValues: map[string]*envoy_config_listener_v3.Listener{
			stats.Name: stats,
		},
	}
}

// Register registers ch to receive a value when Notify is called.
// The value of last is the count of the times Notify has been called on this Cache.
// It functions of a sequence counter, if the value of last supplied to Register
// is less than the Cache's internal counter, then the caller has missed at least
// one notification and will fire immediately.
//
// Sends by the broadcaster to ch must not block, therefor ch must have a capacity
// of at least 1.
func (c *ListenerCache) Register(ch chan int, last int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if last < c.last {
		// notify this channel immediately
		ch <- c.last
		return
	}
	c.waiters = append(c.waiters, ch)
}

// Update replaces the contents of the cache with the supplied map.
func (c *ListenerCache) Update(v map[string]*envoy_config_listener_v3.Listener) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.notify()
}

// notify notifies all registered waiters that an event has occurred.
func (c *ListenerCache) notify() {
	c.last++

	for _, ch := range c.waiters {
		ch <- c.last
	}
	c.waiters = c.waiters[:0]
}

// Contents returns a copy of the cache's contents.
func (c *ListenerCache) Contents() []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, v := range c.values {
		values = append(values, v)
	}
	for _, v := range c.staticValues {
		values = append(values, v)
	}
	sort.Stable(listenersByName(values))
	return values
}

func (c *ListenerCache) Query(names []string) []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, n := range names {
		v, ok := c.values[n]
		if !ok {
			v, ok = c.staticValues[n]
			if !ok {
				// if the listener is not registered in
				// dynamic or static values then skip it
				// as there is no way to return a blank
				// listener because the listener address
				// field is required.
				continue
			}
		}
		values = append(values, v)
	}
	sort.Stable(listenersByName(values))
	return values
}

type listenersByName []proto.Message

func (l listenersByName) Len() int      { return len(l) }
func (l listenersByName) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l listenersByName) Less(i, j int) bool {
	return l[i].(*envoy_config_listener_v3.Listener).Name < l[j].(*envoy_config_listener_v3.Listener).Name
}

func (*ListenerCache) TypeURL() string { return resource.ListenerType }

type listenerVisitor struct {
	*ListenerVisitorConfig

	listeners map[string]*envoy_config_listener_v3.Listener
	// 6-5-2020 - If we find a dag.VirtualHost, we add the listener
	// in visit() just like dag.SecureVirtualHost
	// This simplifies switch/case here and elsewhere
	// http      bool // at least one dag.VirtualHost encountered
}

// Entry-point from builder
func visitListeners(root dag.Vertex, lvc *ListenerVisitorConfig) map[string]*envoy_config_listener_v3.Listener {
	if logger.EL.ELogger != nil {
		logger.EL.ELogger.Debugf("contour:visitListeners()")
	}

	lv := listenerVisitor{
		ListenerVisitorConfig: lvc,
		listeners: map[string]*envoy_config_listener_v3.Listener{
			ENVOY_HTTPS_LISTENER: envoy.Listener(
				ENVOY_HTTPS_LISTENER,
				lvc.httpsAddress(), lvc.httpsPort(),
				secureProxyProtocol(lvc.UseProxyProto),
			),
		},
	}
	lv.visit(root)

	// remove the https listener if there are no vhosts bound to it.
	if len(lv.listeners[ENVOY_HTTPS_LISTENER].FilterChains) == 0 {
		delete(lv.listeners, ENVOY_HTTPS_LISTENER)
	} else {
		// there's some https listeners, we need to sort the filter chains
		// to ensure that the LDS entries are identical.
		sort.SliceStable(lv.listeners[ENVOY_HTTPS_LISTENER].FilterChains,
			func(i, j int) bool {
				// The ServerNames field will only ever have a single entry
				// in our FilterChain config, so it's okay to only sort
				// on the first slice entry.
				return lv.listeners[ENVOY_HTTPS_LISTENER].FilterChains[i].FilterChainMatch.ServerNames[0] < lv.listeners[ENVOY_HTTPS_LISTENER].FilterChains[j].FilterChainMatch.ServerNames[0]
			})
	}

	if logger.EL.ELogger != nil {
		logger.EL.ELogger.Debugf("contour:visitListeners() -> setupHttpFilters()")
	}

	// All the listeners have been setup.
	// Walk through the DAG and update listeners with HttpFilters
	lv.setupHttpFilters(root)

	return lv.listeners
}

func proxyProtocol(useProxy bool) []*envoy_config_listener_v3.ListenerFilter {
	if useProxy {
		return envoy.ListenerFilters(
			envoy.ProxyProtocol(),
		)
	}
	return nil
}

func secureProxyProtocol(useProxy bool) []*envoy_config_listener_v3.ListenerFilter {
	return append(proxyProtocol(useProxy), envoy.TLSInspector())
}

func (v *listenerVisitor) visit(vertex dag.Vertex) {
	switch vh := vertex.(type) {
	case *dag.VirtualHost:
		v.listeners[ENVOY_HTTP_LISTENER] = envoy.Listener(
			ENVOY_HTTP_LISTENER,
			v.httpAddress(), v.httpPort(),
			proxyProtocol(v.UseProxyProto),
			envoy.HTTPConnectionManager(ENVOY_HTTP_LISTENER, v.httpAccessLog(), &vertex),
		)

	case *dag.SecureVirtualHost:

		filters := envoy.Filters(
			envoy.HTTPConnectionManager(ENVOY_HTTPS_LISTENER, v.httpsAccessLog(), &vertex),
		)
		alpnProtos := []string{"h2", "http/1.1"}
		if vh.VirtualHost.TCPProxy != nil {
			filters = envoy.Filters(
				envoy.TCPProxy(ENVOY_HTTPS_LISTENER, vh.VirtualHost.TCPProxy, v.httpsAccessLog()),
			)
			alpnProtos = nil // do not offer ALPN
		}

		fc := envoy.FilterChainTLS(vh.VirtualHost.Name, vh.Secret, filters, vh.MinProtoVersion, alpnProtos...)

		v.listeners[ENVOY_HTTPS_LISTENER].FilterChains = append(v.listeners[ENVOY_HTTPS_LISTENER].FilterChains, fc)
	default:
		// recurse
		vertex.Visit(v.visit)
	}
}
