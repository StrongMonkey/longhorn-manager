package api

import (
	//"bytes"
	//"encoding/json"
	//"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/rancher/longhorn-manager/manager"
)

type NodeIDFunc func(req *http.Request) (string, error)

func NodeIDFromVolume(m *manager.VolumeManager) func(req *http.Request) (string, error) {
	return func(req *http.Request) (string, error) {
		name := mux.Vars(req)["name"]
		volume, err := m.GetVolume(name)
		if err != nil {
			return "", errors.Wrapf(err, "error getting volume '%s'", name)
		}
		if volume == nil {
			return "", nil
		}
		return volume.NodeID, nil
	}
}

type NodeLocator interface {
	GetCurrentNodeID() string
	Node2APIAddress(nodeID string) (string, error)
}

type Fwd struct {
	locator NodeLocator
	proxy   http.Handler
}

func NewFwd(locator NodeLocator) *Fwd {
	return &Fwd{
		locator: locator,
		proxy:   &httputil.ReverseProxy{Director: func(r *http.Request) {}},
	}
}

func (f *Fwd) Handler(getNodeID NodeIDFunc, h HandleFuncWithError) HandleFuncWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		//nodeID, err := getNodeID(copyReq(req))
		nodeID, err := getNodeID(req)
		if err != nil {
			return errors.Wrap(err, "fail to get node ID")
		}
		if nodeID != "" && nodeID != f.locator.GetCurrentNodeID() {
			targetNode, err := f.locator.Node2APIAddress(nodeID)
			if err != nil {
				return errors.Wrapf(err, "cannot find node %v", nodeID)
			}
			if targetNode != req.Host {
				req.Host = targetNode
				req.URL.Host = targetNode
				req.URL.Scheme = "http"
				logrus.Debugf("Forwarding request to %v", targetNode)
				f.proxy.ServeHTTP(w, req)
				return nil
			}
		}
		return h(w, req)
	}
}

//func copyReq(req *http.Request) *http.Request {
//	r := *req
//	buf, _ := ioutil.ReadAll(r.Body)
//	req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
//	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
//	return &r
//}
