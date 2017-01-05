# Snap plugin processor - logs-openstack

## Openstack Log Pattern	

Openstack log messages are expected to be in the following form, where `payload` is the rest of message

```
            <timestamp> <pid> <severity_label> <python_module> <payload>
          
Example:    2016-12-08 03:18:49.626 20 ERROR nova.api.openstack.extensions some_message
```


Openstack `payload` might include request with HTTP context. Patterns used to retrieve all these corresponding values are described in next sections.

### Pattern for Request Context	

Openstack `payload` might include a request context which can take multiple forms:				
a) the case the capture produces nil:
```
 [-]
```
b) the case the capture produces `request_id`:
```
 [req-b571ba10-0b4e-4411-a233-3df02488eae1 - - - - -]
```
c) the case the capture produces `request_id`, `user_id` and `tenant_id`:
```
 [req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - - - ]
```
The `request_id` might be formatted as "req-xxx" or just "xxx" depending on the Openstack project. Also an uuid might be formatted in different form: "b571ba10-0b4e-4411-a233-3df02488eae1" or "b571ba100b4e4411a2333df02488eae1".
	
### Pattern for HTTP Context

Openstack `payload` might include a HTTP request context which produces six values in this form:
	
```
            "<http_method> <http_url> HTTP/<http_version>" status: <http_status> len: <http_response_size> time: <http_response_time>

Example:	`"OPTIONS /example.com HTTP/1.1" status: 200 len: 23011 time: 0.4711170`  
```

These default prefixes "status: ", "len: " and "time: " might not occur.
	
### Pattern for HTTP request IP addresses		

Capture for IP addresses `http_client_ip_addresses` and `http_server_ip_address` from `payload`

### Examples	

There are 3 examples of Openstack logs in different forms which are supported by implemented patterns:

#### Example 1  

a) incoming log:    
```
2016-12-07 03:26:24.254 7 INFO nova.console.websocketproxy [-] handler exception: [Errno 12] Cannot allocate memory
```
        
b) found matches:   
   - timestamp = `2016-12-07 03:26:24.254`  
   - pid = `7`   
   - severity_label_ = `INFO`  
   - python_module = `nova.console.websocketproxy`  
   - payload = `[-] handler exception: [Errno 12] Cannot allocate memory`  


#### Example 2   
a) incoming log:    
```
2016-12-07 03:48:23.505 35 INFO nova.metadata.wsgi.server [req-cacf21a7-2709-444c-97ba-d9d5634db7da - - - - -] (35) wsgi starting up on http://10.0.0.1:8775
```
        
b) found matches:    
   - timestamp = `2016-12-07 03:48:23.505`  
   - pid = `35`   
   - severity_label = `INFO`  
   - python_module = `nova.metadata.wsgi.server`  
   - payload = `[req-cacf21a7-2709-444c-97ba-d9d5634db7da - - - - -] (35) wsgi starting up on http://10.0.0.1:8775`  
   - request_id = `cacf21a7-2709-444c-97ba-d9d5634db7da`  
        
        
#### Example 3  
a) incoming log:     
```
2016-12-07 03:53:55.873 24 INFO nova.osapi_compute.wsgi.server [req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170
```

b) found matches:     
   - timestamp = `2016-12-07 03:53:55.873`  
   - pid = `24`   
   - severity_label = `INFO`  
   - python_module = `nova.osapi_compute.wsgi.server`  
   - payload = `[req-0c0b761c-47b0-4bf5-832c-89ef048fa56a fa2b2986c200431b8119035d4a47d420 b1ad1df9062a4fc682904c6c9b0f4e98 - default default] 10.91.126.38,10.0.0.1 \"GET /v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions HTTP/1.1\" status: 200 len: 23011 time: 0.4711170"` 
   - request_id = `0c0b761c-47b0-4bf5-832c-89ef048fa56a`  
   - user_id = `fa2b2986c200431b8119035d4a47d420`  
   - tenant_id = `b1ad1df9062a4fc682904c6c9b0f4e98`  
   - http_method = `GET`  
   - http_url = `/v2.1/b1ad1df9062a4fc682904c6c9b0f4e98/extensions`  
   - http_version = `1.1`  
   - http_status = `200`  
   - http_response_size = `23011`  
   - http_response_time = `0.4711170`  
   - http_client_ip_address = `10.91.126.38`  
