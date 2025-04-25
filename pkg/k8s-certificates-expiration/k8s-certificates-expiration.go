package k8s_certificates_expiration

import (
	"errors"
	"time"
)

type KubernetesCertificatesExpiration struct {
	lastUpdated          time.Time
	nodesCertsExpiration []NodeCertsExpiration
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
		lastUpdated:          time.Now(),
		nodesCertsExpiration: []NodeCertsExpiration{},
	}
}

func (c *KubernetesCertificatesExpiration) GetLastUpdated() time.Time {
	return c.lastUpdated
}

func (c *KubernetesCertificatesExpiration) GetNodeCertsExpiration(nodeName string) NodeCertsExpiration {
	for _, nodeCertsExpiration := range c.nodesCertsExpiration {
		if nodeCertsExpiration.GetNodeName() == nodeName {
			return nodeCertsExpiration
		}
	}
	return NodeCertsExpiration{}
}

func NewNodeCertsExpiration(nodeName string) *NodeCertsExpiration {
	return &NodeCertsExpiration{}
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
	for _, cert := range n.certificateExpiration {
		expireTime, err := time.Parse(time.RFC3339, cert.Expires)
		if err == nil && time.Now().After(expireTime) {
			return true
		}
	}
	return false
}

func (n *NodeCertsExpiration) HasExpiredCAs() bool {
	for _, ca := range n.certificatesAuthority {
		expireTime, err := time.Parse(time.RFC3339, ca.Expires)
		if err == nil && time.Now().After(expireTime) {
			return true
		}
	}
	return false
}
