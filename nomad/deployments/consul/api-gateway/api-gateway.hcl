# Consul API Gateway Configuration Entry
# This defines the API gateway that routes external traffic to gotak services

Kind = "api-gateway"
Name = "gotak-gateway"

# Listeners define the ports and protocols the gateway accepts
Listeners = [
  {
    Name     = "http-listener"
    Port     = 8443
    Protocol = "http"
  },
  {
    Name     = "cot-listener"
    Port     = 9087
    Protocol = "tcp"
  }
]
