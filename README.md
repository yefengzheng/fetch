## How to install and run
``` 
In windows
go build -o Fetch.exe     # Compiles the program and generates an executable named "Fetch.exe"
./Fetch.exe config_file.yaml   # Runs the "Fetch.exe" executable
```
## How to fix
### Default HTTP Method Handling
```
method := strings.ToUpper(endpoint.Method)
if method == "" {
    method = "GET"
}
```
The original code used endpoint.Method directly, which could be empty and cause invalid requests. This fix sets the method to GET if none is provided.
### Proper Request Body Usage
```reqBody := bytes.NewReader([]byte{})
if method == "POST" || method == "PUT" || method == "PATCH" {
    reqBody = bytes.NewReader([]byte(endpoint.Body))
}
```
The original code incorrectly marshaled the entire Endpoint struct as JSON. This change ensures only the user-defined Body string is sent, and only for methods that support bodies.
### 500ms Timeout Enforcement
```
client := &http.Client{ Timeout: 500 * time.Millisecond }
```
There was no timeout in the original code. This enforces the rule that a response over 500ms should not be considered available.
### Response Time Validation
```
start := time.Now()
resp, err := client.Do(req)
duration := time.Since(start)

if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && duration <= 500*time.Millisecond {
    stats[domain].Success++
}
```
Why: The original version didn't track how long the request took. This checks both response status and time taken before counting as successful.
### Renamed url to rawURL to Avoid Shadowing and add error to track
```
func extractDomain(rawURL string) (string, error) {
    u, err := url.Parse(rawURL)
    ...
}
```
In the original code, the parameter was named url, which shadows Go's built-in url package and can lead to confusion or subtle bugs. Renaming it to rawURL avoids this conflict and improves code readability.
### Domain Extraction with Port Stripping
```
u, err := url.Parse(endpoint.URL)
host := u.Host
if idx := strings.Index(host, ":"); idx != -1 {
    host = host[:idx] // remove port
}
```
The original version used string splits which could include ports. This fix uses proper URL parsing and strips the port to group stats by domain only.
### Drop Decimal in Availability Percentage
```
percentage := 0
if stat.Total > 0 {
    percentage = int(100 * stat.Success / stat.Total)
}
```
 The original used math.Round, which may keep decimal places. This fix uses int() to truncate and conform to the "drop decimals" rule.
 ### Improved Logging
 ```
log.Printf("Error creating request for %s: %v\n", endpoint.URL, err)
log.Printf("Request failed for %s: %v\n", endpoint.URL, err)
 ```
 The original logging was minimal. These changes provide better visibility into errors during domain extraction or HTTP requests.

 ## Final
I used Postman to check the status of each web endpoint listed in the YAML file. I found that 2 requests succeeded and 2 failed, so the availability is 50%. The result I got was:
 ```
dev-sre-take-home-exercise-rubric.us-east-1.recruiting-public.fetchrewards.com has 50% availability
 ```