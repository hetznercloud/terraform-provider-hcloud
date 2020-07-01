package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// LoadBalancer represents a Load Balancer in the Hetzner Cloud.
type LoadBalancer struct {
	ID               int
	Name             string
	PublicNet        LoadBalancerPublicNet
	PrivateNet       []LoadBalancerPrivateNet
	Location         *Location
	LoadBalancerType *LoadBalancerType
	Algorithm        LoadBalancerAlgorithm
	Services         []LoadBalancerService
	Targets          []LoadBalancerTarget
	Protection       LoadBalancerProtection
	Labels           map[string]string
	Created          time.Time
}

// LoadBalancerPublicNet represents a Load Balancer's public network.
type LoadBalancerPublicNet struct {
	Enabled bool
	IPv4    LoadBalancerPublicNetIPv4
	IPv6    LoadBalancerPublicNetIPv6
}

// LoadBalancerPublicNetIPv4 represents a Load Balancer's public IPv4 address.
type LoadBalancerPublicNetIPv4 struct {
	IP net.IP
}

// LoadBalancerPublicNetIPv6 represents a Load Balancer's public IPv6 address.
type LoadBalancerPublicNetIPv6 struct {
	IP net.IP
}

// LoadBalancerPrivateNet represents a Load Balancer's private network.
type LoadBalancerPrivateNet struct {
	Network *Network
	IP      net.IP
}

// LoadBalancerService represents a Load Balancer service.
type LoadBalancerService struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      int
	DestinationPort int
	Proxyprotocol   bool
	HTTP            LoadBalancerServiceHTTP
	HealthCheck     LoadBalancerServiceHealthCheck
}

// LoadBalancerServiceHTTP stores configuration for a service using the HTTP protocol.
type LoadBalancerServiceHTTP struct {
	CookieName     string
	CookieLifetime time.Duration
	Certificates   []*Certificate
	RedirectHTTP   bool
	StickySessions bool
}

// LoadBalancerServiceHealthCheck stores configuration for a service health check.
type LoadBalancerServiceHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     int
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
	HTTP     *LoadBalancerServiceHealthCheckHTTP
}

// LoadBalancerServiceHealthCheckHTTP stores configuration for a service health check
// using the HTTP protocol.
type LoadBalancerServiceHealthCheckHTTP struct {
	Domain      string
	Path        string
	Response    string
	StatusCodes []string
	TLS         bool
}

// LoadBalancerAlgorithmType specifies the algorithm type a Load Balancer
// uses for distributing requests.
type LoadBalancerAlgorithmType string

const (
	// LoadBalancerAlgorithmTypeRoundRobin is an algorithm which distributes
	// requests to targets in a round robin fashion.
	LoadBalancerAlgorithmTypeRoundRobin LoadBalancerAlgorithmType = "round_robin"
	// LoadBalancerAlgorithmTypeLeastConnections is an algorithm which distributes
	// requests to targets with the least number of connections.
	LoadBalancerAlgorithmTypeLeastConnections LoadBalancerAlgorithmType = "least_connections"
)

// LoadBalancerAlgorithm configures the algorithm a Load Balancer uses
// for distributing requests.
type LoadBalancerAlgorithm struct {
	Type LoadBalancerAlgorithmType
}

// LoadBalancerTargetType specifies the type of a Load Balancer target.
type LoadBalancerTargetType string

const (
	// LoadBalancerTargetTypeServer is a target type which points to a specific server.
	LoadBalancerTargetTypeServer LoadBalancerTargetType = "server"
)

// LoadBalancerServiceProtocol specifies the protocol of a Load Balancer service.
type LoadBalancerServiceProtocol string

const (
	// LoadBalancerServiceProtocolTCP specifies a TCP service.
	LoadBalancerServiceProtocolTCP LoadBalancerServiceProtocol = "tcp"
	// LoadBalancerServiceProtocolHTTP specifies an HTTP service.
	LoadBalancerServiceProtocolHTTP LoadBalancerServiceProtocol = "http"
	// LoadBalancerServiceProtocolHTTPS specifies an HTTPS service.
	LoadBalancerServiceProtocolHTTPS LoadBalancerServiceProtocol = "https"
)

// LoadBalancerTarget represents a Load Balancer target.
type LoadBalancerTarget struct {
	Type         LoadBalancerTargetType
	Server       *LoadBalancerTargetServer
	HealthStatus []LoadBalancerTargetHealthStatus
	Targets      []LoadBalancerTarget
	UsePrivateIP bool
}

// LoadBalancerTargetServer configures a Load Balancer target
// pointing to a specific server.
type LoadBalancerTargetServer struct {
	Server *Server
}

// LoadBalancerTargetHealthStatusStatus describes a target's health status.
type LoadBalancerTargetHealthStatusStatus string

const (
	// LoadBalancerTargetHealthStatusStatusUnknown denotes that the health status is unknown.
	LoadBalancerTargetHealthStatusStatusUnknown LoadBalancerTargetHealthStatusStatus = "unknown"
	// LoadBalancerTargetHealthStatusStatusHealthy denotes a healthy target.
	LoadBalancerTargetHealthStatusStatusHealthy LoadBalancerTargetHealthStatusStatus = "healthy"
	// LoadBalancerTargetHealthStatusStatusUnhealthy denotes an unhealthy target.
	LoadBalancerTargetHealthStatusStatusUnhealthy LoadBalancerTargetHealthStatusStatus = "unhealthy"
)

// LoadBalancerTargetHealthStatus describes a target's health for a specific service.
type LoadBalancerTargetHealthStatus struct {
	ListenPort int
	Status     LoadBalancerTargetHealthStatusStatus
}

// LoadBalancerProtection represents the protection level of a Load Balancer.
type LoadBalancerProtection struct {
	Delete bool
}

// LoadBalancerClient is a client for the Load Balancers API.
type LoadBalancerClient struct {
	client *Client
}

// GetByID retrieves a Load Balancer by its ID. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) GetByID(ctx context.Context, id int) (*LoadBalancer, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/load_balancers/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.LoadBalancerGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return LoadBalancerFromSchema(body.LoadBalancer), resp, nil
}

// GetByName retrieves a Load Balancer by its name. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) GetByName(ctx context.Context, name string) (*LoadBalancer, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	LoadBalancer, response, err := c.List(ctx, LoadBalancerListOpts{Name: name})
	if len(LoadBalancer) == 0 {
		return nil, response, err
	}
	return LoadBalancer[0], response, err
}

// Get retrieves a Load Balancer by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Load Balancer by its name. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) Get(ctx context.Context, idOrName string) (*LoadBalancer, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, int(id))
	}
	return c.GetByName(ctx, idOrName)
}

// LoadBalancerListOpts specifies options for listing Load Balancers.
type LoadBalancerListOpts struct {
	ListOpts
	Name string
}

func (l LoadBalancerListOpts) values() url.Values {
	vals := l.ListOpts.values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	return vals
}

// List returns a list of Load Balancers for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *LoadBalancerClient) List(ctx context.Context, opts LoadBalancerListOpts) ([]*LoadBalancer, *Response, error) {
	path := "/load_balancers?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.LoadBalancerListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	LoadBalancers := make([]*LoadBalancer, 0, len(body.LoadBalancers))
	for _, s := range body.LoadBalancers {
		LoadBalancers = append(LoadBalancers, LoadBalancerFromSchema(s))
	}
	return LoadBalancers, resp, nil
}

// All returns all Load Balancers.
func (c *LoadBalancerClient) All(ctx context.Context) ([]*LoadBalancer, error) {
	allLoadBalancer := []*LoadBalancer{}

	opts := LoadBalancerListOpts{}
	opts.PerPage = 50

	_, err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		LoadBalancer, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allLoadBalancer = append(allLoadBalancer, LoadBalancer...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allLoadBalancer, nil
}

// AllWithOpts returns all Load Balancers for the given options.
func (c *LoadBalancerClient) AllWithOpts(ctx context.Context, opts LoadBalancerListOpts) ([]*LoadBalancer, error) {
	var allLoadBalancers []*LoadBalancer

	_, err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		LoadBalancers, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allLoadBalancers = append(allLoadBalancers, LoadBalancers...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allLoadBalancers, nil
}

// LoadBalancerUpdateOpts specifies options for updating a Load Balancer.
type LoadBalancerUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a Load Balancer.
func (c *LoadBalancerClient) Update(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerUpdateOpts) (*LoadBalancer, *Response, error) {
	reqBody := schema.LoadBalancerUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.LoadBalancerUpdateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return LoadBalancerFromSchema(respBody.LoadBalancer), resp, nil
}

// LoadBalancerCreateOpts specifies options for creating a new Load Balancer.
type LoadBalancerCreateOpts struct {
	Name             string
	LoadBalancerType *LoadBalancerType
	Algorithm        *LoadBalancerAlgorithm
	Location         *Location
	NetworkZone      NetworkZone
	Labels           map[string]string
	Targets          []LoadBalancerCreateOptsTarget
	Services         []LoadBalancerCreateOptsService
	PublicInterface  *bool
	Network          *Network
}

// LoadBalancerCreateOptsTarget holds options for specifying a target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTarget struct {
	Type         LoadBalancerTargetType
	Server       LoadBalancerCreateOptsTargetServer
	UsePrivateIP *bool
}

// LoadBalancerCreateOptsTargetServer holds options for specifying a server target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTargetServer struct {
	Server *Server
}

// LoadBalancerCreateOptsService holds options for specifying a service
// when creating a new Load Balancer.
type LoadBalancerCreateOptsService struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      *int
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerCreateOptsServiceHTTP
	HealthCheck     *LoadBalancerCreateOptsServiceHealthCheck
}

// LoadBalancerCreateOptsServiceHTTP holds options for specifying an HTTP service
// when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerCreateOptsServiceHealthCheck holds options for specifying a service
// health check when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerCreateOptsServiceHealthCheckHTTP
}

// LoadBalancerCreateOptsServiceHealthCheckHTTP holds options for specifying a service
// HTTP health check when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// LoadBalancerCreateResult is the result of a create Load Balancer call.
type LoadBalancerCreateResult struct {
	LoadBalancer *LoadBalancer
	Action       *Action
}

// Create creates a new Load Balancer.
func (c *LoadBalancerClient) Create(ctx context.Context, opts LoadBalancerCreateOpts) (LoadBalancerCreateResult, *Response, error) {
	reqBody := loadBalancerCreateOptsToSchema(opts)
	reqBodyData, err := json.Marshal(reqBody)

	if err != nil {
		return LoadBalancerCreateResult{}, nil, err
	}
	req, err := c.client.NewRequest(ctx, "POST", "/load_balancers", bytes.NewReader(reqBodyData))
	if err != nil {
		return LoadBalancerCreateResult{}, nil, err
	}

	respBody := schema.LoadBalancerCreateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return LoadBalancerCreateResult{}, resp, err
	}
	return LoadBalancerCreateResult{
		LoadBalancer: LoadBalancerFromSchema(respBody.LoadBalancer),
		Action:       ActionFromSchema(respBody.Action),
	}, resp, nil
}

// Delete deletes a Load Balancer.
func (c *LoadBalancerClient) Delete(ctx context.Context, loadBalancer *LoadBalancer) (*Response, error) {
	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("/load_balancers/%d", loadBalancer.ID), nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req, nil)
}

func (c *LoadBalancerClient) addTarget(ctx context.Context, loadBalancer *LoadBalancer, reqBody schema.LoadBalancerActionAddTargetRequest) (*Action, *Response, error) {
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/add_target", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.LoadBalancerActionAddTargetResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

func (c *LoadBalancerClient) removeTarget(ctx context.Context, loadBalancer *LoadBalancer, reqBody schema.LoadBalancerActionRemoveTargetRequest) (*Action, *Response, error) {
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/remove_target", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.LoadBalancerActionRemoveTargetResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerAddServerTargetOpts specifies options for adding a server target
// to a Load Balancer.
type LoadBalancerAddServerTargetOpts struct {
	Server       *Server
	UsePrivateIP *bool
}

// AddServerTarget adds a server target to a Load Balancer.
func (c *LoadBalancerClient) AddServerTarget(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddServerTargetOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionAddTargetRequest{
		Type: string(LoadBalancerTargetTypeServer),
		Server: &schema.LoadBalancerActionAddTargetRequestServer{
			ID: opts.Server.ID,
		},
		UsePrivateIP: opts.UsePrivateIP,
	}
	return c.addTarget(ctx, loadBalancer, reqBody)
}

// RemoveServerTarget removes a server target from a Load Balancer.
func (c *LoadBalancerClient) RemoveServerTarget(ctx context.Context, loadBalancer *LoadBalancer, server *Server) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionRemoveTargetRequest{
		Type: string(LoadBalancerTargetTypeServer),
		Server: &schema.LoadBalancerActionRemoveTargetRequestServer{
			ID: server.ID,
		},
	}
	return c.removeTarget(ctx, loadBalancer, reqBody)
}

// LoadBalancerAddServiceOpts specifies options for adding a service to a Load Balancer.
type LoadBalancerAddServiceOpts struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      *int
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerAddServiceOptsHTTP
	HealthCheck     *LoadBalancerAddServiceOptsHealthCheck
}

// LoadBalancerAddServiceOptsHTTP holds options for specifying an HTTP service
// when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerAddServiceOptsHealthCheck holds options for specifying an health check
// when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerAddServiceOptsHealthCheckHTTP
}

// LoadBalancerAddServiceOptsHealthCheckHTTP holds options for specifying an
// HTTP health check when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// AddService adds a service to a Load Balancer.
func (c *LoadBalancerClient) AddService(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddServiceOpts) (*Action, *Response, error) {
	reqBody := loadBalancerAddServiceOptsToSchema(opts)
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/add_service", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.LoadBalancerActionAddServiceResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerUpdateServiceOpts specifies options for updating a service.
type LoadBalancerUpdateServiceOpts struct {
	Protocol        LoadBalancerServiceProtocol
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerUpdateServiceOptsHTTP
	HealthCheck     *LoadBalancerUpdateServiceOptsHealthCheck
}

// LoadBalancerUpdateServiceOptsHTTP specifies options for updating an HTTP(S) service.
type LoadBalancerUpdateServiceOptsHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerUpdateServiceOptsHealthCheck specifies options for updating
// a service's health check.
type LoadBalancerUpdateServiceOptsHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerUpdateServiceOptsHealthCheckHTTP
}

// LoadBalancerUpdateServiceOptsHealthCheckHTTP specifies options for updating
// the HTTP-specific settings of a service's health check.
type LoadBalancerUpdateServiceOptsHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// UpdateService updates a Load Balancer service.
func (c *LoadBalancerClient) UpdateService(ctx context.Context, loadBalancer *LoadBalancer, listenPort int, opts LoadBalancerUpdateServiceOpts) (*Action, *Response, error) {
	reqBody := loadBalancerUpdateServiceOptsToSchema(opts)
	reqBody.ListenPort = listenPort
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/update_service", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.LoadBalancerActionUpdateServiceResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// DeleteService deletes a Load Balancer service.
func (c *LoadBalancerClient) DeleteService(ctx context.Context, loadBalancer *LoadBalancer, listenPort int) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerDeleteServiceRequest{
		ListenPort: listenPort,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/delete_service", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.LoadBalancerDeleteServiceResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerChangeProtectionOpts specifies options for changing the resource protection level of a Load Balancer.
type LoadBalancerChangeProtectionOpts struct {
	Delete *bool
}

// ChangeProtection changes the resource protection level of a Load Balancer.
func (c *LoadBalancerClient) ChangeProtection(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerChangeProtectionOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionChangeProtectionRequest{
		Delete: opts.Delete,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/change_protection", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.LoadBalancerActionChangeProtectionResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}

// LoadBalancerChangeAlgorithmOpts specifies options for changing the algorithm of a Load Balancer.
type LoadBalancerChangeAlgorithmOpts struct {
	Type LoadBalancerAlgorithmType
}

// ChangeAlgorithm changes the algorithm of a Load Balancer.
func (c *LoadBalancerClient) ChangeAlgorithm(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerChangeAlgorithmOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionChangeAlgorithmRequest{
		Type: string(opts.Type),
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/change_algorithm", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.LoadBalancerActionChangeAlgorithmResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}

// LoadBalancerAttachToNetworkOpts specifies options for attaching a Load Balancer to a network.
type LoadBalancerAttachToNetworkOpts struct {
	Network *Network
	IP      net.IP
}

// AttachToNetwork attaches a Load Balancer to a network.
func (c *LoadBalancerClient) AttachToNetwork(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAttachToNetworkOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionAttachToNetworkRequest{
		Network: opts.Network.ID,
	}
	if opts.IP != nil {
		reqBody.IP = String(opts.IP.String())
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/attach_to_network", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.LoadBalancerActionAttachToNetworkResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}

// LoadBalancerDetachFromNetworkOpts specifies options for detaching a Load Balancer from a network.
type LoadBalancerDetachFromNetworkOpts struct {
	Network *Network
}

// DetachFromNetwork detaches a Load Balancer from a network.
func (c *LoadBalancerClient) DetachFromNetwork(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerDetachFromNetworkOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionDetachFromNetworkRequest{
		Network: opts.Network.ID,
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/load_balancers/%d/actions/detach_from_network", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.LoadBalancerActionDetachFromNetworkResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}

// EnablePublicInterface enables the Load Balancer's public network interface.
func (c *LoadBalancerClient) EnablePublicInterface(ctx context.Context, loadBalancer *LoadBalancer) (*Action, *Response, error) {
	path := fmt.Sprintf("/load_balancers/%d/actions/enable_public_interface", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	respBody := schema.LoadBalancerActionEnablePublicInterfaceResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}

// DisablePublicInterface disables the Load Balancer's public network interface.
func (c *LoadBalancerClient) DisablePublicInterface(ctx context.Context, loadBalancer *LoadBalancer) (*Action, *Response, error) {
	path := fmt.Sprintf("/load_balancers/%d/actions/disable_public_interface", loadBalancer.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, nil, err
	}
	respBody := schema.LoadBalancerActionDisablePublicInterfaceResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, err
}
