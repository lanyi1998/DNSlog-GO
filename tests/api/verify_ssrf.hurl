POST http://localhost:8000/api/verifyHttp
{
    "query": "not_domain.com"
}
HTTP 401
{"code":"401","msg":"Invalid token"}

POST http://localhost:8000/api/verifyHttp
Token: {{TOKEN}}
{
    "query": "not_domain.com"
}
HTTP 200
{"code":200,"msg":"Not Found","data":null}

GET http://localhost:8000/dnslog/test

POST http://localhost:8000/api/verifyHttp
Token: {{TOKEN}}
{
    "query": "/dnslog/test"
}
HTTP 200
[Asserts]
jsonpath "$.data.subdomain" == "/dnslog/test"