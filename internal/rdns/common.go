package rdns

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

const IDFormat = "$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS"

type IDResourcePrefix string

const (
	IDResourcePrefixServer IDResourcePrefix = "s"
	IDResourcePrimaryIP    IDResourcePrefix = "p"
	IDResourceFloatingIP   IDResourcePrefix = "f"
	IDResourceLoadBalancer IDResourcePrefix = "l"
)

// ParseID parses the terraform RDNS ID "$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS".
func ParseID(value string) (hcloud.RDNSSupporter, net.IP, error) {
	if value == "" {
		return nil, nil, util.NewInvalidIDError(value, IDFormat)
	}

	parts := strings.SplitN(value, "-", 3)
	if len(parts) != 3 {
		return nil, nil, util.NewInvalidIDError(value, IDFormat)
	}

	// Parse $RESOURCE_ID
	resID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, nil, util.NewInvalidIDError(value, IDFormat).WithHint("is $RESOURCE_ID valid?")
	}

	// Parse $RESOURCE_PREFIX
	var rdns hcloud.RDNSSupporter
	switch IDResourcePrefix(parts[0]) {
	case IDResourcePrefixServer:
		rdns = &hcloud.Server{ID: resID}
	case IDResourcePrimaryIP:
		rdns = &hcloud.PrimaryIP{ID: resID}
	case IDResourceFloatingIP:
		rdns = &hcloud.FloatingIP{ID: resID}
	case IDResourceLoadBalancer:
		rdns = &hcloud.LoadBalancer{ID: resID}
	default:
		return nil, nil, util.NewInvalidIDError(value, IDFormat).WithHint("is $RESOURCE_PREFIX valid?")
	}

	// Parse $IP_ADDRESS
	ip := net.ParseIP(parts[2])
	if ip == nil {
		return nil, nil, util.NewInvalidIDError(value, IDFormat).WithHint("is $IP_ADDRESS valid?")
	}

	return rdns, ip, nil
}

func FormatID(rdns hcloud.RDNSSupporter, ip net.IP) string {
	switch v := rdns.(type) {
	case *hcloud.Server:
		return fmt.Sprintf("%s-%d-%s", IDResourcePrefixServer, v.ID, ip)
	case *hcloud.PrimaryIP:
		return fmt.Sprintf("%s-%d-%s", IDResourcePrimaryIP, v.ID, ip)
	case *hcloud.FloatingIP:
		return fmt.Sprintf("%s-%d-%s", IDResourceFloatingIP, v.ID, ip)
	case *hcloud.LoadBalancer:
		return fmt.Sprintf("%s-%d-%s", IDResourceLoadBalancer, v.ID, ip)
	default:
		return ""
	}
}
