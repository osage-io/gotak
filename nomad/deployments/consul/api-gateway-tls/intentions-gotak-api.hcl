# Service Intentions for API Gateway
# Allows the gateway to route traffic to backend services

# Allow gateway to access gotak-api
Kind = "service-intentions"
Name = "gotak-api"
Sources = [
  {
    Name   = "demoland-gateway"
    Action = "allow"
  }
]
