package k8s_certificates_expiration

import (
	"errors"
	"time"
)

type KubernetesCertificatesExpiration struct {
	lastUpdated         time.Time
	nodeCertsExpiration NodeCertsExpiration
}
type NodeCertsExpiration struct {
	nodeName              string
	missingCertificate    []string
	certificateExpiration []CertsExpiration
	certificatesAuthority []CertsAuthority
}

type CertsExpiration struct {
	Name                 string
	Expires              string
	ResidualTime         string
	CertificateAuthority string
	ExternallyManaged    string
}

type CertsAuthority struct {
	Name              string
	Expires           string
	ResidualTime      string
	ExternallyManaged string
}

func NewKubernetesCertificatesExpiration() *KubernetesCertificatesExpiration {
	return &KubernetesCertificatesExpiration{
		lastUpdated: time.Now(),
	}
}

// UpdateNodeCertificates updates the node certificates information
func (c *KubernetesCertificatesExpiration) UpdateNodeCertificates(node NodeCertsExpiration) {
	c.nodeCertsExpiration = node
	c.lastUpdated = time.Now()
}

// GetLastUpdated returns the last time the certificates were updated
func (c *KubernetesCertificatesExpiration) GetLastUpdated() time.Time {
	return c.lastUpdated
}

func (c *KubernetesCertificatesExpiration) GetNodeCertsExpiration() NodeCertsExpiration {
	return c.nodeCertsExpiration
}

// NewNodeCertsExpiration creates a new NodeCertsExpiration instance
func NewNodeCertsExpiration(nodeName string) *NodeCertsExpiration {
	return &NodeCertsExpiration{
		nodeName:              nodeName,
		missingCertificate:    make([]string, 0),
		certificateExpiration: make([]CertsExpiration, 0),
		certificatesAuthority: make([]CertsAuthority, 0),
	}
}

func (n *NodeCertsExpiration) AddMissingCertificate(certName string) {
	n.missingCertificate = append(n.missingCertificate, certName)
}

func (n *NodeCertsExpiration) AddCertificateExpiration(cert CertsExpiration) {
	n.certificateExpiration = append(n.certificateExpiration, cert)
}

func (n *NodeCertsExpiration) AddCertificateAuthority(ca CertsAuthority) {
	n.certificatesAuthority = append(n.certificatesAuthority, ca)
}

func (n *NodeCertsExpiration) GetNodeName() string {
	return n.nodeName
}

func (n *NodeCertsExpiration) GetMissingCertificates() []string {
	return n.missingCertificate
}

func (n *NodeCertsExpiration) GetCertificateExpirations() []CertsExpiration {
	return n.certificateExpiration
}

func (n *NodeCertsExpiration) GetCertificatesAuthority() []CertsAuthority {
	return n.certificatesAuthority
}

func (n *NodeCertsExpiration) FindCertificateByName(name string) (*CertsExpiration, error) {
	for _, cert := range n.certificateExpiration {
		if cert.Name == name {
			return &cert, nil
		}
	}
	return nil, errors.New("certificate not found")
}

func (n *NodeCertsExpiration) FindCertificateAuthorityByName(name string) (*CertsAuthority, error) {
	for _, ca := range n.certificatesAuthority {
		if ca.Name == name {
			return &ca, nil
		}
	}
	return nil, errors.New("certificate authority not found")
}

func (n *NodeCertsExpiration) HasExpiredCertificates() bool {
	currentTime := time.Now()
	for _, cert := range n.certificateExpiration {
		expireTime, err := time.Parse(time.RFC3339, cert.Expires)
		if err == nil && currentTime.After(expireTime) {
			return true
		}
	}
	return false
}

func (n *NodeCertsExpiration) HasExpiredCAs() bool {
	currentTime := time.Now()
	for _, ca := range n.certificatesAuthority {
		expireTime, err := time.Parse(time.RFC3339, ca.Expires)
		if err == nil && currentTime.After(expireTime) {
			return true
		}
	}
	return false
}
