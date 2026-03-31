# Service Intentions for lurch service
# Allows the gateway to route traffic to lurch

Kind = "service-intentions"
Name = "lurch"
Sources = [
  {
    Name   = "demoland-gateway"
    Action = "allow"
  }
]
