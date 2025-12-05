# Примеры использования тестовых вебхуков

## Содержание
- [JavaScript / Node.js](#javascript--nodejs)
- [Python](#python)
- [PHP](#php)
- [Go](#go)
- [Ruby](#ruby)
- [Java](#java)
- [C#](#c)
- [Bash / cURL](#bash--curl)

---

## JavaScript / Node.js

### Fetch API (браузер)
```javascript
const webhookUrl = 'http://localhost:8080/webhook/test/YOUR_TOKEN';

fetch(webhookUrl, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Custom-Header': 'CustomValue'
  },
  body: JSON.stringify({
    event: 'test',
    data: {
      message: 'Hello from JavaScript',
      timestamp: new Date().toISOString()
    }
  })
})
.then(response => response.json())
.then(data => console.log('Success:', data))
.catch(error => console.error('Error:', error));
```

### Axios (Node.js)
```javascript
const axios = require('axios');

const webhookUrl = 'http://localhost:8080/webhook/test/YOUR_TOKEN';

axios.post(webhookUrl, {
  event: 'test',
  data: {
    message: 'Hello from Node.js',
    timestamp: new Date().toISOString()
  }
}, {
  headers: {
    'Content-Type': 'application/json',
    'X-Custom-Header': 'CustomValue'
  }
})
.then(response => console.log('Success:', response.data))
.catch(error => console.error('Error:', error));
```

---

## Python

### Requests
```python
import requests
import json
from datetime import datetime

webhook_url = 'http://localhost:8080/webhook/test/YOUR_TOKEN'

payload = {
    'event': 'test',
    'data': {
        'message': 'Hello from Python',
        'timestamp': datetime.now().isoformat()
    }
}

headers = {
    'Content-Type': 'application/json',
    'X-Custom-Header': 'CustomValue'
}

response = requests.post(webhook_url, json=payload, headers=headers)
print('Status:', response.status_code)
print('Response:', response.json())
```

### urllib (стандартная библиотека)
```python
import urllib.request
import json
from datetime import datetime

webhook_url = 'http://localhost:8080/webhook/test/YOUR_TOKEN'

payload = {
    'event': 'test',
    'data': {
        'message': 'Hello from Python',
        'timestamp': datetime.now().isoformat()
    }
}

data = json.dumps(payload).encode('utf-8')
headers = {
    'Content-Type': 'application/json',
    'X-Custom-Header': 'CustomValue'
}

req = urllib.request.Request(webhook_url, data=data, headers=headers, method='POST')
with urllib.request.urlopen(req) as response:
    print('Status:', response.status)
    print('Response:', response.read().decode('utf-8'))
```

---

## PHP

### cURL
```php
<?php
$webhookUrl = 'http://localhost:8080/webhook/test/YOUR_TOKEN';

$payload = [
    'event' => 'test',
    'data' => [
        'message' => 'Hello from PHP',
        'timestamp' => date('c')
    ]
];

$ch = curl_init($webhookUrl);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
    'X-Custom-Header: CustomValue'
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

$response = curl_exec($ch);
$statusCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

echo "Status: $statusCode\n";
echo "Response: $response\n";
?>
```

### file_get_contents
```php
<?php
$webhookUrl = 'http://localhost:8080/webhook/test/YOUR_TOKEN';

$payload = [
    'event' => 'test',
    'data' => [
        'message' => 'Hello from PHP',
        'timestamp' => date('c')
    ]
];

$options = [
    'http' => [
        'method' => 'POST',
        'header' => [
            'Content-Type: application/json',
            'X-Custom-Header: CustomValue'
        ],
        'content' => json_encode($payload)
    ]
];

$context = stream_context_create($options);
$response = file_get_contents($webhookUrl, false, $context);

echo "Response: $response\n";
?>
```

---

## Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Payload struct {
    Event string `json:"event"`
    Data  Data   `json:"data"`
}

type Data struct {
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
}

func main() {
    webhookURL := "http://localhost:8080/webhook/test/YOUR_TOKEN"

    payload := Payload{
        Event: "test",
        Data: Data{
            Message:   "Hello from Go",
            Timestamp: time.Now().Format(time.RFC3339),
        },
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Custom-Header", "CustomValue")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer resp.Body.Close()

    fmt.Println("Status:", resp.Status)
}
```

---

## Ruby

```ruby
require 'net/http'
require 'json'
require 'uri'

webhook_url = 'http://localhost:8080/webhook/test/YOUR_TOKEN'

payload = {
  event: 'test',
  data: {
    message: 'Hello from Ruby',
    timestamp: Time.now.iso8601
  }
}

uri = URI.parse(webhook_url)
http = Net::HTTP.new(uri.host, uri.port)

request = Net::HTTP::Post.new(uri.path)
request['Content-Type'] = 'application/json'
request['X-Custom-Header'] = 'CustomValue'
request.body = payload.to_json

response = http.request(request)
puts "Status: #{response.code}"
puts "Response: #{response.body}"
```

---

## Java

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Instant;

public class WebhookTest {
    public static void main(String[] args) throws Exception {
        String webhookUrl = "http://localhost:8080/webhook/test/YOUR_TOKEN";
        
        String payload = String.format(
            "{\"event\":\"test\",\"data\":{\"message\":\"Hello from Java\",\"timestamp\":\"%s\"}}",
            Instant.now().toString()
        );

        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(webhookUrl))
            .header("Content-Type", "application/json")
            .header("X-Custom-Header", "CustomValue")
            .POST(HttpRequest.BodyPublishers.ofString(payload))
            .build();

        HttpResponse<String> response = client.send(request, 
            HttpResponse.BodyHandlers.ofString());

        System.out.println("Status: " + response.statusCode());
        System.out.println("Response: " + response.body());
    }
}
```

---

## C#

```csharp
using System;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;

class Program
{
    static async Task Main(string[] args)
    {
        string webhookUrl = "http://localhost:8080/webhook/test/YOUR_TOKEN";

        var payload = new
        {
            @event = "test",
            data = new
            {
                message = "Hello from C#",
                timestamp = DateTime.UtcNow.ToString("o")
            }
        };

        using (var client = new HttpClient())
        {
            client.DefaultRequestHeaders.Add("X-Custom-Header", "CustomValue");

            var json = JsonConvert.SerializeObject(payload);
            var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await client.PostAsync(webhookUrl, content);
            var responseBody = await response.Content.ReadAsStringAsync();

            Console.WriteLine($"Status: {response.StatusCode}");
            Console.WriteLine($"Response: {responseBody}");
        }
    }
}
```

---

## Bash / cURL

### POST с JSON
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -H "Content-Type: application/json" \
  -H "X-Custom-Header: CustomValue" \
  -d '{
    "event": "test",
    "data": {
      "message": "Hello from cURL",
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }
  }'
```

### GET с параметрами
```bash
curl "http://localhost:8080/webhook/test/YOUR_TOKEN?param1=value1&param2=value2"
```

### POST с XML
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0"?>
<test>
  <message>Hello from cURL</message>
  <timestamp>'$(date -u +%Y-%m-%dT%H:%M:%SZ)'</timestamp>
</test>'
```

### POST с form-data
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -F "message=Hello from cURL" \
  -F "timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

### PUT запрос
```bash
curl -X PUT http://localhost:8080/webhook/test/YOUR_TOKEN \
  -H "Content-Type: text/plain" \
  -d "This is a PUT request"
```

### DELETE запрос
```bash
curl -X DELETE http://localhost:8080/webhook/test/YOUR_TOKEN
```

---

## Дополнительные примеры

### Отправка файла (multipart/form-data)
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -F "file=@/path/to/file.txt" \
  -F "description=Test file upload"
```

### Отправка с Basic Auth
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -u username:password \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

### Отправка с Bearer Token
```bash
curl -X POST http://localhost:8080/webhook/test/YOUR_TOKEN \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

---

## Советы

1. **Замените YOUR_TOKEN** на реальный токен из вашего тестового вебхука
2. **Используйте правильный BASE_URL** для production (не localhost)
3. **Проверяйте Content-Type** - он должен соответствовать формату данных
4. **Добавляйте кастомные заголовки** для тестирования их обработки
5. **Используйте timestamp** для отслеживания времени отправки

## Troubleshooting

- **Connection refused**: Проверьте, что сервер запущен
- **404 Not Found**: Проверьте правильность токена
- **410 Gone**: Вебхук истёк (24 часа)
- **CORS errors**: Используйте серверный код, не браузерный JavaScript для cross-origin запросов
