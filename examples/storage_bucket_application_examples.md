# Bucket API Integration Examples (External Applications)

This guide shows how different application types can call bucket/object APIs.

## 1) Configuration

Use these values in all examples:

```bash
BASE_URL="http://localhost:8000/api/v1/storage"
TENANT_ID="5d82e8fe-83ce-4bc0-8b2c-32723d02fb54"
BUCKET="m"
ACCESS_KEY_ID="<your-access-key-id>"
SECRET_ACCESS_KEY="<your-secret-access-key>"
OBJECT_KEY="docs/report.pdf"
```

Required auth headers for application access keys:

- `X-Storage-Access-Key: <accessKeyId>`
- `X-Storage-Secret-Key: <secretAccessKey>`

---

## 2) Endpoints

- `GET    /buckets/{bucket}/objects?tenantId={tenantId}`
- `GET    /buckets/{bucket}/objects/{objectKey}?tenantId={tenantId}`
- `PUT    /buckets/{bucket}/objects/{objectKey}?tenantId={tenantId}`
- `DELETE /buckets/{bucket}/objects/{objectKey}?tenantId={tenantId}`
- `POST   /buckets/{bucket}/presign?tenantId={tenantId}`
- `POST   /buckets/{bucket}/share-object?tenantId={tenantId}`

---

## 3) cURL (CLI / shell apps)

### List objects

```bash
curl -sS \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  "$BASE_URL/buckets/$BUCKET/objects?tenantId=$TENANT_ID"
```

### Download object

```bash
curl -sS -L \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  "$BASE_URL/buckets/$BUCKET/objects/docs%2Freport.pdf?tenantId=$TENANT_ID" \
  -o report.pdf
```

### Upload object

```bash
curl -sS -X PUT \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  -H "Content-Type: application/octet-stream" \
  --data-binary "@./report.pdf" \
  "$BASE_URL/buckets/$BUCKET/objects/docs%2Freport.pdf?tenantId=$TENANT_ID"
```

### Delete object

```bash
curl -sS -X DELETE \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  "$BASE_URL/buckets/$BUCKET/objects/docs%2Freport.pdf?tenantId=$TENANT_ID"
```

### Generate pre-signed URL

```bash
curl -sS -X POST \
  -H "Content-Type: application/json" \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  -d '{"key":"docs/report.pdf","method":"GET","expires":900,"accessKeyId":"'"$ACCESS_KEY_ID"'"}' \
  "$BASE_URL/buckets/$BUCKET/presign?tenantId=$TENANT_ID"
```

### Generate share URL

```bash
curl -sS -X POST \
  -H "Content-Type: application/json" \
  -H "X-Storage-Access-Key: $ACCESS_KEY_ID" \
  -H "X-Storage-Secret-Key: $SECRET_ACCESS_KEY" \
  -d '{"key":"docs/report.pdf","expires":3600,"accessKeyId":"'"$ACCESS_KEY_ID"'"}' \
  "$BASE_URL/buckets/$BUCKET/share-object?tenantId=$TENANT_ID"
```

---

## 4) JavaScript (Node.js, backend service)

```javascript
// Node 18+ (built-in fetch)
import fs from "node:fs/promises";

const cfg = {
  baseUrl: "http://localhost:8000/api/v1/storage",
  tenantId: "5d82e8fe-83ce-4bc0-8b2c-32723d02fb54",
  bucket: "m",
  accessKeyId: process.env.ACCESS_KEY_ID,
  secretAccessKey: process.env.SECRET_ACCESS_KEY,
};

function headers(contentType = "application/json") {
  return {
    "Content-Type": contentType,
    "X-Storage-Access-Key": cfg.accessKeyId,
    "X-Storage-Secret-Key": cfg.secretAccessKey,
  };
}

async function api(method, path, body, contentType = "application/json") {
  const resp = await fetch(`${cfg.baseUrl}${path}`, {
    method,
    headers: headers(contentType),
    body: body === undefined ? undefined : body,
  });

  const isJson = (resp.headers.get("content-type") || "").includes("application/json");
  const payload = isJson ? await resp.json() : await resp.text();
  if (!resp.ok) throw new Error(typeof payload === "string" ? payload : JSON.stringify(payload));
  return payload;
}

function objectPath(objectKey) {
  return `/buckets/${encodeURIComponent(cfg.bucket)}/objects/${encodeURIComponent(objectKey)}?tenantId=${encodeURIComponent(cfg.tenantId)}`;
}

async function listObjects() {
  return api("GET", `/buckets/${encodeURIComponent(cfg.bucket)}/objects?tenantId=${encodeURIComponent(cfg.tenantId)}`);
}

async function downloadObject(objectKey, localPath) {
  const resp = await fetch(`${cfg.baseUrl}${objectPath(objectKey)}`, {
    method: "GET",
    headers: headers("application/octet-stream"),
  });
  if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
  const bytes = Buffer.from(await resp.arrayBuffer());
  await fs.writeFile(localPath, bytes);
}

async function uploadObject(objectKey, localPath) {
  const bytes = await fs.readFile(localPath);
  return api("PUT", objectPath(objectKey), bytes, "application/octet-stream");
}

async function deleteObject(objectKey) {
  return api("DELETE", objectPath(objectKey));
}

async function generatePresign(objectKey) {
  return api(
    "POST",
    `/buckets/${encodeURIComponent(cfg.bucket)}/presign?tenantId=${encodeURIComponent(cfg.tenantId)}`,
    JSON.stringify({ key: objectKey, method: "GET", expires: 900, accessKeyId: cfg.accessKeyId })
  );
}

async function generateShare(objectKey) {
  return api(
    "POST",
    `/buckets/${encodeURIComponent(cfg.bucket)}/share-object?tenantId=${encodeURIComponent(cfg.tenantId)}`,
    JSON.stringify({ key: objectKey, expires: 3600, accessKeyId: cfg.accessKeyId })
  );
}

// Example run
(async () => {
  console.log(await listObjects());
  console.log(await generatePresign("docs/report.pdf"));
  console.log(await generateShare("docs/report.pdf"));
})();
```

---

## 5) Python (Flask/FastAPI/Django service)

```python
import requests

BASE_URL = "http://localhost:8000/api/v1/storage"
TENANT_ID = "5d82e8fe-83ce-4bc0-8b2c-32723d02fb54"
BUCKET = "m"
ACCESS_KEY_ID = "<your-access-key-id>"
SECRET_ACCESS_KEY = "<your-secret-access-key>"

HEADERS_JSON = {
    "Content-Type": "application/json",
    "X-Storage-Access-Key": ACCESS_KEY_ID,
    "X-Storage-Secret-Key": SECRET_ACCESS_KEY,
}

HEADERS_BIN = {
    "X-Storage-Access-Key": ACCESS_KEY_ID,
    "X-Storage-Secret-Key": SECRET_ACCESS_KEY,
}

def object_url(key: str) -> str:
    from urllib.parse import quote
    return f"{BASE_URL}/buckets/{quote(BUCKET)}/objects/{quote(key, safe='')}?tenantId={TENANT_ID}"

def list_objects():
    url = f"{BASE_URL}/buckets/{BUCKET}/objects?tenantId={TENANT_ID}"
    r = requests.get(url, headers=HEADERS_JSON, timeout=30)
    r.raise_for_status()
    return r.json()

def download_object(key: str, out_file: str):
    r = requests.get(object_url(key), headers=HEADERS_BIN, timeout=60)
    r.raise_for_status()
    with open(out_file, "wb") as f:
        f.write(r.content)

def upload_object(key: str, in_file: str):
    with open(in_file, "rb") as f:
        r = requests.put(object_url(key), headers=HEADERS_BIN, data=f, timeout=60)
    r.raise_for_status()
    return r.text

def delete_object(key: str):
    r = requests.delete(object_url(key), headers=HEADERS_JSON, timeout=30)
    r.raise_for_status()
    return r.text

def generate_presign(key: str):
    url = f"{BASE_URL}/buckets/{BUCKET}/presign?tenantId={TENANT_ID}"
    payload = {"key": key, "method": "GET", "expires": 900, "accessKeyId": ACCESS_KEY_ID}
    r = requests.post(url, headers=HEADERS_JSON, json=payload, timeout=30)
    r.raise_for_status()
    return r.json()

def generate_share(key: str):
    url = f"{BASE_URL}/buckets/{BUCKET}/share-object?tenantId={TENANT_ID}"
    payload = {"key": key, "expires": 3600, "accessKeyId": ACCESS_KEY_ID}
    r = requests.post(url, headers=HEADERS_JSON, json=payload, timeout=30)
    r.raise_for_status()
    return r.json()
```

---

## 6) Go (microservices)

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

const (
    baseURL   = "http://localhost:8000/api/v1/storage"
    tenantID  = "5d82e8fe-83ce-4bc0-8b2c-32723d02fb54"
    bucket    = "m"
    accessKey = "<your-access-key-id>"
    secretKey = "<your-secret-access-key>"
)

func doReq(method, fullURL string, body io.Reader, contentType string) ([]byte, error) {
    req, err := http.NewRequest(method, fullURL, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-Storage-Access-Key", accessKey)
    req.Header.Set("X-Storage-Secret-Key", secretKey)
    if contentType != "" {
        req.Header.Set("Content-Type", contentType)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    payload, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 300 {
        return nil, fmt.Errorf("status=%d body=%s", resp.StatusCode, string(payload))
    }
    return payload, nil
}

func objectURL(key string) string {
    return fmt.Sprintf("%s/buckets/%s/objects/%s?tenantId=%s", baseURL, url.PathEscape(bucket), url.PathEscape(key), url.QueryEscape(tenantID))
}

func main() {
    listURL := fmt.Sprintf("%s/buckets/%s/objects?tenantId=%s", baseURL, url.PathEscape(bucket), url.QueryEscape(tenantID))
    payload, err := doReq(http.MethodGet, listURL, nil, "application/json")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(payload))

    body, _ := json.Marshal(map[string]any{
        "key":         "docs/report.pdf",
        "method":      "GET",
        "expires":     900,
        "accessKeyId": accessKey,
    })

    presignURL := fmt.Sprintf("%s/buckets/%s/presign?tenantId=%s", baseURL, url.PathEscape(bucket), url.QueryEscape(tenantID))
    presignResp, err := doReq(http.MethodPost, presignURL, bytes.NewReader(body), "application/json")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(presignResp))

    _ = objectURL // use for GET/PUT/DELETE object operations as needed
}
```

---

## 7) Java (Spring Boot / Java services)

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class StorageExample {
  static final String BASE_URL = "http://localhost:8000/api/v1/storage";
  static final String TENANT_ID = "5d82e8fe-83ce-4bc0-8b2c-32723d02fb54";
  static final String BUCKET = "m";
  static final String ACCESS_KEY = "<your-access-key-id>";
  static final String SECRET_KEY = "<your-secret-access-key>";

  public static void main(String[] args) throws Exception {
    HttpClient client = HttpClient.newHttpClient();

    HttpRequest listReq = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/buckets/" + BUCKET + "/objects?tenantId=" + TENANT_ID))
        .header("X-Storage-Access-Key", ACCESS_KEY)
        .header("X-Storage-Secret-Key", SECRET_KEY)
        .GET()
        .build();

    HttpResponse<String> listResp = client.send(listReq, HttpResponse.BodyHandlers.ofString());
    System.out.println(listResp.body());

    String presignJson = "{\"key\":\"docs/report.pdf\",\"method\":\"GET\",\"expires\":900,\"accessKeyId\":\"" + ACCESS_KEY + "\"}";
    HttpRequest presignReq = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/buckets/" + BUCKET + "/presign?tenantId=" + TENANT_ID))
        .header("Content-Type", "application/json")
        .header("X-Storage-Access-Key", ACCESS_KEY)
        .header("X-Storage-Secret-Key", SECRET_KEY)
        .POST(HttpRequest.BodyPublishers.ofString(presignJson))
        .build();

    HttpResponse<String> presignResp = client.send(presignReq, HttpResponse.BodyHandlers.ofString());
    System.out.println(presignResp.body());
  }
}
```

---

## 8) C# (.NET backend services)

```csharp
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;

var baseUrl = "http://localhost:8000/api/v1/storage";
var tenantId = "5d82e8fe-83ce-4bc0-8b2c-32723d02fb54";
var bucket = "m";
var accessKeyId = "<your-access-key-id>";
var secretKey = "<your-secret-access-key>";

using var http = new HttpClient();
http.DefaultRequestHeaders.Add("X-Storage-Access-Key", accessKeyId);
http.DefaultRequestHeaders.Add("X-Storage-Secret-Key", secretKey);

var listResp = await http.GetAsync($"{baseUrl}/buckets/{bucket}/objects?tenantId={tenantId}");
listResp.EnsureSuccessStatusCode();
Console.WriteLine(await listResp.Content.ReadAsStringAsync());

var presignBody = """
{
  "key":"docs/report.pdf",
  "method":"GET",
  "expires":900,
  "accessKeyId":""" + accessKeyId + """
}
""";

using var content = new StringContent(presignBody, Encoding.UTF8, "application/json");
var presignResp = await http.PostAsync($"{baseUrl}/buckets/{bucket}/presign?tenantId={tenantId}", content);
presignResp.EnsureSuccessStatusCode();
Console.WriteLine(await presignResp.Content.ReadAsStringAsync());
```

---

## 9) Browser Applications (important)

Do not put `X-Storage-Secret-Key` in browser JavaScript. Instead:

1. Call your own backend API (server-side) to talk to storage using access keys.
2. Or generate pre-signed URLs from your backend and let browser upload/download directly with the pre-signed URL.

---

## 10) Quick troubleshooting

- `401 invalid storage access key`:
  - Access key ID or secret is wrong.
  - Key is revoked or expired.
- `403`:
  - Key role too low for method (for example, reader trying to upload/delete).
  - Bucket scope does not include the target bucket.
  - Prefix scope blocks the object key path.
- `404`:
  - Bucket/object does not exist in that tenant.

---

You can also use the included Postman collections in this folder for no-code API testing.
