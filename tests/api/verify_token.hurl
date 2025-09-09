POST http://localhost:8000/api/verifyToken
{
    "token": "not_token"
}

HTTP 200
{"code":401,"msg":"Invalid token","data":null}



POST http://localhost:8000/api/verifyToken
{
    "token": "{{TOKEN}}"
}
HTTP 200
{"code":200,"msg":"success","data":{"subdomain":"dnslog.demo.com","token":"{{TOKEN}}"}}

GET http://localhost:8000/api/getDnsData
HTTP 401
{"HTTPStatusCode":"401","Msg":"Invalid token"}

GET http://localhost:8000/api/getDnsData
Token: {{TOKEN}}
HTTP 200
{"code":200,"msg":"success","data":[]}