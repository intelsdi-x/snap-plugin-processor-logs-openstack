// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

	Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package processor

import (
	"fmt"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetConfigPolicy(t *testing.T) {
	Convey("Create logs-openstack processor", t, func() {
		So(func() { New() }, ShouldNotPanic)
		processor := New()
		So(processor, ShouldNotBeNil)
		Convey("GetConfigPolicy should return a config policy", func() {
			configPolicy, err := processor.GetConfigPolicy()
			Convey("So config policy should be a plugin.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, plugin.ConfigPolicy{})
			})
			Convey("So err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestProcess(t *testing.T) {
	Convey("Create logs-openstack processor", t, func() {
		So(func() { New() }, ShouldNotPanic)
		processor := New()
		So(processor, ShouldNotBeNil)

		Convey("Process metrics containing openstack log", func() {
			Convey("from Openstack Nova Service", func() {
				for i, mockNovaLog := range mockNovaLogs {
					input := mockNovaLog.input
					expected := mockNovaLog.output
					mt := createMockMetric(input.logFileName, input.logData)
					Convey(fmt.Sprintf("TEST Nova %d", i), func() {
						processedMetrics, err := processor.Process([]plugin.Metric{mt}, nil)
						So(err, ShouldBeNil)
						So(processedMetrics, ShouldNotBeEmpty)
						Convey("verify post-processing metric's values", func() {
							So(processedMetrics[0].Data, ShouldEqual, expected.data)
							So(processedMetrics[0].Tags, ShouldResemble, expected.tags)
						})
					})
				}
			})
			Convey("from Openstack Neutron Service", func() {
				for i, mockNeutronLog := range mockNeutronLogs {
					input := mockNeutronLog.input
					expected := mockNeutronLog.output
					mt := createMockMetric(input.logFileName, input.logData)
					Convey(fmt.Sprintf("TEST Neutron %d", i), func() {
						processedMetrics, err := processor.Process([]plugin.Metric{mt}, nil)
						So(err, ShouldBeNil)
						So(processedMetrics, ShouldNotBeEmpty)
						Convey("verify post-processing metric's values", func() {
							So(processedMetrics[0].Data, ShouldEqual, expected.data)
							So(processedMetrics[0].Tags, ShouldResemble, expected.tags)
						})
					})
				}
			})
			Convey("from Openstack Keystone Service", func() {
				for i, mockKeystoneLog := range mockKeystoneLogs {
					input := mockKeystoneLog.input
					expected := mockKeystoneLog.output
					mt := createMockMetric(input.logFileName, input.logData)
					Convey(fmt.Sprintf("TEST Keystone %d", i), func() {
						processedMetrics, err := processor.Process([]plugin.Metric{mt}, nil)
						So(err, ShouldBeNil)
						So(processedMetrics, ShouldNotBeEmpty)
						Convey("verify post-processing metric's values", func() {
							So(processedMetrics[0].Data, ShouldEqual, expected.data)
							So(processedMetrics[0].Tags, ShouldResemble, expected.tags)
						})
					})
				}
			})
			Convey("from Openstack Heat Service", func() {
				for i, mockHeatLog := range mockHeatLogs {
					input := mockHeatLog.input
					expected := mockHeatLog.output
					mt := createMockMetric(input.logFileName, input.logData)
					Convey(fmt.Sprintf("TEST Heat %d", i), func() {
						processedMetrics, err := processor.Process([]plugin.Metric{mt}, nil)
						So(err, ShouldBeNil)
						So(processedMetrics, ShouldNotBeEmpty)
						Convey("verify post-processing metric's values", func() {
							So(processedMetrics[0].Data, ShouldEqual, expected.data)
							So(processedMetrics[0].Tags, ShouldResemble, expected.tags)
						})
					})
				}
			})
			Convey("from Openstack Glance Service", func() {
				for i, mockGlanceLog := range mockGlanceLogs {
					input := mockGlanceLog.input
					expected := mockGlanceLog.output
					mt := createMockMetric(input.logFileName, input.logData)
					Convey(fmt.Sprintf("TEST Glance %d", i), func() {
						processedMetrics, err := processor.Process([]plugin.Metric{mt}, nil)
						So(err, ShouldBeNil)
						So(processedMetrics, ShouldNotBeEmpty)
						Convey("verify post-processing metric's values", func() {
							So(processedMetrics[0].Data, ShouldEqual, expected.data)
							So(processedMetrics[0].Tags, ShouldResemble, expected.tags)
						})
					})
				}
			})

		})

	})
}

func createMockMetric(logFileName string, logData string) plugin.Metric {
	// see snap-plugin-collector-logs to find how metric's namespace is defined
	ns := plugin.NewNamespace("intel", "logs").
		AddDynamicElement("metric_name", "Metric name defined in config file").
		AddDynamicElement("log_file", "Log file name").AddStaticElement("metric")

	// set values to dynamic element
	ns[2].Value = "mock"
	ns[3].Value = logFileName

	return plugin.Metric{
		Namespace: ns,
		Data:      logData,
		Timestamp: time.Now(),
		Tags:      map[string]string{},
	}
}

//// 	** 	MOCK DATA DEFINITION	**	\\\\

type testInput struct {
	logFileName string
	logData     string
}

type testOutput struct {
	data string
	tags map[string]string
}

type TestCase struct {
	input  testInput
	output testOutput
}

var mockNovaLogs = []*TestCase{
	&TestCase{
		input: testInput{
			logFileName: "nova-novncproxy.log",
			logData:     "2016-12-07 03:26:24.254 7 INFO nova.console.websocketproxy [-] handler exception: [Errno 12] Cannot allocate memory",
		},
		output: testOutput{
			data: "[-] handler exception: [Errno 12] Cannot allocate memory",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "7",
				"python_module":  "nova.console.websocketproxy",
				"logger":         "openstack.nova",
			},
		},
	},
	&TestCase{
		input: testInput{
			logFileName: "nova-api.log",
			logData:     "2016-12-06 09:21:08.488 20 INFO nova.osapi_compute.wsgi.server [req-67440a41-6667-4e07-b546-fa336ab5c3af - - - - -] (20) wsgi starting up on http://10.0.0.1:8774",
		},
		output: testOutput{
			data: "[req-67440a41-6667-4e07-b546-fa336ab5c3af - - - - -] (20) wsgi starting up on http://10.0.0.1:8774",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "20",
				"python_module":  "nova.osapi_compute.wsgi.server",
				"logger":         "openstack.nova",
				"request_id":     "67440a41-6667-4e07-b546-fa336ab5c3af",
			},
		},
	},
	&TestCase{
		input: testInput{
			logFileName: "nova-api.log",
			logData:     "2016-12-07 03:53:55.873 24 INFO nova.osapi_compute.wsgi.server [req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170",
		},
		output: testOutput{
			data: "[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] " +
				"10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170",
			tags: map[string]string{
				"severity_label":         "INFO",
				"severity":               "6",
				"pid":                    "24",
				"python_module":          "nova.osapi_compute.wsgi.server",
				"logger":                 "openstack.nova",
				"request_id":             "0c0b761c-47b0-4bf5-832c-89ef048fa56a",
				"user_id":                "fa2b2986c200431b8119035d4a47d420",
				"tenant_id":              "b1ad1df9062a4fc682904c6c9b0f4e98",
				"http_method":            "GET",
				"http_url":               "/v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions",
				"http_version":           "1.1",
				"http_status":            "200",
				"http_response_size":     "23011",
				"http_response_time":     "0.4711170",
				"http_client_ip_address": "10.91.126.38",
				"http_server_ip_address": "10.0.0.1",
			},
		},
	},
	&TestCase{
		input: testInput{
			logFileName: "nova-novncproxy.log",
			logData: "2016-12-06 09:21:16.749 7 WARNING oslo_reports.guru_meditation_report [-] Guru meditation now registers SIGUSR1 and SIGUSR2 " +
				"by default for backward compatibility. SIGUSR1 will no longer be registered in a future release, so please use SIGUSR2 to generate reports.",
		},
		output: testOutput{
			data: "[-] Guru meditation now registers SIGUSR1 and SIGUSR2 " +
				"by default for backward compatibility. SIGUSR1 will no longer be registered in a future release, so please use SIGUSR2 to generate reports.",
			tags: map[string]string{
				"severity_label": "WARNING",
				"severity":       "4",
				"pid":            "7",
				"python_module":  "oslo_reports.guru_meditation_report",
				"logger":         "openstack.nova",
			},
		},
	},
	&TestCase{
		input: testInput{
			logFileName: "nova-compute.log",
			logData: "2016-12-06 09:22:11.020 6 INFO nova.virt.libvirt.host [req-407244e7-4ef1-4180-b139-372e705eda9e - - - - -] Libvirt host capabilities \n" +
				"<capabilities>\n<host><uuid>03aa02fc-0414-0590-8906-ca0700080009</uuid>\n</host>\n</capabilities>\n",
		},
		output: testOutput{
			data: "[req-407244e7-4ef1-4180-b139-372e705eda9e - - - - -] Libvirt host capabilities ",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "6",
				"python_module":  "nova.virt.libvirt.host",
				"logger":         "openstack.nova",
				"request_id":     "407244e7-4ef1-4180-b139-372e705eda9e",
			},
		},
	},
}

var mockNeutronLogs = []*TestCase{
	&TestCase{
		input: testInput{
			logFileName: "neutron-dhcp-agent.log",
			logData:     "2016-12-07 03:27:26.904 6 INFO neutron.agent.dhcp.agent [req-a5e6e3c6-856a-4a86-80c6-cc35b42e7d83 - - - - -] Agent has just been revived. Scheduling full",
		},
		output: testOutput{
			data: "[req-a5e6e3c6-856a-4a86-80c6-cc35b42e7d83 - - - - -] Agent has just been revived. Scheduling full",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "6",
				"python_module":  "neutron.agent.dhcp.agent",
				"logger":         "openstack.neutron",
				"request_id":     "a5e6e3c6-856a-4a86-80c6-cc35b42e7d83",
			},
		},
	},
}

var mockKeystoneLogs = []*TestCase{
	&TestCase{
		input: testInput{
			logFileName: "keystone.log",
			logData:     "2016-12-06 09:19:48.405 18 INFO migrate.versioning.api [-] 66 -> 67...",
		},
		output: testOutput{
			data: "[-] 66 -> 67...",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "18",
				"python_module":  "migrate.versioning.api",
				"logger":         "openstack.keystone",
			},
		},
	},
}

var mockHeatLogs = []*TestCase{
	&TestCase{
		input: testInput{
			logFileName: "heat-api-cfn.log",
			logData:     "2016-12-07 10:59:19.917 8 INFO eventlet.wsgi.server [-] 10.108.8.212,10.0.0.1 - - [07/Dec/2016 10:59:19] \"GET / HTTP/1.0\" 300 276 0.000323",
		},
		output: testOutput{
			data: "[-] 10.108.8.212,10.0.0.1 - - [07/Dec/2016 10:59:19] \"GET / HTTP/1.0\" 300 276 0.000323",
			tags: map[string]string{
				"severity_label":         "INFO",
				"severity":               "6",
				"pid":                    "8",
				"python_module":          "eventlet.wsgi.server",
				"logger":                 "openstack.heat",
				"http_method":            "GET",
				"http_url":               "/",
				"http_version":           "1.0",
				"http_status":            "300",
				"http_response_size":     "276",
				"http_response_time":     "0.000323",
				"http_client_ip_address": "10.108.8.212",
				"http_server_ip_address": "10.0.0.1",
			},
		},
	},
}

var mockGlanceLogs = []*TestCase{
	&TestCase{
		input: testInput{
			logFileName: "glance-api.log",
			logData:     "2016-12-06 09:20:21.261 18 INFO glance.db.sqlalchemy.metadata [-] File /etc/glance/metadefs/cim-processor-allocation-setting-data.json loaded to database.",
		},
		output: testOutput{
			data: "[-] File /etc/glance/metadefs/cim-processor-allocation-setting-data.json loaded to database.",
			tags: map[string]string{
				"severity_label": "INFO",
				"severity":       "6",
				"pid":            "18",
				"python_module":  "glance.db.sqlalchemy.metadata",
				"logger":         "openstack.glance",
			},
		},
	},
}
