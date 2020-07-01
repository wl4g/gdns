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
package redis

import (
	"fmt"
	redisCon "github.com/go-redis/redis/v7"
	"testing"
	"time"
)

var clusterClient *redisCon.ClusterClient

func TestRedisCollector(t *testing.T) {
	fmt.Printf("Testing %s collector starting ...", "redis")
	clusterClient = redisCon.NewClusterClient(&redisCon.ClusterOptions{
		Addrs: []string{ // 填写master主机
			"10.0.0.160:6379","10.0.0.160:6380","10.0.0.160:6381","10.0.0.162:6379","10.0.0.162:6380","10.0.0.162:6381",
		},
		Password:     "zzx!@#$%",              // 设置密码
		DialTimeout:  5 * time.Second, // 设置连接超时
		ReadTimeout:  5 * time.Second, // 设置读取超时
		WriteTimeout: 5 * time.Second, // 设置写入超时
	})

	// 发送一个ping命令,测试是否通
	//s := clusterClient.Do("ping").String()
	//fmt.Println(s)

	//rLen, err := clusterClient.LLen("_dns:heweijie.top").Result()
	//log.Println(rLen, err)
	////遍历
	//lists, err := clusterClient.LRange("_dns:heweijie.top", 0, rLen-1).Result()
	//log.Println("LRange", lists, err)

	/*s := clusterClient.Do("KEYS","_dns:" + "*").Val()
	//s := clusterClient.Do("KEYS","_dns*").Val()
	result := InterfaceToArray(s);
	fmt.Println(result)*/

	//hget := clusterClient.HGet("_dns:heweijie.top", "host").Val()
	//fmt.Println(hget)

	/*vals := clusterClient.HKeys("_dns:heweijie.top.").Val();//_dns:heweijie.top.
	fmt.Println(vals)*/

	redis := Redis {
		keyPrefix:"_dns:",
		keySuffix:"",
		Ttl:300,
	}

	redis.Connect();
	vals := redis.ClusterClient.HKeys("_dns:heweijie.top.").Val();//_dns:heweijie.top.
	fmt.Println(vals)

}

