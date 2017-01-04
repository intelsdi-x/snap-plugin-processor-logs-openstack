// +build small

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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessorCreation(t *testing.T) {
	Convey("Create processor should not panic", t, func() {
		So(func() { New() }, ShouldNotPanic)
		processor := New()
		Convey("Processor should not be nil", func() {
			So(processor, ShouldNotBeNil)
		})
		Convey("Processor should be of type statisticsProcessor", func() {
			So(processor, ShouldHaveSameTypeAs, &Plugin{})
		})
	})
}

func TestMergeMaps(t *testing.T) {
	Convey("Merge non-empty maps containing different keys", t, func() {
		mapA := map[string]string{
			"key1": "valueA1",
			"key2": "valueA2",
		}
		mapB := map[string]string{
			"key3": "valueB3",
			"key4": "valueB4",
		}
		expectedOut := map[string]string{
			"key1": "valueA1",
			"key2": "valueA2",
			"key3": "valueB3",
			"key4": "valueB4",
		}
		// merging maps, where mapA is a destination, mapB is a source
		mergeMaps(mapA, mapB)

		Convey("Destination map should contain merged data", func() {
			So(mapA, ShouldResemble, expectedOut)
			So(len(mapA), ShouldEqual, len(expectedOut))
		})
		Convey("Source map should be unchanged", func() {
			So(mapB, ShouldEqual, mapB)
			So(len(mapB), ShouldEqual, 2)
		})

	})

	Convey("Merge non-empty maps containing the same key", t, func() {
		mapA := map[string]string{
			"key1": "valueA1",
			"key2": "valueA2",
		}
		mapB := map[string]string{
			"key2": "valueB2",
			"key3": "valueB3",
		}
		expectedOut := map[string]string{
			"key1": "valueA1",
			"key2": "valueB2", // key2 is expected to be overwritten
			"key3": "valueB3",
		}

		mergeMaps(mapA, mapB)

		Convey("Destination map should contain merged data", func() {
			So(mapA, ShouldResemble, expectedOut)

			So(len(mapA), ShouldEqual, len(expectedOut))
			So(mapA["key2"], ShouldEqual, "valueB2")
		})
		Convey("Source map should be unchanged", func() {
			So(mapB, ShouldEqual, mapB)
			So(len(mapB), ShouldEqual, 2)
		})
	})
	Convey("Merge empty map into non-empty map", t, func() {
		mapA := map[string]string{
			"key1": "valueA1",
			"key2": "valueA2",
		}
		mapB := map[string]string{}

		mergeMaps(mapA, mapB)

		Convey("Destination map should be unchanged", func() {
			So(mapA, ShouldResemble, mapA)
		})
		Convey("Source map should be empty (unchanged)", func() {
			So(mapB, ShouldEqual, mapB)
			So(len(mapB), ShouldEqual, 0)
		})
	})
	Convey("Merge non-empty map into empty map", t, func() {
		mapA := map[string]string{}
		mapB := map[string]string{
			"key1": "valueA1",
			"key2": "valueA2",
		}

		mergeMaps(mapA, mapB)

		Convey("Destination map should be unchanged", func() {
			So(mapA, ShouldResemble, mapB)
		})
		Convey("Source map should be empty (unchanged)", func() {
			So(mapB, ShouldEqual, mapB)
			So(len(mapB), ShouldEqual, 2)
		})
	})
}

func TestProcessOpenstackLog(t *testing.T) {
	Convey("Create logs-openstack processor", t, func() {
		So(func() { New() }, ShouldNotPanic)
		processor := New()
		So(processor, ShouldNotBeNil)

		Convey("Process openstack log unsuccessfully", func() {
			Convey("should return an error when log is empty", func() {
				_, _, _, err := processor.processOpenstackLog("")
				So(err, ShouldNotBeNil)
			})
			Convey("should return an error when log is invalid", func() {
				_, _, _, err := processor.processOpenstackLog("invalid")
				So(err, ShouldNotBeNil)
			})
			Convey("should return an error when no payload in log", func() {
				_, _, _, err := processor.processOpenstackLog("2016-12-07 03:26:24.254 7 INFO nova.console.websocketproxy")
				So(err, ShouldNotBeNil)
			})
		})
		Convey("Process openstack log successfully", func() {
			Convey("for payload without request context", func() {
				_, _, _, err := processor.processOpenstackLog("2016-12-07 03:39:17.960 18 INFO nova.wsgi [-] Stopping WSGI server.")
				So(err, ShouldBeNil)
			})
			Convey("for payload with a request context", func() {
				_, _, _, err := processor.processOpenstackLog("2016-12-07 03:39:19.066 33 INFO nova.wsgi [req-cb760354-bbb0-4968-92e6-3312b8a7d223 - - - - -] WSGI server has stopped.")
				So(err, ShouldBeNil)
			})
			Convey("for payload with an HTTP request context", func() {
				_, _, _, err := processor.processOpenstackLog("2016-12-07 03:53:55.873 24 INFO nova.osapi_compute.wsgi.server [req-0c0b761c-47b0-4bf5-832c-89ef048fa56a " +
					"fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 " +
					"\"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170")
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetRequestContext(t *testing.T) {
	Convey("Create a logs-openstack processor", t, func() {
		So(func() {
			New()
		}, ShouldNotPanic)
		processor := New()
		So(processor, ShouldNotBeNil)

		Convey("Unsuccessful getting request context", func() {
			Convey("when message which is empty", func() {
				output := processor.getRequestContext("")
				So(output, ShouldBeNil)
			})
			Convey("when message which does not contain a request", func() {
				output := processor.getRequestContext("[-]")
				So(output, ShouldBeNil)
			})
		})
		Convey("Successful getting request context", func() {
			Convey("when message contains only request_id", func() {
				msg := "[req-cb760354-bbb0-4968-92e6-3312b8a7d223 - - - - -] WSGI server has stopped."
				output := processor.getRequestContext(msg)
				So(output, ShouldNotBeEmpty)
				Convey("so output should contain only request_id", func() {
					So(output, ShouldContainKey, "request_id")
					Convey("so output should not contain others fields", func() {
						So(output, ShouldNotContainKey, "user_id")
						So(output, ShouldNotContainKey, "tenant_id")
					})
					Convey("check the value of request_id", func() {
						So(output["request_id"], ShouldEqual, "cb760354-bbb0-4968-92e6-3312b8a7d223")
					})
				})
			})
			Convey("when message contains full request context", func() {
				msg := "[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] " +
					"10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"
				output := processor.getRequestContext(msg)
				So(output, ShouldNotBeEmpty)
				Convey("so output should contain ids of request, user and tenant", func() {
					So(output, ShouldContainKey, "request_id")
					So(output, ShouldContainKey, "user_id")
					So(output, ShouldContainKey, "tenant_id")
					Convey("check output values ", func() {
						So(output["request_id"], ShouldEqual, "0c0b761c-47b0-4bf5-832c-89ef048fa56a")
						So(output["user_id"], ShouldEqual, "fa2b2986c200431b8119035d4a47d420")
						So(output["tenant_id"], ShouldEqual, "b1ad1df9062a4fc682904c6c9b0f4e98")
					})
				})
			})
		})
	})
}

func TestGetHTTPRequestContext(t *testing.T) {
	Convey("Create a logs-openstack processor", t, func() {
		So(func() {
			New()
		}, ShouldNotPanic)
		processor := New()
		So(processor, ShouldNotBeNil)

		Convey("Unsuccessful getting HTTP request context", func() {
			Convey("when message is empty", func() {
				output := processor.getHTTPRequestContext("")
				So(output, ShouldBeNil)
			})
			Convey("when message does not contain a request", func() {
				output := processor.getHTTPRequestContext("[-]")
				So(output, ShouldBeNil)
			})
			Convey("when message does not contain an HTTP request context", func() {
				output := processor.getHTTPRequestContext("[req-cb760354-bbb0-4968-92e6-3312b8a7d223 - - - - -] WSGI server has stopped.")
				So(output, ShouldBeNil)
			})
		})
		Convey("Successful getting an HTTP request context", func() {
			Convey("when message contains full HTTP request context", func() {
				msg := "[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] " +
					"10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"
				output := processor.getHTTPRequestContext(msg)
				So(output, ShouldNotBeEmpty)
				Convey("so output should contain appropriate fields", func() {
					So(output, ShouldContainKey, "http_method")
					So(output, ShouldContainKey, "http_url")
					So(output, ShouldContainKey, "http_version")
					So(output, ShouldContainKey, "http_status")
					So(output, ShouldContainKey, "http_response_size")
					So(output, ShouldContainKey, "http_response_time")
					So(output, ShouldContainKey, "http_client_ip_address")
					So(output, ShouldContainKey, "http_server_ip_address")
					Convey("check output values ", func() {
						So(output["http_method"], ShouldEqual, "GET")
						So(output["http_url"], ShouldEqual, "/v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions")
						So(output["http_version"], ShouldEqual, "1.1")
						So(output["http_status"], ShouldEqual, "200")
						So(output["http_response_size"], ShouldEqual, "23011")
						So(output["http_response_time"], ShouldEqual, "0.4711170")
						So(output["http_client_ip_address"], ShouldEqual, "10.91.126.38")
						So(output["http_server_ip_address"], ShouldEqual, "10.0.0.1")
					})
				})
			})
			Convey("when message contains HTTP request context without IPs", func() {
				msg := "[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] " +
					"\"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"
				output := processor.getHTTPRequestContext(msg)
				So(output, ShouldNotBeEmpty)
				Convey("so output should contain appropriate fields", func() {
					So(output, ShouldNotContainKey, "http_client_ip_address")
					So(output, ShouldNotContainKey, "http_server_ip_address")
				})
				Convey("so output should contain others appropriate fields", func() {
					So(output, ShouldContainKey, "http_method")
					So(output, ShouldContainKey, "http_url")
					So(output, ShouldContainKey, "http_version")
					So(output, ShouldContainKey, "http_status")
					So(output, ShouldContainKey, "http_response_size")
					So(output, ShouldContainKey, "http_response_time")
					Convey("check output values ", func() {
						So(output["http_method"], ShouldEqual, "GET")
						So(output["http_url"], ShouldEqual, "/v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions")
						So(output["http_version"], ShouldEqual, "1.1")
						So(output["http_status"], ShouldEqual, "200")
						So(output["http_response_size"], ShouldEqual, "23011")
						So(output["http_response_time"], ShouldEqual, "0.4711170")
					})
				})
			})
		})
	})
}
