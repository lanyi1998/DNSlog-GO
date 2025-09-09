POST http://localhost:8000/api/verifyDns
{
    "Query": "not_domain.com"
}
HTTP 401
{"code":"401","msg":"Invalid token"}

POST http://localhost:8000/api/verifyDns
Token: {{TOKEN}}
{
    "Query": "not_domain.com"
}
HTTP 200
{"code":200,"msg":"Not Found","data":null}


#POST http://localhost:8000/api/verifyDns
#Token: {{TOKEN}}
#{
#    "Query": "cer4n.dnslog.demo.com"
#}
#{"code":200,"msg":"success","data":{"ipaddress":"127.0.0.1","subdomain":"cer4n.dnslog.demo.com","time":1757390114,"type":"DNS"}}