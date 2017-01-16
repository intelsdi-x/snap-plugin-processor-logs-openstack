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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	//Name of the plugin
	Name = "logs-openstack"
	//Version of the plugin
	Version = 1

	timeFormat = "2006-01-02 15:04:05 MST"
)

const (
	// ***	1) PATTERN FOR LOG CONTEXT    ***
	// 	Openstack log messages are expected to be in the following form, where `payload` is the rest of message
	// 	<timestamp> <pid> <severity_label> <python_module> <payload>
	//
	// 	Example: 	2016-12-08 03:18:49.626 20 ERROR nova.api.openstack.extensions some_message
	//
	timestampRegexp = `(?P<timestamp>(\d{4})-(\d{2})-(\d{2})[( |T)](\d{2}):(\d{2}):(\d{2})[.]?(\d+)?)`
	logRegexp       = timestampRegexp + `[ ](?P<pid>\d+)[ ](?P<severity_label>\S+)[ ](?P<python_module>\S+)[ ](?P<payload>(\n|.)*)`

	// ***	2) PATTERN FOR REQUEST CONTEXT   ***
	// 	Openstack payload might include a request context which can tak multiple forms:
	//	a) the case the capture produces nil: 					[-]
	// 	b) the case the capture produces request_id: 				[req-b571ba10-0b4e-4411-a233-3df02488eae1 - - - - -]
	// 	c) the case the capture produces request_id, user_id and tenant_id: 	[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - - - ]
	//
	// 	**Notice** that the `request_id` might be formatted as `req-xxx` or just `xxx` depending on the Openstack project
	//	**Notice** that the uuid might be formatted as `b571ba10-0b4e-4411-a233-3df02488eae1` or `b571ba100b4e4411a2333df02488eae1`
	uuidRegexp           = `\S{8}[-]?\S{4}[-]?\S{4}[-]?\S{4}[-]?\S{12}`
	requestContextRegexp = `\[(req-)?(?P<request_id>` + uuidRegexp + `)([ ](?P<user_id>` + uuidRegexp + `)[ ](?P<tenant_id>` + uuidRegexp + `))?.*\]`

	// ***	3) PATTERN FOR HTTP REQUEST CONTEXT   ***
	// 	Openstack payload might include a HTTP request context which produces six values in this form:
	//      "<http_method> <http_url> HTTP/<http_version>" status: <http_status> len: <http_response_size> time: <http_response_time>
	//
	//	Example:	"OPTIONS /example.com HTTP/1.1" status: 200 len: 23011 time: 0.4711170
	//
	// 	**Notice** that the default prefixes "status: ", "len: " and "time: " might not occure
	httpRequestContextRegexp = `"(?P<http_method>\w+)[ ](?P<http_url>.*)[ ]HTTP\/(?P<http_version>\d.\d)"[ ](status: )?(?P<http_status>\d+)[ ](len: )?(?P<http_response_size>\d+)[ ](time: )?(?P<http_response_time>\d+.\d+)`

	// ***	4) PATTERN FOR HTTP REQUEST IP ADDRESSES   ***
	// 	Capture for IP addresses <http_client_ip_addresses> and <http_server_ip_address>
	//
	// 	**Notice** that only `http_client_ip_addresses` is used and stored in metric' tags
	ipAddressesRegexp          = `\d{1,3}[.]\d{1,3}[.]\d{1,3}[.]\d{1,3}`
	httpRequestAddressesRegexp = `(?P<http_client_ip_address>` + ipAddressesRegexp + `)([,])?(?P<http_server_ip_address>` + ipAddressesRegexp + `)?`
)

// Plugin holds regular expressions and timezone needed to process openstack logs
type Plugin struct {
	logRgx                  *regexp.Regexp
	requestContextRgx       *regexp.Regexp
	httpRequestContextRgx   *regexp.Regexp
	httpRequestAddressesRgx *regexp.Regexp
	timezone                string
}

var severity = map[string]int{
	"EMERGENCY": 0,
	"ALERT":     1,
	"CRITICAL":  2,
	"ERROR":     3,
	"WARNING":   4,
	"NOTICE":    5,
	"INFO":      6,
	"DEBUG":     7,
}

// New returns a new instance of the processor logs-openstack plugin with initialized regular expressions using
// to parse log messages
func New() *Plugin {
	p := &Plugin{}

	err := p.init()
	if err != nil {
		panic(err)
	}
	return p
}

// init compiles declared regular expressions to Regexp objects which are held in plugin structure
// and sets timezone for timestamps
func (p *Plugin) init() error {
	var err error
	errors := []error{}

	if p.logRgx, err = regexp.Compile(logRegexp); err != nil {
		log.WithFields(log.Fields{
			"_block": "init",
			"_error": err,
		}).Error("Cannot parse regular expression defined for logs")
		errors = append(errors, err)
	}
	if p.requestContextRgx, err = regexp.Compile(requestContextRegexp); err != nil {
		log.WithFields(log.Fields{
			"_block": "init",
			"_error": err,
		}).Error("Cannot parse regular expression defined for request context")
		errors = append(errors, err)
	}
	if p.httpRequestContextRgx, err = regexp.Compile(httpRequestContextRegexp); err != nil {
		log.WithFields(log.Fields{
			"_block": "init",
			"_error": err,
		}).Error("Cannot parse regular expression defined for HTTP request context")
		errors = append(errors, err)
	}
	if p.httpRequestAddressesRgx, err = regexp.Compile(httpRequestAddressesRegexp); err != nil {
		log.WithFields(log.Fields{
			"_block": "init",
			"_error": err,
		}).Error("Cannot parse regular expression defined for capture IP addresses from HTTP request")
		errors = append(errors, err)
	}

	p.timezone, _ = time.Now().Zone()
	if p.timezone != "" {
		log.WithFields(log.Fields{
			"_block":    "init",
			"_timezone": p.timezone,
		}).Info("Set timezone for timestamps")
	}

	if len(errors) != 0 {
		return fmt.Errorf("Cannot initialize processor plugin, invalid reqular expression(s), errors: %v", errors)
	}

	return nil
}

// GetConfigPolicy returns the config policy
func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, nil
}

// Process processes the data
func (p *Plugin) Process(metrics []plugin.Metric, _ plugin.Config) ([]plugin.Metric, error) {
	for i, m := range metrics {

		logger, err := getLoggerInfo(m.Namespace)
		if err != nil {
			log.WithFields(log.Fields{
				"_block":  "Process",
				"_metric": m.Namespace.Strings(),
				"_data":   m.Data,
				"_error":  err,
			}).Warning("Cannot retrieve logger info")
			continue
		}

		data, ok := m.Data.(string)
		if !ok {
			log.WithFields(log.Fields{
				"_block":  "Process",
				"_metric": m.Namespace.Strings(),
				"_data":   m.Data,
				"_error":  "unexpected data type",
			}).Warning("Plugin processes only string logs")
			continue
		}

		timestamp, msg, fields, err := p.processOpenstackLog(data)
		if err != nil {
			log.WithFields(log.Fields{
				"_block":  "Process",
				"_metric": m.Namespace.Strings(),
				"_data":   m.Data,
				"_error":  err,
			}).Warning("Invalid format of log block")
			continue
		}

		if msg != "" {
			// for not empty msg, do retrieving a request context
			mergeMaps(fields, p.getRequestContext(msg))
			mergeMaps(fields, p.getHTTPRequestContext(msg))
		}

		// overwrite metric's timestamp and data with values retrieved from log
		metrics[i].Timestamp = timestamp
		metrics[i].Data = msg

		// add other info retrieved from log as metric's tag
		metrics[i].Tags["logger"] = logger

		for k, v := range fields {
			metrics[i].Tags[k] = v
		}
	}

	return metrics, nil
}

// parse returns regular expression matches found in incoming data
func parse(data string, rexp *regexp.Regexp) (map[string]string, error) {
	fields := map[string]string{}

	match := rexp.FindStringSubmatch(data)
	if len(match) <= 0 {
		return nil, errors.New("No string match found")
	}

	for i, name := range rexp.SubexpNames() {
		// skip matches without defined name
		if name == "" {
			continue
		}
		// skip the first match which corresponds to full match
		if i > 0 && i < len(match) {
			if match[i] != "" {
				fields[name] = match[i]
			}
		}
	}
	return fields, nil
}

// processOpenstackLog processes incoming openstack log and retrieves based on regular expression `logRgx` such info like
// log's timestamp, message and others fields (i.a. `pid`, `severity_label`, `severity`, `python_module`)
// An error is returned if incoming data does not fit for openstack-log pattern
func (p *Plugin) processOpenstackLog(data string) (timestamp time.Time, msg string, fields map[string]string, err error) {
	fields, err = parse(data, p.logRgx)
	if err != nil {
		return
	}

	// set a timestamp
	timestampStr, exist := fields["timestamp"]
	if !exist {
		err = fmt.Errorf("No timestamp in log")
		return
	}
	delete(fields, "timestamp")

	// parse timestamp to time.Time type
	timestamp, err = time.Parse(timeFormat, fmt.Sprintf("%s %s", timestampStr, p.timezone))
	if err != nil {
		return
	}

	// set a msg which corresponds to `payload`
	msg, exist = fields["payload"]
	if !exist {
		err = fmt.Errorf("No payload in log")
		return
	}
	delete(fields, "payload")

	// set an appropriate `severity` into fields map based on label from `severity_label`
	// for example, for `severity_label` = "INFO", set `severity` = "6"
	// both `severity_label` and `severity` should be kept in fields map
	if label, ok := fields["severity_label"]; ok {
		fields["severity"] = fmt.Sprintf("%d", severity[label])
	}

	return timestamp, msg, fields, err
}

// getRequestContext parses incoming msg to return all found matches of regular expression `requestContextRgx`
// or nil when there is no request context in message
func (p *Plugin) getRequestContext(msg string) map[string]string {
	requestContext, err := parse(msg, p.requestContextRgx)
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "getRequestContext",
			"_msg":   msg,
			"_error": err.Error(),
		}).Info("No request context in log message")
		return nil
	}

	return requestContext
}

// getHTTPRequestContext parses msg to return all matches of regular expressions `httpRequestContextRgx` and
// `httpRequestAddressesRgx` (optional) or nil when there is no request HTTP context in message
func (p *Plugin) getHTTPRequestContext(msg string) map[string]string {
	httpRequestContext, err := parse(msg, p.httpRequestContextRgx)
	if err != nil {
		log.WithFields(log.Fields{
			"_block": "getHTTPRequestContext",
			"_msg":   msg,
			"_error": err,
		}).Info("No HTTP request context in log message")
		return nil
	}

	if httpRequestAddresses, err := parse(msg, p.httpRequestAddressesRgx); err == nil {
		mergeMaps(httpRequestContext, httpRequestAddresses)
	} else {
		log.WithFields(log.Fields{
			"_block": "getHTTPRequestContext",
			"_msg":   msg,
			"_error": err,
		}).Info("No HTTP request adresses in log message")
	}

	return httpRequestContext
}

// getLoggerInfo returns logger in form "openstack.<service_name>", where `service_name` is retrieved from metric's namespace
func getLoggerInfo(ns plugin.Namespace) (string, error) {
	isDynamic, indexes := ns.IsDynamic()
	if !isDynamic {
		return "", fmt.Errorf("Metric `%v` is expected to contain a dynamic element, but it doesn't", ns.Strings())
	}
	// take the last dynamic element which is expected to be named `log_file`
	lde := ns.Element(indexes[len(indexes)-1])
	if lde.Name != "log_file" {
		return "", fmt.Errorf("Metric `%v` is expected to contain a dynamic element `log_file`, but it doesn't", ns.Strings())
	}
	logFileName := strings.TrimSuffix(lde.Value, ".log")

	// serviceName equals the first part of logFileName splitted by the '-' separator
	serviceName := strings.Split(logFileName, "-")[0]

	return fmt.Sprintf("openstack.%s", serviceName), nil
}

// mergeMaps merges `src` map into `dst`, in case they have the same key, dst attributes will be overwritten
// by src attribute values
func mergeMaps(dst map[string]string, src map[string]string) {
	for key, val := range src {
		dst[key] = val
	}
}
