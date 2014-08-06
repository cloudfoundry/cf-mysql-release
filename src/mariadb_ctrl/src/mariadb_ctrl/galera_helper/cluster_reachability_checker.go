package galera_helper

import (
	"strings"
	"net/http"
)

var MakeRequest = http.Get

type Logger interface {
	Log(message string)
}

type ClusterReachabilityChecker interface {
	AnyNodesReachable() (bool)
}

type httpClusterReachabilityChecker struct{
	clusterIps []string
	logger Logger
}

func NewClusterReachabilityChecker(ips string, logger Logger) ClusterReachabilityChecker {
	return httpClusterReachabilityChecker{
		clusterIps:  strings.Split(ips, ","),
		logger: logger,
	}
}

func (h httpClusterReachabilityChecker) AnyNodesReachable() (bool) {
	for _, ip := range h.clusterIps {
		h.logger.Log("Checking if node is reachable: " + ip + "\n")

		resp, _ := MakeRequest("http://"+ip+":9200/")
		if resp != nil && resp.StatusCode == 200 {
			h.logger.Log("At least one node in cluster is reachable.\n")
			return true
		}
	}

	h.logger.Log("No nodes in cluster are reachable.\n")
	return false
}
