# DNSlog-GO API 文档

本文档详细描述了 DNSlog-GO 项目的 API 接口。

## 认证

所有需要认证的 API 端点都需要在 HTTP 请求头中包含一个 `token` 字段。

`token: your_token`

## API 端点

### 1. 验证 Token

*   **URL:** `/api/verifyToken`
*   **Method:** `POST`
*   **Description:** 验证提供的 `token` 是否有效。
*   **Request Body:**

    ```json
    {
        "token": "your_token"
    }
    ```

*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": {
            "subdomain": "your_subdomain.example.com",
            "token": "your_token"
        }
    }
    ```

*   **Error Response (401):**

    ```json
    {
        "code": 401,
        "msg": "Invalid token"
    }
    ```

### 2. 获取 DNS 解析记录

*   **URL:** `/api/getDnsData`
*   **Method:** `GET`
*   **Description:** 获取与 `token` 关联的所有 DNS 解析记录。
*   **Authentication:** Required
*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": [
            {
                "type": "A",
                "subdomain": "test.your_subdomain.example.com",
                "ipaddress": "1.2.3.4",
                "time": 1678886400
            }
        ]
    }
    ```

### 3. 清除 DNS 解析记录

*   **URL:** `/api/clean`
*   **Method:** `GET`
*   **Description:** 清除与 `token` 关联的所有 DNS 解析记录。
*   **Authentication:** Required
*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success"
    }
    ```

### 4. 获取并清除 DNS 解析记录

*   **URL:** `/api/getDnsData_clear`
*   **Method:** `GET`
*   **Description:** 获取与 `token` 关联的所有 DNS 解析记录，然后清除它们。
*   **Authentication:** Required
*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": [
            {
                "type": "A",
                "subdomain": "test.your_subdomain.example.com",
                "ipaddress": "1.2.3.4",
                "time": 1678886400
            }
        ]
    }
    ```

### 5. 验证单个 DNS 解析记录

*   **URL:** `/api/verifyDns`
*   **Method:** `POST`
*   **Description:** 验证指定的子域名是否存在 DNS 解析记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "query": "test.your_subdomain.example.com"
    }
    ```

*   **Success Response (200 - Found):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": {
            "subdomain": "test.your_subdomain.example.com",
            "ipaddress": "1.2.3.4",
            "time": 1678886400,
            "type": "A"
        }
    }
    ```

*   **Success Response (200 - Not Found):**

    ```json
    {
        "code": 200,
        "msg": "Not Found"
    }
    ```

### 6. 批量验证 DNS 解析记录

*   **URL:** `/api/bulkVerifyDns`
*   **Method:** `POST`
*   **Description:** 批量验证指定的子域名是否存在 DNS 解析记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "subdomain": [
            "test1.your_subdomain.example.com",
            "test2.your_subdomain.example.com"
        ]
    }
    ```

*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": [
            "test1.your_subdomain.example.com"
        ]
    }
    ```

### 7. 验证单个 HTTP 记录

*   **URL:** `/api/verifyHttp`
*   **Method:** `POST`
*   **Description:** 验证指定的路径是否存在 HTTP 访问记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "query": "/your_subdomain/path"
    }
    ```

*   **Success Response (200 - Found):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": {
            "subdomain": "/your_subdomain/path",
            "ipaddress": "1.2.3.4",
            "time": 1678886400,
            "type": "HTTP"
        }
    }
    ```

*   **Success Response (200 - Not Found):**

    ```json
    {
        "code": 200,
        "msg": "Not Found"
    }
    ```

### 8. 批量验证 HTTP 记录

*   **URL:** `/api/bulkVerifyHttp`
*   **Method:** `POST`
*   **Description:** 批量验证指定的路径是否存在 HTTP 访问记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "query": [
            "/your_subdomain/path1",
            "/your_subdomain/path2"
        ]
    }
    ```

*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success",
        "data": [
            "/your_subdomain/path1"
        ]
    }
    ```

### 9. 设置 A 记录

*   **URL:** `/api/setARecord`
*   **Method:** `POST`
*   **Description:** 为指定的子域名设置 A 记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "domain": "custom",
        "ip": "1.2.3.4"
    }
    ```

*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success"
    }
    ```

### 10. 设置 TXT 记录

*   **URL:** `/api/setTXTRecord`
*   **Method:** `POST`
*   **Description:** 为指定的子域名设置 TXT 记录。
*   **Authentication:** Required
*   **Request Body:**

    ```json
    {
        "domain": "custom",
        "txt": "your_text_record"
    }
    ```

*   **Success Response (200):**

    ```json
    {
        "code": 200,
        "msg": "success"
    }
    ```
