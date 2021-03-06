// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/northwesternmutual/kanali/spec"
	"github.com/spf13/viper"
)

const (
	k8sNameMaxSize = 253
)

func init() {
	if level, err := logrus.ParseLevel(viper.GetString("log-level")); err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(level)
	}
}

// StartUDPServer will start the udp server that is used to comminute between
// all running Kanali instances.
func StartUDPServer() (e error) {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", viper.GetInt("peer-udp-port")))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	logrus.Infof("upd server listening on :%s", viper.GetString("peer-udp-port"))
	defer func() {
		if err := conn.Close(); err != nil {
			if e != nil {
				e = err
			}
		}
	}()

	// [NAMESPACE],[PROXYNAME],[KEYNAME] <= 761
	buf := make([]byte, k8sNameMaxSize*3+2)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		trafficData := strings.Split(string(buf[0:n]), ",")
		spec.TrafficStore.AddTraffic(trafficData[0], trafficData[1], trafficData[2], time.Now())

	}

}

// Emit will send a message to all other Kanali instances.
func Emit(binding spec.APIKeyBinding, keyName string, currTime time.Time) {

	for _, addr := range spec.KanaliEndpoints.Subsets[0].Addresses {

		if os.Getenv("POD_IP") == addr.IP {
			spec.TrafficStore.AddTraffic(binding.ObjectMeta.Namespace, binding.Spec.APIProxyName, keyName, currTime)
			continue
		}

		go func(ip string) {

			serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, viper.GetInt("peer-udp-port")))
			if err != nil {
				logrus.Warnf("error resolving UDP address for %s:%d", ip, viper.GetInt("peer-udp-port"))
				return
			}

			conn, err := net.DialUDP("udp", nil, serverAddr)
			if err != nil {
				logrus.Warnf("error dialing %s:%d", ip, viper.GetInt("peer-udp-port"))
				return
			}

			_, err = conn.Write([]byte(fmt.Sprintf("%s,%s,%s", binding.ObjectMeta.Namespace, binding.Spec.APIProxyName, keyName)))
			if err != nil {
				logrus.Warnf("error writing traffic to %s:%d", ip, viper.GetString("peer-udp-port"))
				return
			}

			if err := conn.Close(); err != nil {
				logrus.Errorf(err.Error())
			}

		}(addr.IP)

	}

}
