package webhttp

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/saarasio/enroute/enroute-dp/saaras"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Proxy struct {
	Name string `json:"name" xml:"name" form:"name" query:"name"`
}

type Service struct {
	Service_name string `json:"service_name" xml:"service_name" form:"service_name" query:"service_name"`
	Fqdn         string `json:"fqdn" xml:"fqdn" form:"fqdn" query:"fqdn"`
}

type Route struct {
	Route_name   string `json:"route_name" xml:"route_name" form:"route_name" query:"route_name"`
	Route_prefix string `json:"route_prefix" xml:"route_prefix" form:"route_prefix" query:"route_prefix"`
}

type Upstream struct {
	Upstream_name                       string `json:"upstream_name" xml:"upstream_name" form:"upstream_name" query:"upstream_name"`
	Upstream_ip                         string `json:"upstream_ip" xml:"upstream_ip" form:"upstream_ip" query:"upstream_ip"`
	Upstream_port                       string `json:"upstream_port" xml:"upstream_port" form:"upstream_port" query:"upstream_port"`
	Upstream_hc_path                    string `json:"upstream_hc_path" xml:"upstream_hc_path" form:"upstream_hc_path" query:"upstream_hc_path"`
	Upstream_hc_host                    string `json:"upstream_hc_host" xml:"upstream_hc_host" form:"upstream_hc_host" query:"upstream_hc_host"`
	Upstream_weight                     string `json:"upstream_weight" xml:"upstream_weight" form:"upstream_weight" query:"upstream_weight"`
	Upstream_hc_intervalseconds         string `json:"upstream_hc_intervalseconds" xml:"upstream_hc_intervalseconds" form:"upstream_hc_intervalseconds" query:"upstream_hc_intervalseconds"`
	Upstream_hc_unhealthythresholdcount string `json:"upstream_hc_unhealthythresholdcount" xml:"upstream_hc_unhealthythresholdcount" form:"upstream_hc_unhealthythresholdcount" query:"upstream_hc_unhealthythresholdcount"`
	Upstream_hc_healthythresholdcount   string `json:"upstream_hc_healthythresholdcount" xml:"upstream_hc_healthythresholdcount" form:"upstream_hc_healthythresholdcount" query:"upstream_hc_healthythresholdcount"`
	Upstream_strategy                   string `json:"upstream_strategy" xml:"upstream_strategy" form:"upstream_strategy" query:"upstream_strategy"`
	Upstream_validation_cacertificate   string `json:"upstream_validation_cacertificate" xml:"upstream_validation_cacertificate" form:"upstream_validation_cacertificate" query:"upstream_validation_cacertificate"`
	Upstream_validation_subjectname     string `json:"upstream_validation_subjectname" xml:"upstream_validation_subjectname" form:"upstream_validation_subjectname" query:"upstream_validation_subjectname"`
	Upstream_protocol                   string `json:"upstream_protocol" xml:"upstream_protocol" form:"upstream_protocol" query:"upstream_protocol"`
	Upstream_hc_timeoutseconds          string `json:"upstream_hc_timeoutseconds" xml:"upstream_hc_timeoutseconds" form:"upstream_hc_timeoutseconds" query:"upstream_hc_timeoutseconds"`
}

type Secret struct {
	Secret_name string `json:"secret_name" xml:"secret_name" form:"secret_name" query:"secret_name"`
	Secret_key  string `json:"secret_key" xml:"secret_key" form:"secret_key" query:"secret_key"`
	Secret_cert string `json:"secret_cert" xml:"secret_cert" form:"secret_cert" query:"secret_cert"`
	Secret_sni  string `json:"secret_sni" xml:"secret_sni" form:"secret_sni" query:"secret_sni"`
}

var QCreateProxy string = `
    mutation create_proxy($proxy_name : String!){
      insert_saaras_db_proxy(objects: {proxy_name: $proxy_name},
        on_conflict: {constraint: proxy_proxy_name_key, update_columns: create_ts}) {
        affected_rows
      }
    }
`

var QGetProxy string = `
query get_proxies {
  saaras_db_proxy {
    proxy_id
    proxy_name
    create_ts
    update_ts
  }
}
`

var QGetOneProxy string = `
query get_proxies($proxy_name: String!) {
  saaras_db_proxy(where: {proxy_name: {_eq: $proxy_name}}) {
    proxy_id
    proxy_name
    create_ts
    update_ts
  }
}
`

var QDeleteProxy string = `
mutation delete_proxy($proxy_name: String!) {
  delete_saaras_db_proxy(where: {proxy_name: {_eq: $proxy_name}}) {
    affected_rows
  }
}
`

var QCreateProxyService string = `
      mutation create_proxy_service($proxy_name : String!, $fqdn : String!, $service_name : String!) {
				# Create service
        insert_saaras_db_service
        (
          objects:
          {
            fqdn: $fqdn,
            service_name: $service_name
          } on_conflict: {constraint: service_service_name_key, update_columns:[fqdn, service_name]}
        )
        {
          returning
          {
            create_ts
          }
        }

        # Associate a service to a proxy
        insert_saaras_db_proxy_service(
          objects:
          {
            proxy:
            {
              data:
              {
                proxy_name: $proxy_name
              }, on_conflict: {constraint: proxy_proxy_name_key, update_columns: update_ts}
            },
            service:
            {
              data:
              {
                service_name: $service_name
              }, on_conflict: {constraint: service_service_name_key, update_columns: update_ts}
            }
          }
        )
        {
          affected_rows
        }
      }
`

var QGetProxyService string = `
	query get_proxy_service($proxy_name: String!) {
	  saaras_db_service(where: {proxy_services: {proxy: {proxy_name: {_eq: $proxy_name}}}}) {
	    service_id
	    service_name
	    fqdn
	    create_ts
	    update_ts
	  }
	}
`

var QGetProxyServiceAssociation string = `
query get_proxy_service($proxy_name: String!, $service_name: String!) {
  saaras_db_service(where: {_and: 
    {proxy_services: {proxy: {proxy_name: {_eq: $proxy_name}}}}, service_name: {_eq: $service_name}}) {
    service_id
    service_name
    fqdn
    create_ts
    update_ts
  }
}
`

var QDeleteProxyService = `

mutation delete_proxy_service($service_name: String!, $proxy_name: String!) {
  delete_saaras_db_proxy_service(where: {
		_and:
			{
				proxy: {proxy_name: {_eq: $proxy_name}}, 
				service: {service_name: {_eq: $service_name}}
			}
		}) {
    affected_rows
  }
  delete_saaras_db_service(where: {service_name: {_eq: $service_name}}) {
    affected_rows
  }
}

`

var QDeleteProxyServiceAssociation = `

mutation delete_proxy_service($service_name: String!, $proxy_name: String!) {
  delete_saaras_db_proxy_service(where: {
		_and:
			{
				proxy: {proxy_name: {_eq: $proxy_name}}, 
				service: {service_name: {_eq: $service_name}}
			}
		}) {
    affected_rows
  }
}
`

var QCreateProxyServiceAssociation = `
      mutation create_proxy_service($proxy_name : String!, $service_name : String!) {
        # Associate a service to a proxy
        insert_saaras_db_proxy_service(
          objects:
          {
            proxy:
            {
              data:
              {
                proxy_name: $proxy_name
              }, on_conflict: {constraint: proxy_proxy_name_key, update_columns: update_ts}
            },
            service:
            {
              data:
              {
                service_name: $service_name
              }, on_conflict: {constraint: service_service_name_key, update_columns: update_ts}
            }
          }
        )
        {
          affected_rows
        }
      }
`

var QGetAllProxyDetail = `
query get_proxy_detail {
  saaras_db_proxy {
    proxy_id
    proxy_name
    create_ts
    update_ts
    proxy_services {
      service {
        service_id
        service_name
        fqdn
        create_ts
        update_ts
        service_secrets {
          secret {
            secret_id
            secret_name
            secret_key
            secret_cert
            create_ts
            update_ts
          }
        }
        routes {
          route_id
          route_name
          route_prefix
          create_ts
          update_ts
          route_upstreams {
            upstream {
              upstream_id
              upstream_name
              upstream_ip
              upstream_port
              upstream_hc_healthythresholdcount
              upstream_hc_host
              upstream_hc_intervalseconds
              upstream_hc_path
              upstream_hc_timeoutseconds
              upstream_hc_unhealthythresholdcount
              upstream_strategy
              upstream_validation_cacertificate
              upstream_validation_subjectname
              upstream_weight
            }
          }
        }
      }
    }
  }
}
`

var QGetOneProxyDetail = `
query get_one_proxy_detail($proxy_name:String!) {
  saaras_db_proxy(where: {proxy_name: {_eq: $proxy_name}}) {
    proxy_id
    proxy_name
    create_ts
    update_ts
    proxy_services {
      service {
        service_id
        service_name
        fqdn
        create_ts
        update_ts
        service_secrets {
          secret {
            secret_id
            secret_name
            secret_key
            secret_cert
            create_ts
            update_ts
          }
        }
        routes {
          route_id
          route_name
          route_prefix
          create_ts
          update_ts
          route_upstreams {
            upstream {
              upstream_id
              upstream_name
              upstream_ip
              upstream_port
              upstream_hc_path
              upstream_hc_host
              upstream_hc_intervalseconds
              upstream_hc_timeoutseconds
              upstream_hc_unhealthythresholdcount
              upstream_hc_healthythresholdcount
              upstream_strategy
              upstream_validation_cacertificate
              upstream_validation_subjectname
              upstream_weight
    			  create_ts
    			  update_ts
            }
          }
        }
      }
    }
  }
}
`

// Read from DB_HOST and DB_PORT environment variables
var HOST string
var PORT string

var SECRET string

// @Summary Create a proxy
// @Description Create a proxy
// @Tags proxy
// @Accept  json
// @Produce  json
// @Param Name body webhttp.Proxy true "Name of proxy to create"
// @Success 201 {} integer OK
// @Router /proxy [post]
// @Security ApiKeyAuth
func POST_Proxy(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	p := new(Proxy)
	if err := c.Bind(p); err != nil {
		return err
	}

	if len(p.Name) == 0 {
		return c.JSON(http.StatusBadRequest, "Please provide name of proxy using Name field")
	}

	args["proxy_name"] = p.Name
	url := "http://" + HOST + ":" + PORT + "/v1/graphql"

	if err := saaras.RunDBQuery(url, QCreateProxy, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}
	return c.JSON(http.StatusCreated, p)
}

// @Summary List proxies
// @Description Get a list of all proxies
// @Tags proxy
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy [get]
// @Security ApiKeyAuth
func GET_Proxy(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	if err := saaras.RunDBQuery(url, QGetProxy, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// @Summary Get all proxy details
// @Description Get a detailed version of list of proxies
// @Tags proxy, operational-verbs
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/dump [get]
// @Security ApiKeyAuth
func GET_Proxy_Detail(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	if err := saaras.RunDBQuery(url, QGetAllProxyDetail, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// @Summary Get details of specified proxy
// @Description Get a detailed version of specified proxy
// @Tags proxy, operational-verbs
// @Accept  json
// @Produce  json
// @Param proxy_name path string true "Name of proxy for which to list services"
// @Success 200 {} integer OK
// @Router /proxy/dump/{proxy_name} [get]
// @Security ApiKeyAuth
func GET_One_Proxy_Detail(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	proxy_name := c.Param("proxy_name")
	args["proxy_name"] = proxy_name

	if err := saaras.RunDBQuery(url, QGetOneProxyDetail, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// @Summary List services associated with proxy
// @Description Get all services associated with a proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy for which to list services"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name}/service [get]
// @Security ApiKeyAuth
func GET_Proxy_Service(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")
	proxy_name := c.Param("proxy_name")

	args["proxy_name"] = proxy_name
	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	if err := saaras.RunDBQuery(url, QGetProxyService, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// @Summary Delete a proxy
// @Description Delete a proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy to delete"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name} [delete]
// @Security ApiKeyAuth
func DELETE_Proxy(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	proxy_name := c.Param("proxy_name")
	args["proxy_name"] = proxy_name

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	if err := saaras.RunDBQuery(url, QDeleteProxy, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())

}

// @Summary Get information about a proxy
// @Description Get information about a proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy to delete"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name} [get]
// @Security ApiKeyAuth
func GET_One_Proxy(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	proxy_name := c.Param("proxy_name")
	args["proxy_name"] = proxy_name

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"
	if err := saaras.RunDBQuery(url, QGetOneProxy, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())

}

// @Summary Associate a service with proxy
// @Description Associate a service with proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy for which to list service"
// @Param service_name path string true "Name of service to list"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name}/service/{service_name} [post]
// @Security ApiKeyAuth
func POST_Proxy_Service_Association(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	proxy_name := c.Param("proxy_name")
	service_name := c.Param("service_name")

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"

	args["proxy_name"] = proxy_name
	args["service_name"] = service_name

	if err := saaras.RunDBQuery(url, QCreateProxyServiceAssociation, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}
	return c.JSONBlob(http.StatusCreated, buf.Bytes())
}

// @Summary Disassociate a service from proxy
// @Description Disassociate a service from proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy for which to list service"
// @Param service_name path string true "Name of service to list"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name}/service/{service_name} [delete]
// @Security ApiKeyAuth
func DELETE_Proxy_Service_Association(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")

	proxy_name := c.Param("proxy_name")
	service_name := c.Param("service_name")

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"

	args["proxy_name"] = proxy_name
	args["service_name"] = service_name

	if err := saaras.RunDBQuery(url, QDeleteProxyServiceAssociation, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}
	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// @Summary Return specified service associated with this proxy
// @Description Return specified service associated with this proxy
// @Tags proxy
// @Param proxy_name path string true "Name of proxy for which to list service"
// @Param service_name path string true "Name of service to list"
// @Accept  json
// @Produce  json
// @Success 200 {} integer OK
// @Router /proxy/{proxy_name}/service/{service_name} [get]
// @Security ApiKeyAuth
func GET_Proxy_Service_Association(c echo.Context) error {
	var buf bytes.Buffer
	var args map[string]string
	args = make(map[string]string)

	log2 := logrus.StandardLogger()
	log := log2.WithField("context", "web-http")
	proxy_name := c.Param("proxy_name")
	service_name := c.Param("service_name")

	args["proxy_name"] = proxy_name
	args["service_name"] = service_name

	url := "http://" + HOST + ":" + PORT + "/v1/graphql"

	if err := saaras.RunDBQuery(url, QGetProxyServiceAssociation, &buf, args, log); err != nil {
		log.Errorf("Error when running http request [%v]\n", err)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

func GET_Health_Check(c echo.Context) error {
	var buf bytes.Buffer

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

func Add_proxy_routes(e *echo.Echo) {
	// Proxy CRUD
	e.GET("/proxy", GET_Proxy)
	e.POST("/proxy", POST_Proxy)
	e.DELETE("/proxy/:proxy_name", DELETE_Proxy)
	e.GET("/proxy/:proxy_name", GET_One_Proxy)

	// Proxy to Service association with implied service CRUD
	// Only the GET makes sense here?
	e.GET("/proxy/:proxy_name/service", GET_Proxy_Service)

	// Proxy to Service association
	e.POST("/proxy/:proxy_name/service/:service_name", POST_Proxy_Service_Association)
	e.GET("/proxy/:proxy_name/service/:service_name", GET_Proxy_Service_Association)
	e.DELETE("/proxy/:proxy_name/service/:service_name", DELETE_Proxy_Service_Association)

	// Support for operational-verbs
	e.GET("/proxy/dump", GET_Proxy_Detail)
	e.GET("/proxy/dump/:proxy_name", GET_One_Proxy_Detail)

	e.GET("/health", GET_Health_Check)
}
