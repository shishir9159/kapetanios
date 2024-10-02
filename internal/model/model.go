package model

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type Cluster struct {
	Nodes    v1.NodeList `nodes:"type:json"`
	Validity time.Time   `expiration-date:"type:json"`
}
