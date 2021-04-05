/**
 * Copyright 2017 ~ 2025 the original author or authors.
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

import "strings"

/**
 * qname截取出zone
 */
func Qname2Zone(qname string) string {
	s := strings.Split(qname, ".")
	if len(s) <= 3 {
		return qname
	} else { //>3
		for i, _ := range SpecialDomains {
			if strings.HasSuffix(qname, SpecialDomains[i]) {
				d := strings.ReplaceAll(qname, SpecialDomains[i], "")
				t := strings.Split(d, ".")
				return t[len(t)-2] + "." + SpecialDomains[i]
			}
		}
		return s[len(s)-3] + "." + s[len(s)-2] + "."
	}
}

/**
 *  判断一个字符串是否和另一个包含通配符的字符串相等
 *  a 普通字符串
 *  b 带通配符字符串
 */
func ExpressionMatch(a string, b string) bool {
	wildcard := "*"                                     // 通配符
	as := append([]string{""}, strings.Split(a, "")...) // 打断成数组并在前面追加""
	bs := append([]string{""}, strings.Split(b, "")...) // 打断成数组并在前面追加""
	c := make([][]bool, len(bs))                        // 矩形图
	cl := false                                         // 当前位置的左坐标
	cu := false                                         // 当前位置的上坐标
	clu := false                                        // 当前位置的左上坐标
	// 初始化第一列的值
	last := false //上行，上列值
	for i := range bs {
		// 初始化矩阵的每一行
		c[i] = make([]bool, len(as))
		// 上一列的值
		if i-1 < 0 {
			last = true
		} else {
			last = c[i-1][0]
		}
		c[i][0] = (bs[i] == as[0] || wildcard == bs[i]) && last
	}
	// 初始化第一行的值
	for i := range as {
		if i-1 < 0 {
			last = true
		} else {
			last = c[0][i-1]
		}

		c[0][i] = (as[i] == bs[0] || as[i] == wildcard) && last

	}
	// 列行循环初始化所有矩阵值
	for i := 1; i < len(bs); i++ {
		for j := 1; j < len(as); j++ {
			cl = c[i][j-1]
			cu = c[i-1][j]
			clu = c[i-1][j-1]

			if bs[i] == wildcard {
				c[i][j] = cl || cu
			} else {
				c[i][j] = (as[j] == bs[i]) && clu
			}
		}
	}
	return c[len(bs)-1][len(as)-1]
}
