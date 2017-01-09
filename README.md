<!--
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
-->

[![Build Status](https://travis-ci.com/intelsdi-x/snap-plugin-processor-logs-openstack.svg?token=zPZwRQ3ErqZxmoixdxAF&branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-processor-logs-openstack)

# Snap plugin processor - logs-openstack

Snap plugin intended to process openstack logs collected by [snap-plugin-collector-logs](https://github.com/intelsdi-x/snap-plugin-collector-logs).

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Openstack Log Pattern](#openstack-log-pattern)
  * [Openstack Log Processing](#openstack-log-processing)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started
### System Requirements 
* [golang 1.7+](https://golang.org/dl/) (needed only for building)

### Operating systems
All OSs currently supported by snap:
* Linux/amd64

### Installation
#### Download logs-openstack plugin binary:
You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-processor-logs-openstack/releases) page.
Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-processor-logs-openstack

Clone repo into `$GOPATH/src/github/intelsdi-x/`:
```
$ git clone https://github.com/<yourGithubID>/snap-plugin-processor-logs-openstack
```
Build the plugin by running make in repo:
```
$ make
```
This builds the plugin in `./build`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)

## Documentation

### Openstack Log Pattern

The Openstack log is expected to be in this form:
```
    <timestamp> <pid> <severity_label> <python_module> <payload>
```
Example:
```
    2016-12-07 03:53:55.873 24 INFO nova.osapi_compute.wsgi.server _some_message_
```

Find out more about Openstack logs pattern in [LOG_PATTERNS.md](LOG_PATTERNS.md)

### Openstack Log Processing

The intention of this plugin is parsing Openstack logs provided by [snap-plugin-collector-logs](https://github.com/intelsdi-x/snap-plugin-collector-logs) as metric's data
to process them to producing these values:
- `timestamp` which replaces metric's timestamp  
- `pid`  
- `severity_label` and `severity` where:  
  - "EMERGENCY" -> 0  
  - "ALERT"     -> 1  
  - "CRITICAL"  -> 2  
  - "ERROR"     -> 3  
  - "WARNING"   -> 4  
  - "NOTICE"    -> 5  
  - "INFO"      -> 6  
  - "DEBUG"     -> 7  
- `python_module`   
- `payload` which replaces metric's data  
- others request-context related (if occur in `payload`):  
  - `request_id`  
  - `tenant_id`  
  - `user_id`  
  - `http_method`  
  - `http_url`  
  - `http_version`  
  - `http_status`  
  - `http_response_size`  
  - `http_response_time`  
  - `http_client_ip_address`  
- and `logger` in form "openstack.\<service_name\>", where the `service_name` is determined in incoming metric's namespace as a _log_file_ (see [snap-plugin-collector-logs#collected-metrics](https://github.com/intelsdi-x/snap-plugin-collector-logs/blob/master/README.md#collected-metrics)).
     

#### How it works
 
For example, for the following metric which is input provided by [snap-plugin-collector-logs](https://github.com/intelsdi-x/snap-plugin-collector-logs):  

**Before processing**:
- Namespace: 	`/intel/logs/openstack/nova-api.log/metrics`
- Timestamp:    `2016-12-07T10:23:13.509335257+1:00`,
- Data:		     
 `"2016-12-07 03:53:55.873 24 INFO nova.osapi_compute.wsgi.server [req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"`
- Tags:  		
        - "plugin_running_on": "your-hostname"

**After processing**:

- Namespace: 	`/intel/logs/openstack/nova-api.log/metric`
- Timestamp : 	`2016-12-07T03:53:55.873`
- Data:  
`"[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"`
- Tags:  
        - "pid" : "24"  
        - "severity_label" : "INFO"  
	    - "python_module" : "nova.osapi_compute.wsgi.server"  
        - "request_id" : "ee7ac782-32aa-492c-aa48-3d84825814e7"  
        - "user_id" : "fa2b2986c200431b8119035d4a47d420"  
        - "tenant_id" : "b1ad1df9062a4fc682904c6c9b0f4e98"   
        - "http_method":"GET"  
        - "http_url": "/v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions"  
	    - "http_version": "1.1"  
	    - "http_status": "200"  
	    - "http_response_size": "23011"  
	    - "http_response_time": "0.4711170"  
	    - "http_client_ip_address": "10.91.126.38"  
	    - "plugin_running_on": "your-hostname"  
        - "severity" : "6 " 
        - "logger" : "openstack.nova"  


### Examples

This is an example running snap-plugin-collector-logs, processing collected openstack-logs and writing post-processed data to a file. It is assumed that you are using the latest Snap binary and plugins.
The example is run from a directory which includes snaptel, snapteld, along with the plugins and task file.

In one terminal window, open the Snap daemon (in this case with logging set to 1 and trust disabled) with appropriate configuration needed by logs collector. 
To do that properly, please follow the instruction on [snap-plugin-collector-logs](https://github.com/intelsdi-x/snap-plugin-collector-logs).
```
$ snapteld -l 1 -t 0 --config config.json
```
In another terminal window:  

Load logs collector plugin:
```
$ snaptel plugin load snap-plugin-collector-logs
Plugin loaded
Name: logs
Version: 1
Type: collector
Signed: false
Loaded Time: Wed, 04 Jan 2017 15:06:30 CET
```
See available metrics for your system
```
$ snaptel metric list
```
Create a task manifest - see examplary task manifest in [examples/tasks](examples/tasks/):
```json
{
  "version": 1,
  "schedule": {
    "type": "simple",
    "interval": "1s"
  },
  "workflow": {
    "collect": {
      "metrics": {
        "/intel/logs/*": {}
      },
      "process": [
        {
          "plugin_name": "logs-openstack",
          "config": {},
          "publish": [
            {
              "plugin_name": "file",
              "config": {
                "file": "/tmp/published_openstack_logs.log"
              }
            }
          ]
        }
      ]
    }
  }
}

```
Load logs-openstack plugin for processing:
```
$ snaptel plugin load snap-plugin-processor-logs-openstack
Plugin loaded
Name: logs-openstack
Version: 1
Type: processor
Signed: false
Loaded Time: Wed, 04 Jan 2017 15:02:37 CET
```

Load file plugin for publishing:
```
$ snaptel plugin load snap-plugin-publisher-file
Plugin loaded
Name: file
Version: 3
Type: publisher
Signed: false
Loaded Time: Wed, 04 Jan 2017 15:03:25 CET
```

Create a task:
```
$ snaptel task create -t task-openstack-logs.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```

Stop task: 
```
$ snaptel task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-processor-logs-openstack/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-processor-logs-openstack/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Izabella Raulin](https://github.com/IzabellaRaulin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

