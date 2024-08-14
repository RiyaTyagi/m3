// Copyright (c) 2021 Uber Technologies, Inc.
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

package convert

import (
	"encoding/base64"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/m3db/m3/src/x/checked"
	"github.com/m3db/m3/src/x/ident"
	"github.com/m3db/m3/src/x/pool"
	"github.com/m3db/m3/src/x/serialize"
)

type idWithEncodedTags struct {
	id          ident.BytesID
	encodedTags []byte
}

type idWithTags struct {
	id   ident.BytesID
	tags ident.Tags
}

// Samples of series IDs with corresponding tags. Taken from metrics generated by promremotebench.
//
//nolint:lll
var samples = []struct {
	id   string
	tags string
}{
	{
		id:   `{__name__="diskio",arch="x64",datacenter="us-west-2c",hostname="host_78",measurement="reads",os="Ubuntu15.10",rack="87",region="us-west-2",service="11",service_environment="production",service_version="1",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18GAGRpc2tpbwQAYXJjaAMAeDY0CgBkYXRhY2VudGVyCgB1cy13ZXN0LTJjCABob3N0bmFtZQcAaG9zdF83OAsAbWVhc3VyZW1lbnQFAHJlYWRzAgBvcwsAVWJ1bnR1MTUuMTAEAHJhY2sCADg3BgByZWdpb24JAHVzLXdlc3QtMgcAc2VydmljZQIAMTETAHNlcnZpY2VfZW52aXJvbm1lbnQKAHByb2R1Y3Rpb24PAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQIAU0Y=",
	},
	{
		id:   `{__name__="nginx",arch="x64",datacenter="us-west-1a",hostname="host_37",measurement="active",os="Ubuntu16.10",rack="78",region="us-west-1",service="10",service_environment="test",service_version="0",team="LON"}`,
		tags: "dScMAAgAX19uYW1lX18FAG5naW54BABhcmNoAwB4NjQKAGRhdGFjZW50ZXIKAHVzLXdlc3QtMWEIAGhvc3RuYW1lBwBob3N0XzM3CwBtZWFzdXJlbWVudAYAYWN0aXZlAgBvcwsAVWJ1bnR1MTYuMTAEAHJhY2sCADc4BgByZWdpb24JAHVzLXdlc3QtMQcAc2VydmljZQIAMTATAHNlcnZpY2VfZW52aXJvbm1lbnQEAHRlc3QPAHNlcnZpY2VfdmVyc2lvbgEAMAQAdGVhbQMATE9O",
	},
	{
		id:   `{__name__="disk",arch="x64",datacenter="sa-east-1b",hostname="host_54",measurement="inodes_total",os="Ubuntu16.10",rack="88",region="sa-east-1",service="15",service_environment="production",service_version="0",team="CHI"}`,
		tags: "dScMAAgAX19uYW1lX18EAGRpc2sEAGFyY2gDAHg2NAoAZGF0YWNlbnRlcgoAc2EtZWFzdC0xYggAaG9zdG5hbWUHAGhvc3RfNTQLAG1lYXN1cmVtZW50DABpbm9kZXNfdG90YWwCAG9zCwBVYnVudHUxNi4xMAQAcmFjawIAODgGAHJlZ2lvbgkAc2EtZWFzdC0xBwBzZXJ2aWNlAgAxNRMAc2VydmljZV9lbnZpcm9ubWVudAoAcHJvZHVjdGlvbg8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAwBDSEk=",
	},
	{
		id:   `{__name__="net",arch="x86",datacenter="us-east-1b",hostname="host_93",measurement="err_in",os="Ubuntu15.10",rack="37",region="us-east-1",service="12",service_environment="production",service_version="1",team="CHI"}`,
		tags: "dScMAAgAX19uYW1lX18DAG5ldAQAYXJjaAMAeDg2CgBkYXRhY2VudGVyCgB1cy1lYXN0LTFiCABob3N0bmFtZQcAaG9zdF85MwsAbWVhc3VyZW1lbnQGAGVycl9pbgIAb3MLAFVidW50dTE1LjEwBAByYWNrAgAzNwYAcmVnaW9uCQB1cy1lYXN0LTEHAHNlcnZpY2UCADEyEwBzZXJ2aWNlX2Vudmlyb25tZW50CgBwcm9kdWN0aW9uDwBzZXJ2aWNlX3ZlcnNpb24BADEEAHRlYW0DAENISQ==",
	},
	{
		id:   `{__name__="redis",arch="x86",datacenter="eu-central-1a",hostname="host_70",measurement="keyspace_misses",os="Ubuntu16.04LTS",rack="47",region="eu-central-1",service="12",service_environment="staging",service_version="1",team="LON"}`,
		tags: "dScMAAgAX19uYW1lX18FAHJlZGlzBABhcmNoAwB4ODYKAGRhdGFjZW50ZXINAGV1LWNlbnRyYWwtMWEIAGhvc3RuYW1lBwBob3N0XzcwCwBtZWFzdXJlbWVudA8Aa2V5c3BhY2VfbWlzc2VzAgBvcw4AVWJ1bnR1MTYuMDRMVFMEAHJhY2sCADQ3BgByZWdpb24MAGV1LWNlbnRyYWwtMQcAc2VydmljZQIAMTITAHNlcnZpY2VfZW52aXJvbm1lbnQHAHN0YWdpbmcPAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQMATE9O",
	},
	{
		id:   `{__name__="nginx",arch="x86",datacenter="us-east-1b",hostname="host_84",measurement="requests",os="Ubuntu16.04LTS",rack="90",region="us-east-1",service="13",service_environment="test",service_version="0",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18FAG5naW54BABhcmNoAwB4ODYKAGRhdGFjZW50ZXIKAHVzLWVhc3QtMWIIAGhvc3RuYW1lBwBob3N0Xzg0CwBtZWFzdXJlbWVudAgAcmVxdWVzdHMCAG9zDgBVYnVudHUxNi4wNExUUwQAcmFjawIAOTAGAHJlZ2lvbgkAdXMtZWFzdC0xBwBzZXJ2aWNlAgAxMxMAc2VydmljZV9lbnZpcm9ubWVudAQAdGVzdA8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAwBOWUM=",
	},
	{
		id:   `{__name__="mem",arch="x64",datacenter="eu-central-1b",hostname="host_27",measurement="buffered",os="Ubuntu16.04LTS",rack="58",region="eu-central-1",service="0",service_environment="test",service_version="0",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18DAG1lbQQAYXJjaAMAeDY0CgBkYXRhY2VudGVyDQBldS1jZW50cmFsLTFiCABob3N0bmFtZQcAaG9zdF8yNwsAbWVhc3VyZW1lbnQIAGJ1ZmZlcmVkAgBvcw4AVWJ1bnR1MTYuMDRMVFMEAHJhY2sCADU4BgByZWdpb24MAGV1LWNlbnRyYWwtMQcAc2VydmljZQEAMBMAc2VydmljZV9lbnZpcm9ubWVudAQAdGVzdA8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAwBOWUM=",
	},
	{
		id:   `{__name__="kernel",arch="x86",datacenter="us-west-2a",hostname="host_80",measurement="disk_pages_in",os="Ubuntu16.10",rack="42",region="us-west-2",service="13",service_environment="test",service_version="1",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18GAGtlcm5lbAQAYXJjaAMAeDg2CgBkYXRhY2VudGVyCgB1cy13ZXN0LTJhCABob3N0bmFtZQcAaG9zdF84MAsAbWVhc3VyZW1lbnQNAGRpc2tfcGFnZXNfaW4CAG9zCwBVYnVudHUxNi4xMAQAcmFjawIANDIGAHJlZ2lvbgkAdXMtd2VzdC0yBwBzZXJ2aWNlAgAxMxMAc2VydmljZV9lbnZpcm9ubWVudAQAdGVzdA8Ac2VydmljZV92ZXJzaW9uAQAxBAB0ZWFtAgBTRg==",
	},
	{
		id:   `{__name__="disk",arch="x64",datacenter="ap-northeast-1c",hostname="host_77",measurement="inodes_used",os="Ubuntu16.04LTS",rack="84",region="ap-northeast-1",service="5",service_environment="production",service_version="0",team="LON"}`,
		tags: "dScMAAgAX19uYW1lX18EAGRpc2sEAGFyY2gDAHg2NAoAZGF0YWNlbnRlcg8AYXAtbm9ydGhlYXN0LTFjCABob3N0bmFtZQcAaG9zdF83NwsAbWVhc3VyZW1lbnQLAGlub2Rlc191c2VkAgBvcw4AVWJ1bnR1MTYuMDRMVFMEAHJhY2sCADg0BgByZWdpb24OAGFwLW5vcnRoZWFzdC0xBwBzZXJ2aWNlAQA1EwBzZXJ2aWNlX2Vudmlyb25tZW50CgBwcm9kdWN0aW9uDwBzZXJ2aWNlX3ZlcnNpb24BADAEAHRlYW0DAExPTg==",
	},
	{
		id:   `{__name__="postgresl",arch="x64",datacenter="eu-central-1b",hostname="host_27",measurement="xact_rollback",os="Ubuntu16.04LTS",rack="58",region="eu-central-1",service="0",service_environment="test",service_version="0",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18JAHBvc3RncmVzbAQAYXJjaAMAeDY0CgBkYXRhY2VudGVyDQBldS1jZW50cmFsLTFiCABob3N0bmFtZQcAaG9zdF8yNwsAbWVhc3VyZW1lbnQNAHhhY3Rfcm9sbGJhY2sCAG9zDgBVYnVudHUxNi4wNExUUwQAcmFjawIANTgGAHJlZ2lvbgwAZXUtY2VudHJhbC0xBwBzZXJ2aWNlAQAwEwBzZXJ2aWNlX2Vudmlyb25tZW50BAB0ZXN0DwBzZXJ2aWNlX3ZlcnNpb24BADAEAHRlYW0DAE5ZQw==",
	},
	{
		id:   `{__name__="cpu",arch="x64",datacenter="sa-east-1b",hostname="host_43",measurement="usage_nice",os="Ubuntu16.10",rack="95",region="sa-east-1",service="4",service_environment="test",service_version="0",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18DAGNwdQQAYXJjaAMAeDY0CgBkYXRhY2VudGVyCgBzYS1lYXN0LTFiCABob3N0bmFtZQcAaG9zdF80MwsAbWVhc3VyZW1lbnQKAHVzYWdlX25pY2UCAG9zCwBVYnVudHUxNi4xMAQAcmFjawIAOTUGAHJlZ2lvbgkAc2EtZWFzdC0xBwBzZXJ2aWNlAQA0EwBzZXJ2aWNlX2Vudmlyb25tZW50BAB0ZXN0DwBzZXJ2aWNlX3ZlcnNpb24BADAEAHRlYW0CAFNG",
	},
	{
		id:   `{__name__="disk",arch="x64",datacenter="ap-northeast-1c",hostname="host_17",measurement="inodes_total",os="Ubuntu16.10",rack="94",region="ap-northeast-1",service="9",service_environment="staging",service_version="0",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18EAGRpc2sEAGFyY2gDAHg2NAoAZGF0YWNlbnRlcg8AYXAtbm9ydGhlYXN0LTFjCABob3N0bmFtZQcAaG9zdF8xNwsAbWVhc3VyZW1lbnQMAGlub2Rlc190b3RhbAIAb3MLAFVidW50dTE2LjEwBAByYWNrAgA5NAYAcmVnaW9uDgBhcC1ub3J0aGVhc3QtMQcAc2VydmljZQEAORMAc2VydmljZV9lbnZpcm9ubWVudAcAc3RhZ2luZw8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAgBTRg==",
	},
	{
		id:   `{__name__="redis",arch="x86",datacenter="us-west-2a",hostname="host_80",measurement="sync_partial_err",os="Ubuntu16.10",rack="42",region="us-west-2",service="13",service_environment="test",service_version="1",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18FAHJlZGlzBABhcmNoAwB4ODYKAGRhdGFjZW50ZXIKAHVzLXdlc3QtMmEIAGhvc3RuYW1lBwBob3N0XzgwCwBtZWFzdXJlbWVudBAAc3luY19wYXJ0aWFsX2VycgIAb3MLAFVidW50dTE2LjEwBAByYWNrAgA0MgYAcmVnaW9uCQB1cy13ZXN0LTIHAHNlcnZpY2UCADEzEwBzZXJ2aWNlX2Vudmlyb25tZW50BAB0ZXN0DwBzZXJ2aWNlX3ZlcnNpb24BADEEAHRlYW0CAFNG",
	},
	{
		id:   `{__name__="net",arch="x86",datacenter="us-east-1a",hostname="host_79",measurement="drop_out",os="Ubuntu16.04LTS",rack="17",region="us-east-1",service="17",service_environment="staging",service_version="1",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18DAG5ldAQAYXJjaAMAeDg2CgBkYXRhY2VudGVyCgB1cy1lYXN0LTFhCABob3N0bmFtZQcAaG9zdF83OQsAbWVhc3VyZW1lbnQIAGRyb3Bfb3V0AgBvcw4AVWJ1bnR1MTYuMDRMVFMEAHJhY2sCADE3BgByZWdpb24JAHVzLWVhc3QtMQcAc2VydmljZQIAMTcTAHNlcnZpY2VfZW52aXJvbm1lbnQHAHN0YWdpbmcPAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQIAU0Y=",
	},
	{
		id:   `{__name__="redis",arch="x86",datacenter="ap-southeast-2b",hostname="host_100",measurement="used_cpu_user_children",os="Ubuntu16.04LTS",rack="40",region="ap-southeast-2",service="14",service_environment="staging",service_version="1",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18FAHJlZGlzBABhcmNoAwB4ODYKAGRhdGFjZW50ZXIPAGFwLXNvdXRoZWFzdC0yYggAaG9zdG5hbWUIAGhvc3RfMTAwCwBtZWFzdXJlbWVudBYAdXNlZF9jcHVfdXNlcl9jaGlsZHJlbgIAb3MOAFVidW50dTE2LjA0TFRTBAByYWNrAgA0MAYAcmVnaW9uDgBhcC1zb3V0aGVhc3QtMgcAc2VydmljZQIAMTQTAHNlcnZpY2VfZW52aXJvbm1lbnQHAHN0YWdpbmcPAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQMATllD",
	},
	{
		id:   `{__name__="disk",arch="x64",datacenter="ap-southeast-1a",hostname="host_87",measurement="inodes_total",os="Ubuntu15.10",rack="0",region="ap-southeast-1",service="11",service_environment="staging",service_version="0",team="LON"}`,
		tags: "dScMAAgAX19uYW1lX18EAGRpc2sEAGFyY2gDAHg2NAoAZGF0YWNlbnRlcg8AYXAtc291dGhlYXN0LTFhCABob3N0bmFtZQcAaG9zdF84NwsAbWVhc3VyZW1lbnQMAGlub2Rlc190b3RhbAIAb3MLAFVidW50dTE1LjEwBAByYWNrAQAwBgByZWdpb24OAGFwLXNvdXRoZWFzdC0xBwBzZXJ2aWNlAgAxMRMAc2VydmljZV9lbnZpcm9ubWVudAcAc3RhZ2luZw8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAwBMT04=",
	},
	{
		id:   `{__name__="cpu",arch="x64",datacenter="us-west-2a",hostname="host_6",measurement="usage_idle",os="Ubuntu16.10",rack="10",region="us-west-2",service="6",service_environment="test",service_version="0",team="CHI"}`,
		tags: "dScMAAgAX19uYW1lX18DAGNwdQQAYXJjaAMAeDY0CgBkYXRhY2VudGVyCgB1cy13ZXN0LTJhCABob3N0bmFtZQYAaG9zdF82CwBtZWFzdXJlbWVudAoAdXNhZ2VfaWRsZQIAb3MLAFVidW50dTE2LjEwBAByYWNrAgAxMAYAcmVnaW9uCQB1cy13ZXN0LTIHAHNlcnZpY2UBADYTAHNlcnZpY2VfZW52aXJvbm1lbnQEAHRlc3QPAHNlcnZpY2VfdmVyc2lvbgEAMAQAdGVhbQMAQ0hJ",
	},
	{
		id:   `{__name__="nginx",arch="x86",datacenter="us-east-1a",hostname="host_44",measurement="handled",os="Ubuntu16.04LTS",rack="61",region="us-east-1",service="2",service_environment="staging",service_version="1",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18FAG5naW54BABhcmNoAwB4ODYKAGRhdGFjZW50ZXIKAHVzLWVhc3QtMWEIAGhvc3RuYW1lBwBob3N0XzQ0CwBtZWFzdXJlbWVudAcAaGFuZGxlZAIAb3MOAFVidW50dTE2LjA0TFRTBAByYWNrAgA2MQYAcmVnaW9uCQB1cy1lYXN0LTEHAHNlcnZpY2UBADITAHNlcnZpY2VfZW52aXJvbm1lbnQHAHN0YWdpbmcPAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQMATllD",
	},
	{
		id:   `{__name__="nginx",arch="x86",datacenter="us-west-1a",hostname="host_29",measurement="waiting",os="Ubuntu15.10",rack="15",region="us-west-1",service="4",service_environment="test",service_version="1",team="NYC"}`,
		tags: "dScMAAgAX19uYW1lX18FAG5naW54BABhcmNoAwB4ODYKAGRhdGFjZW50ZXIKAHVzLXdlc3QtMWEIAGhvc3RuYW1lBwBob3N0XzI5CwBtZWFzdXJlbWVudAcAd2FpdGluZwIAb3MLAFVidW50dTE1LjEwBAByYWNrAgAxNQYAcmVnaW9uCQB1cy13ZXN0LTEHAHNlcnZpY2UBADQTAHNlcnZpY2VfZW52aXJvbm1lbnQEAHRlc3QPAHNlcnZpY2VfdmVyc2lvbgEAMQQAdGVhbQMATllD",
	},
	{
		id:   `{__name__="diskio",arch="x64",datacenter="ap-northeast-1c",hostname="host_38",measurement="write_time",os="Ubuntu15.10",rack="20",region="ap-northeast-1",service="0",service_environment="staging",service_version="0",team="SF"}`,
		tags: "dScMAAgAX19uYW1lX18GAGRpc2tpbwQAYXJjaAMAeDY0CgBkYXRhY2VudGVyDwBhcC1ub3J0aGVhc3QtMWMIAGhvc3RuYW1lBwBob3N0XzM4CwBtZWFzdXJlbWVudAoAd3JpdGVfdGltZQIAb3MLAFVidW50dTE1LjEwBAByYWNrAgAyMAYAcmVnaW9uDgBhcC1ub3J0aGVhc3QtMQcAc2VydmljZQEAMBMAc2VydmljZV9lbnZpcm9ubWVudAcAc3RhZ2luZw8Ac2VydmljZV92ZXJzaW9uAQAwBAB0ZWFtAgBTRg==",
	},
}

// BenchmarkFromSeriesIDAndTagIter-12    	  772224	      1649 ns/op
func BenchmarkFromSeriesIDAndTagIter(b *testing.B) {
	testData, err := prepareIDAndTags(b)
	require.NoError(b, err)

	b.ResetTimer()
	for i := range testData {
		_, err := FromSeriesIDAndTagIter(testData[i].id, ident.NewTagsIterator(testData[i].tags))
		require.NoError(b, err)
	}
}

// BenchmarkFromSeriesIDAndTags-12       	 1000000	      1311 ns/op
func BenchmarkFromSeriesIDAndTags(b *testing.B) {
	testData, err := prepareIDAndTags(b)
	require.NoError(b, err)

	b.ResetTimer()
	for i := range testData {
		_, err := FromSeriesIDAndTags(testData[i].id, testData[i].tags)
		require.NoError(b, err)
	}
}

func BenchmarkFromSeriesIDAndEncodedTags(b *testing.B) {
	testData, err := prepareIDAndEncodedTags(b)
	require.NoError(b, err)

	b.ResetTimer()
	for i := range testData {
		_, err := FromSeriesIDAndEncodedTags(testData[i].id, testData[i].encodedTags)
		require.NoError(b, err)
	}
}

func BenchmarkFromSeriesIDAndTagIter_TagDecoder(b *testing.B) {
	testData, err := prepareIDAndEncodedTags(b)
	require.NoError(b, err)

	decoderPool := serialize.NewTagDecoderPool(
		serialize.NewTagDecoderOptions(serialize.TagDecoderOptionsConfig{}),
		pool.NewObjectPoolOptions(),
	)
	decoderPool.Init()

	decoder := decoderPool.Get()
	defer decoder.Close()

	b.ResetTimer()
	for i := range testData {
		decoder.Reset(checked.NewBytes(testData[i].encodedTags, nil))
		_, err := FromSeriesIDAndTagIter(testData[i].id, decoder)
		require.NoError(b, err)
	}
}

func prepareIDAndEncodedTags(b *testing.B) ([]idWithEncodedTags, error) {
	var (
		rnd    = rand.New(rand.NewSource(42)) //nolint:gosec
		b64    = base64.StdEncoding
		result = make([]idWithEncodedTags, 0, b.N)
	)

	for i := 0; i < b.N; i++ {
		k := rnd.Intn(len(samples))
		id := clone([]byte(samples[k].id))
		tags, err := b64.DecodeString(samples[k].tags)
		if err != nil {
			return nil, err
		}

		result = append(result, idWithEncodedTags{
			id:          ident.BytesID(id),
			encodedTags: tags,
		})
	}

	return result, nil
}

func prepareIDAndTags(b *testing.B) ([]idWithTags, error) {
	testData, err := prepareIDAndEncodedTags(b)
	if err != nil {
		return nil, err
	}

	decoderPool := serialize.NewTagDecoderPool(
		serialize.NewTagDecoderOptions(serialize.TagDecoderOptionsConfig{}),
		pool.NewObjectPoolOptions(),
	)
	decoderPool.Init()

	bytesPool := pool.NewCheckedBytesPool(nil, nil, func(s []pool.Bucket) pool.BytesPool {
		return pool.NewBytesPool(s, nil)
	})
	bytesPool.Init()

	identPool := ident.NewPool(bytesPool, ident.PoolOptions{})

	tagDecoder := decoderPool.Get()
	defer tagDecoder.Close()

	result := make([]idWithTags, 0, len(testData))
	for i := range testData {
		tagDecoder.Reset(checked.NewBytes(testData[i].encodedTags, nil))
		tags, err := TagsFromTagsIter(testData[i].id, tagDecoder, identPool)
		if err != nil {
			return nil, err
		}
		result = append(result, idWithTags{id: testData[i].id, tags: tags})
	}
	return result, nil
}
