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

	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/proto"
	"github.com/saarasio/enroute/enroute-dp/internal/dag"
	"github.com/saarasio/enroute/enroute-dp/internal/envoy"
)

// RouteCache manages the contents of the gRPC RDS cache.
type RouteCache struct {
	mu      sync.Mutex
	values  map[string]*envoy_config_route_v3.RouteConfiguration
	waiters []chan int
	last    int
}

// Register registers ch to receive a value when Notify is called.
// The value of last is the count of the times Notify has been called on this Cache.
// It functions of a sequence counter, if the value of last supplied to Register
// is less than the Cache's internal counter, then the caller has missed at least
// one notification and will fire immediately.
//
// Sends by the broadcaster to ch must not block, therefor ch must have a capacity
// of at least 1.
func (c *RouteCache) Register(ch chan int, last int) {
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
func (c *RouteCache) Update(v map[string]*envoy_config_route_v3.RouteConfiguration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.notify()
}

// notify notifies all registered waiters that an event has occurred.
func (c *RouteCache) notify() {
	c.last++

	for _, ch := range c.waiters {
		ch <- c.last
	}
	c.waiters = c.waiters[:0]
}

// Contents returns a copy of the cache's contents.
func (c *RouteCache) Contents() []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, v := range c.values {
		values = append(values, v)
	}
	sort.Stable(routeConfigurationsByName(values))
	return values
}

// Query searches the RouteCache for the named RouteConfiguration entries.
func (c *RouteCache) Query(names []string) []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, n := range names {
		v, ok := c.values[n]
		if !ok {
			// if there is no route registered with the cache
			// we return a blank route configuration. This is
			// not the same as returning nil, we're choosing to
			// say "the configuration you asked for _does exists_,
			// but it contains no useful information.
			v = &envoy_config_route_v3.RouteConfiguration{
				Name: n,
			}
		}
		values = append(values, v)
	}
	sort.Stable(routeConfigurationsByName(values))
	return values
}

type routeConfigurationsByName []proto.Message

func (r routeConfigurationsByName) Len() int      { return len(r) }
func (r routeConfigurationsByName) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r routeConfigurationsByName) Less(i, j int) bool {
	return r[i].(*envoy_config_route_v3.RouteConfiguration).Name < r[j].(*envoy_config_route_v3.RouteConfiguration).Name
}

// TypeURL returns the string type of RouteCache Resource.
func (*RouteCache) TypeURL() string { return resource.RouteType }

type routeVisitor struct {
	routes map[string]*envoy_config_route_v3.RouteConfiguration
}

func visitRoutes(root dag.Vertex) map[string]*envoy_config_route_v3.RouteConfiguration {
	rv := routeVisitor{
		routes: map[string]*envoy_config_route_v3.RouteConfiguration{
			"ingress_http": {
				Name: "ingress_http",
			},
			"ingress_https": {
				Name: "ingress_https",
			},
		},
	}
	rv.visit(root)
	for _, v := range rv.routes {
		sort.Stable(virtualHostsByName(v.VirtualHosts))
	}
	return rv.routes
}

func SetupFilters(vh *dag.VirtualHost, vhost *envoy_config_route_v3.VirtualHost, isVh bool, r *dag.Route) {
	// Walk through filters and invoke SetupEnvoyFilters for e
	envoy.SetupEnvoyFilters(vh, vhost, isVh, r)
}

func envoyRouteFromDagRoute(vh *dag.VirtualHost, vhost *envoy_config_route_v3.VirtualHost, isVh bool, r *dag.Route) {
	if len(r.Clusters) < 1 {
		// no services for this route, skip it.
		return
	}

	SetupFilters(vh, vhost, isVh, r)

	rr := &envoy_config_route_v3.Route{
		Match:               envoy.RouteMatchNew(r),
		Action:              envoy.RouteRoute(r),
		RequestHeadersToAdd: envoy.RouteHeaders(),
		// TODO: Only add this if there is a per-route filter configured
		//TypedPerFilterConfig: envoy.RouteTypedFilterConfig(r),
	}

	if isVh == true && r.HTTPSUpgrade {
		rr.Action = envoy.UpgradeHTTPS()
		rr.RequestHeadersToAdd = nil
	}

	vhost.Routes = append(vhost.Routes, rr)
}

func (v *routeVisitor) visit(vertex dag.Vertex) {
	switch l := vertex.(type) {
	case *dag.Listener:
		l.Visit(func(vertex dag.Vertex) {
			switch vh := vertex.(type) {
			case *dag.VirtualHost:
				vhost := envoy.VirtualHost(vh.Name)
				vh.Visit(func(v dag.Vertex) {
					if r, ok := v.(*dag.Route); ok {
						envoyRouteFromDagRoute(vh, vhost, true, r)
					}
				})
				if len(vhost.Routes) < 1 {
					return
				}
				sort.Stable(sort.Reverse(longestRouteFirst(vhost.Routes)))
				v.routes["ingress_http"].VirtualHosts = append(v.routes["ingress_http"].VirtualHosts, vhost)
			case *dag.SecureVirtualHost:
				vhost := envoy.VirtualHost(vh.VirtualHost.Name)
				vh.Visit(func(v dag.Vertex) {
					if r, ok := v.(*dag.Route); ok {
						envoyRouteFromDagRoute(&vh.VirtualHost, vhost, false, r)
					}
				})
				if len(vhost.Routes) < 1 {
					return
				}
				sort.Stable(sort.Reverse(longestRouteFirst(vhost.Routes)))
				v.routes["ingress_https"].VirtualHosts = append(v.routes["ingress_https"].VirtualHosts, vhost)
			default:
				// recurse
				vertex.Visit(v.visit)
			}
		})
	default:
		// recurse
		vertex.Visit(v.visit)
	}
}

type virtualHostsByName []*envoy_config_route_v3.VirtualHost

func (v virtualHostsByName) Len() int           { return len(v) }
func (v virtualHostsByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v virtualHostsByName) Less(i, j int) bool { return v[i].Name < v[j].Name }

type longestRouteFirst []*envoy_config_route_v3.Route

func (l longestRouteFirst) Len() int      { return len(l) }
func (l longestRouteFirst) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l longestRouteFirst) Less(i, j int) bool {
	a, ok := l[i].Match.PathSpecifier.(*envoy_config_route_v3.RouteMatch_Prefix)
	if !ok {
		// ignore non prefix matches
		return false
	}

	b, ok := l[j].Match.PathSpecifier.(*envoy_config_route_v3.RouteMatch_Prefix)
	if !ok {
		// ignore non prefix matches
		return false
	}

	return a.Prefix < b.Prefix
}
