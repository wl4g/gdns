/**
 * Copyright 2017 ~ 2025 the original author or authors[983708408@qq.com].
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package xcloud_dopaas_coredns

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/pkg/log"
	redisCon "github.com/go-redis/redis/v7"
)

var clusterClient *redisCon.ClusterClient

func TestRedisCollector(t *testing.T) {
	fmt.Println("Testing collector starting ...")
	clusterClient = redisCon.NewClusterClient(&redisCon.ClusterOptions{
		Addrs: []string{ // 填写master主机
			"10.0.0.160:6379", "10.0.0.160:6380", "10.0.0.160:6381", "10.0.0.162:6379", "10.0.0.162:6380", "10.0.0.162:6381",
		},
		Password:     "zzx!@#$%",      // 设置密码
		DialTimeout:  5 * time.Second, // 设置连接超时
		ReadTimeout:  5 * time.Second, // 设置读取超时
		WriteTimeout: 5 * time.Second, // 设置写入超时
	})

	hget := clusterClient.HGet("_dns:heweijie.top", "host").Val()
	fmt.Println(hget)

	vals := clusterClient.HKeys("_dns:heweijie.top.").Val() //_dns:heweijie.top.
	fmt.Println(vals)

	smembers := clusterClient.SMembers("_dns_blacklist").Val()
	fmt.Println(smembers)

	hGetAll := clusterClient.HGetAll("_dns:shangmai.com.").Val()
	z := new(Zone)
	z.Name = "shangmai.com"
	z.Locations = make(map[string]*Record)
	for key, val := range hGetAll {
		r := new(Record)
		err := json.Unmarshal([]byte(val), r)
		if err != nil {
			log.Error("parse config error ", val, err)
			continue
		}
		z.Locations[key] = r
	}
	fmt.Println(z)

}

func TestQname2Zone(t *testing.T) {
	s := Qname2Zone("host.heweijie.top.")
	fmt.Println(s)
	s = Qname2Zone("heweijie.com.cn.")
	fmt.Println(s)
	s = Qname2Zone("heweijie.top.")
	fmt.Println(s)
}

func TestSim(t *testing.T) {
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "*baidu.com"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "baidu.com"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "*.baidu.com"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "fanyi.*.com"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "*.baidu.*"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "*.ba*u.*"))
	fmt.Println(ExpressionMatch("fanyi.baidu.com", "*baidu*"))
}

func TestLog(t *testing.T) {
	Error("adsf")
}
